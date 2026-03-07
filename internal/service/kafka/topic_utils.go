// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func TopicConfigsDiffSuppress() planmodifier.String {
	return &topicConfigsDiffSuppress{}
}

type topicConfigsDiffSuppress struct {
}

// Description returns a human-readable description of the plan modifier.
func (m topicConfigsDiffSuppress) Description(_ context.Context) string {
	return "Verifies if computed attributes are the only difference in the topic_configs field."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m topicConfigsDiffSuppress) MarkdownDescription(_ context.Context) string {
	return "Verifies if computed attributes are the only difference in the topic_configs field."
}

func (m *topicConfigsDiffSuppress) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var old Normalized
	diags := req.State.GetAttribute(ctx, path.Root("configs"), &old)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var new Normalized
	diags = req.Plan.GetAttribute(ctx, path.Root("configs"), &new)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	oldMap, err := structure.ExpandJsonFromString(old.ValueString())
	if err != nil {
		return
	}
	newMap, err := structure.ExpandJsonFromString(new.ValueString())
	if err != nil {
		return
	}

	oldMap = removeComputedKeys(oldMap, newMap)

	if reflect.DeepEqual(oldMap, newMap) {
		resp.PlanValue = req.StateValue
	}
}

// removeComputedKeys removes keys from the old map that are not present in the new map.
// This is used to filter out computed attributes that are not defined in the new configuration.
func removeComputedKeys(old, new map[string]any) map[string]any {
	for k := range old {
		if _, exists := new[k]; !exists {
			delete(old, k)
		}
	}
	return old
}

// ----------------------------------------------------------------------------
// A modified implementation of the jsontypes custom type to apply the computed key removal for semantic equality
// ----------------------------------------------------------------------------

var (
	_ basetypes.StringValuable                   = (*Normalized)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*Normalized)(nil)
	_ xattr.ValidateableAttribute                = (*Normalized)(nil)
	_ function.ValidateableParameter             = (*Normalized)(nil)
)

// Normalized represents a valid JSON string (RFC 7159). Semantic equality logic is defined for Normalized
// such that inconsequential differences between JSON strings are ignored (whitespace, property order, etc). If you
// need strict, byte-for-byte, string equality, consider using ExactType.
type Normalized struct {
	basetypes.StringValue
}

// Type returns a NormalizedType.
func (v Normalized) Type(_ context.Context) attr.Type {
	return NormalizedType{}
}

// Equal returns true if the given value is equivalent.
func (v Normalized) Equal(o attr.Value) bool {
	other, ok := o.(Normalized)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals returns true if the given JSON string value is semantically equal to the current JSON string value. When compared,
// these JSON string values are "normalized" by marshalling them to empty Go structs. This prevents Terraform data consistency errors and
// resource drift due to inconsequential differences in the JSON strings (whitespace, property order, etc).
func (v Normalized) StringSemanticEquals(ctx context.Context, stateValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	stateValue, ok := stateValuable.(Normalized)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", stateValuable),
		)

		return false, diags
	}

	//convert the jsons to maps for recursive computed key removal
	planMap, err := structure.ExpandJsonFromString(v.ValueString())
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
	}
	stateMap, err := structure.ExpandJsonFromString(stateValue.ValueString())
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
	}

	stateMap = removeComputedKeys(stateMap, planMap)
	stateMinusComputedJson, err := structure.FlattenJsonToString(stateMap)
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
	}

	result, err := jsonEqual(v.ValueString(), stateMinusComputedJson)
	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return false, diags
	}

	return result, diags
}

// ----------------------------------------------------------------------------
// unmodified below this point -- https://github.com/hashicorp/terraform-plugin-framework-jsontypes
// ----------------------------------------------------------------------------

func jsonEqual(s1, s2 string) (bool, error) {
	s1, err := normalizeJSONString(s1)
	if err != nil {
		return false, err
	}

	s2, err = normalizeJSONString(s2)
	if err != nil {
		return false, err
	}

	return s1 == s2, nil
}

func normalizeJSONString(jsonStr string) (string, error) {
	dec := json.NewDecoder(strings.NewReader(jsonStr))

	// This ensures the JSON decoder will not parse JSON numbers into Go's float64 type; avoiding Go
	// normalizing the JSON number representation or imposing limits on numeric range. See the unit test cases
	// of StringSemanticEquals for examples.
	dec.UseNumber()

	var temp any
	if err := dec.Decode(&temp); err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(&temp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// ValidateAttribute implements attribute value validation. This type requires the value provided to be a String
// value that is valid JSON format (RFC 7159).
func (v Normalized) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsUnknown() || v.IsNull() {
		return
	}

	if ok := json.Valid([]byte(v.ValueString())); !ok {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid JSON String Value",
			"A string value was provided that is not valid JSON string format (RFC 7159).\n\n"+
				"Given Value: "+v.ValueString()+"\n",
		)

		return
	}
}

// ValidateParameter implements provider-defined function parameter value validation. This type requires the value
// provided to be a String value that is valid JSON format (RFC 7159).
func (v Normalized) ValidateParameter(ctx context.Context, req function.ValidateParameterRequest, resp *function.ValidateParameterResponse) {
	if v.IsUnknown() || v.IsNull() {
		return
	}

	if ok := json.Valid([]byte(v.ValueString())); !ok {
		resp.Error = function.NewArgumentFuncError(
			req.Position,
			"Invalid JSON String Value: "+
				"A string value was provided that is not valid JSON string format (RFC 7159).\n\n"+
				"Given Value: "+v.ValueString()+"\n",
		)

		return
	}
}

// Unmarshal calls (encoding/json).Unmarshal with the Normalized StringValue and `target` input. A null or unknown value will produce an error diagnostic.
// See encoding/json docs for more on usage: https://pkg.go.dev/encoding/json#Unmarshal
func (v Normalized) Unmarshal(target any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("Normalized JSON Unmarshal Error", "json string value is null"))
		return diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("Normalized JSON Unmarshal Error", "json string value is unknown"))
		return diags
	}

	err := json.Unmarshal([]byte(v.ValueString()), target)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("Normalized JSON Unmarshal Error", err.Error()))
	}

	return diags
}

// NewNormalizedNull creates a Normalized with a null value. Determine whether the value is null via IsNull method.
func NewNormalizedNull() Normalized {
	return Normalized{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewNormalizedUnknown creates a Normalized with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewNormalizedUnknown() Normalized {
	return Normalized{
		StringValue: basetypes.NewStringUnknown(),
	}
}

// NewNormalizedValue creates a Normalized with a known value. Access the value via ValueString method.
func NewNormalizedValue(value string) Normalized {
	return Normalized{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewNormalizedPointerValue creates a Normalized with a null value if nil or a known value. Access the value via ValueStringPointer method.
func NewNormalizedPointerValue(value *string) Normalized {
	return Normalized{
		StringValue: basetypes.NewStringPointerValue(value),
	}
}

// NormalizedType is an attribute type that represents a valid JSON string (RFC 7159). Semantic equality logic is defined for NormalizedType
// such that inconsequential differences between JSON strings are ignored (whitespace, property order, etc). If you need strict, byte-for-byte,
// string equality, consider using ExactType.
type NormalizedType struct {
	basetypes.StringType
}

// String returns a human readable string of the type name.
func (t NormalizedType) String() string {
	return "NormalizedType"
}

// ValueType returns the Value type.
func (t NormalizedType) ValueType(ctx context.Context) attr.Value {
	return Normalized{}
}

// Equal returns true if the given type is equivalent.
func (t NormalizedType) Equal(o attr.Type) bool {
	other, ok := o.(NormalizedType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t NormalizedType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return Normalized{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t NormalizedType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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
