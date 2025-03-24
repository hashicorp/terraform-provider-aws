// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/shopspring/decimal"
)

const (
	// A common separator to be used for creating resource Ids from a combination of attributes
	ResourceIdSeparator = ","
)

// ExpandStringList the result of flatmap.Expand for an array of strings
// and returns a []*string. Empty strings are skipped.
func ExpandStringList(configured []any) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		if v, ok := v.(string); ok && v != "" { // v != "" may not do anything since in []interface{}, empty string will be nil so !ok
			vs = append(vs, aws.String(v))
		}
	}
	return vs
}

func ExpandStringTimeValueList(configured []any, format string) []time.Time {
	return tfslices.ApplyToAll(ExpandStringValueList(configured), func(v string) time.Time {
		t, _ := time.Parse(format, v)
		return t
	})
}

// ExpandStringValueList takes the result of flatmap.Expand for an array of strings
// and returns a []string
func ExpandStringValueList(configured []any) []string {
	return ExpandStringyValueList[string](configured)
}

func ExpandStringyValueList[E ~string](configured []any) []E {
	vs := make([]E, 0, len(configured))
	for _, v := range configured {
		if val, ok := v.(string); ok && val != "" {
			vs = append(vs, E(val))
		}
	}
	return vs
}

// ExpandStringValueList takes the result of flatmap.Expand for an array of strings
// and returns a []string
func ExpandStringValueListEmpty(configured []any) []string {
	return ExpandStringyValueListEmpty[string](configured)
}

func ExpandStringyValueListEmpty[E ~string](configured []any) []E {
	vs := make([]E, 0, len(configured))
	for _, v := range configured {
		if val, ok := v.(string); ok { // empty string in config turns into nil in []interface{} so !ok
			vs = append(vs, E(val))
		} else {
			vs = append(vs, E(""))
		}
	}
	return vs
}

// Takes list of pointers to strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSetschema.NewSet
func FlattenStringList(list []*string) []any {
	vs := make([]any, 0, len(list))
	for _, v := range list {
		if v != nil {
			vs = append(vs, *v)
		}
	}
	return vs
}

func FlattenTimeStringValueList(list []time.Time, format string) []any {
	return tfslices.ApplyToAll(list, func(v time.Time) any {
		return v.Format(format)
	})
}

// Takes list of strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSetschema.NewSet
func FlattenStringValueList(list []string) []any {
	vs := make([]any, 0, len(list))
	for _, v := range list {
		vs = append(vs, v)
	}
	return vs
}

func FlattenStringyValueList[E ~string](configured []E) []any {
	vs := make([]any, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, string(v))
	}
	return vs
}

// Expands a map of string to interface to a map of string to int32
func ExpandInt32Map(m map[string]any) map[string]int32 {
	return tfmaps.ApplyToAllValues(m, func(v any) int32 {
		return int32(v.(int))
	})
}

// ExpandInt64ValueMap expands a map of string to interface to a map of string to int64
func ExpandInt64ValueMap(m map[string]any) map[string]int64 {
	return tfmaps.ApplyToAllValues(m, func(v any) int64 {
		return int64(v.(int))
	})
}

// ExpandFloat64ValueMap expands a map of string to interface to a map of string to float64
func ExpandFloat64ValueMap(m map[string]any) map[string]float64 {
	return tfmaps.ApplyToAllValues(m, func(v any) float64 {
		return v.(float64)
	})
}

// Expands a map of string to interface to a map of string to *string
func ExpandStringMap(m map[string]any) map[string]*string {
	return tfmaps.ApplyToAllValues(m, func(v any) *string {
		return aws.String(v.(string))
	})
}

func ExpandStringyValueMap[M ~map[K]V, K ~string, V ~string](m M) map[string]string {
	return tfmaps.ApplyToAllKeys(tfmaps.ApplyToAllValues(m,
		func(v V) string {
			return string(v)
		}),
		func(k K) string {
			return string(k)
		})
}

// ExpandStringValueMap expands a string map of interfaces to a string map of strings
func ExpandStringValueMap(m map[string]any) map[string]string {
	return tfmaps.ApplyToAllValues(m, func(v any) string {
		return v.(string)
	})
}

// Expands a map of string to interface to a map of string to *bool
func ExpandBoolValueMap(m map[string]any) map[string]bool {
	return tfmaps.ApplyToAllValues(m, func(v any) bool {
		return v.(bool)
	})
}

// Takes the result of schema.Set of strings and returns a []*string
func ExpandStringSet(configured *schema.Set) []*string {
	return ExpandStringList(configured.List()) // nosemgrep:ci.helper-schema-Set-extraneous-ExpandStringList-with-List
}

func ExpandStringValueSet(configured *schema.Set) []string {
	return ExpandStringValueList(configured.List()) // nosemgrep:ci.helper-schema-Set-extraneous-ExpandStringList-with-List
}

func ExpandStringValueEmptySet(configured *schema.Set) []string {
	return ExpandStringValueListEmpty(configured.List()) // nosemgrep:ci.helper-schema-Set-extraneous-ExpandStringList-with-List
}

func ExpandStringyValueSet[E ~string](configured *schema.Set) []E {
	return ExpandStringyValueList[E](configured.List())
}

func FlattenStringSet(list []*string) *schema.Set {
	return schema.NewSet(schema.HashString, FlattenStringList(list)) // nosemgrep:ci.helper-schema-Set-extraneous-NewSet-with-FlattenStringList
}

func FlattenStringValueSet(list []string) *schema.Set {
	return schema.NewSet(schema.HashString, FlattenStringValueList(list)) // nosemgrep: helper-schema-Set-extraneous-NewSet-with-FlattenStringList
}

func FlattenStringValueSetCaseInsensitive(list []string) *schema.Set {
	return schema.NewSet(sdkv2.StringCaseInsensitiveSetFunc, FlattenStringValueList(list)) // nosemgrep: helper-schema-Set-extraneous-NewSet-with-FlattenStringList
}

func FlattenStringyValueSet[E ~string](list []E) *schema.Set {
	return schema.NewSet(schema.HashString, FlattenStringyValueList[E](list))
}

func FlattenStringValueMap(m map[string]string) map[string]any {
	return tfmaps.ApplyToAllValues(m, func(v string) any {
		return v
	})
}

// Takes the result of schema.Set of strings and returns a []int64
func ExpandInt64ValueSet(configured *schema.Set) []int64 {
	return ExpandInt64ValueList(configured.List())
}

// Takes the result of flatmap.Expand for an array of int32
// and returns a []int32
func ExpandInt32ValueList(configured []any) []int32 {
	return tfslices.ApplyToAll(configured, func(v any) int32 {
		return int32(v.(int))
	})
}

// Takes the result of schema.Set of strings and returns a []int32
func ExpandInt32ValueSet(configured *schema.Set) []int32 {
	return ExpandInt32ValueList(configured.List())
}

func FlattenInt32ValueSet(set []int32) *schema.Set {
	return schema.NewSet(schema.HashInt, FlattenInt32ValueList(set))
}

// Takes the result of flatmap.Expand for an array of int64
// and returns a []int64
func ExpandInt64ValueList(configured []any) []int64 {
	return tfslices.ApplyToAll(configured, func(v any) int64 {
		return int64(v.(int))
	})
}

func ExpandFloat64ValueList(configured []any) []float64 {
	return tfslices.ApplyToAll(configured, func(v any) float64 {
		return v.(float64)
	})
}

func FlattenInt32ValueList(list []int32) []any {
	return tfslices.ApplyToAll(list, func(v int32) any {
		return int(v)
	})
}

// Takes a string of resource attributes separated by the ResourceIdSeparator constant, an expected number of Id Parts, and a boolean specifying if empty parts are to be allowed
// Returns a list of the resource attributes strings used to construct the unique Id or an error message if the resource id does not parse properly
func ExpandResourceId(id string, partCount int, allowEmptyPart bool) ([]string, error) {
	idParts := strings.Split(id, ResourceIdSeparator)

	if len(idParts) <= 1 {
		return nil, fmt.Errorf("unexpected format for ID (%v), expected more than one part", idParts)
	}

	if len(idParts) != partCount {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected (%d) parts separated by (%s)", id, partCount, ResourceIdSeparator)
	}

	if !allowEmptyPart {
		var emptyPart bool
		emptyParts := make([]int, 0, partCount)
		for index, part := range idParts {
			if part == "" {
				emptyPart = true
				emptyParts = append(emptyParts, index)
			}
		}

		if emptyPart {
			return nil, fmt.Errorf("unexpected format for ID (%[1]s), the following id parts indexes are blank (%v)", id, emptyParts)
		}
	}
	return idParts, nil
}

// Takes a list of the resource attributes as strings used to construct the unique Id, an expected number of Id Parts, and a boolean specifying if empty parts are to be allowed
// Returns a string of resource attributes separated by the ResourceIdSeparator constant or an error message if the id parts do not parse properly
func FlattenResourceId(idParts []string, partCount int, allowEmptyPart bool) (string, error) {
	if len(idParts) <= 1 {
		return "", fmt.Errorf("unexpected format for ID parts (%v), expected more than one part", idParts)
	}

	if len(idParts) != partCount {
		return "", fmt.Errorf("unexpected format for ID parts (%v), expected (%d) parts", idParts, partCount)
	}

	if !allowEmptyPart {
		var emptyPart bool
		emptyParts := make([]int, 0, len(idParts))
		for index, part := range idParts {
			if part == "" {
				emptyPart = true
				emptyParts = append(emptyParts, index)
			}
		}

		if emptyPart {
			return "", fmt.Errorf("unexpected format for ID parts (%v), the following id parts indexes are blank (%v)", idParts, emptyParts)
		}
	}

	return strings.Join(idParts, ResourceIdSeparator), nil
}

// BoolToStringValue converts a bool pointer to a Go string value.
func BoolToStringValue(v *bool) string {
	return strconv.FormatBool(aws.ToBool(v))
}

// BoolValueToString converts a Go bool value to a string pointer.
func BoolValueToString(v bool) *string {
	return aws.String(strconv.FormatBool(v))
}

// StringToBoolValue converts a string pointer to a Go bool value.
// Only the string "true" is converted to true, all other values return false.
func StringToBoolValue(v *string) bool {
	return aws.ToString(v) == strconv.FormatBool(true)
}

// Float32ValueToFloat64Value converts a float32 value to a Go float64 value.
func Float32ValueToFloat64Value(v float32) float64 {
	return decimal.NewFromFloat32(v).InexactFloat64()
}

// Float32ToFloat64Value converts a float32 pointer to a Go float64 value.
func Float32ToFloat64Value(v *float32) float64 {
	return Float32ValueToFloat64Value(aws.ToFloat32(v))
}

// Float64ValueToFloat32 converts a float64 value to a float32 pointer.
func Float64ValueToFloat32(v float64) *float32 {
	return aws.Float32(float32(v))
}

// Float64ToStringValue converts a float64 pointer to a Go string value.
func Float64ToStringValue(v *float64) string {
	return strconv.FormatFloat(aws.ToFloat64(v), 'f', -1, 64)
}

// IntValueToString converts a Go int value to a string pointer.
func IntValueToString(v int) *string {
	return aws.String(strconv.Itoa(v))
}

// Int32ToStringValue converts an int32 pointer to a Go string value.
func Int32ToStringValue(v *int32) string {
	return strconv.FormatInt(int64(aws.ToInt32(v)), 10)
}

// Int32ValueToStringValue converts an int32 value to a Go string value.
func Int32ValueToStringValue(v int32) string {
	return strconv.FormatInt(int64(v), 10)
}

// Int64ToStringValue converts an int64 pointer to a Go string value.
func Int64ToStringValue(v *int64) string {
	return strconv.FormatInt(aws.ToInt64(v), 10)
}

// Int64ValueToString converts a Go int64 value to a string pointer.
func Int64ValueToString(v int64) *string {
	return aws.String(strconv.FormatInt(v, 10))
}

// StringToIntValue converts a string pointer to a Go int value.
// Invalid integer strings are converted to 0.
func StringToIntValue(v *string) int {
	i, _ := strconv.Atoi(aws.ToString(v))
	return i
}

// StringToInt32Value converts a string pointer to a Go int32 value.
// Invalid integer strings are converted to 0.
func StringToInt32Value(v *string) int32 {
	return StringValueToInt32Value(aws.ToString(v))
}

// StringValueToBase64String converts a string to a Go base64 string pointer.
func StringValueToBase64String(v string) *string {
	return aws.String(itypes.Base64EncodeOnce([]byte(v)))
}

// StringValueToInt64 converts a string to a Go int32 pointer.
// Invalid integer strings are converted to 0.
func StringValueToInt32(v string) *int32 {
	return aws.Int32(StringValueToInt32Value(v))
}

// StringValueToInt32Value converts a string to a Go int32 value.
// Invalid integer strings are converted to 0.
func StringValueToInt32Value(v string) int32 {
	i, _ := strconv.ParseInt(v, 0, 32)
	return int32(i)
}

// StringValueToInt64 converts a string to a Go int64 pointer.
// Invalid integer strings are converted to 0.
func StringValueToInt64(v string) *int64 {
	return aws.Int64(StringValueToInt64Value(v))
}

// StringValueToInt64Value converts a string to a Go int64 value.
// Invalid integer strings are converted to 0.
func StringValueToInt64Value(v string) int64 {
	i, _ := strconv.ParseInt(v, 0, 64)
	return i
}

// Takes a string of resource attributes separated by the ResourceIdSeparator constant
// returns the number of parts
func ResourceIdPartCount(id string) int {
	idParts := strings.Split(id, ResourceIdSeparator)
	return len(idParts)
}

// DiffStringValueMaps returns the set of keys and values that must be created, the set of keys
// and values that must be destroyed, and the set of keys and values that are unchanged.
func DiffStringValueMaps(oldMap, newMap map[string]any) (map[string]string, map[string]string, map[string]string) {
	// First, we're creating everything we have.
	add := ExpandStringValueMap(newMap)

	// Build the maps of what to remove and what is unchanged.
	remove := make(map[string]string)
	unchanged := make(map[string]string)
	for k, v := range oldMap {
		v := v.(string)
		if old, ok := add[k]; !ok || old != v {
			// Delete it!
			remove[k] = v
		} else if ok {
			unchanged[k] = v
			// Already present, so remove from new.
			delete(add, k)
		}
	}

	return add, remove, unchanged
}

func DiffSlices[E any](old []E, new []E, eq func(E, E) bool) ([]E, []E, []E) {
	// First, we're creating everything we have.
	add := new

	// Build the slices of what to remove and what is unchanged.
	remove := make([]E, 0)
	unchanged := make([]E, 0)
	for _, e := range old {
		eq := func(v E) bool { return eq(v, e) }
		if !slices.ContainsFunc(new, eq) {
			// Delete it!
			remove = append(remove, e)
		} else {
			unchanged = append(unchanged, e)
			// Already present, so remove from new.
			add = slices.DeleteFunc(add, eq)
		}
	}

	return add, remove, unchanged
}

// DiffSlicesWithModify is a variant of DiffSlices which can account for
// cases when a partially equal item should be modified, rather than
// deleted and re-created
func DiffSlicesWithModify[E any](old []E, new []E, eq func(E, E) bool, modifyEq func(E, E) bool) ([]E, []E, []E, []E) {
	// First, we're creating everything we have.
	add := new

	// Build the slices of what to remove, modify, and what is unchanged.
	remove := make([]E, 0)
	modify := make([]E, 0)
	unchanged := make([]E, 0)
	for _, e := range old {
		eq := func(v E) bool { return eq(v, e) }
		modifyEq := func(v E) bool { return modifyEq(v, e) }

		if slices.ContainsFunc(new, eq) {
			// Unchanged, remove from add.
			unchanged = append(unchanged, e)
			add = slices.DeleteFunc(add, eq)
		} else {
			if i := slices.IndexFunc(new, modifyEq); i != -1 {
				// Modify, grab the indexed item from new, remove from add.
				modify = append(modify, new[i])
				add = slices.DeleteFunc(add, modifyEq)
			} else {
				// Delete it!
				remove = append(remove, e)
			}
		}
	}

	return add, remove, modify, unchanged
}
