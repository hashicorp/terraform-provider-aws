// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqbusiness "github.com/hashicorp/terraform-provider-aws/internal/service/qbusiness"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQBusinessIndex_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var index qbusiness.GetIndexOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIndex(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					resource.TestCheckResourceAttrSet(resourceName, "application_id"),
					resource.TestCheckResourceAttrSet(resourceName, "index_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Index name"),
					resource.TestCheckResourceAttr(resourceName, "capacity_configuration.0.units", "1"),
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
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceIndex, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessIndex_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var index qbusiness.GetIndexOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIndex(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_tags(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIndexConfig_tags(rName, "key1", "value1new", "key2", "value2new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1new"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2new"),
				),
			},
		},
	})
}

func TestAccQBusinessIndex_documentAttributeConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	var index qbusiness.GetIndexOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_index.test"
	attr1 := "foo1"
	attr2 := "foo2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckIndex(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexConfig_documentAttributeConfigurations(rName, attr1, attr2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
					resource.TestCheckResourceAttr(resourceName, "document_attribute_configuration.#", "2"),
				),
			},
			{
				Config: testAccIndexConfig_documentAttributeConfigurations(rName, attr2, attr1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIndexExists(ctx, resourceName, &index),
				),
				ExpectNonEmptyPlan: false,
				PlanOnly:           true,
			},
		},
	})
}

func testAccPreCheckIndex(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

	input := &qbusiness.ListApplicationsInput{}

	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckIndexDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

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

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

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
data "aws_ssoadmin_instances" "test" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }
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
  application_id = aws_qbusiness_app.test.id
  display_name   = %[1]q
  capacity_configuration {
    units = 1
  }
  description = "Index name"
}
`, rName)
}

func testAccIndexConfig_documentAttributeConfigurations(rName, attr1, attr2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_ssoadmin_instances" "test" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }
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
  application_id = aws_qbusiness_app.test.id
  display_name   = %[1]q
  capacity_configuration {
    units = 1
  }
  description = %[1]q
  document_attribute_configuration {
    name   = %[2]q
    search = "ENABLED"
    type   = "STRING"
  }
  document_attribute_configuration {
    name   = %[3]q
    search = "ENABLED"
    type   = "STRING"
  }
}
`, rName, attr1, attr2)
}

func testAccIndexConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_ssoadmin_instances" "test" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }
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
  application_id = aws_qbusiness_app.test.id
  display_name   = %[1]q
  capacity_configuration {
    units = 1
  }
  description = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
