package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/hoisie/mustache"
)

var (
	inputName  = flag.String("i", "", "input filename")
	tmplName   = flag.String("t", "", "template filename")
	tmplXpn    = flag.String("tx", "", "template filename expansion")
	outputName = flag.String("o", "", "output filename")
	outputXpn  = flag.String("ox", "", "output filename expansion")
	append     = flag.Bool("a", false, "append to file")
)

func main() {
	flag.Parse()

	in, err := getInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	out, err := getOutputFactory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	tmpl, err := getTemplateFactory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	if err := expand(in, out, tmpl); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func getInput() (io.Reader, error) {
	if *inputName == "" {
		return os.Stdin, nil
	}
	in, err := os.Open(*inputName)
	if err != nil {
		return nil, fmt.Errorf("could not open file %q for reading: %s", *inputName, err)
	}
	return in, nil
}

func getOutputFactory() (writerFactory, error) {
	if *outputName == "" && *outputXpn == "" {
		return &staticWriterFactory{os.Stdout}, nil
	}
	if *outputName != "" && *outputXpn != "" {
		return nil, fmt.Errorf("-o and -ox are mutally exclusive")
	}
	if *outputName != "" {
		out, err := openFile(*outputName, *append)
		if err != nil {
			return nil, fmt.Errorf("could not open file %q for writing: %s", *outputName, err)
		}
		return &staticWriterFactory{out}, nil
	}
	fnTmpl, err := mustache.ParseString(*outputXpn)
	if err != nil {
		return nil, fmt.Errorf("could not parse output path template %q: %s", *tmplXpn, err)
	}
	return &dynamicWriterFactory{fnTmpl: fnTmpl, append: *append}, nil
}

func getTemplateFactory() (templateFactory, error) {
	if *tmplName == "" && *tmplXpn == "" && flag.NArg() == 0 {
		return nil, errors.New("you must supply a template or -t")
	}
	if *tmplName != "" && *tmplXpn != "" {
		return nil, fmt.Errorf("-t and -tx are mutually exclusive")
	}
	if *tmplName == "" && *tmplXpn == "" {
		tmpl, err := mustache.ParseString(flag.Arg(0))
		if err != nil {
			return nil, fmt.Errorf("could not parse template %q: %s", flag.Arg(0), err)
		}
		return &staticTemplateFactory{tmpl}, nil
	}

	if *tmplName != "" {
		tmpl, err := mustache.ParseFile(*tmplName)
		if err != nil {
			return nil, fmt.Errorf("could not parse template file %q: %s", *tmplName, err)
		}
		return &staticTemplateFactory{tmpl}, nil
	}

	fnTmpl, err := mustache.ParseString(*tmplXpn)
	if err != nil {
		return nil, fmt.Errorf("could not parse template path template %q: %s", *tmplXpn, err)
	}
	return &dynamicTemplateFactory{fnTmpl: fnTmpl}, nil
}

// writerFactory allows choosing output sources based on the JSON input
type writerFactory interface {
	getWriter(xpn map[string]interface{}) (io.Writer, error)
}

type staticWriterFactory struct {
	writer io.Writer
}

func (f *staticWriterFactory) getWriter(xpn map[string]interface{}) (io.Writer, error) {
	return f.writer, nil
}

type dynamicWriterFactory struct {
	fnTmpl *mustache.Template // filename template
	append bool				  // append to file rather than truncate
	fn     string             // path to current template
	writer *os.File           // current writer
}

func (f *dynamicWriterFactory) getWriter(xpn map[string]interface{}) (io.Writer, error) {
	fn := f.fnTmpl.Render(xpn)
	if fn == f.fn {
		return f.writer, nil
	}
	if f.writer != nil {
		f.writer.Close()
	}
	writer, err := openFile(fn, f.append)
	if err != nil {
		return nil, err
	}
	f.fn = fn
	f.writer = writer
	return writer, nil
}

// extracted for cleanliness
func openFile(fn string, append bool) (*os.File, error) {
	md := os.O_TRUNC
	if append {
		md = os.O_APPEND
	}
	return os.OpenFile(fn, md|os.O_CREATE|os.O_WRONLY, 0666)
}

// templateFactory allows choosing templates based on the JSON input
type templateFactory interface {
	getTemplate(xpn map[string]interface{}) (*mustache.Template, error)
}

type staticTemplateFactory struct {
	tmpl *mustache.Template
}

func (f *staticTemplateFactory) getTemplate(xpn map[string]interface{}) (*mustache.Template, error) {
	return f.tmpl, nil
}

type dynamicTemplateFactory struct {
	fnTmpl *mustache.Template // filename template
	fn     string             // path to current template
	tmpl   *mustache.Template // current template
}

func (dtf *dynamicTemplateFactory) getTemplate(xpn map[string]interface{}) (*mustache.Template, error) {
	fn := dtf.fnTmpl.Render(xpn)
	if fn == dtf.fn {
		return dtf.tmpl, nil
	}
	tmpl, err := mustache.ParseFile(fn)
	if err != nil {
		return nil, err
	}
	dtf.tmpl = tmpl
	return tmpl, nil
}

// expand combines JSON input with templates to produce output.
func expand(in io.Reader, outFact writerFactory, tmplFact templateFactory) error {
	dec := json.NewDecoder(in)
	var j interface{}
	for {
		if err := dec.Decode(&j); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		xpn := getExpn(j)

		tmpl, err := tmplFact.getTemplate(xpn)
		if err != nil {
			return err
		}

		out, err := outFact.getWriter(xpn)
		if err != nil {
			return err
		}

		out.Write([]byte(tmpl.Render(xpn)))
		out.Write([]byte("\n"))
	}
        return nil
}

// getExpn transforms various JSON datatypes into mustache expansion dictionaries.
func getExpn(j interface{}) map[string]interface{} {
	switch j.(type) {
	case map[string]interface{}:
		xpn := map[string]interface{}{};
		for k, v := range j.(map[string]interface{}) {
			xpn[k] = terminalForm(v)
		}
		return xpn
	case []interface{}:
		xpn := map[string]interface{}{}
		for i, v := range j.([]interface{}) {
			xpn[strconv.Itoa(i + 1)] = terminalForm(v)
		}
		return xpn
	default:
		return map[string]interface{}{"1": terminalForm(j)}
	}
	return nil
}


func terminalForm(j interface{}) interface{} {
	switch j.(type) {
	case map[string]interface{}:
		v, err := json.Marshal(j)
		if (err != nil) {
			panic("Original form should have come from JSON so something is very wrong")
		}
		return string(v)
	case []interface{}:
		v, err := json.Marshal(j)
		if (err != nil) {
			panic("Original form should have come from JSON so something is very wrong")
		}
		return string(v)
	case string:
		return j.(string)
	case bool:
		return j.(bool)
	case float64:
		return j.(float64)
	default:
		if j == nil {
                   return j
                }
		log.Fatal("Should be unreachable")
	}
	return nil
}

