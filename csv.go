package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
)

type CSVFormat struct {
}

func (f *CSVFormat) GetExtensions() []string {
	return []string{
		"csv",
	}
}

func (f *CSVFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      true,  // TODO
		Can_prettyprint: false, // TODO
		Can_scalar:      false,
		Can_array:       true,
		Can_object:      false,
		Is_binary:       false,
		Arbitrary_tree:  false,
		Universal:       false, // header lines are not universal
	}
}

func (f *CSVFormat) Input(in []byte) (any, error) {
	// always returns a slice of string maps.
	b := bytes.NewBuffer(in)
	r := csv.NewReader(b)
	// header row
	headers, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("could not read CSV header row: %v", err)
	}

	rownum := 1
	rv := []any{}
	for {
		rownum += 1
		rowarray, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("csv.Reader.Read returned %v", err)
		}
		if len(rowarray) > len(headers) {
			return nil, fmt.Errorf("csv row %d has more fields than the header", rownum)
		}
		rowmap := map[string]any{}
		for i, value := range rowarray {
			header := headers[i]
			rowmap[header] = value
		}
		rv = append(rv, rowmap)
	}
	return rv, nil
}

type stringable interface {
	String() string
}

func (f *CSVFormat) Output(a any, _ bool) ([]byte, error) {
	// only supports a slice of string maps, or a slice of scalar slices.
	b := bytes.NewBuffer([]byte{})

	slice, ok := a.([]any)
	if !ok {
		return nil, fmt.Errorf("csv output only supports arrays at the top level")
	}

	if len(slice) == 0 {
		// easy
		return []byte{'\n'}, nil
	}

	// entries must all be maps, or all be slices
	found_map := false
	found_slice := false
	found_neither := false
	for _, thing := range slice {
		if _, ok := thing.(map[string]any); ok {
			found_map = true
		} else if _, ok := thing.([]any); ok {
			found_slice = true
		} else {
			found_neither = true
		}
	}
	if found_map && found_slice {
		return nil, fmt.Errorf("csv output only supports row arrays or row objects, not both")
	}
	if found_neither {
		return nil, fmt.Errorf("csv output only supports row arrays or row objects, this is neither")
	}
	if !found_map && !found_slice {
		return nil, fmt.Errorf("csv internal error: nothing to output?")
	}

	var headers []string
	// if maps, find the key names
	if found_map {
		headermap := map[string]bool{}
		for _, anything := range slice {
			mapthing, ok := anything.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("csv could not treat %v as map", anything)
			}
			for key := range mapthing {
				headermap[key] = true
			}
		}
		headers = make([]string, 0, len(headermap))
		for key := range headermap {
			headers = append(headers, key)
		}
		sort.Strings(headers)
	}

	if found_slice {
		maxlen := 0
		for _, anything := range slice {
			slicething, ok := anything.([]any)
			if !ok {
				return nil, fmt.Errorf("csv could not treat %v as slice", anything)
			}
			// fmt.Printf("slice has %d elements, maxlen is %d\n", len(slicething), maxlen)
			if len(slicething) > maxlen {
				maxlen = len(slicething)
			}
		}
		// fmt.Printf("producing headers, maxlen is %d\n", maxlen)
		strlen := len(strconv.Itoa(maxlen))
		headers = make([]string, maxlen)
		for i := 0; i < maxlen; i++ {
			key := strconv.Itoa(i)
			for len(key) < strlen {
				key = "0" + key
			}
			headers[i] = key
		}
		// fmt.Printf("headers is %v\n", headers)
	}

	w := csv.NewWriter(b)
	// emit header
	w.Write(headers)

	for _, rowthing := range slice {
		rowany := []any{}

		if found_map {
			// put values in the right order
			rowmap, ok := rowthing.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("csv could not treat %v as map", rowthing)
			}
			for _, key := range headers {
				value, ok := rowmap[key]
				if ok {
					rowany = append(rowany, value)
				} else {
					rowany = append(rowany, "")
				}
			}
		}
		if found_slice {
			// put values in as-is
			rowslice, ok := rowthing.([]any)
			if !ok {
				return nil, fmt.Errorf("csv could not treat %v as slice", rowthing)
			}
			rowany = rowslice
		}

		// convert any-array to string-array
		rowstrings := make([]string, len(rowany))
		for i, thing := range rowany {
			if strthing, ok := thing.(string); ok {
				rowstrings[i] = strthing
			} else if strable, ok := thing.(stringable); ok {
				rowstrings[i] = strable.String()
			} else {
				rowstrings[i] = fmt.Sprintf("%v", strthing)
			}
		}
		w.Write(rowstrings)
	}
	w.Flush()
	return b.Bytes(), nil
}
