// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const attrAccount = "account"

// accountSchema returns the top-level account argument. One instance is shared by every resource and data source.
// Like region, it is recorded in state for every resource, so that a resource stays in its account
// when the provider's own account changes. A change of account replaces the resource.
func accountSchema(ctx context.Context, provider *schema.Provider) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ForceNew:     true,
		Description:  "AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials.",
		ValidateFunc: verify.ValidAccountID,
		// A state from before the argument existed has no account: the resource is in the provider's own.
		DiffSuppressFunc: func(_, o, n string, d *schema.ResourceData) bool {
			c, ok := provider.Meta().(*conns.AWSClient)
			return ok && d.Id() != "" && o == "" && n == c.AccountID(ctx)
		},
	}
}

// injectAccountSchema adds the account argument to a resource or data source and reports whether it did.
// The one that already has an argument of that name, with another meaning, is left as it is.
func injectAccountSchema(typeName string, r *schema.Resource, account *schema.Schema) bool {
	if typeName == "aws_redshift_endpoint_authorization" {
		return false
	}

	if f := r.SchemaFunc; f != nil {
		r.SchemaFunc = func() map[string]*schema.Schema {
			s := f()
			s[attrAccount] = account
			return s
		}
	} else {
		r.Schema[attrAccount] = account
	}

	return true
}

// bootstrapAccount puts any per-resource account into ctx and validates it,
// so that an account the provider cannot manage fails the operation before any API client is created.
func bootstrapAccount(ctx context.Context, getAttribute getAttributeFunc, meta any) (context.Context, error) {
	if getAttribute == nil {
		return ctx, nil // The resource has no account argument, or the operation has no data.
	}
	v, _ := getAttribute(attrAccount)
	accountID, _ := v.(string)
	if accountID == "" {
		return ctx, nil
	}

	ctx = conns.WithOverrideAccount(ctx, accountID)
	if c, ok := meta.(*conns.AWSClient); ok {
		if err := c.ValidateInContextAccount(ctx); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

// defaultAccount plans a new resource without an account in the provider's own account.
// An existing resource in another account must name it: without an account it would move if it were replaced.
func defaultAccount() customizeDiffInterceptor {
	return interceptorFunc1[*schema.ResourceDiff, error](func(ctx context.Context, opts customizeDiffInterceptorOptions) error {
		switch d, when, why := opts.d, opts.when, opts.why; when {
		case Before:
			switch why {
			case CustomizeDiff:
				if !d.GetRawConfig().GetAttr(attrAccount).IsNull() {
					return nil
				}

				providerAccountID := opts.c.AccountID(conns.WithOverrideAccount(ctx, ""))
				if d.Id() == "" {
					return d.SetNew(attrAccount, providerAccountID)
				}
				if v := d.Get(attrAccount).(string); v != "" && v != providerAccountID {
					return fmt.Errorf("the resource is in AWS account %[1]s, which is not the provider's account: set account = %[1]q to keep it there", v)
				}
			}
		}

		return nil
	})
}

// setAccountInState records the account a resource or data source was read from.
func setAccountInState() crudInterceptor {
	return interceptorFunc1[schemaResourceData, diag.Diagnostics](func(ctx context.Context, opts crudInterceptorOptions) diag.Diagnostics {
		var diags diag.Diagnostics

		switch d, when, why := opts.d, opts.when, opts.why; when {
		case After:
			switch why {
			case Read:
				// The resource no longer exists.
				if d.Id() == "" {
					return diags
				}

				if err := d.Set(attrAccount, opts.c.AccountID(ctx)); err != nil {
					return sdkdiag.AppendErrorf(diags, "setting %s: %s", attrAccount, err)
				}
			}
		}

		return diags
	})
}

// assumeRoleInAccountSchema returns the provider-level assume_role_in_account block:
// the session arguments of assume_role, with a role name in place of a role ARN and the accounts the block applies to.
func assumeRoleInAccountSchema() *schema.Schema {
	s := assumeRoleSchema()

	elem := s.Elem.(*schema.Resource).Schema
	delete(elem, "role_arn")
	elem["account_ids"] = &schema.Schema{
		Type:        schema.TypeSet,
		Optional:    true,
		Description: "AWS account IDs the block applies to. The block without `account_ids` applies to every other account.",
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: verify.ValidAccountID,
		},
	}
	elem["role_name"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		Description:  "Name, with any path, of the IAM role to assume in the resource's account.",
		ValidateFunc: validation.All(validation.StringIsNotEmpty, validation.StringDoesNotContainAny(":")),
	}

	return s
}

func expandAssumeRoleInAccount(ctx context.Context, tfList []any) ([]conns.AccountRole, diag.Diagnostics) {
	var diags diag.Diagnostics
	apiObjects := make([]conns.AccountRole, 0, len(tfList))

	for i, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		// The role ARN is built from the resource's account and the role name.
		tfMap = maps.Clone(tfMap)
		tfMap["role_arn"] = "-"
		assumeRole, d := expandAssumeRole(ctx, cty.GetAttrPath("assume_role_in_account").IndexInt(i), tfMap)
		diags = append(diags, d...)
		assumeRole.RoleARN = ""
		apiObject := conns.AccountRole{AssumeRole: assumeRole}
		apiObject.RoleName, _ = tfMap["role_name"].(string)
		if v, ok := tfMap["account_ids"].(*schema.Set); ok {
			apiObject.AccountIDs = flex.ExpandStringValueSet(v)
		}
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, diags
}
