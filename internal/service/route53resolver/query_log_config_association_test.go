// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverQueryLogConfigAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResolverQueryLogConfigAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_query_log_config_association.test"
	queryLogConfigResourceName := "aws_route53_resolver_query_log_config.test"
	vpcResourceName := "aws_vpc.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryLogConfigAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLogConfigAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "resolver_query_log_config_id", queryLogConfigResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, vpcResourceName, names.AttrID),
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

func TestAccRoute53ResolverQueryLogConfigAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResolverQueryLogConfigAssociation
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_query_log_config_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueryLogConfigAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccQueryLogConfigAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryLogConfigAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53resolver.ResourceQueryLogConfigAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQueryLogConfigAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53ResolverClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_query_log_config_association" {
				continue
			}

			_, err := tfroute53resolver.FindResolverQueryLogConfigAssociationByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Resolver Query Log Config Association still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckQueryLogConfigAssociationExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResolverQueryLogConfigAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver Query Log Config Association ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).Route53ResolverClient(ctx)

		output, err := tfroute53resolver.FindResolverQueryLogConfigAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccQueryLogConfigAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route53_resolver_query_log_config" "test" {
  name            = %[1]q
  destination_arn = aws_cloudwatch_log_group.test.arn
}

resource "aws_route53_resolver_query_log_config_association" "test" {
  resolver_query_log_config_id = aws_route53_resolver_query_log_config.test.id
  resource_id                  = aws_vpc.test.id
}
`, rName)
}
