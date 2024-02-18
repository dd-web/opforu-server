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
	PostLinkPatterns = map[string]*regexp.Regexp{
		"post-internal-thread":  regexp.MustCompile(`(?m)&gt;&gt;([[:digit:]]{1,9})[[:blank:]]`),                                        // 1 = post_num
		"thread-internal-board": regexp.MustCompile(`(?m)&gt;&gt;([[:alnum:]]{8,12})[[:blank:]]`),                                       // 1 = threadslug
		"post-internal-board":   regexp.MustCompile(`(?m)&gt;&gt;([[:alnum:]]{8,12})/([[:digit:]]{1,9})[[:blank:]]`),                    // 1 = threadslug, 2 = post_num
		"thread-external-board": regexp.MustCompile(`(?m)&gt;&gt;([[:alpha:]]{2,5})/([[:alnum:]]{8,12})[[:blank:]]`),                    // 1 = board short, 2 = threadslug
		"post-external-board":   regexp.MustCompile(`(?m)&gt;&gt;([[:alpha:]]{2,5})/([[:alnum:]]{8,12})/([[:digit:]]{1,9})[[:blank:]]`), // 1 = board short, 2 = threadslug, 3 = post_num
	}
	NewLineFeedLimit = regexp.MustCompile(`(?m)[[:cntrl:]]{2,}`)
)

type TemplateStore struct {
	// html templates replace all html character codes, our case is iterative which means we must use text templates
	// for any html we wish to inject because our previous iterations become invalid html. This is used to invalidate
	// any html a user may have submitted. Use it once and once only against any submitted content. (sanitize first)
	HtmlReplTempl *template.Template
	PostLinks     map[string]*texttempl.Template
	Text          map[string]*texttempl.Template
	PostLinkKinds []PostLink
}

func NewTemplateStore() *TemplateStore {
	t := &TemplateStore{
		HtmlReplTempl: &template.Template{},
		PostLinks:     map[string]*texttempl.Template{},
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

	for _, v := range ts.PostLinkKinds {
		tmpl, err := texttempl.New(string(v)).Parse(`<button class="{{ .ClassList }}">{{ .Content }}</button>`)
		if err != nil {
			panic(err)
		}
		ts.PostLinks[string(v)] = tmpl // (v) post link
	}

	replacement, err := template.New("utf-8-replace").Parse("{{ .Content }}")
	if err != nil {
		panic(err)
	}
	ts.HtmlReplTempl = replacement // utf-8 char code replacements

	paragraphs, err := texttempl.New("paragraph").Parse("<p>{{ .Content }}</p>")
	if err != nil {
		panic(err)
	}
	ts.Text["paragraph"] = paragraphs // <p> tag wrapper

	wrapper, err := texttempl.New("wrapper").Parse(`<div class="{{ .ClassList }}">{{ .Content }}</div>`)
	if err != nil {
		panic(err)
	}
	ts.Text["wrapper"] = wrapper // <div> wrapper

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
	parsed := text

	for _, postlinktype := range ts.PostLinkKinds {
		tmpl, ok := ts.PostLinks[string(postlinktype)]
		if !ok {
			return "", fmt.Errorf("unresolvable template: %s", string(postlinktype))
		}

		rxp, ok := PostLinkPatterns[string(postlinktype)]
		if !ok {
			return "", fmt.Errorf("unresolvable pattern match: %s", string(postlinktype))
		}

		matchList := map[string]string{} // map instead of array is to proactively prevent subtle bug

		for _, match := range rxp.FindAllString(text, -1) {
			innert := &innerTemplate{
				Content:   "",
				ClassList: fmt.Sprintf("%s post-link", string(postlinktype)),
			}
			buf := new(bytes.Buffer)

			content := strings.ReplaceAll(match, "&gt;", "")
			content = strings.ReplaceAll(content, " ", "")
			innert.Content = content

			if innert.Content != "" {
				err := tmpl.Execute(buf, innert)
				if err != nil {
					return "", err
				}
				matchList[match] = buf.String() + " "
			}
		}

		for k, v := range matchList {
			parsed = strings.ReplaceAll(parsed, k, v)
		}
	}

	return parsed, nil
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

// wraps whitespace deliminated text with paragraph tags
func (ts *TemplateStore) WrapParagraphs(text string) (string, error) {
	t, ok := ts.Text["paragraph"]
	if !ok {
		return "", fmt.Errorf("unresolvable template %s", "paragraph")
	}

	strRepl := text
	for _, match := range NewLineFeedLimit.FindAllString(text, -1) {
		strRepl = strings.ReplaceAll(strRepl, match, "<br>")
	}

	matchSplit := strings.Split(strRepl, "<br>")
	finished := ""

	for _, v := range matchSplit {
		innert := &innerTemplate{
			Content: v,
		}
		if v != "" {
			buf := new(bytes.Buffer)
			err := t.Execute(buf, innert)
			if err != nil {
				return "", err
			}
			finished = finished + buf.String()

		}
	}
	return finished, nil
}

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
