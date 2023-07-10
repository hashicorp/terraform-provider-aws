// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package organizations_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccPolicyAttachment_Account(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy_attachment.test"
	policyIdResourceName := "aws_organizations_policy.test"
	targetIdResourceName := "aws_organizations_organization.test"

	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`
	tagPolicyContent := `{ "tags": { "Product": { "tag_key": { "@@assign": "Product" }, "enforced_for": { "@@assign": [ "ec2:instance" ] } } } }`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_account(rName, organizations.PolicyTypeServiceControlPolicy, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", policyIdResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", targetIdResourceName, "master_account_id"),
				),
			},
			{
				Config: testAccPolicyAttachmentConfig_account(rName, organizations.PolicyTypeTagPolicy, tagPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", policyIdResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", targetIdResourceName, "master_account_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func testAccPolicyAttachment_OrganizationalUnit(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy_attachment.test"
	policyIdResourceName := "aws_organizations_policy.test"
	targetIdResourceName := "aws_organizations_organizational_unit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_organizationalUnit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", policyIdResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", targetIdResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func testAccPolicyAttachment_Root(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy_attachment.test"
	policyIdResourceName := "aws_organizations_policy.test"
	targetIdResourceName := "aws_organizations_organization.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_root(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", policyIdResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", targetIdResourceName, "roots.0.id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"skip_destroy"},
			},
		},
	})
}

func testAccPolicyAttachment_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy_attachment.test"
	policyIdResourceName := "aws_organizations_policy.test"
	targetIdResourceName := "aws_organizations_organization.test"

	serviceControlPolicyContent := `{"Version": "2012-10-17", "Statement": { "Effect": "Allow", "Action": "*", "Resource": "*"}}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentNoDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_skipDestroy(rName, organizations.PolicyTypeServiceControlPolicy, serviceControlPolicyContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_id", policyIdResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_id", targetIdResourceName, "master_account_id"),
					resource.TestCheckResourceAttr(resourceName, "skip_destroy", "true"),
				),
			},
		},
	})
}

func testAccPolicyAttachment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_organizations_policy_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, organizations.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyAttachmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyAttachmentConfig_organizationalUnit(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyAttachmentExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tforganizations.ResourcePolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyAttachmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_policy_attachment" {
				continue
			}

			targetID, policyID, err := tforganizations.DecodePolicyAttachmentID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tforganizations.FindPolicyAttachmentByTwoPartKey(ctx, conn, targetID, policyID)

			if tfresource.NotFound(err) {
				continue
			}

			if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Organizations Policy Attachment %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccCheckPolicyAttachmentNoDestroy is a variant of the CheckDestroy function to be used when
// skip_destroy is true and the attachment should still exist after destroy completes
func testAccCheckPolicyAttachmentNoDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_organizations_policy_attachment" {
				continue
			}

			targetID, policyID, err := tforganizations.DecodePolicyAttachmentID(rs.Primary.ID)
			if err != nil {
				return err
			}

			_, err = tforganizations.FindPolicyAttachmentByTwoPartKey(ctx, conn, targetID, policyID)
			if tfawserr.ErrCodeEquals(err, organizations.ErrCodeAWSOrganizationsNotInUseException) {
				// The organization was destroyed, so we can safely assume the policy attachment
				// skipped during destruction was as well
				continue
			}
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccCheckPolicyAttachmentExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Organizations Policy Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OrganizationsConn(ctx)

		targetID, policyID, err := tforganizations.DecodePolicyAttachmentID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tforganizations.FindPolicyAttachmentByTwoPartKey(ctx, conn, targetID, policyID)

		return err
	}
}

func testAccPolicyAttachmentConfig_account(rName, policyType, policyContent string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY", "TAG_POLICY"]
}

resource "aws_organizations_policy" "test" {
  depends_on = [aws_organizations_organization.test]

  name    = "%s"
  type    = "%s"
  content = %s
}

resource "aws_organizations_policy_attachment" "test" {
  policy_id = aws_organizations_policy.test.id
  target_id = aws_organizations_organization.test.master_account_id
}
`, rName, policyType, strconv.Quote(policyContent))
}

func testAccPolicyAttachmentConfig_organizationalUnit(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY"]
}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}

resource "aws_organizations_policy" "test" {
  depends_on = [aws_organizations_organization.test]

  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = %[1]q
}

resource "aws_organizations_policy_attachment" "test" {
  policy_id = aws_organizations_policy.test.id
  target_id = aws_organizations_organizational_unit.test.id
}
`, rName)
}

func testAccPolicyAttachmentConfig_root(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY"]
}

resource "aws_organizations_policy" "test" {
  depends_on = [aws_organizations_organization.test]

  content = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF

  name = %[1]q
}

resource "aws_organizations_policy_attachment" "test" {
  policy_id = aws_organizations_policy.test.id
  target_id = aws_organizations_organization.test.roots[0].id
}
`, rName)
}

func testAccPolicyAttachmentConfig_skipDestroy(rName, policyType, policyContent string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  enabled_policy_types = ["SERVICE_CONTROL_POLICY", "TAG_POLICY"]
}

resource "aws_organizations_policy" "test" {
  depends_on = [aws_organizations_organization.test]

  name    = "%s"
  type    = "%s"
  content = %s

  skip_destroy = true
}

resource "aws_organizations_policy_attachment" "test" {
  policy_id = aws_organizations_policy.test.id
  target_id = aws_organizations_organization.test.master_account_id

  skip_destroy = true
}
`, rName, policyType, strconv.Quote(policyContent))
}
