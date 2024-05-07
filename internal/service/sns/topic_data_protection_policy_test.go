// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sns_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsns "github.com/hashicorp/terraform-provider-aws/internal/service/sns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSNSTopicDataProtectionPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_data_protection_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicDataProtectionPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, "aws_sns_topic.test", &attributes),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, "aws_sns_topic.test", names.AttrARN),
					resource.TestMatchResourceAttr(resourceName, names.AttrPolicy, regexache.MustCompile(fmt.Sprintf("\"Sid\":\"%[1]s\"", rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSNSTopicDataProtectionPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var attributes map[string]string
	resourceName := "aws_sns_topic_data_protection_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SNSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTopicDataProtectionPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTopicDataProtectionPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTopicExists(ctx, "aws_sns_topic.test", &attributes),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsns.ResourceTopicDataProtectionPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTopicDataProtectionPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SNSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sns_topic_data_protection_policy" {
				continue
			}

			_, err := tfsns.FindTopicAttributesByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SNS Data Protection Topic Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTopicDataProtectionPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_sns_topic_data_protection_policy" "test" {
  arn = aws_sns_topic.test.arn
  policy = jsonencode(
    {
      "Description" = "Default data protection policy"
      "Name"        = "__default_data_protection_policy"
      "Statement" = [
        {
          "DataDirection" = "Inbound"
          "DataIdentifier" = [
            "arn:${data.aws_partition.current.partition}:dataprotection::aws:data-identifier/EmailAddress",
          ]
          "Operation" = {
            "Deny" = {}
          }
          "Principal" = [
            "*",
          ]
          "Sid" = %[1]q
        },
      ]
      "Version" = "2021-06-01"
    }
  )
}
`, rName)
}
