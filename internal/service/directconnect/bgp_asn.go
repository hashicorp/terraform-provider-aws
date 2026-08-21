// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	bgpASNAttributeName     = "bgp_asn"
	bgpASNLongAttributeName = "bgp_asn_long"
	maxBGPASN               = 2147483646
	maxBGPASNLong           = 4294967294
)

func bgpASNAttributeSchema(long bool) *schema.Schema {
	exactlyOneOf := []string{bgpASNAttributeName, bgpASNLongAttributeName}

	if long {
		return &schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ForceNew:     true,
			ExactlyOneOf: exactlyOneOf,
			ValidateFunc: validateBGPASNLong,
		}
	}

	return &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: exactlyOneOf,
		ValidateFunc: validation.IntBetween(1, maxBGPASN),
	}
}

func validateBGPASNLong(v any, k string) ([]string, []error) {
	value := v.(string)

	if value == "" || value[0] == '0' {
		return nil, []error{fmt.Errorf("%q must be an ASN in asplain format between 1 and %d", k, maxBGPASNLong)}
	}

	for _, r := range value {
		if r < '0' || r > '9' {
			return nil, []error{fmt.Errorf("%q must be an ASN in asplain format between 1 and %d", k, maxBGPASNLong)}
		}
	}

	asn, err := strconv.ParseInt(value, 10, 64)
	if err != nil || asn < 1 || asn > maxBGPASNLong {
		return nil, []error{fmt.Errorf("%q must be an ASN in asplain format between 1 and %d", k, maxBGPASNLong)}
	}

	return nil, nil
}

func expandBGPASN(d *schema.ResourceData) (int32, *int64) {
	if v, ok := d.GetOk(bgpASNLongAttributeName); ok {
		asnLong, _ := strconv.ParseInt(v.(string), 10, 64)
		return 0, aws.Int64(asnLong)
	}

	return int32(d.Get(bgpASNAttributeName).(int)), nil
}

func effectiveBGPASN(asn int32, asnLong *int64) int64 {
	if asnLong != nil {
		return aws.ToInt64(asnLong)
	}

	return int64(asn)
}

func setBGPASN(d *schema.ResourceData, asn int32, asnLong *int64) error {
	effectiveASN := effectiveBGPASN(asn, asnLong)

	if _, ok := d.GetOk(bgpASNLongAttributeName); ok || asnLong != nil && effectiveASN > maxBGPASN {
		return d.Set(bgpASNLongAttributeName, strconv.FormatInt(effectiveASN, 10))
	}

	return d.Set(bgpASNAttributeName, int(effectiveASN))
}
