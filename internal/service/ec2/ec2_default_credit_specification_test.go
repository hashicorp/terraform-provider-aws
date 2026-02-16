// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2DefaultCreditSpecification_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var defaultcreditspecification awstypes.InstanceFamilyCreditSpecification
	resourceName := "aws_ec2_default_credit_specification.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultCreditSpecificationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDefaultCreditSpecificationExists(ctx, resourceName, &defaultcreditspecification),
					resource.TestCheckResourceAttr(resourceName, "cpu_credits", "standard"),
					resource.TestCheckResourceAttr(resourceName, "instance_family", "t4g"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccDefaultCreditSpecificationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "instance_family",
			},
		},
	})
}

func testAccCheckDefaultCreditSpecificationExists(ctx context.Context, n string, v *awstypes.InstanceFamilyCreditSpecification) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindDefaultCreditSpecificationByInstanceFamily(ctx, conn, awstypes.UnlimitedSupportedInstanceFamily(rs.Primary.Attributes["instance_family"]))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDefaultCreditSpecificationImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return rs.Primary.Attributes["instance_family"], nil
	}
}

func testAccDefaultCreditSpecificationConfig_basic() string {
	return `
resource "aws_ec2_default_credit_specification" "test" {
  cpu_credits     = "standard"
  instance_family = "t4g"
}
`
}
