package yamlp

import (
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

type DocKind uint32

const (
	Plain DocKind = 1 << iota
	Ref
	Refs
	RefMerge
	RefsMerge
	Export
)

func (dk DocKind) String() string {
	switch dk {
	case Plain:
		return "plain"
	case Ref:
		return "ref"
	case Refs:
		return "ref[]"
	case RefMerge:
		return "<<ref"
	case Export:
		return "export"
	default:
		return ""
	}
}

type Doc struct {
	Kind DocKind
	Val  string
}

func ToDocKind(str string) DocKind {
	switch str {
	case "ref":
		return Ref
	case "refs":
		return Refs
	case "<<ref":
		return RefMerge
	case "<<ref[]":
		return RefsMerge
	case "export":
		return Export
	default:
		return Plain
	}
}

func determineDoc(n *yqlib.CandidateNode) Doc {
	doc := Doc{Plain, ""}

	fixHeadComment(n)
	if n.HeadComment == "" {
		return doc
	}

	trimmed := strings.TrimLeft(n.HeadComment, "# ")

	tokens := strings.Split(trimmed, "/")
	kind := tokens[0]
	doc.Kind = ToDocKind(kind)

	if doc.Kind == Export {
		if len(tokens) == 2 {
			doc.Val = tokens[1]
		} else {
			doc.Val = "."
		}
	}

	n.HeadComment = ""
	return doc
}

// required due to bug in yaml.v3 module
// https://github.com/go-yaml/yaml/issues/801
func fixHeadComment(n *yqlib.CandidateNode) {
	if n.HeadComment != "" {
		return
	}

	if len(n.Content) > 0 {
		n.HeadComment = n.Content[0].HeadComment
		n.Content[0].HeadComment = ""
	}
}
