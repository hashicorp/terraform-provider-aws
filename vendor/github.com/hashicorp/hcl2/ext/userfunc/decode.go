package userfunc

import (
	"github.com/hashicorp/hcl2/gohcl"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var funcBodySchema = &hcl.BodySchema{
	Attributes: []hcl.AttributeSchema{
		{
			Name:     "params",
			Required: true,
		},
		{
			Name:     "variadic_param",
			Required: false,
		},
		{
			Name:     "result",
			Required: true,
		},
	},
}

func decodeUserFunctions(body hcl.Body, blockType string, contextFunc ContextFunc) (funcs map[string]function.Function, remain hcl.Body, diags hcl.Diagnostics) {
	schema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       blockType,
				LabelNames: []string{"name"},
			},
		},
	}

	content, remain, diags := body.PartialContent(schema)
	if diags.HasErrors() {
		return nil, remain, diags
	}

	// first call to getBaseCtx will populate context, and then the same
	// context will be used for all subsequent calls. It's assumed that
	// all functions in a given body should see an identical context.
	var baseCtx *hcl.EvalContext
	getBaseCtx := func() *hcl.EvalContext {
		if baseCtx == nil {
			if contextFunc != nil {
				baseCtx = contextFunc()
			}
		}
		// baseCtx might still be nil here, and that's okay
		return baseCtx
	}

	funcs = make(map[string]function.Function)
	for _, block := range content.Blocks {
		name := block.Labels[0]
		funcContent, funcDiags := block.Body.Content(funcBodySchema)
		diags = append(diags, funcDiags...)
		if funcDiags.HasErrors() {
			continue
		}

		paramsExpr := funcContent.Attributes["params"].Expr
		resultExpr := funcContent.Attributes["result"].Expr
		var varParamExpr hcl.Expression
		if funcContent.Attributes["variadic_param"] != nil {
			varParamExpr = funcContent.Attributes["variadic_param"].Expr
		}

		var params []string
		var varParam string

		paramsDiags := gohcl.DecodeExpression(paramsExpr, nil, &params)
		diags = append(diags, paramsDiags...)
		if paramsDiags.HasErrors() {
			continue
		}
		if varParamExpr != nil {
			paramsDiags := gohcl.DecodeExpression(varParamExpr, nil, &varParam)
			diags = append(diags, paramsDiags...)
			if paramsDiags.HasErrors() {
				continue
			}
		}

		spec := &function.Spec{}
		for _, paramName := range params {
			spec.Params = append(spec.Params, function.Parameter{
				Name: paramName,
				Type: cty.DynamicPseudoType,
			})
		}
		if varParamExpr != nil {
			spec.VarParam = &function.Parameter{
				Name: varParam,
				Type: cty.DynamicPseudoType,
			}
		}
		impl := func(args []cty.Value) (cty.Value, error) {
			ctx := getBaseCtx()
			ctx = ctx.NewChild()
			ctx.Variables = make(map[string]cty.Value)

			// The cty function machinery guarantees that we have at least
			// enough args to fill all of our params.
			for i, paramName := range params {
				ctx.Variables[paramName] = args[i]
			}
			if spec.VarParam != nil {
				varArgs := args[len(params):]
				ctx.Variables[varParam] = cty.TupleVal(varArgs)
			}

			result, diags := resultExpr.Value(ctx)
			if diags.HasErrors() {
				// Smuggle the diagnostics out via the error channel, since
				// a diagnostics sequence implements error. Caller can
				// type-assert this to recover the individual diagnostics
				// if desired.
				return cty.DynamicVal, diags
			}
			return result, nil
		}
		spec.Type = func(args []cty.Value) (cty.Type, error) {
			val, err := impl(args)
			return val.Type(), err
		}
		spec.Impl = func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			return impl(args)
		}
		funcs[name] = function.New(spec)
	}

	return funcs, remain, diags
}
