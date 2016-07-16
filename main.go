package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/hoisie/mustache"
	"github.com/urfave/cli"
)

var Version string;

func main() {
	app := cli.NewApp()
	app.Name = "jx"
	app.Version = Version
	app.Usage = "Uses JSON objects to expand string templates."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "input, i",
			Usage: "Read input from `FILE` instead of stdin",
			Value: "",
		},
		cli.StringFlag{
			Name:  "template, t",
			Usage: "Read template from `FILE`",
			Value: "",
		},
		cli.StringFlag{
			Name:  "template-file-template, tx",
			Usage: "Read template from expanded `FILENAME_TEMPLATE` based on the current JSON object",
			Value: "",
		},
		cli.StringFlag{
			Name:  "output, o",
			Usage: "Write to `FILE` instead of stdout",
			Value: "",
		},
		cli.StringFlag{
			Name:  "output-file-template, ox",
			Usage: "Write to expanded `FILENAME_TEMPLATE` based on the current JSON object",
		},
		cli.BoolTFlag{
			Name: "append, a",
			Usage: "Append new output to file instead of overwriting",
		},
	}
	app.Action = run;
	app.Run(os.Args)
}

func run(ctx *cli.Context) error {
	in, err := getInput(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	out, err := getOutputFactory(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	tmpl, err := getTemplateFactory(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(127)
	}

	if err := expand(in, out, tmpl); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	return nil
}

func getInput(ctx *cli.Context) (io.Reader, error) {
	if ctx.String("input") == "" {
		return os.Stdin, nil
	} else {
		in, err := os.Open(ctx.String("input"))
		if err != nil {
			return nil, fmt.Errorf("could not open file %q for reading: %s", ctx.String("input"), err)
		}
		return in, nil
	}
}

func getOutputFactory(ctx *cli.Context) (writerFactory, error) {
	if ctx.String("output") != "" && ctx.String("output-file-template") != "" {
		return nil, fmt.Errorf("only one destination possible")
	}
	if ctx.String("output") != "" {
		out, err := openFile(ctx.String("output"), ctx.BoolT("append"))
		if err != nil {
			return nil, fmt.Errorf("could not open file %q for writing: %s", ctx.String("output"), err)
		}
		return &staticWriterFactory{out}, nil
	} else if (ctx.String("output-file-template") != "") {
		fnTmpl, err := mustache.ParseString(ctx.String("output-file-template"))
		if err != nil {
			return nil, fmt.Errorf("could not parse output path template %q: %s", ctx.String("output-file-template"), err)
		}
		return &dynamicWriterFactory{fnTmpl: fnTmpl, append: ctx.BoolT("append")}, nil
	} else {
		return &staticWriterFactory{os.Stdout}, nil
	}
}

func getTemplateFactory(ctx *cli.Context) (templateFactory, error) {
	nSrc := 0
	if (ctx.String("template") != "") {
		nSrc += 1
	}
	if (ctx.String("template-file-template") != "") {
		nSrc += 1
	}
	if (ctx.NArg() > 0) {
		nSrc += 1
	}
	if (nSrc == 0) {
		return nil, errors.New("you must supply a template")
	}
	if (nSrc > 1) {
		return nil, fmt.Errorf("only one template source possible")
	}
	if ctx.String("template") != "" {
		tmpl, err := mustache.ParseFile(ctx.String("template"))
		if err != nil {
			return nil, fmt.Errorf("could not parse template file %q: %s", ctx.String("template"), err)
		}
		return &staticTemplateFactory{tmpl}, nil
	} else if (ctx.String("template-file-template") != "") {
		fnTmpl, err := mustache.ParseString(ctx.String("template-file-template"))
		if err != nil {
			return nil, fmt.Errorf("could not parse template path template %q: %s", ctx.String("template-file-template"), err)
		}
		return &dynamicTemplateFactory{fnTmpl: fnTmpl}, nil
	} else {
		tmpl, err := mustache.ParseString(ctx.Args().Get(0))
		if err != nil {
			return nil, fmt.Errorf("could not parse template %q: %s", ctx.Args().Get(0), err)
		}
		return &staticTemplateFactory{tmpl}, nil
	}
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
