// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package listresource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type identityInterceptorSDK struct {
	attributes []inttypes.IdentityAttribute
}

func IdentityInterceptorSDK(attributes []inttypes.IdentityAttribute) identityInterceptorSDK {
	return identityInterceptorSDK{
		attributes: attributes,
	}
}

func (r identityInterceptorSDK) Read(ctx context.Context, params InterceptorParamsSDK) diag.Diagnostics {
	var diags diag.Diagnostics

	switch params.When {
	case After:
		if err := r.readAfter(ctx, params); err != nil {
			diags.Append(diag.NewErrorDiagnostic(
				"Error Listing Remote Resources",
				"An unexpected error occurred setting resource identity. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					"Error: "+err.Error(),
			))
			return diags
		}
	}

	return diags
}

func (r identityInterceptorSDK) readAfter(ctx context.Context, params InterceptorParamsSDK) error {
	identity, err := params.ResourceData.Identity()
	if err != nil {
		return err
	}

	awsClient := params.C

	for _, attr := range r.attributes {
		switch attr.Name() {
		case names.AttrAccountID:
			if err := identity.Set(attr.Name(), awsClient.AccountID(ctx)); err != nil {
				return err
			}

		case names.AttrRegion:
			if err := identity.Set(attr.Name(), awsClient.Region(ctx)); err != nil {
				return err
			}

		default:
			val, ok := getAttributeOk(params.ResourceData, attr.ResourceAttributeName())
			if !ok {
				continue
			}
			if err := identity.Set(attr.Name(), val); err != nil {
				return err
			}
		}
	}

	return nil
}

type resourceData interface {
	Id() string
	GetOk(string) (any, bool)
}

func getAttributeOk(d resourceData, name string) (any, bool) {
	if name == "id" {
		return d.Id(), true
	}
	if v, ok := d.GetOk(name); !ok {
		return nil, false
	} else {
		return v, true
	}
}
