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
	outputName = flag.String("o", "", "output filename")
)


func main() {
	flag.Parse()

	in, err := getInput()
	if err != nil {
		log.Printf("error: %s", err)
		os.Exit(127)
	}

	out, err := getOutput()
	if err != nil {
		log.Printf("error: %s", err)
		os.Exit(127)
	}

	tmpl, err := getTemplate()
	if err != nil {
		log.Printf("error: %s", err)
		os.Exit(127)
	}

	expand(in, out, tmpl)
}

func getTemplate() (*mustache.Template, error) {
	if *tmplName == "" && flag.NArg() == 0 {
		return nil, errors.New("you must supply a template or -t")
	}
	if *tmplName == "" {
		tmpl, err := mustache.ParseString(flag.Arg(0))
		if err != nil {
			return nil, fmt.Errorf("could not parse template %q: %s", flag.Arg(0), err)
		}
		return tmpl, nil
	}

	tmpl, err := mustache.ParseFile(*tmplName)
	if err != nil {
		return nil, fmt.Errorf("could not parse template file %q: %s", *tmplName, err)
	}
	return tmpl, nil
}

func getInput() (io.Reader, error) {
	if *inputName == "" {
		return os.Stdin, nil
	}
	in, err := os.Open(*inputName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not open file %q for reading: %s", *inputName, err))
	}
	return in, nil
}

func getOutput() (io.Writer, error) {
	if *outputName == "" {
		return os.Stdout, nil
	}
	out, err := os.OpenFile(*outputName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not open file %q for writing: %s", *outputName, err))
	}
	return out, nil
}

func expand(in io.Reader, out io.Writer, tmpl *mustache.Template) error {
	dec := json.NewDecoder(in)
	var j interface{}
	for {
		if err := dec.Decode(&j); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		out.Write([]byte(tmpl.Render(getExpn(j))))
		out.Write([]byte("\n"))
	}
}

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
