package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"text/template"

	"encoding/base64"

	_ "github.com/mattn/go-sqlite3"
	yaml "gopkg.in/yaml.v2"
)

var opts = struct {
	db        string
	yaml      string
	showPlots bool
	port      int
}{
	port: 8080,
}

type Tuple struct {
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
}

type Series struct {
	Title string `yaml:"title"`

	SQL  string  `yaml:"sql"`
	Data []Tuple `yaml:"data"`
}

type YamlFile struct {
	Settings map[string]string `yaml:"settings"`

	Plots []struct {
		Series   []Series          `yaml:"series"`
		Settings map[string]string `yaml:"settings"`
	} `yaml:"plots"`
}

func parseArgs() {
	flag.BoolVar(&opts.showPlots, "showPlots", false, "show plot scripts")
	flag.StringVar(&opts.db, "db", "", "sqlite3 database")
	flag.StringVar(&opts.yaml, "yaml", "", "plot yaml")
	flag.IntVar(&opts.port, "port", opts.port, "HTTP port")

	flag.Parse()

	if len(opts.db) == 0 && false {
		log.Fatal("Need to set -db")
	}

	if len(opts.yaml) == 0 {
		log.Fatal("Need to set -yaml")
	}
}

type gnuplot struct {
	Settings map[string]string
	Series   []Series
}

func (g *gnuplot) plot() string {
	cmd := exec.Command("gnuplot")
	writer, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	reader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	plotScript := g.generatePlotfile()
	if opts.showPlots {
		log.Printf("Plot:\n%v", plotScript)
	}

	if _, err := writer.Write([]byte(plotScript)); err != nil {
		log.Fatal(err)
	}
	writer.Close()

	image, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}

	return base64.StdEncoding.EncodeToString(image)
}

func (g *gnuplot) generatePlotfile() string {
	const plot = `{{range $k, $v := .Settings}}set {{$k}} {{$v}}
{{end}}
{{range $i, $series := .Series}}{{if $i}},{{else}}plot{{end}} '-' using 1:2 with lines title '{{.Title}}'{{end}}
{{range $_, $series := .Series}}{{range $_, $data := .Data}}
{{.X}} {{.Y}}{{end}}
e
{{end}}`
	tmpl, err := template.New("plot").Parse(plot)
	if err != nil {
		log.Fatal(err)
	}

	buf := &bytes.Buffer{}
	err = tmpl.Execute(buf, g)
	if err != nil {
		log.Fatal(err)
	}

	return string(buf.Bytes())
}

type plots struct {
	Images []string
	Type   string
}

func renderHTML(p *plots) string {
	const html = `<html>
<head><title>pfs</title></head>
<body>{{$type := .Type}}
{{range $_, $data := .Images}}
	<img src="data:image/{{$type}};base64,{{$data}}"/>
{{end}}
</body>
</html>`
	tmpl, err := template.New("html").Parse(html)
	if err != nil {
		log.Fatal(err)
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, p); err != nil {
		log.Fatal(err)
	}

	return string(buf.Bytes())
}

func maybeQuoteSetting(key string, val string) string {
	switch key {
	case "title":
		return "'" + val + "'"
	default:
		return val
	}
}

func makeSeries(db *sql.DB, series *Series) Series {
	if len(series.SQL) != 0 {
		rows, err := db.Query(series.SQL)
		if err != nil {
			log.Fatal(err)
		}
		series.Data = []Tuple{}
		for rows.Next() {
			tuple := Tuple{}
			if err := rows.Scan(&tuple.X, &tuple.Y); err != nil {
				log.Fatal(err)
			}
			series.Data = append(series.Data, tuple)
		}
	}

	return *series
}

func plotFile(db *sql.DB, filename string) string {
	yamlBytes, err := ioutil.ReadFile(opts.yaml)
	if err != nil {
		panic(err)
	}

	desc := &YamlFile{}
	if err := yaml.Unmarshal(yamlBytes, &desc); err != nil {
		panic(err)
	}

	plots := &plots{Type: "png"}
	for _, plot := range desc.Plots {
		series := []Series{}
		for _, inSeries := range plot.Series {
			series = append(series, makeSeries(db, &inSeries))
		}
		g := gnuplot{
			Settings: map[string]string{},
			Series:   series,
		}
		for k, v := range desc.Settings {
			g.Settings[k] = maybeQuoteSetting(k, v)
		}
		for k, v := range plot.Settings {
			g.Settings[k] = maybeQuoteSetting(k, v)
		}
		plots.Images = append(plots.Images, g.plot())
	}

	return renderHTML(plots)
}

func main() {
	parseArgs()

	var db *sql.DB
	if len(opts.db) != 0 {
		var err error
		db, err = sql.Open("sqlite3", opts.db)
		if err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "%s", plotFile(db, opts.yaml))
	})
	log.Printf("Starting HTTP server on :%v", opts.port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", opts.port), nil))
}
