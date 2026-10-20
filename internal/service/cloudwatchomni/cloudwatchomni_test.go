// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni_test

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Tests are serialized because an account can have at most one CloudWatch Omni domain.
func TestAccCloudWatchOmni_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AccessGrant": {
			acctest.CtBasic:          testAccAccessGrant_basic,
			acctest.CtDisappears:     testAccAccessGrant_disappears,
			acctest.CtName:           testAccAccessGrant_name,
			"permission":             testAccAccessGrant_permission,
			"principal":              testAccAccessGrant_principal,
			"principalAccessProfile": testAccAccessGrant_principalAccessProfile,
			"scopedAction":           testAccAccessGrant_scopedAction,
			"validation":             testAccAccessGrant_validation,
			"tags":                   testAccCloudWatchOmniAccessGrant_tagsSerial,
			"Identity":               testAccCloudWatchOmniAccessGrant_identitySerial,
			"List_basic":             testAccAccessGrant_List_basic,
			"List_includeResource":   testAccAccessGrant_List_includeResource,
			"List_regionOverride":    testAccAccessGrant_List_regionOverride,
		},
		"AccessProfile": {
			acctest.CtBasic:        testAccAccessProfile_basic,
			acctest.CtDisappears:   testAccAccessProfile_disappears,
			acctest.CtName:         testAccAccessProfile_name,
			"description":          testAccAccessProfile_description,
			"Identity":             testAccCloudWatchOmniAccessProfile_identitySerial,
			"List_basic":           testAccAccessProfile_List_basic,
			"List_includeResource": testAccAccessProfile_List_includeResource,
			"List_regionOverride":  testAccAccessProfile_List_regionOverride,
		},
		"Domain": {
			acctest.CtBasic:        testAccDomain_basic,
			acctest.CtDisappears:   testAccDomain_disappears,
			acctest.CtName:         testAccDomain_name,
			"identityCenter":       testAccDomain_identityCenter,
			"Identity":             testAccCloudWatchOmniDomain_identitySerial,
			"List_basic":           testAccDomain_List_basic,
			"List_includeResource": testAccDomain_List_includeResource,
			"List_regionOverride":  testAccDomain_List_regionOverride,
		},
		"Space": {
			acctest.CtBasic:               testAccSpace_basic,
			acctest.CtDisappears:          testAccSpace_disappears,
			"agentCore":                   testAccSpace_agentCore,
			acctest.CtName:                testAccSpace_name,
			"encryption.AWS_OWNED":        testAccSpace_encryptionCreateAWS_OWNED,
			"encryption.CUSTOMER_MANAGED": testAccSpace_encryptionCreateCUSTOMER_MANAGED,
			"Identity":                    testAccCloudWatchOmniSpace_identitySerial,
			"List_basic":                  testAccSpace_List_basic,
			"List_includeResource":        testAccSpace_List_includeResource,
			"List_regionOverride":         testAccSpace_List_regionOverride,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).CloudWatchOmniClient(ctx)

	var input cloudwatchomni.ListDomainsInput
	out, err := conn.ListDomains(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Creating a domain fails if the account already has one.
	if len(out.Items) > 0 {
		t.Skip("skipping acceptance testing: account already has a CloudWatch Omni domain")
	}
}
