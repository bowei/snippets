SRCS := $(shell find . -name \*.cc)
BINS := $(SRCS:.cc=.exe)

all: $(BINS)

%.exe: %.cc
	g++ $< -g -O0 -std=c++0x -o $@
