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

func TestAccQBusinessApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckApplication(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QBusinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.0.attachments_control_mode", string(types.AttachmentsControlModeEnabled)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"identity_center_instance_arn"},
			},
		},
	})
}

func TestAccQBusinessApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckApplication(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QBusinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceApplication, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessApplication_update(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := names.AttrDescription
	descriptionUpdated := "description updated"
	resourceName := "aws_qbusiness_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckApplication(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QBusinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_update(rName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"identity_center_instance_arn"},
			},
			{
				Config: testAccApplicationConfig_update(rName, descriptionUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
				),
			},
		},
	})
}

func TestAccQBusinessApplication_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckApplication(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QBusinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"identity_center_instance_arn"},
			},
			{
				Config: testAccApplicationConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccApplicationConfig_tags1(rName, acctest.CtKey2, "value2updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, "value2updated"),
				),
			},
		},
	})
}

func TestAccQBusinessApplication_attachmentsConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var application qbusiness.GetApplicationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckApplication(ctx, t)
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QBusinessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationConfig_attachmentsConfiguration(rName, string(types.AttachmentsControlModeEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.0.attachments_control_mode", string(types.AttachmentsControlModeEnabled)),
				),
			},
			{
				Config: testAccApplicationConfig_attachmentsConfiguration(rName, string(types.AttachmentsControlModeDisabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationExists(ctx, resourceName, &application),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "attachments_configuration.0.attachments_control_mode", string(types.AttachmentsControlModeDisabled)),
				),
			},
		},
	})
}

func testAccPreCheckApplication(ctx context.Context, t *testing.T) {
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

func testAccCheckApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_application" {
				continue
			}

			_, err := tfqbusiness.FindApplicationByID(ctx, conn, rs.Primary.ID)

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

func testAccCheckApplicationExists(ctx context.Context, n string, v *qbusiness.GetApplicationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		output, err := tfqbusiness.FindApplicationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccApplicationConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_ssoadmin_instances" "test" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Principal = {
          Service = "qbusiness.${data.aws_partition.current.dns_suffix}"
        }
        Effect = "Allow"
        Sid    = ""
      }
    ]
  })
}
`, rName)
}

func testAccApplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_application" "test" {
  display_name                 = %[1]q
  description                  = %[1]q
  iam_service_role_arn         = aws_iam_role.test.arn
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }
}
`, rName))
}

func testAccApplicationConfig_update(rName, description string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_application" "test" {
  display_name                 = %[1]q
  description                  = %[2]q
  iam_service_role_arn         = aws_iam_role.test.arn
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }
}
`, rName, description))
}

func testAccApplicationConfig_attachmentsConfiguration(rName, mode string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_application" "test" {
  display_name                 = %[1]q
  iam_service_role_arn         = aws_iam_role.test.arn
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = %[2]q
  }
}
`, rName, mode))
}

func testAccApplicationConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_application" "test" {
  display_name                 = %[1]q
  iam_service_role_arn         = aws_iam_role.test.arn
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccApplicationConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccApplicationConfig_base(rName), fmt.Sprintf(`
resource "aws_qbusiness_application" "test" {
  display_name                 = %[1]q
  iam_service_role_arn         = aws_iam_role.test.arn
  identity_center_instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  attachments_configuration {
    attachments_control_mode = "ENABLED"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
