// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package create

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
)

// Name returns in order the name if non-empty, a prefix generated name if non-empty, or fully generated name prefixed with terraform-
func Name(name string, namePrefix string) string {
	return NewNameGenerator(WithConfiguredName(name), WithConfiguredPrefix(namePrefix)).Generate()
}

// hasResourceUniqueIDPlusAdditionalSuffix returns true if the string has the built-in unique ID suffix plus an additional suffix
func hasResourceUniqueIDPlusAdditionalSuffix(s string, additionalSuffix string) bool {
	re := regexache.MustCompile(fmt.Sprintf("[[:xdigit:]]{%d}%s$", id.UniqueIDSuffixLength, additionalSuffix))
	return re.MatchString(s)
}

// NamePrefixFromName returns a name prefix if the string matches prefix criteria
//
// The input to this function must be strictly the "name" and not any
// additional information such as a full Amazon Resource Name (ARN).
//
// An expected usage might be:
//
//	d.Set("name_prefix", create.NamePrefixFromName(d.Id()))
func NamePrefixFromName(name string) *string {
	return NamePrefixFromNameWithSuffix(name, "")
}

func NamePrefixFromNameWithSuffix(name, nameSuffix string) *string {
	if !hasResourceUniqueIDPlusAdditionalSuffix(name, nameSuffix) {
		return nil
	}

	namePrefixIndex := len(name) - id.UniqueIDSuffixLength - len(nameSuffix)

	if namePrefixIndex <= 0 {
		return nil
	}

	namePrefix := name[:namePrefixIndex]

	return &namePrefix
}

type nameGenerator struct {
	configuredName   string
	configuredPrefix string
	defaultPrefix    string
	suffix           string
}

// nameGeneratorOptionsFunc is a type alias for a name generator functional option.
type NameGeneratorOptionsFunc func(*nameGenerator)

// WithConfiguredName is a helper function to construct functional options
// that set a name generator's configured name value.
// An empty ("") configured name inidicates that no name was configured.
func WithConfiguredName(name string) NameGeneratorOptionsFunc {
	return func(g *nameGenerator) {
		g.configuredName = name
	}
}

// WithConfiguredPrefix is a helper function to construct functional options
// that set a name generator's configured prefix value.
// An empty ("") configured prefix inidicates that no prefix was configured.
func WithConfiguredPrefix(prefix string) NameGeneratorOptionsFunc {
	return func(g *nameGenerator) {
		g.configuredPrefix = prefix
	}
}

// WithDefaultPrefix is a helper function to construct functional options
// that set a name generator's default prefix value.
func WithDefaultPrefix(prefix string) NameGeneratorOptionsFunc {
	return func(g *nameGenerator) {
		g.defaultPrefix = prefix
	}
}

// WithSuffix is a helper function to construct functional options
// that set a name generator's suffix value.
func WithSuffix(suffix string) NameGeneratorOptionsFunc {
	return func(g *nameGenerator) {
		g.suffix = suffix
	}
}

// NewNameGenerator returns a new name generator from the specified varidaic list of functional options.
func NewNameGenerator(optFns ...NameGeneratorOptionsFunc) *nameGenerator {
	g := &nameGenerator{defaultPrefix: id.UniqueIdPrefix}

	for _, optFn := range optFns {
		optFn(g)
	}

	return g
}

// Generate generates a new name.
func (g *nameGenerator) Generate() string {
	if g.configuredName != "" {
		return g.configuredName
	}

	prefix := g.defaultPrefix
	if g.configuredPrefix != "" {
		prefix = g.configuredPrefix
	}
	return id.PrefixedUniqueId(prefix) + g.suffix
}
