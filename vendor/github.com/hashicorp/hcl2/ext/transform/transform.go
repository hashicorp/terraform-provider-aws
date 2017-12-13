package transform

import (
	"github.com/hashicorp/hcl2/hcl"
)

// Shallow is equivalent to calling transformer.TransformBody(body), and
// is provided only for completeness of the top-level API.
func Shallow(body hcl.Body, transformer Transformer) hcl.Body {
	return transformer.TransformBody(body)
}

// Deep applies the given transform to the given body and then
// wraps the result such that any descendent blocks that are decoded will
// also have the transform applied to their bodies.
//
// This allows for language extensions that define a particular block type
// for a particular body and all nested blocks within it.
//
// Due to the wrapping behavior, the body resulting from this function
// will not be of the type returned by the transformer. Callers may call
// only the methods defined for interface hcl.Body, and may not type-assert
// to access other methods.
func Deep(body hcl.Body, transformer Transformer) hcl.Body {
	return deepWrapper{
		Transformed: transformer.TransformBody(body),
		Transformer: transformer,
	}
}

// deepWrapper is a hcl.Body implementation that ensures that a given
// transformer is applied to another given body when content is extracted,
// and that it recursively applies to any child blocks that are extracted.
type deepWrapper struct {
	Transformed hcl.Body
	Transformer Transformer
}

func (w deepWrapper) Content(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Diagnostics) {
	content, diags := w.Transformed.Content(schema)
	content = w.transformContent(content)
	return content, diags
}

func (w deepWrapper) PartialContent(schema *hcl.BodySchema) (*hcl.BodyContent, hcl.Body, hcl.Diagnostics) {
	content, remain, diags := w.Transformed.PartialContent(schema)
	content = w.transformContent(content)
	return content, remain, diags
}

func (w deepWrapper) transformContent(content *hcl.BodyContent) *hcl.BodyContent {
	if len(content.Blocks) == 0 {
		// Easy path: if there are no blocks then there are no child bodies to wrap
		return content
	}

	// Since we're going to change things here, we'll be polite and clone the
	// structure so that we don't risk impacting any internal state of the
	// original body.
	ret := &hcl.BodyContent{
		Attributes:       content.Attributes,
		MissingItemRange: content.MissingItemRange,
		Blocks:           make(hcl.Blocks, len(content.Blocks)),
	}

	for i, givenBlock := range content.Blocks {
		// Shallow-copy the block so we can mutate it
		newBlock := *givenBlock
		newBlock.Body = Deep(newBlock.Body, w.Transformer)
		ret.Blocks[i] = &newBlock
	}

	return ret
}

func (w deepWrapper) JustAttributes() (hcl.Attributes, hcl.Diagnostics) {
	// Attributes can't have bodies or nested blocks, so this is just a thin wrapper.
	return w.Transformed.JustAttributes()
}

func (w deepWrapper) MissingItemRange() hcl.Range {
	return w.Transformed.MissingItemRange()
}
