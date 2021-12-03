package verify

// This code started as jen20/awspolicyequivalence

// Needing to continue development and maintenance, and the
// inability of the jen20 owner to do so (https://github.com/jen20/awspolicyequivalence/issues/10#issuecomment-931409953),
// HashiCorp is incorporating and building on in conformance with MPL 2.0.

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/mitchellh/mapstructure"
)

// PoliciesAreEquivalent tests for the structural equivalence of two
// AWS policies. It does not read into the semantics, other than treating
// single element string arrays as equivalent to a string without an
// array, as the AWS endpoints do.
//
// It will, however, detect reordering and ignore whitespace.
//
// Returns true if the policies are structurally equivalent, false
// otherwise. If either of the input strings are not valid JSON,
// false is returned along with an error.
func PoliciesAreEquivalent(policy1, policy2 string) (bool, error) {
	policy1intermediate := &intermediatePolicyDocument{}
	if err := json.Unmarshal([]byte(policy1), policy1intermediate); err != nil {
		return false, fmt.Errorf("Error unmarshaling policy: %s", err)
	}

	policy2intermediate := &intermediatePolicyDocument{}
	if err := json.Unmarshal([]byte(policy2), policy2intermediate); err != nil {
		return false, fmt.Errorf("Error unmarshaling policy: %s", err)
	}

	policy1Doc, err := policy1intermediate.document()
	if err != nil {
		return false, fmt.Errorf("Error parsing policy: %s", err)
	}
	policy2Doc, err := policy2intermediate.document()
	if err != nil {
		return false, fmt.Errorf("Error parsing policy: %s", err)
	}

	return policy1Doc.equals(policy2Doc), nil
}

type intermediatePolicyDocument struct {
	Version    string      `json:",omitempty"`
	Id         string      `json:",omitempty"`
	Statements interface{} `json:"Statement"`
}

func (intermediate *intermediatePolicyDocument) document() (*policyDocument, error) {
	var statements []*policyStatement

	switch s := intermediate.Statements.(type) {
	case []interface{}:
		if err := mapstructure.Decode(s, &statements); err != nil {
			return nil, fmt.Errorf("Error parsing statement: %s", err)
		}
	case map[string]interface{}:
		var singleStatement *policyStatement
		if err := mapstructure.Decode(s, &singleStatement); err != nil {
			return nil, fmt.Errorf("Error parsing statement: %s", err)
		}
		statements = append(statements, singleStatement)
	default:
		return nil, errors.New("Unknown error parsing statement")
	}

	document := &policyDocument{
		Version:    intermediate.Version,
		Id:         intermediate.Id,
		Statements: statements,
	}

	return document, nil
}

type policyDocument struct {
	Version    string
	Id         string
	Statements []*policyStatement
}

func (doc *policyDocument) equals(other *policyDocument) bool {
	// Check the basic fields of the document
	if doc.Version != other.Version {
		return false
	}
	if doc.Id != other.Id {
		return false
	}

	// If we have different number of statements we are very unlikely
	// to have them be equivalent.
	if len(doc.Statements) != len(other.Statements) {
		return false
	}

	// If we have the same number of statements in the policy, does
	// each statement in the intermediate have a corresponding statement in
	// other which is equal? If no, policies are not equal, if yes,
	// then they may be.
	for _, ours := range doc.Statements {
		found := false
		for _, theirs := range other.Statements {
			if ours.equals(theirs) {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	// Now we need to repeat this process the other way around to
	// ensure we don't have any matching errors.
	for _, theirs := range other.Statements {
		found := false
		for _, ours := range doc.Statements {
			if theirs.equals(ours) {
				found = true
			}
		}

		if !found {
			return false
		}
	}

	return true
}

type policyStatement struct {
	Sid           string                            `json:",omitempty" mapstructure:"Sid"`
	Effect        string                            `json:",omitempty" mapstructure:"Effect"`
	Actions       interface{}                       `json:"Action,omitempty" mapstructure:"Action"`
	NotActions    interface{}                       `json:"NotAction,omitempty" mapstructure:"NotAction"`
	Resources     interface{}                       `json:"Resource,omitempty" mapstructure:"Resource"`
	NotResources  interface{}                       `json:"NotResource,omitempty" mapstructure:"NotResource"`
	Principals    interface{}                       `json:"Principal,omitempty" mapstructure:"Principal"`
	NotPrincipals interface{}                       `json:"NotPrincipal,omitempty" mapstructure:"NotPrincipal"`
	Conditions    map[string]map[string]interface{} `json:"Condition,omitempty" mapstructure:"Condition"`
}

func (statement *policyStatement) equals(other *policyStatement) bool {
	if statement.Sid != other.Sid {
		return false
	}

	if strings.ToLower(statement.Effect) != strings.ToLower(other.Effect) {
		return false
	}

	ourActions := newStringSet(statement.Actions)
	theirActions := newStringSet(other.Actions)
	if !StringSlicesEqualIgnoreOrder(ourActions, theirActions) {
		return false
	}

	ourNotActions := newStringSet(statement.NotActions)
	theirNotActions := newStringSet(other.NotActions)
	if !StringSlicesEqualIgnoreOrder(ourNotActions, theirNotActions) {
		return false
	}

	ourResources := newStringSet(statement.Resources)
	theirResources := newStringSet(other.Resources)
	if !StringSlicesEqualIgnoreOrder(ourResources, theirResources) {
		return false
	}

	ourNotResources := newStringSet(statement.NotResources)
	theirNotResources := newStringSet(other.NotResources)
	if !StringSlicesEqualIgnoreOrder(ourNotResources, theirNotResources) {
		return false
	}

	ourConditionsBlock := conditionsBlock(statement.Conditions)
	theirConditionsBlock := conditionsBlock(other.Conditions)
	if !ourConditionsBlock.Equals(theirConditionsBlock) {
		return false
	}

	if statement.Principals != nil || other.Principals != nil {
		stringPrincipalsEqual := stringPrincipalsEqual(statement.Principals, other.Principals)
		mapPrincipalsEqual := mapPrincipalsEqual(statement.Principals, other.Principals)
		if !(stringPrincipalsEqual || mapPrincipalsEqual) {
			return false
		}
	}

	if statement.NotPrincipals != nil || other.NotPrincipals != nil {
		stringNotPrincipalsEqual := stringPrincipalsEqual(statement.NotPrincipals, other.NotPrincipals)
		mapNotPrincipalsEqual := mapPrincipalsEqual(statement.NotPrincipals, other.NotPrincipals)
		if !(stringNotPrincipalsEqual || mapNotPrincipalsEqual) {
			return false
		}
	}

	return true
}

func mapPrincipalsEqual(ours, theirs interface{}) bool {
	ourPrincipalMap, oursOk := ours.(map[string]interface{})
	theirPrincipalMap, theirsOk := theirs.(map[string]interface{})

	oursNormalized := make(map[string]principalStringSet)
	if oursOk {
		for key, val := range ourPrincipalMap {
			var tmp = newPrincipalStringSet(val)
			if len(tmp) > 0 {
				oursNormalized[key] = tmp
			}
		}
	}

	theirsNormalized := make(map[string]principalStringSet)
	if theirsOk {
		for key, val := range theirPrincipalMap {
			var tmp = newPrincipalStringSet(val)
			if len(tmp) > 0 {
				theirsNormalized[key] = newPrincipalStringSet(val)
			}
		}
	}

	for key, ours := range oursNormalized {
		theirs, ok := theirsNormalized[key]
		if !ok {
			return false
		}

		if !ours.equals(theirs) {
			return false
		}
	}

	for key, theirs := range theirsNormalized {
		ours, ok := oursNormalized[key]
		if !ok {
			return false
		}

		if !theirs.equals(ours) {
			return false
		}
	}

	return true
}

func stringPrincipalsEqual(ours, theirs interface{}) bool {
	ourPrincipal, oursIsString := ours.(string)
	theirPrincipal, theirsIsString := theirs.(string)

	if !(oursIsString && theirsIsString) {
		return false
	}

	if ourPrincipal == theirPrincipal {
		return true
	}

	// Handle AWS converting account ID principal to root IAM user ARN
	// ACCOUNTID == arn:PARTITION:iam::ACCOUNTID:root
	accountIDRegex := regexp.MustCompile(`^[0-9]{12}$`)

	if accountIDRegex.MatchString(ourPrincipal) {
		if theirArn, err := arn.Parse(theirPrincipal); err == nil {
			if theirArn.Service == "iam" && theirArn.Resource == "root" && theirArn.AccountID == ourPrincipal {
				return true
			}
		}
	}

	if accountIDRegex.MatchString(theirPrincipal) {
		if ourArn, err := arn.Parse(ourPrincipal); err == nil {
			if ourArn.Service == "iam" && ourArn.Resource == "root" && ourArn.AccountID == theirPrincipal {
				return true
			}
		}
	}

	return false
}

type conditionsBlock map[string]map[string]interface{}

func (conditions conditionsBlock) Equals(other conditionsBlock) bool {
	if conditions == nil && other != nil || other == nil && conditions != nil {
		return false
	}

	if len(conditions) != len(other) {
		return false
	}

	oursNormalized := make(map[string]map[string]stringSet)
	for key, condition := range conditions {
		normalizedCondition := make(map[string]stringSet)
		for innerKey, val := range condition {
			normalizedCondition[innerKey] = newStringSet(val)
		}
		oursNormalized[key] = normalizedCondition
	}

	theirsNormalized := make(map[string]map[string]stringSet)
	for key, condition := range other {
		normalizedCondition := make(map[string]stringSet)
		for innerKey, val := range condition {
			normalizedCondition[innerKey] = newStringSet(val)
		}
		theirsNormalized[key] = normalizedCondition
	}

	for key, ours := range oursNormalized {
		theirs, ok := theirsNormalized[key]
		if !ok {
			return false
		}

		for innerKey, oursInner := range ours {
			theirsInner, ok := theirs[innerKey]
			if !ok {
				return false
			}

			if !oursInner.equals(theirsInner) {
				return false
			}
		}
	}

	for key, theirs := range theirsNormalized {
		ours, ok := oursNormalized[key]
		if !ok {
			return false
		}

		for innerKey, theirsInner := range theirs {
			oursInner, ok := ours[innerKey]
			if !ok {
				return false
			}

			if !theirsInner.equals(oursInner) {
				return false
			}
		}
	}

	return true
}

type stringSet []string
type principalStringSet stringSet

// newStringSet constructs an stringSet from an interface{} - which
// may be nil, a single string, or []interface{} (each of which is a string).
// This corresponds with how structures come off the JSON unmarshaler
// without any custom encoding rules.
func newStringSet(members interface{}) stringSet {
	if members == nil {
		return stringSet{}
	}

	if single, ok := members.(string); ok {
		return stringSet{single}
	}

	if multiple, ok := members.([]interface{}); ok {
		if len(multiple) == 0 {
			return stringSet{}
		}
		actions := make([]string, len(multiple))
		for i, action := range multiple {
			if _, ok := action.(string); !ok {
				return nil
			}

			actions[i] = action.(string)
		}
		return stringSet(actions)
	}

	return nil
}

func newPrincipalStringSet(members interface{}) principalStringSet {
	return principalStringSet(newStringSet(members))
}

func (actions stringSet) equals(other stringSet) bool {
	if actions == nil || other == nil {
		return false
	}

	if len(actions) != len(other) {
		return false
	}

	ourMap := map[string]struct{}{}
	theirMap := map[string]struct{}{}

	for _, action := range actions {
		ourMap[action] = struct{}{}
	}

	for _, action := range other {
		theirMap[action] = struct{}{}
	}

	return reflect.DeepEqual(ourMap, theirMap)
}

func (ours principalStringSet) equals(theirs principalStringSet) bool {
	if len(ours) != len(theirs) {
		return false
	}

	for _, ourPrincipal := range ours {
		matches := false
		for _, theirPrincipal := range theirs {
			if stringPrincipalsEqual(ourPrincipal, theirPrincipal) {
				matches = true
				break
			}
		}
		if !matches {
			return false
		}
	}

	return true
}
