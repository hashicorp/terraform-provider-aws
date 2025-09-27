// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransferWebApp_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var webappBefore, webappAfter awstypes.DescribedWebApp
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappBefore),
					resource.TestMatchResourceAttr(resourceName, "access_endpoint", regexache.MustCompile(`^https:\/\/.*.aws$`)),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebAppConfig_basic(rName+"-tag-changed", rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestMatchResourceAttr(resourceName, "access_endpoint", regexache.MustCompile(`^https:\/\/.*.aws$`)),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName+"-tag-changed"),
				),
			},
			{
				Config: testAccWebAppConfig_basic(rName+"-tag-changed", rName+"-tag-changed"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestMatchResourceAttr(resourceName, "access_endpoint", regexache.MustCompile(`^https:\/\/.*.aws$`)),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName+"-tag-changed"),
				),
			},
		},
	})
}

func TestAccTransferWebApp_webAppUnits(t *testing.T) {
	ctx := acctest.Context(t)

	var webappBefore, webappAfter awstypes.DescribedWebApp
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_webAppUnits(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappBefore),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_endpoint_policy", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebAppConfig_webAppUnits(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "2"),
					resource.TestCheckResourceAttr(resourceName, "web_app_endpoint_policy", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccWebAppConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "2"),
					resource.TestCheckResourceAttr(resourceName, "web_app_endpoint_policy", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccTransferWebApp_accessEndpoint(t *testing.T) {
	ctx := acctest.Context(t)

	var webappBefore, webappAfter awstypes.DescribedWebApp
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_accessEndPoint(rName, "https://example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappBefore),
					resource.TestCheckResourceAttr(resourceName, "access_endpoint", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_endpoint_policy", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebAppConfig_accessEndPoint(rName, "https://example2.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, "access_endpoint", "https://example2.com"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_endpoint_policy", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccWebAppConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, "access_endpoint", "https://example2.com"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider_details.0.identity_center_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttrPair(resourceName, "identity_provider_details.0.identity_center_config.0.role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_units.0.provisioned", "1"),
					resource.TestCheckResourceAttr(resourceName, "web_app_endpoint_policy", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccTransferWebApp_tags(t *testing.T) {
	ctx := acctest.Context(t)

	var webappBefore, webappAfter awstypes.DescribedWebApp
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappBefore),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccWebAppConfig_noTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWebAppConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccWebAppConfig_multipleTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Env", rName),
				),
			},
			{
				Config: testAccWebAppConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webappAfter),
					testAccCheckWebAppNotRecreated(&webappBefore, &webappAfter),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccTransferWebApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var webapp awstypes.DescribedWebApp
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_transfer_web_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.TransferEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TransferServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebAppConfig_basic(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebAppExists(ctx, resourceName, &webapp),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tftransfer.ResourceWebApp, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckWebAppDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_web_app" {
				continue
			}

			_, err := tftransfer.FindWebAppByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Transfer, create.ErrActionCheckingDestroyed, tftransfer.ResNameWebApp, rs.Primary.ID, err)
			}

			return create.Error(names.Transfer, create.ErrActionCheckingDestroyed, tftransfer.ResNameWebApp, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckWebAppExists(ctx context.Context, name string, webapp *awstypes.DescribedWebApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Transfer, create.ErrActionCheckingExistence, tftransfer.ResNameWebApp, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Transfer, create.ErrActionCheckingExistence, tftransfer.ResNameWebApp, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferClient(ctx)

		resp, err := tftransfer.FindWebAppByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Transfer, create.ErrActionCheckingExistence, tftransfer.ResNameWebApp, rs.Primary.ID, err)
		}

		*webapp = *resp

		return nil
	}
}

func testAccCheckWebAppNotRecreated(before, after *awstypes.DescribedWebApp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.WebAppId), aws.ToString(after.WebAppId); before != after {
			return create.Error(names.Transfer, create.ErrActionCheckingNotRecreated, tftransfer.ResNameWebApp, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccWebAppConfig_base(roleName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

data "aws_iam_policy_document" "assume_role_transfer" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
      "sts:SetContext"
    ]
    principals {
      type        = "Service"
      identifiers = ["transfer.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "aws:SourceAccount"
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.assume_role_transfer.json
}

data "aws_iam_policy_document" "web_app_identity_bearer" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetDataAccess",
      "s3:ListCallerAccessGrants",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:s3:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:access-grants/*"
    ]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "s3:ResourceAccount"
    }
  }
  statement {
    effect = "Allow"
    actions = [
      "s3:ListAccessGrantsInstances"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      values   = [data.aws_caller_identity.current.account_id]
      variable = "s3:ResourceAccount"
    }
  }
}

resource "aws_iam_role_policy" "web_app_identity_bearer" {
  policy = data.aws_iam_policy_document.web_app_identity_bearer.json
  role   = aws_iam_role.test.name
}
`, roleName)
}

func testAccWebAppConfig_basic(rName, roleName string) string {
	return acctest.ConfigCompose(
		testAccWebAppConfig_base(roleName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccWebAppConfig_webAppUnits(rName string, webAppUnitsProvisioned int) string {
	return acctest.ConfigCompose(
		testAccWebAppConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
  web_app_units {
    provisioned = %[2]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, webAppUnitsProvisioned))
}

func testAccWebAppConfig_accessEndPoint(rName, accessEndPoint string) string {
	return acctest.ConfigCompose(
		testAccWebAppConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
  access_endpoint = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, accessEndPoint))
}

func testAccWebAppConfig_noTags(rName string) string {
	return acctest.ConfigCompose(
		testAccWebAppConfig_base(rName), `
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }
}
`)
}

func testAccWebAppConfig_multipleTags(rName string) string {
	return acctest.ConfigCompose(
		testAccWebAppConfig_base(rName), fmt.Sprintf(`
resource "aws_transfer_web_app" "test" {
  identity_provider_details {
    identity_center_config {
      instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
      role         = aws_iam_role.test.arn
    }
  }

  tags = {
    Name = %[1]q
    Env  = %[1]q
  }

}
`, rName))
}
