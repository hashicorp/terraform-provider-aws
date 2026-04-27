// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// add sweeper to delete known test servicecat service actions

func TestAccServiceCatalogServiceAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_service_action.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceActionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"accept_language",
				},
			},
		},
	})
}

func TestAccServiceCatalogServiceAction_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_service_action.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceActionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfservicecatalog.ResourceServiceAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogServiceAction_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_service_action.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceActionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccServiceActionConfig_update(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.assume_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			}},
	})
}

func testAccCheckServiceActionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_service_action" {
				continue
			}

			input := &servicecatalog.DescribeServiceActionInput{
				Id: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeServiceAction(ctx, input)

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error getting Service Catalog Service Action (%s): %w", rs.Primary.ID, err)
			}

			if output != nil {
				return fmt.Errorf("Service Catalog Service Action (%s) still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckServiceActionExists(ctx context.Context, t *testing.T, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.ProviderMeta(ctx, t).ServiceCatalogClient(ctx)

		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeServiceAction(ctx, input)

		if err != nil {
			return fmt.Errorf("error describing Service Catalog Service Action (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccServiceActionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_servicecatalog_service_action" "test" {
  accept_language = "en"
  description     = %[1]q
  name            = %[1]q

  definition {
    name    = "AWS-RestartEC2Instance"
    version = "1"
  }
}
`, rName)
}

func testAccServiceActionConfig_update(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    principals {
      type = "Service"

      identifiers = [
        "servicecatalog.${data.aws_region.current.region}.${data.aws_partition.current.dns_suffix}",
      ]
    }

    actions = [
      "sts:AssumeRole",
    ]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_servicecatalog_service_action" "test" {
  description = %[1]q
  name        = %[1]q

  definition {
    assume_role = aws_iam_role.test.arn
    name        = "AWS-RestartEC2Instance"
    version     = "1"
  }
}
`, rName)
}
