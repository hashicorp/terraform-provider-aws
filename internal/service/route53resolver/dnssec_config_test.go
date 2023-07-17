// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccRoute53ResolverDNSSECConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDNSSECConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSSECConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECConfigExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "route53resolver", regexp.MustCompile(`resolver-dnssec-config/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "validation_status", "ENABLED"),
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

func TestAccRoute53ResolverDNSSECConfig_disappear(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, route53resolver.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDNSSECConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSSECConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECConfigExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53resolver.ResourceDNSSECConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDNSSECConfigDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_dnssec_config" {
				continue
			}

			_, err := tfroute53resolver.FindResolverDNSSECConfigByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Resolver DNSSEC Config still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDNSSECConfigExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver DNSSEC Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn(ctx)

		_, err := tfroute53resolver.FindResolverDNSSECConfigByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccDNSSECConfigConfig_basic(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), `
resource "aws_route53_resolver_dnssec_config" "test" {
  resource_id = aws_vpc.test.id
}
`)
}
