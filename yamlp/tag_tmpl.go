package yamlp

import (
	"bytes"
	"text/template"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("!tmpl", tmplResolver)
}

func tmplResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	val, err := renderTemplate(rc.Target.Value, rc.Ctx)
	if err != nil {
		return nil, err
	}

	node := &yqlib.CandidateNode{}
	node.Kind = yqlib.ScalarNode
	node.Tag = "!!str"
	node.Value = val

	return node, nil
}

func renderTemplate(tmpl string, ctx *ContextNode) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}

	t.Option("missingkey=error")

	data, err := ctx.Interface()
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
