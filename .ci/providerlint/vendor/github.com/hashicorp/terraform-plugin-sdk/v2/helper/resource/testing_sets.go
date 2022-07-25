// These test helpers were developed by the AWS provider team at HashiCorp.

package resource

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	sentinelIndex = "*"
)

// TestCheckTypeSetElemNestedAttrs ensures a subset map of values is stored in
// state for the given name and key combination of attributes nested under a
// list or set block. Use this TestCheckFunc in preference over non-set
// variants to simplify testing code and ensure compatibility with indicies,
// which can easily change with schema changes. State value checking is only
// recommended for testing Computed attributes and attribute defaults.
//
// For managed resources, the name parameter is a combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//     resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//     data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the sentinel value '*' to replace the element indexing into
// a list or set. The sentinel value can be used for each list or set index, if
// there are multiple lists or sets in the attribute path.
//
// The values parameter is the map of attribute names to attribute values
// expected to be nested under the list or set.
//
// You may check for unset nested attributes, however this will also match keys
// set to an empty string. Use a map with at least 1 non-empty value.
//
//     map[string]string{
//       "key1": "value",
//       "key2": "",
//     }
//
// If the values map is not granular enough, it is possible to match an element
// you were not intending to in the set. Provide the most complete mapping of
// attributes possible to be sure the unique element exists.
func TestCheckTypeSetElemNestedAttrs(name, attr string, values map[string]string) TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		attrParts := strings.Split(attr, ".")
		if attrParts[len(attrParts)-1] != sentinelIndex {
			return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
		}
		// account for cases where the user is trying to see if the value is unset/empty
		// there may be ambiguous scenarios where a field was deliberately unset vs set
		// to the empty string, this will match both, which may be a false positive.
		var matchCount int
		for _, v := range values {
			if v != "" {
				matchCount++
			}
		}
		if matchCount == 0 {
			return fmt.Errorf("%#v has no non-empty values", values)
		}

		if testCheckTypeSetElemNestedAttrsInState(is, attrParts, matchCount, values) {
			return nil
		}
		return fmt.Errorf("%q no TypeSet element %q, with nested attrs %#v in state: %#v", name, attr, values, is.Attributes)
	}
}

// TestMatchTypeSetElemNestedAttrs ensures a subset map of values, compared by
// regular expressions, is stored in state for the given name and key
// combination of attributes nested under a list or set block. Use this
// TestCheckFunc in preference over non-set variants to simplify testing code
// and ensure compatibility with indicies, which can easily change with schema
// changes. State value checking is only recommended for testing Computed
// attributes and attribute defaults.
//
// For managed resources, the name parameter is a combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//     resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//     data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the sentinel value '*' to replace the element indexing into
// a list or set. The sentinel value can be used for each list or set index, if
// there are multiple lists or sets in the attribute path.
//
// The values parameter is the map of attribute names to regular expressions
// for matching attribute values expected to be nested under the list or set.
//
// You may check for unset nested attributes, however this will also match keys
// set to an empty string. Use a map with at least 1 non-empty value.
//
//     map[string]*regexp.Regexp{
//       "key1": regexp.MustCompile(`^value`),
//       "key2": regexp.MustCompile(`^$`),
//     }
//
// If the values map is not granular enough, it is possible to match an element
// you were not intending to in the set. Provide the most complete mapping of
// attributes possible to be sure the unique element exists.
func TestMatchTypeSetElemNestedAttrs(name, attr string, values map[string]*regexp.Regexp) TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		attrParts := strings.Split(attr, ".")
		if attrParts[len(attrParts)-1] != sentinelIndex {
			return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
		}
		// account for cases where the user is trying to see if the value is unset/empty
		// there may be ambiguous scenarios where a field was deliberately unset vs set
		// to the empty string, this will match both, which may be a false positive.
		var matchCount int
		for _, v := range values {
			if v != nil {
				matchCount++
			}
		}
		if matchCount == 0 {
			return fmt.Errorf("%#v has no non-empty values", values)
		}

		if testCheckTypeSetElemNestedAttrsInState(is, attrParts, matchCount, values) {
			return nil
		}
		return fmt.Errorf("%q no TypeSet element %q, with the regex provided, match in state: %#v", name, attr, is.Attributes)
	}
}

// TestCheckTypeSetElemAttr is a TestCheckFunc that accepts a resource
// name, an attribute path, which should use the sentinel value '*' for indexing
// into a TypeSet. The function verifies that an element matches the provided
// value.
//
// Use this function over SDK provided TestCheckFunctions when validating a
// TypeSet where its elements are a simple value

// TestCheckTypeSetElemAttr ensures a specific value is stored in state for the
// given name and key combination under a list or set. Use this TestCheckFunc
// in preference over non-set variants to simplify testing code and ensure
// compatibility with indicies, which can easily change with schema changes.
// State value checking is only recommended for testing Computed attributes and
// attribute defaults.
//
// For managed resources, the name parameter is a combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//     resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//     data "myprovider_thing" "example" { ... }
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the sentinel value '*' to replace the element indexing into
// a list or set. The sentinel value can be used for each list or set index, if
// there are multiple lists or sets in the attribute path.
//
// The value parameter is the stringified data to check at the given key. Use
// the following attribute type rules to set the value:
//
//     - Boolean: "false" or "true".
//     - Float/Integer: Stringified number, such as "1.2" or "123".
//     - String: No conversion necessary.
func TestCheckTypeSetElemAttr(name, attr, value string) TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, name)
		if err != nil {
			return err
		}

		err = testCheckTypeSetElem(is, attr, value)
		if err != nil {
			return fmt.Errorf("%q error: %s", name, err)
		}

		return nil
	}
}

// TestCheckTypeSetElemAttrPair ensures value equality in state between the
// first given name and key combination and the second name and key
// combination. State value checking is only recommended for testing Computed
// attributes and attribute defaults.
//
// For managed resources, the name parameter is a combination of the resource
// type, a period (.), and the name label. The name for the below example
// configuration would be "myprovider_thing.example".
//
//     resource "myprovider_thing" "example" { ... }
//
// For data sources, the name parameter is a combination of the keyword "data",
// a period (.), the data source type, a period (.), and the name label. The
// name for the below example configuration would be
// "data.myprovider_thing.example".
//
//     data "myprovider_thing" "example" { ... }
//
// The first and second names may use any combination of managed resources
// and/or data sources.
//
// The key parameter is an attribute path in Terraform CLI 0.11 and earlier
// "flatmap" syntax. Keys start with the attribute name of a top-level
// attribute. Use the sentinel value '*' to replace the element indexing into
// a list or set. The sentinel value can be used for each list or set index, if
// there are multiple lists or sets in the attribute path.
func TestCheckTypeSetElemAttrPair(nameFirst, keyFirst, nameSecond, keySecond string) TestCheckFunc {
	return func(s *terraform.State) error {
		isFirst, err := primaryInstanceState(s, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := primaryInstanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vSecond, okSecond := isSecond.Attributes[keySecond]
		if !okSecond {
			return fmt.Errorf("%s: Attribute %q not set, cannot be checked against TypeSet", nameSecond, keySecond)
		}

		return testCheckTypeSetElemPair(isFirst, keyFirst, vSecond)
	}
}

func testCheckTypeSetElem(is *terraform.InstanceState, attr, value string) error {
	attrParts := strings.Split(attr, ".")
	if attrParts[len(attrParts)-1] != sentinelIndex {
		return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
	}
	for stateKey, stateValue := range is.Attributes {
		if stateValue == value {
			stateKeyParts := strings.Split(stateKey, ".")
			if len(stateKeyParts) == len(attrParts) {
				for i := range attrParts {
					if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
						break
					}
					if i == len(attrParts)-1 {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("no TypeSet element %q, with value %q in state: %#v", attr, value, is.Attributes)
}

func testCheckTypeSetElemPair(is *terraform.InstanceState, attr, value string) error {
	attrParts := strings.Split(attr, ".")
	for stateKey, stateValue := range is.Attributes {
		if stateValue == value {
			stateKeyParts := strings.Split(stateKey, ".")
			if len(stateKeyParts) == len(attrParts) {
				for i := range attrParts {
					if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
						break
					}
					if i == len(attrParts)-1 {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("no TypeSet element %q, with value %q in state: %#v", attr, value, is.Attributes)
}

// testCheckTypeSetElemNestedAttrsInState is a helper function
// to determine if nested attributes and their values are equal to those
// in the instance state. Currently, the function accepts a "values" param of type
// map[string]string or map[string]*regexp.Regexp.
// Returns true if all attributes match, else false.
func testCheckTypeSetElemNestedAttrsInState(is *terraform.InstanceState, attrParts []string, matchCount int, values interface{}) bool {
	matches := make(map[string]int)

	for stateKey, stateValue := range is.Attributes {
		stateKeyParts := strings.Split(stateKey, ".")
		// a Set/List item with nested attrs would have a flatmap address of
		// at least length 3
		// foo.0.name = "bar"
		if len(stateKeyParts) < 3 || len(attrParts) > len(stateKeyParts) {
			continue
		}
		var pathMatch bool
		for i := range attrParts {
			if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
				break
			}
			if i == len(attrParts)-1 {
				pathMatch = true
			}
		}
		if !pathMatch {
			continue
		}
		id := stateKeyParts[len(attrParts)-1]
		nestedAttr := strings.Join(stateKeyParts[len(attrParts):], ".")

		var match bool
		switch t := values.(type) {
		case map[string]string:
			if v, keyExists := t[nestedAttr]; keyExists && v == stateValue {
				match = true
			}
		case map[string]*regexp.Regexp:
			if v, keyExists := t[nestedAttr]; keyExists && v != nil && v.MatchString(stateValue) {
				match = true
			}
		}
		if match {
			matches[id] = matches[id] + 1
			if matches[id] == matchCount {
				return true
			}
		}
	}
	return false
}
