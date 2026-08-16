// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestValidateBGPASNLong(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value string
		valid bool
	}{
		"minimum":         {value: "1", valid: true},
		"legacy range":    {value: "65000", valid: true},
		"four byte range": {value: "4294967294", valid: true},
		"empty":           {value: "", valid: false},
		"zero":            {value: "0", valid: false},
		"leading zero":    {value: "01", valid: false},
		"sign":            {value: "+1", valid: false},
		"whitespace":      {value: " 1", valid: false},
		"asdot":           {value: "1.10", valid: false},
		"out of range":    {value: "4294967295", valid: false},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, errors := validateBGPASNLong(tc.value, bgpASNLongAttributeName)
			if got := len(errors) == 0; got != tc.valid {
				t.Fatalf("validateBGPASNLong(%q) valid = %t, want %t: %v", tc.value, got, tc.valid, errors)
			}
		})
	}
}

func TestBGPASNAttributeSchema(t *testing.T) {
	t.Parallel()

	expectedExactlyOneOf := []string{bgpASNAttributeName, bgpASNLongAttributeName}

	legacy := bgpASNAttributeSchema(false)
	if legacy.Type != schema.TypeInt {
		t.Errorf("legacy Type = %v, want %v", legacy.Type, schema.TypeInt)
	}
	if !legacy.Optional || legacy.Required || !legacy.ForceNew {
		t.Errorf("legacy Optional, Required, ForceNew = %t, %t, %t, want true, false, true", legacy.Optional, legacy.Required, legacy.ForceNew)
	}
	if !reflect.DeepEqual(legacy.ExactlyOneOf, expectedExactlyOneOf) {
		t.Errorf("legacy ExactlyOneOf = %v, want %v", legacy.ExactlyOneOf, expectedExactlyOneOf)
	}
	for _, value := range []int{0, maxBGPASN + 1} {
		if _, errors := legacy.ValidateFunc(value, bgpASNAttributeName); len(errors) == 0 {
			t.Errorf("legacy validation accepted %d", value)
		}
	}
	if _, errors := legacy.ValidateFunc(maxBGPASN, bgpASNAttributeName); len(errors) != 0 {
		t.Errorf("legacy validation rejected %d: %v", maxBGPASN, errors)
	}

	long := bgpASNAttributeSchema(true)
	if long.Type != schema.TypeString {
		t.Errorf("long Type = %v, want %v", long.Type, schema.TypeString)
	}
	if !long.Optional || long.Required || !long.ForceNew {
		t.Errorf("long Optional, Required, ForceNew = %t, %t, %t, want true, false, true", long.Optional, long.Required, long.ForceNew)
	}
	if !reflect.DeepEqual(long.ExactlyOneOf, expectedExactlyOneOf) {
		t.Errorf("long ExactlyOneOf = %v, want %v", long.ExactlyOneOf, expectedExactlyOneOf)
	}
}

func TestBGPASNPrivateVirtualInterfaceSchema(t *testing.T) {
	t.Parallel()

	resourceSchema := resourcePrivateVirtualInterface().SchemaFunc()

	if resourceSchema[bgpASNAttributeName].Type != schema.TypeInt {
		t.Errorf("%s Type = %v, want %v", bgpASNAttributeName, resourceSchema[bgpASNAttributeName].Type, schema.TypeInt)
	}
	if resourceSchema[bgpASNLongAttributeName].Type != schema.TypeString {
		t.Errorf("%s Type = %v, want %v", bgpASNLongAttributeName, resourceSchema[bgpASNLongAttributeName].Type, schema.TypeString)
	}
	if !reflect.DeepEqual(resourceSchema[bgpASNAttributeName].ExactlyOneOf, resourceSchema[bgpASNLongAttributeName].ExactlyOneOf) {
		t.Errorf("BGP ASN attributes have mismatched ExactlyOneOf values")
	}
}

func TestExpandBGPASN(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		raw           map[string]any
		wantASN       int32
		wantASNLong   int64
		wantLongIsNil bool
	}{
		"legacy": {
			raw:           map[string]any{bgpASNAttributeName: 65000},
			wantASN:       65000,
			wantLongIsNil: true,
		},
		"long": {
			raw:         map[string]any{bgpASNLongAttributeName: "4294967294"},
			wantASNLong: 4294967294,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			asn, asnLong := expandBGPASN(testBGPASNResourceData(t, tc.raw))
			if asn != tc.wantASN {
				t.Errorf("Asn = %d, want %d", asn, tc.wantASN)
			}
			if got := asnLong == nil; got != tc.wantLongIsNil {
				t.Fatalf("AsnLong nil = %t, want %t", got, tc.wantLongIsNil)
			}
			if asnLong != nil && aws.ToInt64(asnLong) != tc.wantASNLong {
				t.Errorf("AsnLong = %d, want %d", aws.ToInt64(asnLong), tc.wantASNLong)
			}
		})
	}
}

func TestEffectiveBGPASN(t *testing.T) {
	t.Parallel()

	if got := effectiveBGPASN(65000, nil); got != 65000 {
		t.Errorf("effective legacy ASN = %d, want 65000", got)
	}
	if got := effectiveBGPASN(0, aws.Int64(4294967294)); got != 4294967294 {
		t.Errorf("effective long ASN = %d, want 4294967294", got)
	}
}

func TestSetBGPASN(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		raw           map[string]any
		asn           int32
		asnLong       *int64
		wantLegacy    int
		wantLong      string
		wantLongIsSet bool
	}{
		"legacy configuration": {
			raw:        map[string]any{bgpASNAttributeName: 65000},
			asn:        65000,
			wantLegacy: 65000,
		},
		"long configuration in legacy range": {
			raw:           map[string]any{bgpASNLongAttributeName: "65000"},
			asn:           65000,
			asnLong:       aws.Int64(65000),
			wantLong:      "65000",
			wantLongIsSet: true,
		},
		"imported long ASN": {
			asnLong:       aws.Int64(4294967294),
			wantLong:      "4294967294",
			wantLongIsSet: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := testBGPASNResourceData(t, tc.raw)
			if err := setBGPASN(d, tc.asn, tc.asnLong); err != nil {
				t.Fatalf("setBGPASN returned error: %s", err)
			}
			if got := d.Get(bgpASNAttributeName).(int); got != tc.wantLegacy {
				t.Errorf("%s = %d, want %d", bgpASNAttributeName, got, tc.wantLegacy)
			}
			gotLong, longIsSet := d.GetOk(bgpASNLongAttributeName)
			if longIsSet != tc.wantLongIsSet {
				t.Fatalf("%s set = %t, want %t", bgpASNLongAttributeName, longIsSet, tc.wantLongIsSet)
			}
			if longIsSet && gotLong.(string) != tc.wantLong {
				t.Errorf("%s = %q, want %q", bgpASNLongAttributeName, gotLong, tc.wantLong)
			}
		})
	}
}

func testBGPASNResourceData(t *testing.T, raw map[string]any) *schema.ResourceData {
	t.Helper()

	return schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		bgpASNAttributeName:     bgpASNAttributeSchema(false),
		bgpASNLongAttributeName: bgpASNAttributeSchema(true),
	}, raw)
}
