// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftAuthenticationProfile_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_authentication_profile.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthenticationProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthenticationProfileConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthenticationProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_profile_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAuthenticationProfileConfig_basic(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthenticationProfileExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_profile_name", rName),
				),
			},
		},
	})
}

func TestAccRedshiftAuthenticationProfile_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_authentication_profile.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAuthenticationProfileDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAuthenticationProfileConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAuthenticationProfileExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfredshift.ResourceAuthenticationProfile(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAuthenticationProfileDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_authentication_profile" {
				continue
			}

			_, err := tfredshift.FindAuthenticationProfileByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Authentication Profile %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAuthenticationProfileExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Authentication Profile ID is not set")
		}

		conn := acctest.ProviderMeta(ctx, t).RedshiftClient(ctx)

		_, err := tfredshift.FindAuthenticationProfileByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccAuthenticationProfileConfig_basic(rName, id string) string {
	return fmt.Sprintf(`
resource "aws_redshift_authentication_profile" "test" {
  authentication_profile_name = %[1]q
  authentication_profile_content = jsonencode(
    {
      AllowDBUserOverride = "1"
      Client_ID           = "ExampleClientID"
      App_ID              = %[2]q
    }
  )
}
`, rName, id)
}
