// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type setRegionInterceptorSDK struct{}

func SetRegionInterceptorSDK() setRegionInterceptorSDK {
	return setRegionInterceptorSDK{}
}

// Copied from resourceSetRegionInStateInterceptor.read()
func (r setRegionInterceptorSDK) Read(ctx context.Context, params InterceptorParamsSDK) diag.Diagnostics {
	var diags diag.Diagnostics

	switch params.When {
	case After:
		if err := params.ResourceData.Set(names.AttrRegion, params.C.Region(ctx)); err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"Error Listing Remote Resources",
				"An unexpected error occurred. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					"Error: "+err.Error(),
			))
			return diags
		}
	}

	return diags
}
