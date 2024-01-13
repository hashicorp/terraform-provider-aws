// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/qbusiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqbusiness "github.com/hashicorp/terraform-provider-aws/internal/service/qbusiness"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccQBusinessIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var index qbusiness.GetIndexOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIndex(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, qbusiness.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Index name"),
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

func TestAccQBusinessIndex_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var index qbusiness.GetIndexOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIndex(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, qbusiness.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceIndex(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessIndex_documentAttributeConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	var index qbusiness.GetIndexOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIndex(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, qbusiness.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Index name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIndexConfig_documentAttributeConfigurations(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "document_attribute_configurations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "document_attribute_configurations.0.attribute.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "document_attribute_configurations.0.attribute.0.name", "foo1"),
					resource.TestCheckResourceAttr(resourceName, "document_attribute_configurations.0.attribute.1.name", "foo2"),
				),
			},
		},
	})
}

func testAccPreCheckIndex(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessConn(ctx)

	input := &qbusiness.ListApplicationsInput{}

	_, err := conn.ListApplicationsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckIndexDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_index" {
				continue
			}

			_, err := tfqbusiness.FindIndexByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amazon Q Index %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIndexExists(ctx context.Context, n string, v *qbusiness.GetIndexOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessConn(ctx)

		output, err := tfqbusiness.FindIndexByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccIndexConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Action": "sts:AssumeRole",
	"Principal": {
		"Service": "qbusiness.${data.aws_partition.current.dns_suffix}"
	},
	"Effect": "Allow",
	"Sid": ""
	}
	]
}
EOF
}

resource "aws_qbusiness_index" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  display_name         = %[1]q
  capacity_configuration {
    units = 1
  }
  description          = "Index name"
}
`, rName)
}

func testAccIndexConfig_documentAttributeConfigurations(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Action": "sts:AssumeRole",
	"Principal": {
		"Service": "qbusiness.${data.aws_partition.current.dns_suffix}"
	},
	"Effect": "Allow",
	"Sid": ""
	}
	]
}
EOF
}

resource "aws_qbusiness_index" "test" {
  application_id       = aws_qbusiness_app.test.application_id
  display_name         = %[1]q
  capacity_configuration {
    units = 1
  }
  description          = "Index name"
  document_attribute_configurations {
    attribute {
        name   = "foo1"
        search = "ENABLED"
        type   = "STRING"
	}
	attribute {
        name   = "foo2"
        search = "ENABLED"
        type   = "STRING"
	}
  }
}
`, rName)
}
