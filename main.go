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
	inputName = flag.String("i", "", "input filename")
	tmplName = flag.String("t", "", "template filename")
	tmplXpn = flag.String("tx", "", "template filename expansion")
	outputName = flag.String("o", "", "output filename")
)

func main() {
	flag.Parse()

	in, err := getInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	out, err := getOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	tmpl, err := getTemplateFactory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	expand(in, out, tmpl)
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

func getOutput() (io.Writer, error) {
	if *outputName == "" {
		return os.Stdout, nil
	}
	out, err := os.OpenFile(*outputName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("could not open file %q for writing: %s", *outputName, err)
	}
	return out, nil
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

// Nasty mechanisms to manage templates
type templateFactory interface {
	getTemplate(xpn map[string]interface{}) (*mustache.Template, error)
}

type staticTemplateFactory struct {
	tmpl *mustache.Template
}

func (stf *staticTemplateFactory)getTemplate(xpn map[string]interface{}) (*mustache.Template, error) {
	return stf.tmpl, nil
}

type dynamicTemplateFactory struct {
	fnTmpl *mustache.Template // filename template
	fn string // path to current template
	tmpl *mustache.Template // current template
}

func (dtf *dynamicTemplateFactory)getTemplate(xpn map[string]interface{}) (*mustache.Template, error) {
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
func expand(in io.Reader, out io.Writer, tmplFact templateFactory) error {
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
		out.Write([]byte(tmpl.Render(xpn)))
		out.Write([]byte("\n"))
	}
}

// getExpn transforms various JSON datatypes into mustache expansion dictionaries.
func getExpn(j interface{}) map[string]interface{} {
	switch j.(type) {
	case map[string]interface{}:
		return j.(map[string]interface{})
	case []interface{}:
		xpn := map[string]interface{}{}
		for i, v := range j.([]interface{}) {
			xpn[strconv.Itoa(i+1)] = v
		}
		return xpn
	case string:
		return map[string]interface{}{"1": j.(string)}
	case bool:
		return map[string]interface{}{"1": j.(bool)}
	case float64:
		return map[string]interface{}{"1": j.(float64)}
	default:
		if j == nil {
			return map[string]interface{}{"1": nil}
		}
		log.Fatal("Should be unreachable")
	}
	return nil
}
