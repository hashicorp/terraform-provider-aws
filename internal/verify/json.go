// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/YakDriver/regexache"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func SuppressEquivalentPolicyDiffs(k, old, new string, d *schema.ResourceData) bool {
	if strings.TrimSpace(old) == "" && strings.TrimSpace(new) == "" {
		return true
	}

	if strings.TrimSpace(old) == "{}" && strings.TrimSpace(new) == "" {
		return true
	}

	if strings.TrimSpace(old) == "" && strings.TrimSpace(new) == "{}" {
		return true
	}

	if strings.TrimSpace(old) == "{}" && strings.TrimSpace(new) == "{}" {
		return true
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)
	if err != nil {
		return false
	}

	return equivalent
}

func SuppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	return JSONStringsEqual(old, new)
}

func SuppressEquivalentJSONOrYAMLDiffs(k, old, new string, d *schema.ResourceData) bool {
	normalizedOld, err := NormalizeJSONOrYAMLString(old)

	if err != nil {
		log.Printf("[WARN] Unable to normalize Terraform state CloudFormation template body: %s", err)
		return false
	}

	normalizedNew, err := NormalizeJSONOrYAMLString(new)

	if err != nil {
		log.Printf("[WARN] Unable to normalize Terraform configuration CloudFormation template body: %s", err)
		return false
	}

	return normalizedOld == normalizedNew
}

func NormalizeJSONOrYAMLString(templateString interface{}) (string, error) {
	if looksLikeJSONString(templateString) {
		return structure.NormalizeJsonString(templateString.(string))
	}

	return checkYAMLString(templateString)
}

func looksLikeJSONString(s interface{}) bool {
	return regexache.MustCompile(`^\s*{`).MatchString(s.(string))
}

func JSONStringsEqual(s1, s2 string) bool {
	b1 := bytes.NewBufferString("")
	if err := json.Compact(b1, []byte(s1)); err != nil {
		return false
	}

	b2 := bytes.NewBufferString("")
	if err := json.Compact(b2, []byte(s2)); err != nil {
		return false
	}

	return JSONBytesEqual(b1.Bytes(), b2.Bytes())
}

func JSONBytesEqual(b1, b2 []byte) bool {
	var o1 interface{}
	if err := json.Unmarshal(b1, &o1); err != nil {
		return false
	}

	var o2 interface{}
	if err := json.Unmarshal(b2, &o2); err != nil {
		return false
	}

	return reflect.DeepEqual(o1, o2)
}

func SecondJSONUnlessEquivalent(old, new string) (string, error) {
	// valid empty JSON is "{}" not "" so handle special case to avoid
	// Error unmarshaling policy: unexpected end of JSON input
	if strings.TrimSpace(new) == "" {
		return "", nil
	}

	if strings.TrimSpace(new) == "{}" {
		return "{}", nil
	}

	if strings.TrimSpace(old) == "" || strings.TrimSpace(old) == "{}" {
		return new, nil
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)

	if err != nil {
		return "", err
	}

	if equivalent {
		return old, nil
	}

	return new, nil
}

// PolicyToSet returns the existing policy if the new policy is equivalent.
// Otherwise, it returns the new policy. Either policy is normalized.
func PolicyToSet(exist, new string) (string, error) {
	policyToSet, err := SecondJSONUnlessEquivalent(exist, new)
	if err != nil {
		return "", fmt.Errorf("while checking equivalency of existing policy (%s) and new policy (%s), encountered: %w", exist, new, err)
	}

	policyToSet, err = structure.NormalizeJsonString(policyToSet)
	if err != nil {
		return "", fmt.Errorf("policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	return policyToSet, nil
}

// LegacyPolicyNormalize returns a "normalized" JSON policy document except
// the Version element is first in the JSON as required by AWS in many places.
// Version not being first is one reason for this error:
// MalformedPolicyDocument: The policy failed legacy parsing
func LegacyPolicyNormalize(policy interface{}) (string, error) {
	if policy == nil || policy.(string) == "" {
		return "", nil
	}

	np, err := structure.NormalizeJsonString(policy)
	if err != nil {
		return policy.(string), fmt.Errorf("legacy policy (%s) is invalid JSON: %w", policy, err)
	}

	m := regexache.MustCompile(`(?s)^(\{\n?)(.*?)(,\s*)?(  )?("Version":\s*"2012-10-17")(,)?(\n)?(.*?)(\})`)

	n := m.ReplaceAllString(np, `$1$4$5$3$2$6$7$8$9`)

	_, err = structure.NormalizeJsonString(n)
	if err != nil {
		return policy.(string), fmt.Errorf("LegacyPolicyNormalize created a policy (%s) that is invalid JSON: %w", n, err)
	}

	return n, nil
}

// LegacyPolicyToSet returns the existing policy if the new policy is equivalent.
// Otherwise, it returns the new policy. Either policy is legacy normalized.
func LegacyPolicyToSet(exist, new string) (string, error) {
	policyToSet, err := SecondJSONUnlessEquivalent(exist, new)
	if err != nil {
		return "", fmt.Errorf("while checking equivalency of existing policy (%s) and new policy (%s), encountered: %w", exist, new, err)
	}

	policyToSet, err = LegacyPolicyNormalize(policyToSet)
	if err != nil {
		return "", fmt.Errorf("legacy policy (%s) is invalid JSON: %w", policyToSet, err)
	}

	return policyToSet, nil
}
