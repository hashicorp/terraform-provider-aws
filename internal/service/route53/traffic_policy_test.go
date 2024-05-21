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

func TestAccRoute53TrafficPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficPolicy
	resourceName := "aws_route53_traffic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTrafficPolicy(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "A"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccTrafficPolicyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccRoute53TrafficPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficPolicy
	resourceName := "aws_route53_traffic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTrafficPolicy(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfroute53.ResourceTrafficPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53TrafficPolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.TrafficPolicy
	resourceName := "aws_route53_traffic_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	commentUpdated := `comment updated`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTrafficPolicy(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTrafficPolicyDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.Route53ServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccTrafficPolicyConfig_complete(rName, names.AttrComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, names.AttrComment),
				),
			},
			{
				Config: testAccTrafficPolicyConfig_complete(rName, commentUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTrafficPolicyExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrComment, commentUpdated),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccTrafficPolicyImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckTrafficPolicyExists(ctx context.Context, n string, v *awstypes.TrafficPolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		output, err := tfroute53.FindTrafficPolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTrafficPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route53_traffic_policy" {
				continue
			}

			_, err := tfroute53.FindTrafficPolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route53 Traffic Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTrafficPolicyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes[names.AttrID], rs.Primary.Attributes[names.AttrVersion]), nil
	}
}

func testAccTrafficPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccTrafficPolicyConfig_complete(rName, comment string) string {
	return fmt.Sprintf(`
resource "aws_route53_traffic_policy" "test" {
  name     = %[1]q
  comment  = %[2]q
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
`, rName, comment)
}
