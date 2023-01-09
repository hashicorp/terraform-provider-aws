package flex

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// A common separator to be used for creating resource Ids from a combination of attributes
	ResourceIdSeparator = ","
)

// Takes the result of flatmap.Expand for an array of strings
// and returns a []*string
func ExpandStringList(configured []interface{}) []*string {
	vs := make([]*string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, aws.String(v.(string)))
		}
	}
	return vs
}

// ExpandStringValueList takes the result of flatmap.Expand for an array of strings
// and returns a []string
func ExpandStringValueList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}

// Takes list of pointers to strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSetschema.NewSet
func FlattenStringList(list []*string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, *v)
	}
	return vs
}

// Takes list of pointers to strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSetschema.NewSet
func FlattenStringValueList(list []string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, v)
	}
	return vs
}

// Expands a map of string to interface to a map of string to *int32
func ExpandInt32Map(m map[string]interface{}) map[string]int32 {
	intMap := make(map[string]int32, len(m))
	for k, v := range m {
		intMap[k] = int32(v.(int))
	}
	return intMap
}

// Expands a map of string to interface to a map of string to *string
func ExpandStringMap(m map[string]interface{}) map[string]*string {
	stringMap := make(map[string]*string, len(m))
	for k, v := range m {
		stringMap[k] = aws.String(v.(string))
	}
	return stringMap
}

// ExpandStringValueMap expands a string map of interfaces to a string map of strings
func ExpandStringValueMap(m map[string]interface{}) map[string]string {
	stringMap := make(map[string]string, len(m))
	for k, v := range m {
		stringMap[k] = v.(string)
	}
	return stringMap
}

// Expands a map of string to interface to a map of string to *bool
func ExpandBoolMap(m map[string]interface{}) map[string]*bool {
	boolMap := make(map[string]*bool, len(m))
	for k, v := range m {
		boolMap[k] = aws.Bool(v.(bool))
	}
	return boolMap
}

// Takes the result of schema.Set of strings and returns a []*string
func ExpandStringSet(configured *schema.Set) []*string {
	return ExpandStringList(configured.List()) // nosemgrep:ci.helper-schema-Set-extraneous-ExpandStringList-with-List
}

func ExpandStringValueSet(configured *schema.Set) []string {
	return ExpandStringValueList(configured.List()) // nosemgrep:ci.helper-schema-Set-extraneous-ExpandStringList-with-List
}

func FlattenStringSet(list []*string) *schema.Set {
	return schema.NewSet(schema.HashString, FlattenStringList(list)) // nosemgrep:ci.helper-schema-Set-extraneous-NewSet-with-FlattenStringList
}

func FlattenStringValueSet(list []string) *schema.Set {
	return schema.NewSet(schema.HashString, FlattenStringValueList(list)) // nosemgrep: helper-schema-Set-extraneous-NewSet-with-FlattenStringList
}

// Takes the result of schema.Set of strings and returns a []*int64
func ExpandInt64Set(configured *schema.Set) []*int64 {
	return ExpandInt64List(configured.List())
}

func FlattenInt64Set(list []*int64) *schema.Set {
	return schema.NewSet(schema.HashInt, FlattenInt64List(list))
}

// Takes the result of flatmap.Expand for an array of int64
// and returns a []*int64
func ExpandInt64List(configured []interface{}) []*int64 {
	vs := make([]*int64, 0, len(configured))
	for _, v := range configured {
		vs = append(vs, aws.Int64(int64(v.(int))))
	}
	return vs
}

// Takes list of pointers to int64s. Expand to an array
// of raw ints and returns a []interface{}
// to keep compatibility w/ schema.NewSet
func FlattenInt64List(list []*int64) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, int(aws.Int64Value(v)))
	}
	return vs
}

func PointersMapToStringList(pointers map[string]*string) map[string]interface{} {
	list := make(map[string]interface{}, len(pointers))
	for i, v := range pointers {
		list[i] = *v
	}
	return list
}

// Takes a string of resource attributes separated by the ResourceIdSeparator constant and an expected number of Id Parts
// Returns a list of the resource attributes strings used to construct the unique Id or an error message if the resource id does not parse properly
func ExpandResourceId(id string, partCount int) ([]string, error) {
	idParts := strings.Split(id, ResourceIdSeparator)

	if len(idParts) <= 1 {
		return nil, fmt.Errorf("unexpected format for ID (%v), expected more than one part", idParts)
	}

	if len(idParts) != partCount {
		return nil, fmt.Errorf("unexpected format for ID (%s), expected (%d) parts separated by (%s)", id, partCount, ResourceIdSeparator)
	}

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

	return idParts, nil
}

// Takes a list of the resource attributes as strings used to construct the unique Id and an expected number of Id Parts
// Returns a string of resource attributes separated by the ResourceIdSeparator constant or an error message if the id parts do not parse properly
func FlattenResourceId(idParts []string, partCount int) (string, error) {
	if len(idParts) <= 1 {
		return "", fmt.Errorf("unexpected format for ID parts (%v), expected more than one part", idParts)
	}

	if len(idParts) != partCount {
		return "", fmt.Errorf("unexpected format for ID parts (%v), expected (%d) parts", idParts, partCount)
	}

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

	return strings.Join(idParts, ResourceIdSeparator), nil
}
