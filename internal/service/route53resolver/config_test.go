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
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRoute53ResolverConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResolverConfig
	resourceName := "aws_route53_resolver_config.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConfig_basic(rName, "DISABLE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "autodefined_reverse_flag", string(awstypes.AutodefinedReverseFlagDisable)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, vpcResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigConfig_basic(rName, "ENABLE"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckConfigExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "autodefined_reverse_flag", string(awstypes.AutodefinedReverseFlagEnable)),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceID, vpcResourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccRoute53ResolverConfig_Disappears_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.ResolverConfig
	resourceName := "aws_route53_resolver_config.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigConfig_basic(rName, "ENABLE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfec2.ResourceVPC(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigExists(ctx context.Context, t *testing.T, n string, v *awstypes.ResolverConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route53 Resolver Config ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).Route53ResolverClient(ctx)

		output, err := tfroute53resolver.FindResolverConfigByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccConfigConfig_basic(rName, autodefinedReverseFlag string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 0), fmt.Sprintf(`
resource "aws_route53_resolver_config" "test" {
  autodefined_reverse_flag = %[1]q
  resource_id              = aws_vpc.test.id
}
`, autodefinedReverseFlag))
}
