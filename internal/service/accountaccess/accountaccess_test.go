// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// serializeDelay is applied between serialized subtests. Account Access allows
// only one Application per IAM Identity Center instance, and a test account has
// a single instance, so every Application-creating test contends for it.
const serializeDelay = 5 * time.Second

// TestAccAccountAccess_serial runs all Account Access acceptance tests
// sequentially because Account Access permits only one Application per IAM
// Identity Center instance.
func TestAccAccountAccess_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Application": {
			acctest.CtBasic:        testAccApplication_basic,
			acctest.CtDisappears:   testAccApplication_disappears,
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

func testAccPreCheck(ctx context.Context, t *testing.T) {
	acctest.PreCheckSSOAdminInstances(ctx, t)
	acctest.PreCheckOrganizationsEnabledServicePrincipal(ctx, t, "account-access.amazonaws.com")
}

// testAccPrerequisitesConfig creates the user, group, and IAM role needed by
// entitlement acceptance cases. The randomized name is required by the
// Identity Store and IAM APIs; the Identity Center instance is pre-existing.
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
