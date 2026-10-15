// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type setResourceStateSDK struct{}

func SetResourceStateSDK() setResourceStateSDK {
	return setResourceStateSDK{}
}

func (r setResourceStateSDK) Read(ctx context.Context, params InterceptorParamsSDK) diag.Diagnostics {
	var diags diag.Diagnostics

	switch params.When {
	case After:
		if params.IncludeResource {
			tfTypeResource, err := params.ResourceData.TfTypeResourceState()
			if err != nil {
				diags.Append(diag.NewErrorDiagnostic(
					"Error Listing Remote Resources",
					"An unexpected error occurred converting resource state. "+
						"This is always an error in the provider. "+
						"Please report the following to the provider developer:\n\n"+
						"Error: "+err.Error(),
				))
				return diags
			}

			diags.Append(params.Result.Resource.Set(ctx, *tfTypeResource)...)
		}
	}

	return diags
}
