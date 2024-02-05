package types

import (
	"regexp"
)

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
