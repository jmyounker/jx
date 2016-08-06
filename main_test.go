package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hoisie/mustache"
)

var inputTests = []struct{in, tmpl, want string}{
	{"", "foo", ""},
	{"{}", "", "\n"},
	{"{\"a\": \"foo\"}", "{{a}}", "foo\n"},
	{"{\"a\": \"foo\"}{\"a\":\"bar\"}", "{{a}}", "foo\nbar\n"},
	{"[\"foo\"]", "{{1}}", "foo\n"},
	{"\"foo\"", "{{1}}", "foo\n"},
	{"true", "{{1}}", "true\n"},
	{"42", "{{1}}", "42\n"},
	{"null", "{{1}}", "null\n"},
	{"null", "{{{1}}}", "null\n"},
	{"{\"a\":[1,2]}", "{{{a}}}", "[1,2]\n"},
	{"{\"a\":[\"b\",\"c\"]}", "{{{a}}}", "[\"b\",\"c\"]\n"},
	{"{\"a\":{\"b\":\"c\"}}", "{{{a}}}", "{\"b\":\"c\"}\n"},
}

func TestExpandInput(t *testing.T) {
	for _, tc := range inputTests {
		in := strings.NewReader(tc.in)
		tmpl, err := mustache.ParseString(tc.tmpl)
		assertNoError(t, err)
		out := bytes.Buffer{}
		outFact := &staticWriterFactory{&out}
		assertNoError(t, expand(in, outFact, &staticTemplateFactory{tmpl}))
		if (out.String() != tc.want) {
			t.Fatalf("Expected %q but got %q", tc.want, out.String())
		}
	}
}

func assertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error but got: %s", err)
	}
}
