// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfaccountaccess "github.com/hashicorp/terraform-provider-aws/internal/service/accountaccess"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountAccessApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var v accountaccess.GetApplicationOutput
	resourceName := "aws_accountaccess_application.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, "identity_center_application_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
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

func testAccAccountAccessApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var v accountaccess.GetApplicationOutput
	resourceName := "aws_accountaccess_application.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfaccountaccess.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckApplicationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_accountaccess_application" {
				continue
			}

			_, err := tfaccountaccess.FindApplicationByARN(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Account Access Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckApplicationExists(ctx context.Context, t *testing.T, n string, v *accountaccess.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return errors.New("No Account Access Application ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

		output, err := tfaccountaccess.FindApplicationByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPrerequisitesConfig(rName), `
resource "aws_accountaccess_application" "test" {
  identity_center_instance_arn = local.instance_arn
}
`)
}
