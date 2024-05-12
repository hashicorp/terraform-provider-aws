// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	"github.com/aws/aws-sdk-go-v2/service/qbusiness/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqbusiness "github.com/hashicorp/terraform-provider-aws/internal/service/qbusiness"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQBusinessApp_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.0.attachments_control_mode", string(types.AttachmentsControlModeEnabled)),
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

func TestAccQBusinessApp_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessApp_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_tags(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAppConfig_tags(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccQBusinessApp_attachmentsConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_app.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckApp(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAppDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAppConfig_attachmentsConfiguration(rName, string(types.AttachmentsControlModeEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.0.attachments_control_mode", string(types.AttachmentsControlModeEnabled)),
				),
			},
			{
				Config: testAccAppConfig_attachmentsConfiguration(rName, string(types.AttachmentsControlModeDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAppExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.0.attachments_control_mode", string(types.AttachmentsControlModeDisabled)),
				),
			},
		},
	})
}

func testAccPreCheckApp(ctx context.Context, t *testing.T) {
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

func testAccCheckAppDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_app" {
				continue
			}

			_, err := tfqbusiness.FindAppByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amazon Q App %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckAppExists(ctx context.Context, n string, v *qbusiness.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		output, err := tfqbusiness.FindAppByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccAppConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_ssoadmin_instances" "test" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  description          = %[1]q
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

`, rName)
}

func testAccAppConfig_attachmentsConfiguration(rName, mode string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_ssoadmin_instances" "test" {}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn

  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = %[2]q
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

`, rName, mode)
}

func testAccAppConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
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

`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
