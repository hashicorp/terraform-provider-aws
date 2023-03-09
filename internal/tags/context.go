package tags

import (
	"context"
)

// InContext represents the tagging information kept in Context.
type InContext struct {
	DefaultConfig *DefaultConfig
	IgnoreConfig  *IgnoreConfig
}

// NewContext returns a Context enhanced with tagging information.
func NewContext(ctx context.Context, defaultConfig *DefaultConfig, ignoreConfig *IgnoreConfig) context.Context {
	v := InContext{
		DefaultConfig: defaultConfig,
		IgnoreConfig:  ignoreConfig,
	}

	return context.WithValue(ctx, tagKey, &v)
}

func FromContext(ctx context.Context) (*InContext, bool) {
	v, ok := ctx.Value(tagKey).(*InContext)
	return v, ok
}

func MergedTagsFromContext(ctx context.Context) (*KeyValueTags, bool) {
	v, ok := ctx.Value(MergedTagsKey).(*KeyValueTags)
	return v, ok
}

type key int

var tagKey key
var MergedTagsKey key
