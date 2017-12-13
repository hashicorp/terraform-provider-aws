package include

import (
	"github.com/hashicorp/hcl2/ext/transform"
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
)

// Transformer builds a transformer that finds any "include" blocks in a body
// and produces a merged body that contains the original content plus the
// content of the other bodies referenced by the include blocks.
//
// blockType specifies the type of block to interpret. The conventional type name
// is "include".
//
// ctx provides an evaluation context for the path expressions in include blocks.
// If nil, path expressions may not reference variables nor functions.
//
// The given resolver is used to translate path strings (after expression
// evaluation) into bodies. FileResolver returns a reasonable implementation for
// applications that read configuration files from local disk.
//
// The returned Transformer can either be used directly to process includes
// in a shallow fashion on a single body, or it can be used with
// transform.Deep (from the sibling transform package) to allow includes
// at all levels of a nested block structure:
//
//    transformer = include.Transformer("include", nil, include.FileResolver(".", parser))
//    body = transform.Deep(body, transformer)
//    // "body" will now have includes resolved in its own content and that
//    // of any descendent blocks.
//
func Transformer(blockType string, ctx *hcl.EvalContext, resolver Resolver) transform.Transformer {
	return &transformer{
		Schema: &hcl.BodySchema{
			Blocks: []hcl.BlockHeaderSchema{
				{
					Type: blockType,
				},
			},
		},
		Ctx:      ctx,
		Resolver: resolver,
	}
}

type transformer struct {
	Schema   *hcl.BodySchema
	Ctx      *hcl.EvalContext
	Resolver Resolver
}

func (t *transformer) TransformBody(in hcl.Body) hcl.Body {
	content, remain, diags := in.PartialContent(t.Schema)

	if content == nil || len(content.Blocks) == 0 {
		// Nothing to do!
		return transform.BodyWithDiagnostics(remain, diags)
	}

	bodies := make([]hcl.Body, 1, len(content.Blocks)+1)
	bodies[0] = remain // content in "remain" takes priority over includes
	for _, block := range content.Blocks {
		incContent, incDiags := block.Body.Content(includeBlockSchema)
		diags = append(diags, incDiags...)
		if incDiags.HasErrors() {
			continue
		}

		pathExpr := incContent.Attributes["path"].Expr
		var path string
		incDiags = gohcl.DecodeExpression(pathExpr, t.Ctx, &path)
		diags = append(diags, incDiags...)
		if incDiags.HasErrors() {
			continue
		}

		incBody, incDiags := t.Resolver.ResolveBodyPath(path, pathExpr.Range())
		bodies = append(bodies, transform.BodyWithDiagnostics(incBody, incDiags))
	}

	return hcl.MergeBodies(bodies)
}

var includeBlockSchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "path",
			Required: true,
		},
	},
}
