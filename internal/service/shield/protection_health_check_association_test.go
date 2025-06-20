// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfshield "github.com/hashicorp/terraform-provider-aws/internal/service/shield"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccShieldProtectionHealthCheckAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_protection_health_check_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionHealthCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionHealthCheckAssociationConfig_protectionaHealthCheckAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionHealthCheckAssociationExists(ctx, resourceName),
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

func TestAccShieldProtectionHealthCheckAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_shield_protection_health_check_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ShieldEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ShieldServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProtectionHealthCheckAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProtectionHealthCheckAssociationConfig_protectionaHealthCheckAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProtectionHealthCheckAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfshield.ResourceProtectionHealthCheckAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckProtectionHealthCheckAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_shield_protection_health_check_association" {
				continue
			}

			protectionId, _, err := tfshield.ProtectionHealthCheckAssociationParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			input := &shield.DescribeProtectionInput{
				ProtectionId: aws.String(protectionId),
			}

			resp, err := conn.DescribeProtection(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			if resp != nil && resp.Protection != nil && len(resp.Protection.HealthCheckIds) == 0 {
				return fmt.Errorf("The Shield protection HealthCheck with IDs %v still exists", resp.Protection.HealthCheckIds)
			}
		}

		return nil
	}
}

func testAccCheckProtectionHealthCheckAssociationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Shield Protection and Route53 Health Check Association ID is set")
		}

		protectionId, _, err := tfshield.ProtectionHealthCheckAssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ShieldClient(ctx)

		input := &shield.DescribeProtectionInput{
			ProtectionId: aws.String(protectionId),
		}

		resp, err := conn.DescribeProtection(ctx, input)

		if err != nil {
			return err
		}

		if resp == nil || resp.Protection == nil {
			return fmt.Errorf("The Shield protection does not exist")
		}

		if resp.Protection.HealthCheckIds == nil || len(resp.Protection.HealthCheckIds) != 1 {
			return fmt.Errorf("The Shield protection HealthCheck does not exist")
		}

		return nil
	}
}

func testAccProtectionHealthCheckAssociationConfig_protectionaHealthCheckAssociation(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    foo  = "bar"
    Name = %[1]q
  }
}
resource "aws_shield_protection" "test" {
  name         = %[1]q
  resource_arn = "arn:${data.aws_partition.current.partition}:ec2:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:eip-allocation/${aws_eip.test.id}"
}
resource "aws_route53_health_check" "test" {
  fqdn              = "example.com"
  port              = 80
  type              = "HTTP"
  resource_path     = "/"
  failure_threshold = "5"
  request_interval  = "30"
  tags = {
    Name = "tf-test-health-check"
  }
}
resource "aws_shield_protection_health_check_association" "test" {
  shield_protection_id = aws_shield_protection.test.id
  health_check_arn     = aws_route53_health_check.test.arn
}
`, rName)
}
