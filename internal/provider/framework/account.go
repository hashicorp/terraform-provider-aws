// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"encoding/json"
	"iter"
	"maps"

	aschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	eschema "github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	pschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

const (
	attrAccount        = "account"
	accountDescription = "AWS account ID where this resource will be managed. Defaults to the account of the provider's credentials."
)

func accountValidators() []validator.String {
	return []validator.String{fwvalidators.AWSAccountID()}
}

// The account attribute, in the types of each kind of schema.
// Like region, it is recorded in state for every resource, so that a resource stays in its account
// when the provider's own account changes. Actions and list resources have no state.

func resourceAccountAttribute() rschema.Attribute {
	return rschema.StringAttribute{Optional: true, Computed: true, Description: accountDescription, Validators: accountValidators()}
}

func dataSourceAccountAttribute() dsschema.Attribute {
	return dsschema.StringAttribute{Optional: true, Computed: true, Description: accountDescription, Validators: accountValidators()}
}

func ephemeralResourceAccountAttribute() eschema.Attribute {
	return eschema.StringAttribute{Optional: true, Computed: true, Description: accountDescription, Validators: accountValidators()}
}

func actionAccountAttribute() aschema.Attribute {
	return aschema.StringAttribute{Optional: true, Description: accountDescription, Validators: accountValidators()}
}

func listResourceAccountAttribute() listschema.Attribute {
	return listschema.StringAttribute{Optional: true, Description: accountDescription, Validators: accountValidators()}
}

// assumeRoleInAccountBlock returns the provider-level assume_role_in_account block:
// the session arguments of assume_role, with a role name in place of a role ARN and the accounts the block applies to.
func assumeRoleInAccountBlock(assumeRole pschema.ListNestedBlock) pschema.Block {
	attributes := maps.Clone(assumeRole.NestedObject.Attributes)
	delete(attributes, "role_arn")
	attributes["account_ids"] = pschema.SetAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Description: "AWS account IDs the block applies to. The block without `account_ids` applies to every other account.",
	}
	attributes["role_name"] = pschema.StringAttribute{
		Required:    true,
		Description: "Name, with any path, of the IAM role to assume in the resource's account.",
	}

	return pschema.ListNestedBlock{
		NestedObject: pschema.NestedBlockObject{Attributes: attributes},
	}
}

// schemaData is any of the Terraform Plugin Framework's (raw value, schema) pairs.
//
// Resource implementations decode whole configurations, plans and states into model structs,
// which fails on an attribute the model does not declare.
// The wrappers therefore hand the inner resource data without the injected account attribute,
// typed for the resource's own schema, and put the attribute back on the way out.
type schemaData interface {
	tfsdk.Config | tfsdk.Plan | tfsdk.State | tfsdk.EphemeralResultData | tfsdk.Resource
}

// addAccountAttribute adds the attribute to a schema's attributes.
// The few schemas that already have an argument of this name are left as they are.
func addAccountAttribute[A any](attributes map[string]A, attribute A) map[string]A {
	if _, ok := attributes[attrAccount]; ok {
		return attributes
	}
	if attributes == nil {
		attributes = make(map[string]A, 1)
	}
	attributes[attrAccount] = attribute
	return attributes
}

func withoutAccount[V any](m map[string]V) map[string]V {
	m = maps.Clone(m)
	delete(m, attrAccount)
	return m
}

// withoutInjectedAccount returns a schema's attributes without the injected one, and whether there was one.
// A schema's own attribute of that name is not optional but required or computed, and stays.
func withoutInjectedAccount[A interface{ IsOptional() bool }](attributes map[string]A) (map[string]A, bool) {
	if a, ok := attributes[attrAccount]; !ok || !a.IsOptional() {
		return attributes, false
	}
	return withoutAccount(attributes), true
}

// stripAccount removes the injected account attribute from data and returns its value.
// Data whose schema has no injected attribute is left unchanged.
func stripAccount[T schemaData](ctx context.Context, data *T, diags *diag.Diagnostics) tftypes.Value {
	var account tftypes.Value
	v := tfsdk.State(*data)

	var ok bool
	switch s := v.Schema.(type) {
	case rschema.Schema:
		s.Attributes, ok = withoutInjectedAccount(s.Attributes)
		v.Schema = s
	case dsschema.Schema:
		s.Attributes, ok = withoutInjectedAccount(s.Attributes)
		v.Schema = s
	case eschema.Schema:
		s.Attributes, ok = withoutInjectedAccount(s.Attributes)
		v.Schema = s
	case aschema.Schema:
		s.Attributes, ok = withoutInjectedAccount(s.Attributes)
		v.Schema = s
	case listschema.Schema:
		s.Attributes, ok = withoutInjectedAccount(s.Attributes)
		v.Schema = s
	}
	typ, isObject := v.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok || !isObject {
		return account
	}

	switch {
	case v.Raw.IsNull():
		v.Raw = tftypes.NewValue(typ, nil)
	case !v.Raw.IsKnown():
		v.Raw = tftypes.NewValue(typ, tftypes.UnknownValue)
	default:
		var m map[string]tftypes.Value
		if err := v.Raw.As(&m); err != nil {
			diags.AddError("Removing account", err.Error())
			return account
		}
		account = m[attrAccount]
		v.Raw = tftypes.NewValue(typ, withoutAccount(m))
	}
	*data = T(v)

	return account
}

// restoreAccount is the inverse of stripAccount. outer supplies the schema with the injected attribute.
// A zero account restores a null attribute.
func restoreAccount[T schemaData](ctx context.Context, data *T, outer T, account tftypes.Value, diags *diag.Diagnostics) {
	v := tfsdk.State(outer)
	inner := tfsdk.State(*data)

	typ, ok := v.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		return
	}
	if innerType, ok := inner.Schema.Type().TerraformType(ctx).(tftypes.Object); !ok || len(innerType.AttributeTypes) == len(typ.AttributeTypes) {
		return // Nothing was removed.
	}

	switch raw := inner.Raw; {
	case raw.IsNull():
		v.Raw = tftypes.NewValue(typ, nil)
	case !raw.IsKnown():
		v.Raw = tftypes.NewValue(typ, tftypes.UnknownValue)
	default:
		var m map[string]tftypes.Value
		if err := raw.As(&m); err != nil {
			diags.AddError("Restoring account", err.Error())
			return
		}
		if account.Type() == nil {
			account = tftypes.NewValue(tftypes.String, nil)
		}
		m = maps.Clone(m)
		m[attrAccount] = account
		v.Raw = tftypes.NewValue(typ, m)
	}
	*data = T(v)
}

// accountID returns the AWS account ID in an account attribute value: "" if it is not set or not yet known.
func accountID(account tftypes.Value) string {
	var v string
	if account.Type() == nil || !account.IsKnown() || account.IsNull() {
		return ""
	}
	if err := account.As(&v); err != nil {
		return "" // Not reached: the attribute is a string.
	}
	return v
}

// withAccount puts the account in an account attribute value into ctx and validates it,
// so that an account the provider cannot manage fails the operation before any API client is created.
// It returns the value to record for the resource: the account it is managed in.
func withAccount(ctx context.Context, account tftypes.Value, c *conns.AWSClient, diags *diag.Diagnostics) (context.Context, tftypes.Value) {
	ctx = conns.WithOverrideAccount(ctx, accountID(account))
	if c == nil || (account.Type() != nil && !account.IsKnown()) {
		return ctx, account
	}
	if err := c.ValidateInContextAccount(ctx); err != nil {
		diags.AddAttributeError(path.Root(attrAccount), "Invalid account Value", err.Error())
	}
	if v := c.AccountID(ctx); v != "" {
		account = tftypes.NewValue(tftypes.String, v)
	}

	return ctx, account
}

// requiresReplaceIfAccountChanges forces resource replacement if the resource's account changes.
// A state from before the argument existed has no account: the resource is in the provider's own.
// An account that is not yet known may be any account.
func requiresReplaceIfAccountChanges(ctx context.Context, c *conns.AWSClient, plan, state tftypes.Value, response *resource.ModifyPlanResponse) {
	effective := func(account tftypes.Value) string {
		return c.AccountID(conns.WithOverrideAccount(ctx, accountID(account)))
	}
	if (plan.Type() != nil && !plan.IsKnown()) || effective(plan) != effective(state) {
		response.RequiresReplace = append(response.RequiresReplace, path.Root(attrAccount))
	}
}

// rawStateAccount returns the account stored in a state that is being upgraded or moved,
// so that the resource is still managed in its account. It returns a zero value if there is none.
func rawStateAccount(raw *tfprotov6.RawState) tftypes.Value {
	var account tftypes.Value
	if raw == nil {
		return account
	}

	var m map[string]json.RawMessage
	var v *string
	if err := json.Unmarshal(raw.JSON, &m); err == nil && json.Unmarshal(m[attrAccount], &v) == nil && v != nil {
		account = tftypes.NewValue(tftypes.String, *v)
	}

	return account
}

// configValidatorWithoutAccount hands a resource's configuration validator the configuration without the injected attribute.
type configValidatorWithoutAccount struct {
	resource.ConfigValidator
}

func (v configValidatorWithoutAccount) ValidateResource(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	stripAccount(ctx, &request.Config, &response.Diagnostics)
	v.ConfigValidator.ValidateResource(ctx, request, response)
}

// restoreListResults puts the account that was listed back into every listed resource.
func restoreListResults(ctx context.Context, results iter.Seq[list.ListResult], outer tfsdk.Resource, account tftypes.Value) iter.Seq[list.ListResult] {
	return func(yield func(list.ListResult) bool) {
		for result := range results {
			if result.Resource != nil {
				restoreAccount(ctx, result.Resource, outer, account, &result.Diagnostics)
			}
			if !yield(result) {
				return
			}
		}
	}
}
