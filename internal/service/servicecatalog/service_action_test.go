// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfservicecatalog "github.com/hashicorp/terraform-provider-aws/internal/service/servicecatalog"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// add sweeper to delete known test servicecat service actions

func TestAccServiceCatalogServiceAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_service_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfservicecatalog.ResourceServiceAction(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceCatalogServiceAction_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_service_action.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceActionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttr(resourceName, "definition.0.name", "AWS-RestartEC2Instance"),
					resource.TestCheckResourceAttr(resourceName, "definition.0.version", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				Config: testAccServiceActionConfig_update(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceActionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "accept_language", tfservicecatalog.AcceptLanguageEnglish),
					resource.TestCheckResourceAttrPair(resourceName, "definition.0.assume_role", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
				),
			}},
	})
}

func testAccCheckServiceActionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_servicecatalog_service_action" {
				continue
			}

			input := &servicecatalog.DescribeServiceActionInput{
				Id: aws.String(rs.Primary.ID),
			}

			output, err := conn.DescribeServiceActionWithContext(ctx, input)

			if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
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

func testAccCheckServiceActionExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceCatalogConn(ctx)

		input := &servicecatalog.DescribeServiceActionInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeServiceActionWithContext(ctx, input)

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
        "servicecatalog.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}",
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
