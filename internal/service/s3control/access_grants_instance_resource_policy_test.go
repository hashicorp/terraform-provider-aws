// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3control "github.com/hashicorp/terraform-provider-aws/internal/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccessGrantsInstanceResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_grants_instance_resource_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckAccessGrantsInstanceResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsInstanceResourcePolicyConfig_basic(`"s3:ListAccessGrants","s3:ListAccessGrantsLocations","s3:GetDataAccess"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccessGrantsInstanceResourcePolicyExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
		},
	})
}

func testAccAccessGrantsInstanceResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3control_access_grants_instance_resource_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAlternateAccount(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ControlServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckAccessGrantsInstanceResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessGrantsInstanceResourcePolicyConfig_basic(`"s3:ListAccessGrants","s3:ListAccessGrantsLocations","s3:GetDataAccess"`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessGrantsInstanceResourcePolicyExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfs3control.ResourceAccessGrantsInstanceResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAccessGrantsInstanceResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_s3control_access_grants_instance_resource_policy" {
				continue
			}

			_, err := tfs3control.FindAccessGrantsInstanceResourcePolicy(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("S3 Access Grants Instance Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAccessGrantsInstanceResourcePolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3ControlClient(ctx)

		_, err := tfs3control.FindAccessGrantsInstanceResourcePolicy(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccAccessGrantsInstanceResourcePolicyConfig_basic(action string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateAccountProvider(), fmt.Sprintf(`
data "aws_caller_identity" "target" {
  provider = "awsalternate"
}

resource "aws_s3control_access_grants_instance" "test" {}

resource "aws_s3control_access_grants_instance_resource_policy" "test" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "S3AccessGrantsPolicy",
  "Statement": [{
    "Sid": "AllowAccessToS3AccessGrants",
    "Effect": "Allow",
    "Principal": {
      "AWS": "${data.aws_caller_identity.target.account_id}"
    },
    "Action": [%[1]s],
    "Resource": "${aws_s3control_access_grants_instance.test.access_grants_instance_arn}"
  }]
}
EOF
}
`, action))
}
