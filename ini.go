package main

import (
	"fmt"

	"github.com/zieckey/goini"
)

type INIFormat struct {
}

func (f *INIFormat) GetExtensions() []string {
	return []string{
		"ini",
	}
}

func (f *INIFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      false, // TODO
		Can_prettyprint: false,
		Can_scalar:      false,
		Can_array:       false,
		Can_object:      true,
		Is_binary:       false,
		Arbitrary_tree:  false,
		Universal:       true,
	}
}

func (f *INIFormat) Input(b []byte) (any, error) {
	ini := goini.New()
	ini.SetParseSection(true)
	ini.SetSkipCommits(true) // they mean "comments"
	err := ini.Parse(b, goini.DefaultLineSeparator, goini.DefaultKeyValueSeparator)
	if err != nil {
		return nil, err
	}
	sectionmap := ini.GetAll()
	rv := map[string]any{}
	// for some reason this doesn't work
	// rv := map[string]map[string]string(sectionmap)
	// so do it manually
	for sectionkey, kvmap := range sectionmap {
		section := map[string]any{}
		rv[sectionkey] = section
		for key, value := range kvmap {
			section[key] = value
		}
	}
	return rv, err
}

func (f *INIFormat) Output(a any, _ bool) ([]byte, error) {
	return nil, fmt.Errorf("INIFormat.Output not implemented")
}
