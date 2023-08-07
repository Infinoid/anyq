package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	apex_log "github.com/apex/log"
	apex_cli "github.com/apex/log/handlers/cli"
)

type DataNode map[string]any

type FormatFeatures struct {
	Can_input      bool // supports being used as an input format
	Can_output     bool // supports being used as an output format
	Arbitrary_tree bool // can represent as deep a hierarchy as you like
	Opinionated    bool // converts data in a non-universal way
}

type Format interface {
	GetFeatures() FormatFeatures
	GetExtensions() []string
	Input(b []byte) (DataNode, error)
}

var formats = []Format{
	&JSONFormat{},
}

type App struct {
	log *apex_log.Entry

	infmtname  string
	outfmtname string

	input_filename  string
	output_filename string

	infmt  *Format
	outfmt *Format
}

// Detect if called as "tomlq", and return "toml".
// Returns "auto" if detection fails.
func detect_infmt_by_exename() string {
	exename := filepath.Base(os.Args[0])
	ext := filepath.Ext(exename)
	exename = strings.TrimSuffix(strings.TrimSuffix(exename, ext), ".") // remove ".exe" or whatever
	if !strings.HasSuffix(exename, "q") {
		goto fail
	}
	exename = strings.TrimSuffix(exename, "q")
	for _, format := range formats {
		exts := format.GetExtensions()
		for _, ext := range exts {
			if exename == ext {
				return ext
			}
		}
	}

fail:
	return "auto"
}

// Detect if given an explicit input filename like "conf.toml", and return "toml".
// Returns "auto" if detection fails.
func (a *App) detect_fmt_by_fn(fn string) string {
	infnext := filepath.Ext(fn)
	for _, format := range formats {
		exts := format.GetExtensions()
		for _, ext := range exts {
			if infnext == ext {
				return ext
			}
		}
	}

	return "auto"
}

func (a *App) list_formats() {
	a.log.Fatal("TODO: list_formats()")
}

func NewApp() *App {
	apex_log.SetHandler(apex_cli.Default)
	log := apex_log.WithFields(apex_log.Fields{})
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel != "" {
		apex_log.SetLevelFromString(loglevel)
	}

	infmtdefault := detect_infmt_by_exename()

	infmtarg := flag.String("infmt", infmtdefault, "input file format")
	outfmtarg := flag.String("outfmt", "auto", "output file format")
	outfnarg := flag.String("o", "-", "output filename")
	formatsarg := flag.Bool("formats", false, "list all supported formats")
	flag.Parse()

	// figure out the input filename
	infn := "-"
	if len(flag.Args()) > 1 {
		log.Fatal("At most one input file may be specified.")
	} else if len(flag.Args()) == 1 {
		infn = flag.Arg(0)
	}
	outfn := *outfnarg

	a := &App{
		infmtname:       *infmtarg,
		outfmtname:      *outfmtarg,
		input_filename:  infn,
		output_filename: outfn,
		log:             log,
	}

	if *formatsarg {
		a.list_formats()
	}

	// figure out the input file type if "auto"
	if a.infmtname == "auto" {
		a.infmtname = a.detect_fmt_by_fn(a.input_filename)
	}
	if a.infmtname == "auto" {
		fmt.Printf("Unable to determine input file format.  Please provide an --infmt= parameter.\n")
		os.Exit(1)
	}
	// figure out the output file type if "auto"
	if a.outfmtname == "auto" {
		a.outfmtname = a.detect_fmt_by_fn(a.output_filename)
	}
	if a.outfmtname == "auto" {
		a.outfmtname = a.infmtname
	}

	return a
}
func main() {
	a := NewApp()
	a.log.Debugf("app is %v\n", a)
}
