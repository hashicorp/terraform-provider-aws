// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfresiliencehubv2 "github.com/hashicorp/terraform-provider-aws/internal/service/resiliencehubv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccResilienceHubV2Assertion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assertion awstypes.Assertion
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_assertion.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAssertionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAssertionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAssertionExists(ctx, t, resourceName, &assertion),
					resource.TestCheckResourceAttr(resourceName, "text", "The service must recover within 5 minutes"),
					resource.TestCheckResourceAttrSet(resourceName, "assertion_id"),
				),
			},
		},
	})
}

func testAccCheckAssertionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_assertion" {
				continue
			}

			_, err := tfresiliencehubv2.FindAssertionByID(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["assertion_id"])
			if err == nil {
				return fmt.Errorf("Resilience Hub V2 Assertion %s still exists", rs.Primary.Attributes[names.AttrID])
			}
		}

		return nil
	}
}

func testAccCheckAssertionExists(ctx context.Context, t *testing.T, n string, v *awstypes.Assertion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Assertion not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindAssertionByID(ctx, conn, rs.Primary.Attributes["service_arn"], rs.Primary.Attributes["assertion_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAssertionConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_resiliencehubv2_policy" "test" {
  name = "%[1]s-policy"

  availability_slo {
    target = 99.9
  }
}

resource "aws_resiliencehubv2_service" "test" {
  name    = "%[1]s-service"
  regions = [data.aws_region.current.name]

  policy_arn = aws_resiliencehubv2_policy.test.arn

  permission_model {
    invoker_role_name = "AWSResilienceHubAssessmentRole"
  }
}

resource "aws_resiliencehubv2_assertion" "test" {
  service_arn = aws_resiliencehubv2_service.test.arn
  text        = "The service must recover within 5 minutes"
}
`, rName)
}
