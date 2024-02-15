package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/dd-web/opforu-server/internal/types"
)

var (
	replcharIn  = "../data/replacechars/input.txt"
	replcharOut = "../data/replacechars/output.txt"
	paraIn      = "../data/paragraphs/input.txt"
	paraOut     = "../data/paragraphs/output.txt"
)

type test struct {
	name  string
	input string
	want  string
	fn    func(string) (string, error)
}

func TestHTMLWrapper(t *testing.T) {
	tstore := types.NewTemplateStore()
	tests := []*test{}

	// utf-16 -> utf-8 & html char code replacement
	wrapTest, err := newTest("html wrap - character replacement", replcharIn, replcharOut, tstore.ReplaceChars)
	if err != nil {
		panic(err)
	}
	tests = append(tests, wrapTest)

	// wraps into <p> tags
	paragraphs, err := newTest("html wrap - paragraph wrapping", paraIn, paraOut, tstore.WrapParagraphs)
	if err != nil {
		panic(err)
	}
	tests = append(tests, paragraphs)

	for _, testcase := range tests {
		got, err := testcase.fn(testcase.input)
		if err != nil {
			t.Fatalf("Test case %s encountered an error:\n%+v", testcase.name, err)
		}

		if !reflect.DeepEqual(testcase.want, got) {
			t.Fatalf("%v Failed:\n got:\n%+v\n\nwant:\n%+v\n\n", testcase.name, got, testcase.want)
		}
	}

}

// reads file at provided path and returns it's contents as a string
func loadFile(path string) (string, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

// makes a new test for provided arguments
func newTest(name string, inpath, outpath string, fn func(string) (string, error)) (*test, error) {
	instr, err := loadFile(inpath)
	if err != nil {
		return nil, err
	}

	outstr, err := loadFile(outpath)
	if err != nil {
		return nil, err
	}

	return &test{
		name:  name,
		input: instr,
		want:  outstr,
		fn:    fn,
	}, nil
}
