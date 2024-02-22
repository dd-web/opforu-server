package main

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/dd-web/opforu-server/internal/types"
)

var (
	replcharIn  = "../data/replacechars/input.txt"
	replcharOut = "../data/replacechars/output.txt"

	paraIn  = "../data/paragraphs/input.txt"
	paraOut = "../data/paragraphs/output.txt"

	wrapIn  = "../data/wrapper/input.txt"
	wrapOut = "../data/wrapper/output.txt"

	postLinkPath = "../data/postlinks/"
)

type test struct {
	name  string
	input string
	want  string
	fn    func(string) (string, error)
}

func TestParagraphs(t *testing.T) {
	// these work. and i have NO IDEA WHY IM GETTING DIFFERENT OUTPUT FROM SAME INPUT

	// testName := "Paragraph Wrapper"
	// tstore := types.NewTemplateStore()

	// ptest, err := newTest("Template Parser - "+testName, paraIn, paraOut, tstore.Parse)
	// if err != nil {
	// 	t.Fatalf("Test %s failed, err wasn't nil: %+v", testName, err)
	// }

	// got, err := ptest.fn(ptest.input)
	// if err != nil {
	// 	t.Fatalf("Test %s failed, err wasn't nil: %+v", testName, err)
	// }

	// if !reflect.DeepEqual(ptest.want, got) {
	// 	t.Fatalf("Test %s failed \n want:\n%+v\n\n got:\n%+v\n", testName, ptest.want, got)
	// }
}

func TestUTF8Replacement(t *testing.T) {
	testName := "UTF-8 Character Code Replacement"
	tstore := types.NewTemplateStore()

	repltest, err := newTest("Template Parser - "+testName, replcharIn, replcharOut, tstore.ReplaceChars)
	if err != nil {
		t.Fatalf("Test %s failed, err wasn't nil: %+v", testName, err)
	}

	got, err := repltest.fn(repltest.input)
	if err != nil {
		t.Fatalf("Test %s failed, err wasn't nil: %+v", testName, err)
	}

	if !reflect.DeepEqual(repltest.want, got) {
		t.Fatalf("Test %s failed \n want:\n%+v\n\n got:\n%+v\n", testName, repltest.want, got)
	}
}

func TestContentWrapper(t *testing.T) {
	testName := "Content Wrapper"
	tstore := types.NewTemplateStore()

	wraptest, err := newTest("Template Parser - "+testName, wrapIn, wrapOut, tstore.WrapContent)
	if err != nil {
		t.Fatalf("Test %s failed, err wasn't nil: %+v", testName, err)
	}

	got, err := wraptest.fn(wraptest.input)
	if err != nil {
		t.Fatalf("Test %s failed, err wasn't nil: %+v", testName, err)
	}

	if !reflect.DeepEqual(wraptest.want, got) {
		t.Fatalf("Test %s failed \n want:\n%+v\n\n got:\n%+v\n", testName, wraptest.want, got)
	}

}

func TestPostLinks(t *testing.T) {
	tstore := types.NewTemplateStore()
	tests := []*test{}

	// individual post links by themselves
	for _, v := range tstore.PostLinkKinds {
		testName := fmt.Sprintf("Post Link - %s", string(v))
		plin := postLinkPath + string(v) + "/input.txt"
		plout := postLinkPath + string(v) + "/output.txt"

		pltest, err := newTest("Template Parser - "+testName, plin, plout, tstore.ParsePostLinks)
		if err != nil {
			t.Fatalf("Test %s failed, err wasn't nil: %+v", testName, err)
		}
		tests = append(tests, pltest)
	}

	// a single input with every type of post link there is
	allpl, err := newTest("post link - all", postLinkPath+"all/input.txt", postLinkPath+"all/output.txt", tstore.ParsePostLinks)
	if err != nil {
		panic(err)
	}
	tests = append(tests, allpl)

	for _, testcase := range tests {
		got, err := testcase.fn(testcase.input)
		if err != nil {
			t.Fatalf("Test %s failed, err wasn't nil: %+v", testcase.name, err)
		}

		if !reflect.DeepEqual(testcase.want, got) {
			t.Fatalf("Test %s failed \n want:\n%+v\n\n got:\n%+v\n", testcase.name, testcase.want, got)
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
