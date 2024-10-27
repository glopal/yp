//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"text/template/parse"
)

type ValProvider struct {
	v string
}

func (vp *ValProvider) String() string {
	return "OTHER"
}
func main() {
	t, err := template.New("").Parse(`BLA {{ range $i, $v := .items }}{{ if eq .app.one.two "aa" }}{{.test.ggg}}{{else if eq .else "bla"}}{{.else.content}}{{else}}{{.final}}{{end}}{{end}}{{.outside}}`)
	if err != nil {
		panic(err)
	}

	fmt.Println(getRootVariables(t.Root.Nodes))
	// ctx := map[string]*ValProvider{
	// 	"app": {
	// 		v: "TESTING",
	// 	},
	// }

	t.Funcs(template.FuncMap{
		"app": func() string {
			// fmt.Println(args)
			return "FUNC"
		},
	})
	buf := new(bytes.Buffer)
	err = t.Execute(buf, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println(buf.String())

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
