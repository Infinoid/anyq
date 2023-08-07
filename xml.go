package main

import (
	"bytes"

	"github.com/momiji/xqml"
)

type XMLFormat struct {
}

func (f *XMLFormat) GetExtensions() []string {
	return []string{
		"xml",
		"xhtml",
		"xsd",
		"xsl",
		"xslt",
		// TODO: probably lots of others (even excluding hundreds of app-specific extensions, like .glade)
	}
}

func (f *XMLFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      true,
		Can_prettyprint: true,
		Can_scalar:      true,  // inserts a <root> element around it
		Can_array:       false, // inserts a <root> element around it
		Can_object:      true,  // if top level object has multiple keys, inserts a <root> element around it
		Is_binary:       false,
		Arbitrary_tree:  true,
		Universal:       false, // tags are mapped cleanly; attributes and text can be handled in multiple ways
	}
}

func (f *XMLFormat) Input(in []byte) (any, error) {
	b := bytes.NewBuffer(in)
	dec := xqml.NewDecoder(b)
	dec.Namespaces = false // maybe parameterize this
	var data map[string]any
	err := dec.Decode(&data)
	return data, err
}

func (f *XMLFormat) Output(a any, prettyprint bool) ([]byte, error) {
	b := bytes.NewBuffer([]byte{})
	enc := xqml.NewEncoder(b)
	if prettyprint {
		indent := "  "
		enc.Indent = indent
	}
	err := enc.Encode(a)
	if prettyprint {
		b.WriteString("\n")
	}
	return b.Bytes(), err
}
