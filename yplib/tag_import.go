package yplib

import (
	"fmt"

	"github.com/mikefarah/yq/v4/pkg/yqlib"
)

func init() {
	AddTagResolver("import", importResolver)
}

func importResolver(rc ResolveContext) (*yqlib.CandidateNode, error) {
	n, exists := rc.Vars[rc.Target.Value]
	if !exists {
		return nil, fmt.Errorf("failed to import '%s': not found", rc.Target.Value)
	}

	return n.candidateNode, nil
}
