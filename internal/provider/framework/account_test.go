// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStripRestoreAccount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schema := rschema.Schema{Attributes: addAccountAttribute(map[string]rschema.Attribute{
		"name": rschema.StringAttribute{Required: true},
	}, resourceAccountAttribute())}
	typ := schema.Type().TerraformType(ctx).(tftypes.Object)
	want := tftypes.NewValue(tftypes.String, "222222222222")
	raw := tftypes.NewValue(typ, map[string]tftypes.Value{
		"name":      tftypes.NewValue(tftypes.String, "test"),
		attrAccount: want,
	})

	var diags diag.Diagnostics
	outer := tfsdk.State{Raw: raw, Schema: schema}
	state := outer
	account := stripAccount(ctx, &state, &diags)
	if diags.HasError() {
		t.Fatalf("stripAccount: %v", diags)
	}
	if !account.Equal(want) {
		t.Errorf("stripAccount = %s, want %s", account, want)
	}
	if _, ok := state.Raw.Type().(tftypes.Object).AttributeTypes[attrAccount]; ok {
		t.Error("stripAccount left the attribute in the data")
	}
	if _, ok := state.Schema.(rschema.Schema).Attributes[attrAccount]; ok {
		t.Error("stripAccount left the attribute in the schema")
	}
	if _, ok := schema.Attributes[attrAccount]; !ok {
		t.Error("stripAccount modified the original schema")
	}

	restoreAccount(ctx, &state, outer, account, &diags)
	if diags.HasError() {
		t.Fatalf("restoreAccount: %v", diags)
	}
	if !state.Raw.Equal(raw) {
		t.Errorf("restoreAccount(stripAccount(x)) = %s, want %s", state.Raw, raw)
	}

	// A null attribute is restored when the account is not known to the caller, e.g. on import.
	stripAccount(ctx, &state, &diags)
	restoreAccount(ctx, &state, outer, tftypes.Value{}, &diags)
	var m map[string]tftypes.Value
	if err := state.Raw.As(&m); err != nil {
		t.Fatal(err)
	}
	if !m[attrAccount].IsNull() {
		t.Errorf("restoreAccount without an account = %s, want null", m[attrAccount])
	}

	// A resource that is removed or not yet known stays so.
	for name, v := range map[string]tftypes.Value{"null": tftypes.NewValue(typ, nil), "unknown": tftypes.NewValue(typ, tftypes.UnknownValue)} {
		state := tfsdk.State{Raw: v, Schema: schema}
		stripAccount(ctx, &state, &diags)
		restoreAccount(ctx, &state, outer, account, &diags)
		if diags.HasError() || !state.Raw.Equal(v) {
			t.Errorf("%s: restoreAccount(stripAccount(x)) = %s, %v, want %s", name, state.Raw, diags, v)
		}
	}
}

// A schema's own account attribute, which is required or computed, is not the injected one.
func TestStripAccount_ownAttribute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	schema := dsschema.Schema{Attributes: addAccountAttribute(map[string]dsschema.Attribute{
		attrAccount: dsschema.StringAttribute{Computed: true},
	}, dataSourceAccountAttribute())}
	raw := tftypes.NewValue(schema.Type().TerraformType(ctx), map[string]tftypes.Value{
		attrAccount: tftypes.NewValue(tftypes.String, "222222222222"),
	})

	var diags diag.Diagnostics
	outer := tfsdk.State{Raw: raw, Schema: schema}
	state := outer
	if account := stripAccount(ctx, &state, &diags); account.Type() != nil {
		t.Errorf("stripAccount = %s, want a zero value", account)
	}
	restoreAccount(ctx, &state, outer, tftypes.Value{}, &diags)
	if diags.HasError() || !state.Raw.Equal(raw) {
		t.Errorf("data = %s, %v, want it unchanged", state.Raw, diags)
	}
}

func TestResourceAccountID(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value tftypes.Value
		want  string
	}{
		"zero":    {},
		"null":    {value: tftypes.NewValue(tftypes.String, nil)},
		"unknown": {value: tftypes.NewValue(tftypes.String, tftypes.UnknownValue)},
		"known":   {value: tftypes.NewValue(tftypes.String, "222222222222"), want: "222222222222"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := accountID(tc.value); got != tc.want {
				t.Errorf("accountID = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRawStateAccount(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		raw  *tfprotov6.RawState
		want string
	}{
		"no state":   {},
		"no account": {raw: &tfprotov6.RawState{JSON: []byte(`{"old_name":"test"}`)}},
		"null":       {raw: &tfprotov6.RawState{JSON: []byte(`{"account":null}`)}},
		"account":    {raw: &tfprotov6.RawState{JSON: []byte(`{"old_name":"test","account":"222222222222"}`)}, want: "222222222222"},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := accountID(rawStateAccount(tc.raw)); got != tc.want {
				t.Errorf("rawStateAccount = %q, want %q", got, tc.want)
			}
		})
	}
}
