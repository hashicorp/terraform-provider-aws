// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53resolver_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53resolver/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverDNSSECConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDNSSECConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSSECConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECConfigExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "route53resolver", regexache.MustCompile(`resolver-dnssec-config/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrResourceID),
					resource.TestCheckResourceAttr(resourceName, "validation_status", string(awstypes.ResolverDNSSECValidationStatusEnabled)),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDNSSECConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDNSSECConfigConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDNSSECConfigExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfroute53resolver.ResourceDNSSECConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDNSSECConfigDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).Route53ResolverClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_resolver_dnssec_config" {
				continue
			}

			_, err := tfroute53resolver.FindResolverDNSSECConfigByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func testAccCheckDNSSECConfigExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver DNSSEC Config ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).Route53ResolverClient(ctx)

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
