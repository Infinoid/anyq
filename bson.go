package main

import "go.mongodb.org/mongo-driver/bson"

type BSONFormat struct {
}

func (f *BSONFormat) GetExtensions() []string {
	return []string{
		// > by convention, using ".bson" extension
		// -- https://bsonspec.org/faq.html
		"bson",
	}
}

func (f *BSONFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      true,
		Can_prettyprint: false,
		Can_scalar:      false,
		Can_array:       false,
		Can_object:      true,
		Is_binary:       true,
		Arbitrary_tree:  true,
		Universal:       true,
	}
}

func (f *BSONFormat) Input(b []byte) (any, error) {
	var data any
	err := bson.Unmarshal(b, &data)
	return data, err
}

func (f *BSONFormat) Output(a any, _ bool) ([]byte, error) {
	rv, err := bson.Marshal(a)
	return rv, err
}
