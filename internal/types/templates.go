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
		"post-internal-thread":  regexp.MustCompile(`(?m)>>([[:digit:]]{1,9})[[:blank:]]`),                                        // 1 = post_num
		"thread-internal-board": regexp.MustCompile(`(?m)>>([[:alnum:]]{8,12})[[:blank:]]`),                                       // 1 = threadslug
		"post-internal-board":   regexp.MustCompile(`(?m)>>([[:alnum:]]{8,12})/([[:digit:]]{1,9})[[:blank:]]`),                    // 1 = threadslug, 2 = post_num
		"thread-external-board": regexp.MustCompile(`(?m)>>([[:alpha:]]{2,5})/([[:alnum:]]{8,12})[[:blank:]]`),                    // 1 = board short, 2 = threadslug
		"post-external-board":   regexp.MustCompile(`(?m)>>([[:alpha:]]{2,5})/([[:alnum:]]{8,12})/([[:digit:]]{1,9})[[:blank:]]`), // 1 = board short, 2 = threadslug, 3 = post_num
	}
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
		tmpl, err := texttempl.New(string(v)).Parse(fmt.Sprintf(`<button class="%s {{ .ClassList }}">{{ .Content }}</button>`, string(v)))
		if err != nil {
			panic(err)
		}
		ts.PostLinks[string(v)] = tmpl
	}

	replacement, err := template.New("wrapper").Parse("{{ .Content }}")
	if err != nil {
		panic(err)
	}
	ts.HtmlReplTempl = replacement

	paragraphs, err := texttempl.New("paragraph").Parse("<p>{{ .Content }}</p>")
	if err != nil {
		panic(err)
	}
	ts.Text["paragraph"] = paragraphs

}

type PostLink string

const (
	PostInternalThread  PostLink = "post-internal-thread"
	ThreadInternalBoard PostLink = "thread-internal-board"
	PostInternalBoard   PostLink = "post-internal-board"
	ThreadExternalBoard PostLink = "thread-external-board"
	PostExternalBoard   PostLink = "post-external-board"
)

type PostLinkTemplate struct {
	Kind       PostLink
	PostNumber int    // referenced post
	ThreadSlug string // referenced thread
	BoardShort string // referenced board
	ClassList  string
	Content    string
}

func (plt *PostLinkTemplate) InnerContent() string {
	switch plt.Kind {
	case PostInternalThread:
		return fmt.Sprintf("%d", plt.PostNumber)
	case ThreadInternalBoard:
		return plt.ThreadSlug
	case PostInternalBoard:
		return fmt.Sprintf("%s/%d", plt.ThreadSlug, plt.PostNumber)
	case ThreadExternalBoard:
		return fmt.Sprintf("%s/%s", plt.BoardShort, plt.ThreadSlug)
	case PostExternalBoard:
		return fmt.Sprintf("%s/%s/%d", plt.BoardShort, plt.ThreadSlug, plt.PostNumber)
	default:
		return "unknown link"
	}
}

func NewPostLinkTemplate(kind PostLink, post int, slug, short string) *PostLinkTemplate {
	pl := &PostLinkTemplate{
		Kind:       kind,
		PostNumber: post,
		ThreadSlug: slug,
		BoardShort: short,
		ClassList:  "post-link",
	}
	pl.Content = pl.InnerContent()
	return pl
}

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

		matchList := map[string]string{}

		for _, match := range rxp.FindAllString(text, -1) {
			innert := &innerTemplate{
				Content:   "",
				ClassList: "post-link",
			}
			buf := new(bytes.Buffer)

			content := strings.ReplaceAll(match, ">", "")
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

func (ts *TemplateStore) WrapParagraphs(text string) (string, error) {
	t, ok := ts.Text["paragraph"]
	if !ok {
		return "", fmt.Errorf("unresolvable template %s", "paragraph")
	}

	str := ""

	splits := strings.Split(text, "\n\n")
	for i, v := range splits {
		ic := &innerTemplate{
			Content: v,
		}
		buf := new(bytes.Buffer)
		err := t.Execute(buf, ic)
		if err != nil {
			return "", err
		}
		if i > 0 {
			str = str + "\n"
		}
		str = str + buf.String()
	}

	return str, nil
}

// to be removed - everything below this line - after more robust template implementation complete
type TRegSub struct {
	Reg *regexp.Regexp
	Sub string
}

var (
	rxp_post_internal_thread *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)>>(\d+)\s`),
		Sub: `<button class="post-internal-thread post-link">${1}</button>`,
	}

	rxp_thread_internal_board *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)>>([[:alnum:]]+[[:alpha:]]+[[:alnum:]]+)\s`),
		Sub: `<button class="thread-internal-board post-link">${1}</button>`,
	}

	rxp_post_internal_board *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)>>([[:alnum:]]+[[:alpha:]]+[[:alnum:]]+\/\d+)\s`),
		Sub: `<button class="post-internal-board post-link">${1}</button>`,
	}

	rxp_thread_external_board *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)>>([[:alpha:]]+\/[[:alnum:]]+[[:alpha:]]+[[:alnum:]]+)\s`),
		Sub: `<button class="thread-external-board post-link">${1}</button>`,
	}

	rxp_post_external_board *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)>>([[:alpha:]]+\/[[:alnum:]]+[[:alpha:]]+[[:alnum:]]+\/\d+)\s`),
		Sub: `<button class="post-external-board post-link">${1}</button>`,
	}

	// whitespace reduction on beginning of strings
	rxp_ws_start *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)^[[:blank:]]{1,}`),
		Sub: "",
	}

	// whitespace preservation between lines
	rxp_ws_mid *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)[\n\r]{3,}`),
		Sub: "\n",
	}

	// whitespace reduction on end of strings
	rxp_ws_end *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)[[:blank:]]{1,}$`),
		Sub: "",
	}

	// template regex quotes
	rxp_quote *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)^>([^>].+)$`),
		Sub: `<blockquote class="reply-quote">${1}</blockquote>`,
	}
)

// Template Thread Reply
// used as a wrapper for the entirety of the reply contents, including furthur nested templates
type TemplateThreadReply struct {
	Content string
	RegOps  []*TRegSub
}

func NewTemplateThreadReply(content string) *TemplateThreadReply {
	return &TemplateThreadReply{
		Content: content,
		RegOps: []*TRegSub{
			rxp_ws_start,
			rxp_ws_end,
			rxp_ws_mid,
			rxp_post_internal_thread,
			rxp_thread_internal_board,
			rxp_post_internal_board,
			rxp_thread_external_board,
			rxp_post_external_board,
			rxp_quote,
		},
	}
}

func (ttr *TemplateThreadReply) Parse() string {
	str := ttr.Content

	for _, v := range ttr.RegOps {
		str = v.Reg.ReplaceAllString(str, v.Sub)
	}

	return str
}
