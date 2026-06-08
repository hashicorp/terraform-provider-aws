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

func TestAccResilienceHubV2UserJourney_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var uj awstypes.UserJourney
	rName := acctest.RandomWithPrefix(t, "tf-acc-test")
	resourceName := "aws_resiliencehubv2_user_journey.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ResilienceHubV2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserJourneyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserJourneyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserJourneyExists(ctx, t, resourceName, &uj),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "user_journey_id"),
					resource.TestCheckResourceAttrSet(resourceName, "system_arn"),
				),
			},
		},
	})
}

func testAccCheckUserJourneyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_resiliencehubv2_user_journey" {
				continue
			}

			_, err := tfresiliencehubv2.FindUserJourneyByID(ctx, conn, rs.Primary.Attributes["system_arn"], rs.Primary.Attributes["user_journey_id"])
			if err == nil {
				return fmt.Errorf("Resilience Hub V2 User Journey %s still exists", rs.Primary.Attributes[names.AttrID])
			}
		}

		return nil
	}
}

func testAccCheckUserJourneyExists(ctx context.Context, t *testing.T, n string, v *awstypes.UserJourney) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("User Journey not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ResilienceHubV2Client(ctx)

		output, err := tfresiliencehubv2.FindUserJourneyByID(ctx, conn, rs.Primary.Attributes["system_arn"], rs.Primary.Attributes["user_journey_id"])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccUserJourneyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_resiliencehubv2_system" "test" {
  name = "%[1]s-system"
}

resource "aws_resiliencehubv2_user_journey" "test" {
  name       = %[1]q
  system_arn = aws_resiliencehubv2_system.test.arn
}
`, rName)
}
