// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestOpenIDConnectProviderClientID_resourceName(t *testing.T) {
	t.Parallel()

	if got, want := tfiam.ResNameOpenIDConnectProviderClientID, "Open ID Connect Provider Client ID"; got != want {
		t.Errorf("ResNameOpenIDConnectProviderClientID = %q, want %q", got, want)
	}
}

func TestAccIAMOpenIDConnectProviderClientID_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_openid_connect_provider_client_id.test"
	providerResourceName := "aws_iam_openid_connect_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderClientIDDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderClientIDConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenIDConnectProviderClientIDExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrClientID, "sts.amazonaws.com"),
					resource.TestCheckResourceAttrPair(resourceName, "openid_connect_provider_arn", providerResourceName, names.AttrARN),
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

func TestAccIAMOpenIDConnectProviderClientID_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_openid_connect_provider_client_id.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckOpenIDConnectProviderClientIDDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccOpenIDConnectProviderClientIDConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOpenIDConnectProviderClientIDExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfiam.ResourceOpenIDConnectProviderClientID, resourceName),
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

func testAccCheckOpenIDConnectProviderClientIDDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_openid_connect_provider_client_id" {
				continue
			}

			providerARN := rs.Primary.Attributes["openid_connect_provider_arn"]
			clientID := rs.Primary.Attributes[names.AttrClientID]

			err := tfiam.FindOpenIDConnectProviderClientID(ctx, conn, providerARN, clientID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameOpenIDConnectProviderClientID, clientID, err)
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameOpenIDConnectProviderClientID, clientID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckOpenIDConnectProviderClientIDExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameOpenIDConnectProviderClientID, name, errors.New("not found"))
		}

		providerARN := rs.Primary.Attributes["openid_connect_provider_arn"]
		clientID := rs.Primary.Attributes[names.AttrClientID]

		if providerARN == "" || clientID == "" {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameOpenIDConnectProviderClientID, name, errors.New("openid_connect_provider_arn or client_id not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		return tfiam.FindOpenIDConnectProviderClientID(ctx, conn, providerARN, clientID)
	}
}

func testAccOpenIDConnectProviderClientIDConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_openid_connect_provider" "test" {
  url             = "https://accounts.testle.com/%[1]s"
  thumbprint_list = ["cf23df2207d99a74fbe169e3eba035e633b65d94"]
}

resource "aws_iam_openid_connect_provider_client_id" "test" {
  openid_connect_provider_arn = aws_iam_openid_connect_provider.test.arn
  client_id                   = "sts.amazonaws.com"
}
`, rName)
}
