package yamlp

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
)

var decoder = yqlib.NewYamlDecoder(yqlib.YamlPreferences{
	Indent:                      2,
	ColorsEnabled:               false,
	LeadingContentPreProcessing: false,
	PrintDocSeparators:          true,
	UnwrapScalar:                true,
	EvaluateTogether:            false,
})

func init() {
	// disable yqlib debug logging
	leveled := logging.AddModuleLevel(logging.NewLogBackend(os.Stderr, "", 0))
	leveled.SetLevel(logging.ERROR, "")
	yqlib.GetLogger().SetBackend(leveled)
}

func LoadDir(dir string, opts ...func(*loadOptions)) (*Nodes, error) {
	options := defaultLoadOptions()
	for _, o := range opts {
		o(options)
	}

	nodes := NewNodes()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !IsYamlFile(path) || options.omitFunc(path) {
			return nil
		}

		n, err := LoadFile(path)
		if err != nil {
			return err
		}

		nodes.Append(n)

		return nil
	})

	return nodes, err
}

func LoadFile(file string) (*Nodes, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	nodes := NewNodes()

	err = decoder.Init(f)
	if err != nil {
		return nil, err
	}

	for {
		node, err := decoder.Decode()

		// check it was parsed
		if err == nil {
			fixHeadComment(node)

			n := &Node{
				CandidateNode: node,
				NodeContext: NodeContext{
					Dir:  filepath.Dir(file),
					Name: strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)),
				},
			}

			if ref := parseRef(node.HeadComment); ref != "" {
				n.NodeContext.IsRef = true
				n.NodeContext.Name = ref
				nodes.refs[ref] = n
			} else {
				nodes.nodes = append(nodes.nodes, n)
			}

			continue
		}
		// break the loop in case of EOF
		if errors.Is(err, io.EOF) {
			break
		}
	}

	return nodes, nil
}

func IsYamlFile(file string) bool {
	ext := filepath.Ext(file)
	return ext == ".yml" || ext == ".yaml"
}

// required due to bug in yaml.v3 module
// https://github.com/go-yaml/yaml/issues/801
func fixHeadComment(n *yqlib.CandidateNode) {
	if n.HeadComment != "" {
		return
	}

	if len(n.Content) > 0 {
		n.HeadComment = n.Content[0].HeadComment
	}
}

func parseRef(headComment string) string {
	if headComment == "" || !strings.HasPrefix(strings.TrimLeft(headComment, "# "), "ref/") {
		return ""
	}

	tokens := strings.Split(headComment, "/")
	if len(tokens) != 2 {
		return ""
	}

	return tokens[1]
}
