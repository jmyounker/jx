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
	{"null", "{{1}}", "&lt;nil&gt;\n"},
}

func TestExpandInput(t *testing.T) {
	for _, tc := range inputTests {
		in := strings.NewReader(tc.in)
		tmpl, err := mustache.ParseString(tc.tmpl)
		assertNoError(t, err)
		out := bytes.Buffer{}
		assertNoError(t, expand(in, &out, tmpl))
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
