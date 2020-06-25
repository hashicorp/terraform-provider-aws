package terraform

import (
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
)

// GraphNodeSubPath says that a node is part of a graph with a
// different path, and the context should be adjusted accordingly.
type GraphNodeSubPath interface {
	Path() addrs.ModuleInstance
}
