package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"text/template"

	"encoding/base64"

	_ "github.com/mattn/go-sqlite3"
	yaml "gopkg.in/yaml.v2"
)

var opts struct {
	db        string
	yaml      string
	showPlots bool
}

type Series struct {
	Data []struct {
		X float64 `yaml:"x"`
		Y float64 `yaml:"y"`
	}
	Title string `yaml:"title"`
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
}

func renderHTML(p *plots) string {
	const html = `<html>
<head><title>pfs</title></head>
<body>
{{range $_, $data := .Images}}
	<img src="data:image/png;base64,{{$data}}"/>
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

func main() {
	parseArgs()
	/*
		db, err := sql.Open("sqlite3", opts.db)
		if err != nil {
			panic(err)
		}
	*/
	yamlBytes, err := ioutil.ReadFile(opts.yaml)
	if err != nil {
		panic(err)
	}

	desc := &YamlFile{}
	if err := yaml.Unmarshal(yamlBytes, &desc); err != nil {
		panic(err)
	}

	plots := &plots{}
	for _, plot := range desc.Plots {
		g := gnuplot{
			Settings: map[string]string{},
			Series:   plot.Series,
		}
		for k, v := range desc.Settings {
			g.Settings[k] = maybeQuoteSetting(k, v)
		}
		for k, v := range plot.Settings {
			g.Settings[k] = maybeQuoteSetting(k, v)
		}
		plots.Images = append(plots.Images, g.plot())
	}

	fmt.Println(renderHTML(plots))
}
