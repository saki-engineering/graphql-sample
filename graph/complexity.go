package graph

import (
	"github.com/saki-engineering/graphql-sample/internal"
)

func ComplexityConfig() internal.ComplexityRoot {
	var c internal.ComplexityRoot

	c.Query.Node = func(childComplexity int, id string) int {
		return 1
	}
	c.ProjectV2.Title = func(childComplexity int) int {
		return 1
	}
	return c
}
