package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// InterceptorFunc is functionality invoked during the CRUD request lifecycle.
// If a Before interceptor returns Diagnostics indicating an error occurred then
// no further interceptors in the chain are run and neither is the schema's method.
// In other cases all interceptors in the chain are run.
type InterceptorFunc func(context.Context, *schema.ResourceData, any, diag.Diagnostics) (context.Context, diag.Diagnostics)

// Interceptor represents a single interceptor.
type Interceptor struct {
	When When
	Why  Why
	Func InterceptorFunc
}

// When represents the point in the CRUD request lifecycle that an interceptor is run.
// Multiple values can be ORed together.
type When uint16

const (
	Before  When = 1 << iota // Interceptor is invoked before call to method in schema
	After                    // Interceptor is invoked after successful call to method in schema
	OnError                  // Interceptor is invoked after unsuccessful call to method in schema
	Finally                  // Interceptor is invoked after After or OnError
)

// Why represents the CRUD operation(s) that an interceptor is run.
// Multiple values can be ORed together.
type Why uint16

const (
	Create Why = 1 << iota // Interceptor is invoked for a Create call
	Read                   // Interceptor is invoked for a Read call
	Update                 // Interceptor is invoked for a Update call
	Delete                 // Interceptor is invoked for a Delete call
)

type Interceptors []Interceptor

func (v *Interceptors) Append(when When, why Why, f InterceptorFunc) {
	interceptor := Interceptor{
		When: when,
		Why:  why,
		Func: f,
	}

	if v == nil {
		*v = Interceptors{interceptor}
	} else {
		*v = append(*v, interceptor)
	}
}

func (v *Interceptors) Why(why Why) Interceptors {
	var interceptors Interceptors

	for _, v := range *v {
		if v.Why&why != 0 {
			interceptors = append(interceptors, v)
		}
	}

	return interceptors
}

func InvokeHandler[F ~func(context.Context, *schema.ResourceData, any) diag.Diagnostics](ctx context.Context, d *schema.ResourceData, meta any, interceptors []Interceptor, f F) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, v := range interceptors {
		if v.When&Before != 0 {
			ctx, diags = v.Func(ctx, d, meta, diags)

			// Short circuit if any Before interceptor errors.
			if diags.HasError() {
				return diags
			}
		}
	}

	reversed := slices.Reverse(interceptors)
	diags = f(ctx, d, meta)

	if diags.HasError() {
		for _, v := range reversed {
			if v.When&OnError != 0 {
				ctx, diags = v.Func(ctx, d, meta, diags)
			}
		}
	} else {
		for _, v := range reversed {
			if v.When&After != 0 {
				ctx, diags = v.Func(ctx, d, meta, diags)
			}
		}
	}

	for _, v := range reversed {
		if v.When&Finally != 0 {
			ctx, diags = v.Func(ctx, d, meta, diags)
		}
	}

	return diags
}
