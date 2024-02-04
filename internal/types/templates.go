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
	// replacements
	tmpl_blockquote string = `<blockquote class="post-quote">${1}</blockquote>`
	tmpl_epl        string = `<button class="post-link epl" id="epl-${1}-${2}">${1} / ${2}</button>`
	tmpl_ipl        string = `<button class="post-link ipl" id="ipl-${1}">${1}</button>`

	// whitespace reduction on beginning of strings.
	rxp_ws_start *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)^[[:blank:]]{1,}`),
		Sub: "",
	}

	// whitespace preservation between lines
	rxp_ws_mid *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)[\n\r]{1,}`),
		Sub: "<br>",
	}

	// whitespace reduction on end of strings.
	// takes care of excessive newline feeds as well
	rxp_ws_end *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)[\n\r\s]{2,}$`),
		Sub: "<br>",
	}

	// template regex quotes
	rxp_quote *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)^>([^>].*)$`),
		Sub: tmpl_blockquote,
	}

	// same thread post links
	rxp_ipl *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)>>(\d{1,})`),
		Sub: tmpl_ipl,
	}

	// external post links
	rxp_epl *TRegSub = &TRegSub{
		Reg: regexp.MustCompile(`(?m)>>/([a-zA-Z]*)/(\d*)`),
		Sub: tmpl_epl,
	}
)

// Template Thread Reply
// used as a wrapper for the entirety of the reply contents, including furthur nested templates.
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
			rxp_quote,
			rxp_ipl,
			rxp_epl,
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
