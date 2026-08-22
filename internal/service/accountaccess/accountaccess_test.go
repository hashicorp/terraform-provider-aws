// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// serializeDelay is applied between serialized subtests. Account Access allows
// only one Application per IAM Identity Center instance, and a test account has
// a single instance, so every Application-creating test contends for it. The
// tests must run serially (via TestAccAccountAccess_serial), and a short delay
// smooths the delete→create transition on the shared instance.
const serializeDelay = 5 * time.Second

// TestAccAccountAccess_serial runs every Application acceptance group
// sequentially. AWS Account Access enforces a 1:1 Application-to-Identity-
// Center-instance constraint, so concurrent CreateApplication calls against the
// shared organization instance can fail with AlreadyCreatedException. Each
// group is independently runnable and its CheckDestroy verifies cleanup before
// the next group begins.
func TestAccAccountAccess_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Application": {
			acctest.CtBasic:        testAccAccountAccessApplication_basic,
			acctest.CtDisappears:   testAccAccountAccessApplication_disappears,
			"tags":                 testAccAccountAccessApplication_tagsSerial,
			"Identity":             testAccAccountAccessApplication_identitySerial,
			"List_basic":           testAccAccountAccessApplication_List_basic,
			"List_includeResource": testAccAccountAccessApplication_List_includeResource,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, serializeDelay)
}

// testAccPreCheck verifies the test environment can reach Account Access and
// that the prerequisite IAM Identity Center instance exists in the test
// account. IAM Identity Center is an organization-level prerequisite that the
// test discovers rather than provisions.
func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckSSOAdminInstances(ctx, t)

	conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

	_, err := conn.ListApplications(ctx, &accountaccess.ListApplicationsInput{})
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// testAccPrerequisitesConfig returns the IAM Identity Center instance
// discovery shared by the inline Application acceptance configurations.
func testAccPrerequisitesConfig(_ string) string {
	return `
data "aws_ssoadmin_instances" "test" {}

locals {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`
}
