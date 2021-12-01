package verify

import (
	"bytes"
	"encoding/json"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func SuppressEquivalentPolicyDiffs(k, old, new string, d *schema.ResourceData) bool {
	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)
	if err != nil {
		return false
	}

	return equivalent
}

func SuppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	ob := bytes.NewBufferString("")
	if err := json.Compact(ob, []byte(old)); err != nil {
		return false
	}

	nb := bytes.NewBufferString("")
	if err := json.Compact(nb, []byte(new)); err != nil {
		return false
	}

	return JSONBytesEqual(ob.Bytes(), nb.Bytes())
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
	return regexp.MustCompile(`^\s*{`).MatchString(s.(string))
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

	if strings.TrimSpace(old) == "" {
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
