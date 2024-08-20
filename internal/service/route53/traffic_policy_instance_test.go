// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53 "github.com/hashicorp/terraform-provider-aws/internal/service/route53"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreCheckTrafficPolicy(t *testing.T) {
	acctest.PreCheckPartitionNot(t, names.USGovCloudPartitionID)
}

func TestAccRoute53TrafficPolicyInstance_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficPolicyInstance
	resourceName := "aws_route53_traffic_policy_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	zoneName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTrafficPolicy(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyInstanceDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyInstanceConfig_basic(rName, zoneName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, fmt.Sprintf("%s.%s", rName, zoneName)),
					resource.TestCheckResourceAttr(resourceName, "ttl", "3600"),
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

func TestAccRoute53TrafficPolicyInstance_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficPolicyInstance
	resourceName := "aws_route53_traffic_policy_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	zoneName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTrafficPolicy(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyInstanceDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyInstanceConfig_basic(rName, zoneName, 360),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceTrafficPolicyInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53TrafficPolicyInstance_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficPolicyInstance
	resourceName := "aws_route53_traffic_policy_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	zoneName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTrafficPolicy(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyInstanceDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyInstanceConfig_basic(rName, zoneName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ttl", "3600"),
				),
			},
			{
				Config: testAccTrafficPolicyInstanceConfig_basic(rName, zoneName, 7200),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyInstanceExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ttl", "7200"),
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

func testAccCheckTrafficPolicyInstanceExists(ctx context.Context, n string, v *awstypes.TrafficPolicyInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		output, err := tfroute53.FindTrafficPolicyInstanceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTrafficPolicyInstanceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_traffic_policy_instance" {
				continue
			}

			_, err := tfroute53.FindTrafficPolicyInstanceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Traffic Policy Instance %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrafficPolicyInstanceConfig_basic(rName, zoneName string, ttl int) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = %[2]q
}

resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  document = <<-EOT
{
    "AWSPolicyFormatVersion":"2015-10-01",
    "RecordType":"A",
    "Endpoints":{
        "endpoint-start-NkPh":{
            "Type":"value",
            "Value":"10.0.0.1"
        }
    },
    "StartEndpoint":"endpoint-start-NkPh"
}
EOT
}

resource "aws_route53_traffic_policy_instance" "test" {
  hosted_zone_id         = aws_route53_zone.test.zone_id
  name                   = "%[1]s.%[2]s"
  traffic_policy_id      = aws_route53_traffic_policy.test.id
  traffic_policy_version = aws_route53_traffic_policy.test.version
  ttl                    = %[3]d
}
`, rName, zoneName, ttl)
}
