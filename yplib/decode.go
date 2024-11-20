package yplib

import "errors"

type Decoder[T any] struct {
	Decode func() []T
}

func Decode[T any](ns *Nodes) ([]T, error) {
	outNodes, err := ns.resolveOut()
	if err != nil {
		return nil, err
	}

	if len(outNodes) != 1 {
		return nil, errors.New("decoder operates on a single out node list, might change this later")
	}

	decoded := *new([]T)

	for _, n := range outNodes[0].node.Content {
		yn, err := n.MarshalYAML()
		if err != nil {
			return nil, err
		}
		var i T
		err = yn.Decode(&i)
		if err != nil {
			return nil, err
		}

		decoded = append(decoded, i)
	}

	return decoded, nil
}
