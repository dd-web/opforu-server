package types

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	texttempl "text/template"
)

type innerTemplate struct {
	Content   string
	ClassList string
}

var (
	PostLinkRegex = map[string]*regexp.Regexp{
		"post-internal-thread":  regexp.MustCompile(`(?m)&gt;&gt;([[:digit:]]{1,9})&lt;`),                                        // 1 = post_num
		"thread-internal-board": regexp.MustCompile(`(?m)&gt;&gt;([[:alnum:]]{8,12})&lt;`),                                       // 1 = threadslug
		"post-internal-board":   regexp.MustCompile(`(?m)&gt;&gt;([[:alnum:]]{8,12})/([[:digit:]]{1,9})&lt;`),                    // 1 = threadslug, 2 = post_num
		"thread-external-board": regexp.MustCompile(`(?m)&gt;&gt;([[:alpha:]]{2,5})/([[:alnum:]]{8,12})&lt;`),                    // 1 = board short, 2 = threadslug
		"post-external-board":   regexp.MustCompile(`(?m)&gt;&gt;([[:alpha:]]{2,5})/([[:alnum:]]{8,12})/([[:digit:]]{1,9})&lt;`), // 1 = board short, 2 = threadslug, 3 = post_num
	}
	// paragraph delimiting
	CtrlCharReplace     = regexp.MustCompile(`(?m)[[:cntrl:]]`)
	ExcessiveNewLineFix = regexp.MustCompile(`(?m)\n{2,}`)

	// whitespace fixes
	LineStartEndSpaceFix = regexp.MustCompile(`(?m)^[[:blank:]]+|[[:blank:]]+$`)
	ExtraSpaceLimit      = regexp.MustCompile(`(?m)[[:blank:]]+`)
)

type TemplateStore struct {
	// html templates replace all html character codes, our case is iterative which means we must use text templates
	// for any html we wish to inject because our previous iterations become invalid html. This is used to invalidate
	// any html a user may have submitted. Use it once and once only against any submitted content. (sanitize first)
	HtmlReplTempl *template.Template
	PostLinkTempl *texttempl.Template
	Text          map[string]*texttempl.Template
	PostLinkKinds []PostLink
}

func NewTemplateStore() *TemplateStore {
	t := &TemplateStore{
		HtmlReplTempl: &template.Template{},
		PostLinkTempl: &texttempl.Template{},
		Text:          map[string]*texttempl.Template{},
	}
	t.Hydrate()
	return t
}

func (ts *TemplateStore) Hydrate() {
	ts.PostLinkKinds = []PostLink{
		PostInternalThread,
		ThreadInternalBoard,
		PostInternalBoard,
		ThreadExternalBoard,
		PostExternalBoard,
	}

	postLink, err := texttempl.New(string("post-link")).Parse(`<button class="{{ .ClassList }}">{{ .Content }}</button>`)
	if err != nil {
		panic(err)
	}
	ts.PostLinkTempl = postLink

	replacement, err := template.New("utf-8-replace").Parse("{{ .Content }}")
	if err != nil {
		panic(err)
	}
	ts.HtmlReplTempl = replacement

	paragraphs, err := texttempl.New("paragraph").Parse("<p>{{ .Content }}</p>")
	if err != nil {
		panic(err)
	}
	ts.Text["paragraph"] = paragraphs

	wrapper, err := texttempl.New("wrapper").Parse(`<div class="{{ .ClassList }}">{{ .Content }}</div>`)
	if err != nil {
		panic(err)
	}
	ts.Text["wrapper"] = wrapper

}

type PostLink string

const (
	PostInternalThread  PostLink = "post-internal-thread"
	ThreadInternalBoard PostLink = "thread-internal-board"
	PostInternalBoard   PostLink = "post-internal-board"
	ThreadExternalBoard PostLink = "thread-external-board"
	PostExternalBoard   PostLink = "post-external-board"
)

// parses entire input's post links, all instances will be replaced
func (ts *TemplateStore) ParsePostLinks(text string) (string, error) {
	t := ts.PostLinkTempl
	result := text

	for _, postlink := range ts.PostLinkKinds {
		rxp, ok := PostLinkRegex[string(postlink)]
		if !ok {
			panic(fmt.Sprintf("unresolvable regex - %s", postlink))
		}
		result = rxp.ReplaceAllStringFunc(result, postLinkReplWrapper(postlink, t))
	}
	return result, nil
}

// func constructor for RAS regexp func
func postLinkReplWrapper(kind PostLink, tpl *texttempl.Template) func(string) string {
	return func(s string) string {
		content := strings.ReplaceAll(s, "&gt;", "")
		content = strings.ReplaceAll(content, "&lt;", "")

		innert := &innerTemplate{
			Content:   content,
			ClassList: fmt.Sprintf("%s post-link", string(kind)),
		}

		buf := new(bytes.Buffer)
		err := tpl.Execute(buf, innert)
		if err != nil {
			panic(fmt.Sprintf("post links template execution failed: \n kind: %s\n content: %s\n classlist: %s\n err: %+v\n", kind, s, innert.ClassList, err))
		}

		return buf.String()
	}
}

// uses go's template system to sanitize html into their character codes utf-8 (js uses utf-16)
// raw text should already be sanitized for destructive content earlier
func (ts *TemplateStore) ReplaceChars(text string) (string, error) {
	t := ts.HtmlReplTempl
	if t == nil {
		return "", fmt.Errorf("unresolvable template %s", "wrapper")
	}

	innert := &innerTemplate{
		Content: text,
	}

	buf := new(bytes.Buffer)
	err := t.Execute(buf, innert)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// wraps content in a container div with the content-body css class. this occurs for all
// types of submitted content as article body text
func (ts *TemplateStore) WrapContent(text string) (string, error) {
	t, ok := ts.Text["wrapper"]
	if !ok {
		return "", fmt.Errorf("unresolvable template %s", "wrapper")
	}

	ic := &innerTemplate{
		Content:   text,
		ClassList: "content-body",
	}

	buf := new(bytes.Buffer)
	err := t.Execute(buf, ic)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// parses out excessive new lines & whitespace and deliminates lines into paragraph tags
// any line without text (line is only newline char) is the space between paragraphs, which is why we
// parse out extra nonsesne here to make it easier. also normalizes line endings to LF style endings
func (ts *TemplateStore) WrapParagraphs(text string) (string, error) {
	passage := ts.NormalizeLineEndings(text)
	passage = ExtraSpaceLimit.ReplaceAllLiteralString(passage, " ")
	passage = LineStartEndSpaceFix.ReplaceAllLiteralString(passage, "")
	passage = ExcessiveNewLineFix.ReplaceAllLiteralString(passage, "<#delim#>") // this is okay. tags are parsed out earlier
	passage = CtrlCharReplace.ReplaceAllLiteralString(passage, "<br>")

	ptagArr := strings.Split(passage, "<#delim#>")
	resultStr := ""

	for _, v := range ptagArr {
		resultStr += ts.executeTemplateParagraph(v)
	}

	return resultStr, nil
}

// normalizes line endings between windows/mac to all use linux LF style endings
func (ts *TemplateStore) NormalizeLineEndings(text string) string {
	bytestr := []byte(text)
	bytestr = bytes.Replace(bytestr, []byte{13, 10}, []byte{10}, -1)
	bytestr = bytes.Replace(bytestr, []byte{13}, []byte{10}, -1)
	return string(bytestr)
}

// parses user content to generate an html output
func (ts *TemplateStore) Parse(text string) (string, error) {
	str, err := ts.ReplaceChars(text)
	if err != nil {
		return "", err
	}

	paras, err := ts.WrapParagraphs(str)
	if err != nil {
		return "", err
	}

	postlinks, err := ts.ParsePostLinks(paras)
	if err != nil {
		return "", err
	}

	wrapped, err := ts.WrapContent(postlinks)
	if err != nil {
		return "", err
	}

	return wrapped, nil
}

func (ts *TemplateStore) executeTemplateParagraph(text string) string {
	t, ok := ts.Text["paragraph"]
	if !ok {
		panic("paragraph template is unresolvable")
	}

	innert := &innerTemplate{
		Content: text,
	}

	buf := new(bytes.Buffer)
	err := t.Execute(buf, innert)
	if err != nil {
		return ""
	}

	return buf.String()
}
