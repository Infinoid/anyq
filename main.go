package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	apex_log "github.com/apex/log"
	apex_cli "github.com/apex/log/handlers/cli"
	"github.com/itchyny/gojq"
	"github.com/mattn/go-isatty"
)

type DataNode map[string]any

type FormatFeatures struct {
	Can_input       bool // supports being used as an input format
	Can_output      bool // supports being used as an output format
	Can_prettyprint bool // supports laying out things nicely for humans to read (e.g. indentation)
	Can_scalar      bool // supports a scalar value (string/int/bool/float/null) at the top level
	Can_array       bool // supports array (slice/list) at the top level
	Can_object      bool // supports object (struct/dict/map) at the top level
	Is_binary       bool // binary formats cannot be safely printed to a terminal
	Arbitrary_tree  bool // can represent as deep a hierarchy as you like
	Universal       bool // data is converted in a standardized/universal way
}

type Format interface {
	GetFeatures() FormatFeatures
	GetExtensions() []string
	Input(b []byte) (any, error)
	Output(a any, pretty bool) ([]byte, error)
}

var formats = map[string]Format{
	// all keys in lower case
	// format args are passed through ToLower() before looking them up here
	"bson":    &BSONFormat{},
	"csv":     &CSVFormat{},
	"ini":     &INIFormat{},
	"json":    &JSONFormat{},
	"msgpack": &MsgPackFormat{},
	"toml":    &TOMLFormat{},
	"xml":     &XMLFormat{},
	"yaml":    &YAMLFormat{},
}

type App struct {
	log *apex_log.Entry

	infmtname  string
	outfmtname string

	prettyprint bool

	input_filename  string
	output_filename string

	expr *gojq.Code

	infmt  Format
	outfmt Format
}

// Detect if called as "tomlq", and return "toml".
// Returns "auto" if detection fails.
func detect_fmt_by_exename() string {
	exename := filepath.Base(os.Args[0])
	ext := filepath.Ext(exename)
	exename = strings.TrimSuffix(strings.TrimSuffix(exename, ext), ".") // remove ".exe" or whatever
	exename = strings.TrimSuffix(exename, "q")                          // tomlq → toml
	if _, ok := formats[exename]; ok {
		return exename
	}

	return "auto"
}

// Detect if given an explicit input filename like "conf.toml", and return "toml".
// Returns "auto" if detection fails.
func (a *App) detect_fmt_by_fn(fn string) string {
	fnext := filepath.Ext(fn)
	fnext = strings.TrimPrefix(fnext, ".")
	for fmtname, format := range formats {
		exts := format.GetExtensions()
		for _, ext := range exts {
			if fnext == ext {
				return fmtname
			}
		}
	}

	return "auto"
}

func (a *App) list_formats() {
	// inspired by `ffmpeg -codecs`
	order := "POISAsoaB"
	type flagdesc struct {
		test func(f FormatFeatures) bool
		desc string
	}
	flags := map[rune]flagdesc{
		'I': {desc: "Input supported", test: func(f FormatFeatures) bool { return f.Can_input }},
		'O': {desc: "Output supported", test: func(f FormatFeatures) bool { return f.Can_output }},
		'P': {desc: "Pretty-printing/indentation supported", test: func(f FormatFeatures) bool { return f.Can_prettyprint }},
		'A': {desc: "Arbitrary tree depths / data layouts supported", test: func(f FormatFeatures) bool { return f.Arbitrary_tree }},
		'S': {desc: "Converts data in a standard/universal way", test: func(f FormatFeatures) bool { return f.Universal }},
		's': {desc: "Supports a scalar value (string/int/float/bool/null) at the top level", test: func(f FormatFeatures) bool { return f.Can_scalar }},
		'o': {desc: "Supports object (struct/dict/map) at the top level", test: func(f FormatFeatures) bool { return f.Can_object }},
		'a': {desc: "Supports array (list/slice) at the top level", test: func(f FormatFeatures) bool { return f.Can_array }},
		'B': {desc: "Binary format (cannot safely write to stdout)", test: func(f FormatFeatures) bool { return f.Is_binary }},
	}
	format_names := []string{}
	for fmtname := range formats {
		format_names = append(format_names, fmtname)
	}
	sort.Strings(format_names)
	a.log.Info("Format features:")
	template := ""
	separator := "-"
	for i := 0; i < len(order); i++ {
		template += "."
		separator += "-"
	}
	for i, c := range order {
		desc := flags[c]
		featurehdr := []rune(template)
		featurehdr[i] = c

		a.log.Infof("%s = %s", string(featurehdr), desc.desc)
	}
	a.log.Infof("%s  Supported formats:", separator)
	for _, fmtname := range format_names {
		format := formats[fmtname]
		featurestr := []rune(template)
		features := format.GetFeatures()
		for i, c := range order {
			desc := flags[c]
			if desc.test(features) {
				featurestr[i] = c
			}
		}
		flagstr := string(featurestr)
		extensionslist := []string{}
		for _, ext := range format.GetExtensions() {
			extensionslist = append(extensionslist, fmt.Sprintf(".%s", ext))
		}
		extensionstr := strings.Join(extensionslist, ", ")
		a.log.Infof("%s  %s has file extensions %s", flagstr, fmtname, extensionstr)
	}
	os.Exit(0)
}

func NewApp() *App {
	apex_log.SetHandler(apex_cli.Default)
	log := apex_log.WithFields(apex_log.Fields{})
	loglevel := os.Getenv("LOG_LEVEL")
	if loglevel != "" {
		apex_log.SetLevelFromString(loglevel)
	}

	infmtarg := flag.String("input-format", "auto", "input file format")
	outfmtarg := flag.String("output-format", "auto", "output file format")
	outfnarg := flag.String("o", "-", "output filename")
	formatsarg := flag.Bool("formats", false, "list all supported formats")
	prettyprintarg := flag.Bool("pretty-print", true, "pretty-print formats which support it")
	forcebinoutarg := flag.Bool("force-binary-output", false, "force writing to stdout if output format is binary")
	flag.Parse()

	if *formatsarg {
		a := &App{log: log}
		a.list_formats()
	}

	// gojq expression
	if len(flag.Args()) < 1 {
		log.Error("A 'jq' expression is required as the first positional argument.")
		flag.Usage()
		os.Exit(1)
	}
	exprstr := flag.Arg(0)
	parsed, err := gojq.Parse(exprstr)
	if err != nil {
		log.Fatalf("Could not parse jq expression: %v", err)
	}
	compiled, err := gojq.Compile(parsed)
	if err != nil {
		log.Fatalf("Could not compile jq expression: %v", err)
	}

	// figure out the input filename

	infn := "-"
	if len(flag.Args()) > 2 {
		log.Fatal("At most one input file may be specified.")
	} else if len(flag.Args()) == 2 {
		infn = flag.Arg(1)
	}
	outfn := *outfnarg

	a := &App{
		infmtname:       strings.ToLower(*infmtarg),
		outfmtname:      strings.ToLower(*outfmtarg),
		input_filename:  infn,
		output_filename: outfn,
		prettyprint:     *prettyprintarg,
		expr:            compiled,
		log:             log,
	}

	// figure out the input file type if "auto"
	if a.infmtname == "auto" {
		a.infmtname = a.detect_fmt_by_fn(a.input_filename)
	}
	if a.infmtname == "auto" {
		a.infmtname = detect_fmt_by_exename()
	}
	if a.infmtname == "auto" {
		a.log.Fatal("Unable to determine input file format.  Please provide an --input-format= parameter.")
	}
	if _, ok := formats[a.infmtname]; !ok {
		a.log.Fatalf("I don't know how to speak the '%s' format.  See --formats for a list.", a.infmtname)
	}
	a.infmt = formats[a.infmtname]
	if features := a.infmt.GetFeatures(); !features.Can_input {
		a.log.Fatalf("The '%s' format doesn't know how to handle input.", a.infmtname)
	}

	// figure out the output file type if "auto"
	if a.outfmtname == "auto" {
		a.outfmtname = a.detect_fmt_by_fn(a.output_filename)
	}
	if a.outfmtname == "auto" {
		a.outfmtname = detect_fmt_by_exename()
	}
	if a.outfmtname == "auto" {
		a.outfmtname = a.infmtname
	}
	if _, ok := formats[a.outfmtname]; !ok {
		a.log.Fatalf("I don't know how to speak the '%s' format.  See --formats for a list.", a.outfmtname)
	}
	a.outfmt = formats[a.outfmtname]
	outfeatures := a.outfmt.GetFeatures()
	if !outfeatures.Can_output {
		a.log.Fatalf("The '%s' format doesn't know how to handle output.  Try passing something else with --output-format= parameter.", a.outfmtname)
	}
	if outfn == "-" && outfeatures.Is_binary && !*forcebinoutarg && isatty.IsTerminal(os.Stdout.Fd()) {
		a.log.Fatalf("Preventing binary (%s) output to terminal.  Use --force-binary-output if you're sure you want that.", a.outfmtname)
	}

	return a
}

func (a *App) Run() {
	// read input
	var rawinput []byte
	var err error
	if a.input_filename == "-" {
		buf := bytes.NewBuffer([]byte{})
		_, err = buf.ReadFrom(os.Stdin)
		rawinput = buf.Bytes()
	} else {
		rawinput, err = os.ReadFile(a.input_filename)
	}
	if err != nil {
		a.log.WithError(err).Fatalf("could not read input file %s", a.input_filename)
	}
	input, err := a.infmt.Input(rawinput)
	if err != nil {
		a.log.WithError(err).Fatalf("could not decode input as %s", a.infmtname)
	}
	// execute gojq things
	outiter := a.expr.Run(input)
	output, ok := outiter.Next()
	if !ok {
		a.log.Fatalf("gojq didn't return anything?")
	}
	if err, ok := output.(error); ok {
		a.log.Infof("gojq returned an error; err is %v", err)
		a.log.WithError(err).Fatal("unable to execute gojq expression")
	}
	// a.log.Infof("output: %#v", output)

	// write output
	rawoutput, err := a.outfmt.Output(output, a.prettyprint)
	if err != nil {
		a.log.WithError(err).Fatalf("could not encode output as %s", a.outfmtname)
	}
	if a.output_filename == "-" {
		_, err = os.Stdout.Write(rawoutput)
	} else {
		err = os.WriteFile(a.output_filename, rawoutput, fs.ModePerm)
	}
	if err != nil {
		a.log.WithError(err).Fatalf("could not write file %s", a.output_filename)
	}
}

func main() {
	a := NewApp()
	a.Run()
}
