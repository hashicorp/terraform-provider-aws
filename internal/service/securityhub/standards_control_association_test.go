// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package securityhub_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccStandardsControlAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var standardsControlAssociation awstypes.StandardsControlAssociationSummary
	resourceName := "aws_securityhub_standards_control_association.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SecurityHubServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccStandardsControlAssociationConfig_associationStatusEnabled(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStandardsControlAssociationExists(ctx, t, resourceName, &standardsControlAssociation),
					resource.TestCheckResourceAttr(resourceName, "association_status", string(awstypes.AssociationStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, "security_control_id", "IAM.1"),
					resource.TestCheckResourceAttrPair(resourceName, "standards_arn", "aws_securityhub_standards_subscription.test", "standards_arn"),
				),
			},
			{
				Config: testAccStandardsControlAssociationConfig_associationStatusDisabled(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStandardsControlAssociationExists(ctx, t, resourceName, &standardsControlAssociation),
					resource.TestCheckResourceAttr(resourceName, "association_status", string(awstypes.AssociationStatusDisabled)),
					resource.TestCheckResourceAttr(resourceName, "security_control_id", "IAM.1"),
					resource.TestCheckResourceAttrPair(resourceName, "standards_arn", "aws_securityhub_standards_subscription.test", "standards_arn"),
					resource.TestCheckResourceAttr(resourceName, "updated_reason", "test"),
				),
			},
		},
	})
}

func testAccCheckStandardsControlAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.StandardsControlAssociationSummary) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).SecurityHubClient(ctx)

		output, err := tfsecurityhub.FindStandardsControlAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["security_control_id"], rs.Primary.Attributes["standards_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

const testAccStandardsControlAssociation_base = `
data "aws_partition" "current" {}

resource "aws_securityhub_account" "test" {
  enable_default_standards = false
}

resource "aws_securityhub_standards_subscription" "test" {
  standards_arn = "arn:${data.aws_partition.current.partition}:securityhub:::ruleset/cis-aws-foundations-benchmark/v/1.2.0"
  depends_on    = [aws_securityhub_account.test]
}
`

func testAccStandardsControlAssociationConfig_associationStatusEnabled() string {
	return acctest.ConfigCompose(testAccStandardsControlAssociation_base, `
resource "aws_securityhub_standards_control_association" "test" {
  security_control_id = "IAM.1"
  standards_arn       = aws_securityhub_standards_subscription.test.standards_arn
  association_status  = "ENABLED"
}
`)
}

func testAccStandardsControlAssociationConfig_associationStatusDisabled() string {
	return acctest.ConfigCompose(testAccStandardsControlAssociation_base, `
resource "aws_securityhub_standards_control_association" "test" {
  security_control_id = "IAM.1"
  standards_arn       = aws_securityhub_standards_subscription.test.standards_arn
  association_status  = "DISABLED"
  updated_reason      = "test"
}
`)
}
