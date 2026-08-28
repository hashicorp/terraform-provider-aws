// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package accountaccess_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/accountaccess"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfaccountaccess "github.com/hashicorp/terraform-provider-aws/internal/service/accountaccess"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	checkApplicationARN = tfknownvalue.RegionalARNRegexp("account-access", regexache.MustCompile(`application/[a-zA-Z0-9-]+`))
)

func testAccAccountAccessApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v accountaccess.GetApplicationOutput
	resourceName := "aws_accountaccess_application.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Application/basic/"),
				ConfigVariables: config.Variables{},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkApplicationARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_application_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("identity_center_instance_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory:   config.StaticDirectory("testdata/Application/basic/"),
				ConfigVariables:   config.Variables{},
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

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AccountAccessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Application/basic/"),
				ConfigVariables: config.Variables{},
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

			_, err := tfaccountaccess.FindApplicationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("Account Access Application %s still exists", rs.Primary.Attributes[names.AttrARN])
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

		conn := acctest.ProviderMeta(ctx, t).AccountAccessClient(ctx)

		output, err := tfaccountaccess.FindApplicationByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}
