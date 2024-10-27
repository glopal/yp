package yamlp

import (
	"bytes"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("!tmpl", tmplResolver)
}

func tmplResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	t, err := template.New("").Parse(rc.Target.Value)
	if err != nil {
		return nil, err
	}

	t.Option("missingkey=error")

	data, err := getRefInterface(t, rc.Node.NodeContext, rc.Refs)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, data)
	if err != nil {
		panic(err)
	}

	node := &yqlib.CandidateNode{}
	node.Kind = yqlib.ScalarNode
	node.Tag = "!!str"
	node.Value = buf.String()

	return node, nil
}

func getRefInterface(t *template.Template, nc NodeContext, refs map[string]*Node) (map[string]interface{}, error) {
	refMap := map[string]interface{}{}

	for _, rootVar := range getRootVariables(t.Root.Nodes) {
		if n, exists := refs[rootVar]; exists {
			if !nc.IsRef || nc.Name != rootVar {
				err := n.Resolve(refs)
				if err != nil {
					return nil, err
				}
			}

			i, err := n.Interface()
			if err != nil {
				return nil, err
			}

			refMap[rootVar] = i
		}
	}

	return refMap, nil
}

func getRootVariables(nodes []parse.Node) []string {
	rootVars := []string{}
	for _, n := range nodes {

		switch node := n.(type) {
		case *parse.ActionNode:
			for _, cmd := range node.Pipe.Cmds {
				rootVars = append(rootVars, extractRootVars(cmd.String())...)
			}
		case *parse.BreakNode:
		case *parse.CommentNode:
		case *parse.ContinueNode:
		case *parse.IfNode:
			for _, cmd := range node.BranchNode.Pipe.Cmds {
				rootVars = append(rootVars, extractRootVars(cmd.String())...)
			}

			rootVars = append(rootVars, getRootVariables(node.List.Nodes)...)

			if node.ElseList != nil {
				rootVars = append(rootVars, getRootVariables(node.ElseList.Nodes)...)
			}
		case *parse.ListNode:
		case *parse.RangeNode:
			for _, cmd := range node.BranchNode.Pipe.Cmds {
				rootVars = append(rootVars, extractRootVars(cmd.String())...)
			}
			rootVars = append(rootVars, getRootVariables(node.List.Nodes)...)
		case *parse.TemplateNode:
		case *parse.TextNode:
		case *parse.WithNode:
		default:
		}
	}

	return unique(rootVars)
}

func extractRootVars(cmd string) []string {
	rootVars := []string{}

	for _, token := range strings.Split(cmd, " ") {
		if strings.HasPrefix(token, ".") {
			if vt := strings.Split(token, "."); len(vt) > 1 {
				rootVars = append(rootVars, vt[1])
			}

		}
	}

	return rootVars
}

func unique(items []string) []string {
	m := make(map[string]struct{}, len(items))

	for _, item := range items {
		m[item] = struct{}{}
	}

	out := make([]string, 0, len(m))

	for k, _ := range m {
		out = append(out, k)
	}

	return out
}
