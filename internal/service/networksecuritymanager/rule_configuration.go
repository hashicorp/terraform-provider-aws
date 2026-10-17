// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/document"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
)

var (
	_ basetypes.StringTypable                    = (*ruleConfigurationType)(nil)
	_ basetypes.StringValuable                   = (*ruleConfiguration)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*ruleConfiguration)(nil)
	_ xattr.ValidateableAttribute                = (*ruleConfiguration)(nil)
	_ fwtypes.SmithyDocumentValue                = (*ruleConfiguration)(nil)
)

// ruleConfigurationType is the type of a rule's configuration: a free-form JSON
// document passed to the API as a Smithy document.
//
// It is a normalized JSON string type whose semantic equality also treats a JSON
// number and its decimal string representation as equal, because the service
// stores and returns every numeric value in a configuration as a string
// (for example "ImmunityTime": 300 is returned as "ImmunityTime": "300") while
// rejecting strings on input.
type ruleConfigurationType struct {
	jsontypes.NormalizedType
}

func (t ruleConfigurationType) Equal(o attr.Type) bool {
	other, ok := o.(ruleConfigurationType)
	if !ok {
		return false
	}

	return t.NormalizedType.Equal(other.NormalizedType)
}

func (ruleConfigurationType) String() string {
	return "RuleConfigurationType"
}

func (t ruleConfigurationType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return ruleConfigurationNull(), diags
	}
	if in.IsUnknown() {
		return ruleConfigurationUnknown(), diags
	}

	return ruleConfigurationValue(in.ValueString()), diags
}

func (t ruleConfigurationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (ruleConfigurationType) ValueType(context.Context) attr.Value {
	return ruleConfiguration{}
}

type ruleConfiguration struct {
	jsontypes.Normalized
}

func (v ruleConfiguration) Equal(o attr.Value) bool {
	other, ok := o.(ruleConfiguration)
	if !ok {
		return false
	}

	return v.Normalized.Equal(other.Normalized)
}

func (ruleConfiguration) Type(context.Context) attr.Type {
	return ruleConfigurationType{}
}

func (v ruleConfiguration) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(ruleConfiguration)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	oldDocument, err := canonicalRuleConfiguration(v.ValueString())
	if err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
		return false, diags
	}
	newDocument, err := canonicalRuleConfiguration(newValue.ValueString())
	if err != nil {
		diags.AddError("Semantic Equality Check Error", err.Error())
		return false, diags
	}

	return reflect.DeepEqual(oldDocument, newDocument), diags
}

// ToSmithyObjectDocument returns the configuration as the Smithy document sent to the API.
func (v ruleConfiguration) ToSmithyObjectDocument(context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v.IsNull() || v.IsUnknown() {
		return nil, diags
	}

	doc, err := tfsmithy.DocumentFromJSONString(v.ValueString(), document.NewLazyDocument)
	if err != nil {
		diags.AddError(
			"JSON Unmarshal Error",
			"An unexpected error occurred while unmarshalling a JSON string. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return nil, diags
	}

	return doc, diags
}

func ruleConfigurationNull() ruleConfiguration {
	return ruleConfiguration{Normalized: jsontypes.NewNormalizedNull()}
}

func ruleConfigurationUnknown() ruleConfiguration {
	return ruleConfiguration{Normalized: jsontypes.NewNormalizedUnknown()}
}

func ruleConfigurationValue(value string) ruleConfiguration {
	return ruleConfiguration{Normalized: jsontypes.NewNormalizedValue(value)}
}

// canonicalRuleConfiguration decodes a JSON document with every number
// represented as its decimal string, so that documents differing only in
// whitespace, key order, or number-versus-string encoding compare equal.
func canonicalRuleConfiguration(s string) (any, error) {
	decoder := json.NewDecoder(strings.NewReader(s))
	decoder.UseNumber()

	var v any
	if err := decoder.Decode(&v); err != nil {
		return nil, err
	}

	return stringifyJSONNumbers(v), nil
}

func stringifyJSONNumbers(v any) any {
	switch t := v.(type) {
	case json.Number:
		return canonicalJSONNumber(t)
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, e := range t {
			m[k] = stringifyJSONNumbers(e)
		}
		return m
	case []any:
		l := make([]any, len(t))
		for i, e := range t {
			l[i] = stringifyJSONNumbers(e)
		}
		return l
	default:
		return v
	}
}

// canonicalJSONNumber renders a JSON number the way the service returns it: an
// integral value as its shortest decimal form, so that 300, 300.0 and 3e2 all
// match the returned "300". A number that is not integral, or an integer too
// large to be held exactly by a float64, keeps its literal spelling: the
// service echoes those back unchanged, and rewriting them would lose digits.
func canonicalJSONNumber(n json.Number) string {
	if i, err := n.Int64(); err == nil {
		return strconv.FormatInt(i, 10)
	}
	if f, err := n.Float64(); err == nil && f == math.Trunc(f) && math.Abs(f) < 1<<53 {
		return strconv.FormatInt(int64(f), 10)
	}

	return n.String()
}
