package tags

import (
	"context"
)

// InContext represents the tagging information kept in Context.
type InContext struct {
	DefaultConfig *DefaultConfig
	IgnoreConfig  *IgnoreConfig
}

func FromContext(ctx context.Context) (*InContext, bool) {
	v, ok := ctx.Value(TagKey).(*InContext)
	return v, ok
}

type key int

var TagKey key
