package types

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"
)

type TemplateStore struct {
	PostLinks map[string]*template.Template
}

func NewTemplateStore() *TemplateStore {
	t := &TemplateStore{
		PostLinks: map[string]*template.Template{},
	}
	t.Hydrate()
	return t
}

func (ts *TemplateStore) Hydrate() {
	postLinkKinds := []PostLink{
		PostInternalThread,
		ThreadInternalBoard,
		PostInternalBoard,
		ThreadExternalBoard,
		PostExternalBoard,
	}

	for _, v := range postLinkKinds {
		ts.PostLinks[string(v)] = template.New(fmt.Sprintf(`<button class="%s {{ .ClassList }}">{{ .Content }}</button>`, string(v)))
	}

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
	return &PostLinkTemplate{
		Kind:       kind,
		PostNumber: post,
		ThreadSlug: slug,
		BoardShort: short,
		ClassList:  string(kind),
	}
}

func (plt *PostLinkTemplate) Parse(ts *TemplateStore) error {
	t, ok := ts.PostLinks[string(plt.Kind)]
	if !ok {
		return fmt.Errorf("unresolvable link type %s", plt.Kind)
	}
	buf := new(bytes.Buffer)
	err := t.Execute(buf, plt)
	if err != nil {
		return err
	}
	plt.Content = buf.String()

	return nil
}

type InternalTemplate interface {
	HTML() string
}

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
