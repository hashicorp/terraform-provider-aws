// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/jmespath/go-jmespath"
)

const (
	policyModelMarshallJSONStartSliceSize = 2
)

type IAMPolicyDoc struct {
	Version    string                `json:",omitempty"`
	Id         string                `json:",omitempty"`
	Statements []*IAMPolicyStatement `json:"Statement,omitempty"`
}

type IAMPolicyStatement struct {
	Sid           string                         `json:",omitempty"`
	Effect        string                         `json:",omitempty"`
	Actions       interface{}                    `json:"Action,omitempty"`
	NotActions    interface{}                    `json:"NotAction,omitempty"`
	Resources     interface{}                    `json:"Resource,omitempty"`
	NotResources  interface{}                    `json:"NotResource,omitempty"`
	Principals    IAMPolicyStatementPrincipalSet `json:"Principal,omitempty"`
	NotPrincipals IAMPolicyStatementPrincipalSet `json:"NotPrincipal,omitempty"`
	Conditions    IAMPolicyStatementConditionSet `json:"Condition,omitempty"`
}

type IAMPolicyStatementPrincipal struct {
	Type        string
	Identifiers interface{}
}

type IAMPolicyStatementCondition struct {
	Test     string
	Variable string
	Values   interface{}
}

type IAMPolicyStatementPrincipalSet []IAMPolicyStatementPrincipal
type IAMPolicyStatementConditionSet []IAMPolicyStatementCondition

func (s *IAMPolicyDoc) Merge(newDoc *IAMPolicyDoc) {
	// adopt newDoc's Id
	if len(newDoc.Id) > 0 {
		s.Id = newDoc.Id
	}

	// let newDoc upgrade our Version
	if newDoc.Version > s.Version {
		s.Version = newDoc.Version
	}

	// merge in newDoc's statements, overwriting any existing Sids
	var seen bool
	for _, newStatement := range newDoc.Statements {
		if len(newStatement.Sid) == 0 {
			s.Statements = append(s.Statements, newStatement)
			continue
		}
		seen = false
		for i, existingStatement := range s.Statements {
			if existingStatement.Sid == newStatement.Sid {
				s.Statements[i] = newStatement
				seen = true
				break
			}
		}
		if !seen {
			s.Statements = append(s.Statements, newStatement)
		}
	}
}

func (ps IAMPolicyStatementPrincipalSet) MarshalJSON() ([]byte, error) {
	raw := map[string]interface{}{}

	// Although IAM documentation says, that "*" and {"AWS": "*"} are equivalent
	// (https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html),
	// in practice they are not for IAM roles. IAM will return an error if trust
	// policy have "*" or {"*": "*"} as principal, but will accept {"AWS": "*"}.
	// Only {"*": "*"} should be normalized to "*".
	if len(ps) == 1 {
		p := ps[0]
		if p.Type == "*" {
			if sv, ok := p.Identifiers.(string); ok && sv == "*" {
				return []byte(`"*"`), nil
			}

			if av, ok := p.Identifiers.([]string); ok && len(av) == 1 && av[0] == "*" {
				return []byte(`"*"`), nil
			}
		}
	}

	for _, p := range ps {
		switch i := p.Identifiers.(type) {
		case []string:
			switch v := raw[p.Type].(type) {
			case nil:
				raw[p.Type] = make([]string, 0, len(i))
			case string:
				// Convert to []string to prevent panic
				raw[p.Type] = make([]string, 0, len(i)+1)
				raw[p.Type] = append(raw[p.Type].([]string), v)
			}
			sort.Sort(sort.Reverse(sort.StringSlice(i)))
			raw[p.Type] = append(raw[p.Type].([]string), i...)
		case string:
			switch v := raw[p.Type].(type) {
			case nil:
				raw[p.Type] = i
			case string:
				// Convert to []string to stop drop of principals
				raw[p.Type] = make([]string, 0, policyModelMarshallJSONStartSliceSize)
				raw[p.Type] = append(raw[p.Type].([]string), v)
				raw[p.Type] = append(raw[p.Type].([]string), i)
			case []string:
				raw[p.Type] = append(raw[p.Type].([]string), i)
			}
		default:
			return []byte{}, fmt.Errorf("Unsupported data type %T for IAMPolicyStatementPrincipalSet", i)
		}
	}

	return json.Marshal(&raw)
}

func (ps *IAMPolicyStatementPrincipalSet) UnmarshalJSON(b []byte) error {
	var out IAMPolicyStatementPrincipalSet

	var data interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	switch t := data.(type) {
	case string:
		out = append(out, IAMPolicyStatementPrincipal{Type: "*", Identifiers: []string{"*"}})
	case map[string]interface{}:
		for key, value := range data.(map[string]interface{}) {
			switch vt := value.(type) {
			case string:
				out = append(out, IAMPolicyStatementPrincipal{Type: key, Identifiers: value.(string)})
			case []interface{}:
				values := []string{}
				for _, v := range value.([]interface{}) {
					values = append(values, v.(string))
				}
				sort.Strings(values)
				out = append(out, IAMPolicyStatementPrincipal{Type: key, Identifiers: values})
			default:
				return fmt.Errorf("Unsupported data type %T for IAMPolicyStatementPrincipalSet.Identifiers", vt)
			}
		}
	default:
		return fmt.Errorf("Unsupported data type %T for IAMPolicyStatementPrincipalSet", t)
	}

	*ps = out
	return nil
}

func (cs IAMPolicyStatementConditionSet) MarshalJSON() ([]byte, error) {
	raw := map[string]map[string]interface{}{}

	for _, c := range cs {
		if _, ok := raw[c.Test]; !ok {
			raw[c.Test] = map[string]interface{}{}
		}
		if _, ok := raw[c.Test][c.Variable]; !ok {
			raw[c.Test][c.Variable] = []string{}
		}
		switch i := c.Values.(type) {
		case []string:
			// order matters with values so not sorting here
			raw[c.Test][c.Variable] = append(raw[c.Test][c.Variable].([]string), i...)
		case string:
			raw[c.Test][c.Variable] = append(raw[c.Test][c.Variable].([]string), i)
		default:
			return nil, fmt.Errorf("Unsupported data type for IAMPolicyStatementConditionSet: %s", i)
		}
	}

	// flatten entries with a single item to match AWS IAM syntax
	for k1 := range raw {
		for k2 := range raw[k1] {
			items := raw[k1][k2].([]string)
			if len(items) == 1 {
				raw[k1][k2] = items[0]
			}
		}
	}

	return json.Marshal(&raw)
}

func (cs *IAMPolicyStatementConditionSet) UnmarshalJSON(b []byte) error {
	var out IAMPolicyStatementConditionSet

	var data map[string]map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}

	for test_key, test_value := range data {
		for var_key, var_values := range test_value {
			switch var_values := var_values.(type) {
			case string:
				out = append(out, IAMPolicyStatementCondition{Test: test_key, Variable: var_key, Values: []string{var_values}})
			case bool:
				out = append(out, IAMPolicyStatementCondition{Test: test_key, Variable: var_key, Values: strconv.FormatBool(var_values)})
			case []interface{}:
				values := []string{}
				for _, v := range var_values {
					values = append(values, v.(string))
				}
				out = append(out, IAMPolicyStatementCondition{Test: test_key, Variable: var_key, Values: values})
			}
		}
	}

	*cs = out
	return nil
}

func policyDecodeConfigStringList(lI []interface{}) interface{} {
	if len(lI) == 1 {
		return lI[0].(string)
	}
	ret := make([]string, len(lI))
	for i, vI := range lI {
		ret[i] = vI.(string)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(ret)))
	return ret
}

// PolicyHasValidAWSPrincipals validates that the Principals in an IAM Policy are valid
// Assumes that non-"AWS" Principals are valid
// The value can be a single string or a slice of strings
// Valid strings are either an ARN or an AWS account ID
func PolicyHasValidAWSPrincipals(policy string) (bool, error) { // nosemgrep:ci.aws-in-func-name
	var policyData any
	err := json.Unmarshal([]byte(policy), &policyData)
	if err != nil {
		return false, fmt.Errorf("parsing policy: %w", err)
	}

	result, err := jmespath.Search("Statement[*].Principal.AWS", policyData)
	if err != nil {
		return false, fmt.Errorf("parsing policy: %w", err)
	}

	principals, ok := result.([]any)
	if !ok {
		return false, fmt.Errorf(`parsing policy: unexpected result: (%[1]T) "%[1]v"`, result)
	}

	for _, principal := range principals {
		switch x := principal.(type) {
		case string:
			if !IsValidPolicyAWSPrincipal(x) {
				return false, nil
			}
		case []string:
			for _, s := range x {
				if !IsValidPolicyAWSPrincipal(s) {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

// IsValidPolicyAWSPrincipal returns true if a string is a valid AWS Princial for an IAM Policy document
// That is: either an ARN, an AWS account ID, or `*`
func IsValidPolicyAWSPrincipal(principal string) bool { // nosemgrep:ci.aws-in-func-name
	if principal == "*" {
		return true
	}
	if arn.IsARN(principal) {
		return true
	}
	if regexache.MustCompile(`^\d{12}$`).MatchString(principal) {
		return true
	}
	return false
}
