// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"fmt"
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

// TestAccAccountAccess_serial runs all Account Access acceptance tests
// sequentially. This is required because AWS Account Access enforces a 1:1
// Application-to-Identity-Center-instance constraint (a second CreateApplication
// against the same instance returns AlreadyCreatedException). Running the tests
// in parallel — the provider's CI default — would cause nondeterministic
// AlreadyCreatedException collisions on the single shared org instance, so the
// individual test functions are unexported and funneled through this one
// parallel-eligible entry point. Mirrors the upstream AppFabric pattern.
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
		"Entitlement": {
			"user":                 testAccAccountAccessEntitlement_user,
			"group":                testAccAccountAccessEntitlement_group,
			acctest.CtDisappears:   testAccAccountAccessEntitlement_disappears,
			"Identity":             testAccAccountAccessEntitlement_identitySerial,
			"List_basic":           testAccAccountAccessEntitlement_List_basic,
			"List_includeResource": testAccAccountAccessEntitlement_List_includeResource,
		},
		"EntitlementsDataSource": {
			"byPrincipal": testAccAccountAccessEntitlementsDataSource_byPrincipal,
			"byRole":      testAccAccountAccessEntitlementsDataSource_byRole,
			"byAccount":   testAccAccountAccessEntitlementsDataSource_byAccount,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, serializeDelay)
}

// testAccPreCheck verifies the test environment can reach Account Access and
// that the prerequisite IAM Identity Center instance exists in the test
// account. It is the standard PreCheck for every acceptance test in this
// package.
//
// Prerequisites that CANNOT be provisioned by Terraform in-test (and so are
// asserted here):
//   - IAM Identity Center must be enabled with an instance in the test region
//     (org-level setup). PreCheckSSOAdminInstances asserts this.
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

// testAccPrerequisitesConfig returns HCL that self-provisions everything an
// Account Access acceptance test needs, EXCEPT the IAM Identity Center
// instance (which is an org-level prerequisite discovered via data source):
//
//   - data.aws_ssoadmin_instances — the IdC instance ARN + identity store ID
//   - aws_identitystore_user       — a USER principal
//   - aws_identitystore_group      — a GROUP principal
//   - aws_iam_role                 — a target role with the Account Access
//     trust policy (account-access.amazonaws.com + sts:AssumeRole,
//     sts:SetContext, sts:TagSession — see CONTEXT.md §4/§5)
//
// Outputs are referenced by callers as:
//
//	data.aws_ssoadmin_instances.test
//	aws_identitystore_user.test.user_id
//	aws_identitystore_group.test.group_id
//	aws_iam_role.test.arn
func testAccPrerequisitesConfig(rName string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

locals {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  instance_arn      = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_identitystore_user" "test" {
  identity_store_id = local.identity_store_id

  display_name = "%[1]s"
  user_name    = "%[1]s"

  name {
    given_name  = "Acceptance"
    family_name = "Test"
  }

  emails {
    value = "%[2]s"
  }
}

resource "aws_identitystore_group" "test" {
  identity_store_id = local.identity_store_id
  display_name      = "%[1]s"
  description       = "Account Access acceptance test group"
}

resource "aws_iam_role" "test" {
  name = "%[1]s"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "account-access.amazonaws.com"
        }
        Action = [
          "sts:AssumeRole",
          "sts:SetContext",
          "sts:TagSession",
        ]
      },
    ]
  })
}
`, rName, acctest.DefaultEmailAddress))
}
