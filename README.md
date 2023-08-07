# anyq

This is a silly proof-of-concept, which acts as an alternate frontend to the
excellent `gojq` library, with support for many input/output formats.

The goal is to be able to work with various data formats, converting between
them freely and manipulating them all in the same way, allowing the user to
navigate the data without having to care much about the details.

# Humble expectations

Honestly, this project is a cheap hack.  It achieves the above goal, but it
has very little polish.  It has proved the concept; I'm pretty happy with it.

I'm not planning on developing it much further.  I would much rather have
features like this added to the excellent tools we already have, alongside
the other features of those tools (which this project does not support).
If/when that happens, I will happily link to them here.

The only thing I might add here is more weird data formats.

# Features

This project does the bare minimum to achieve the goal, as stated above.  So
it's missing quite a lot of stuff.

Things that are missing:
* [compiler options](https://github.com/itchyny/gojq/blob/main/option.go)
* multiple input files
* even more formats
    * protobuf?
    * url-encoded form values
    * mime-encoded form values
    * readonly support for crazy things
        * scalar parts of data buffer formats (because why not)
            * python pickle format / torch.save() files
            * .npy, .npz files
            * .cdf, .hdf5, .jld files
        * media tags (because why not)
            * vorbis comments
            * matroska tags
            * mp3/mp4/m4a/id3 tags
* colorization
* "raw" output (omitting quotes on a scalar string value)
* lots of other jq/gojq command line parameters
* regression tests

Actual features:
* Inferring the default input format based on the input filename
* Inferring the default input/output formats based on symlinks (e.g. `yamlq` is `anyq` with yaml defaults)
* data formats
    * BSON (only structs at the top level)
    * CSV (only two-level with an array at the top)
    * INI (only two-level struct of structs)
    * JSON
    * MessagePack
    * TOML (only structs at the top level)
    * XML
    * YAML
* a `--formats` parameter listing the formats and supported features

# Formats

```
% ./anyq --formats
   • Format features:
   • P........ = Pretty-printing/indentation supported
   • .O....... = Output supported
   • ..I...... = Input supported
   • ...S..... = Converts data in a standard/universal way
   • ....A.... = Arbitrary tree depths / data layouts supported
   • .....s... = Supports a scalar value (string/int/float/bool/null) at the top level
   • ......o.. = Supports object (struct/dict/map) at the top level
   • .......a. = Supports array (list/slice) at the top level
   • ........B = Binary format (cannot safely write to stdout)
   • ----------  Supported formats:
   • .OISA.o.B  bson has file extensions .bson
   • .OI....a.  csv has file extensions .csv
   • ..IS..o..  ini has file extensions .ini
   • POISAsoa.  json has file extensions .json, .js
   • .OISAsoaB  msgpack has file extensions .msgpack, .mpk
   • POISA.o..  toml has file extensions .toml
   • POI.Aso..  xml has file extensions .xml, .xhtml, .xsd, .xsl, .xslt
   • POISAsoa.  yaml has file extensions .yaml, .yml
```

# Examples

Extracting some `<a>` tags from an XSLT file:

```
% anyq '.stylesheet.template[1].ul.li|map(.a)' /usr/share/bison/xslt/xml2xhtml.xsl
<root>
  <element href="#reductions">Reductions</element>
  <element href="#conflicts">Conflicts</element>
  <element href="#grammar">Grammar</element>
  <element href="#automaton">Automaton</element>
</root>
```

And the same data as a CSV file, because why not:

```
% csvq '.stylesheet.template[1].ul.li|map(.a)' /usr/share/bison/xslt/xml2xhtml.xsl
#text,@href
Reductions,#reductions
Conflicts,#conflicts
Grammar,#grammar
Automaton,#automaton
```
