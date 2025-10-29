// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	awsfis "github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffis "github.com/hashicorp/terraform-provider-aws/internal/service/fis"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Acceptance tests

func TestAccFISTargetAccountConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_target_account_configuration.test"
	var conf awstypes.TargetAccountConfiguration

	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FISEndpointID)
			testAccPreCheckTargetAccountConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAccountConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetAccountConfigurationConfig_basic(rName, "Initial description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttrPair(resourceName, "account_id", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
					resource.TestCheckResourceAttrPair(resourceName, "experiment_template_id", "aws_fis_experiment_template.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.target", names.AttrARN),
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

func TestAccFISTargetAccountConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_target_account_configuration.test"
	var conf awstypes.TargetAccountConfiguration

	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FISEndpointID)
			testAccPreCheckTargetAccountConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAccountConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetAccountConfigurationConfig_basic(rName, "Initial description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffis.ResourceTargetAccountConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFISTargetAccountConfiguration_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_target_account_configuration.test"
	var before, after awstypes.TargetAccountConfiguration

	if testing.Short() {
		t.Skip("skipping acceptance test in short mode")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.FISEndpointID)
			testAccPreCheckTargetAccountConfiguration(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.FISServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTargetAccountConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTargetAccountConfigurationConfig_basic(rName, "Initial description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
				),
			},
			{
				Config: testAccTargetAccountConfigurationConfig_basic(rName, "Updated description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTargetAccountConfigurationExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
		},
	})
}

func testAccCheckTargetAccountConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fis_target_account_configuration" {
				continue
			}

			_, err := tffis.FindTargetAccountConfigurationByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.FIS, create.ErrActionCheckingDestroyed, tffis.ResNameTargetAccountConfiguration, rs.Primary.ID, err)
			}

			return create.Error(names.FIS, create.ErrActionCheckingDestroyed, tffis.ResNameTargetAccountConfiguration, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTargetAccountConfigurationExists(ctx context.Context, name string, conf *awstypes.TargetAccountConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameTargetAccountConfiguration, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameTargetAccountConfiguration, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)

		out, err := tffis.FindTargetAccountConfigurationByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.FIS, create.ErrActionCheckingExistence, tffis.ResNameTargetAccountConfiguration, rs.Primary.ID, err)
		}

		*conf = *out

		return nil
	}
}

func testAccPreCheckTargetAccountConfiguration(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)

	input := &awsfis.ListTargetAccountConfigurationsInput{}

	_, err := conn.ListTargetAccountConfigurations(ctx, input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTargetAccountConfigurationConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "template" {
  name = "%[1]s-template"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role" "target" {
  name = "%[1]s-target"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = "Acceptance test template"
  role_arn    = aws_iam_role.template.arn

  stop_condition {
    source = "none"
  }

  experiment_options {
    account_targeting = "multi-account"
  }

  action {
    name      = "example"
    action_id = "aws:ec2:stop-instances"

    target {
      key   = "Instances"
      value = "to-stop"
    }
  }

  target {
    name           = "to-stop"
    resource_type  = "aws:ec2:instance"
    selection_mode = "ALL"

    resource_tag {
      key   = "env"
      value = "%[1]s"
    }
  }
}

resource "aws_fis_target_account_configuration" "test" {
  account_id             = data.aws_caller_identity.current.account_id
  description            = %[2]q
  experiment_template_id = aws_fis_experiment_template.test.id
  role_arn               = aws_iam_role.target.arn
}
`, rName, description)
}
