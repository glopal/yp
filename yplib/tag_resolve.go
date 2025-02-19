package yplib

import (
	"errors"
	"strings"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("!resolve", resolveResolver)
}

func resolveResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	tokens := strings.Split(strings.TrimSpace(rc.Target.Value), " ")
	if len(tokens) != 2 {
		return nil, errors.New("expected 2 expressions")
	}

	ctx, err := yq(createContext(rc), tokens[0])
	if err != nil {
		return nil, err
	}

	partial, err := yq(createContext(rc), tokens[1])
	if err != nil {
		return nil, err
	}

	partialNode := &Node{
		Dir:           rc.Node.Dir,
		File:          rc.Node.File,
		CandidateNode: partial,
		tagNodes:      getTagNodes(partial),
	}

	clone := partialNode.Clone()
	err = clone.Resolve(NewContextNode(ctx), rc.Vars, rc.Opts)
	if err != nil {
		return nil, err
	}

	return clone.CandidateNode, nil
}
