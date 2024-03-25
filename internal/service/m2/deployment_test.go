// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	"github.com/aws/aws-sdk-go-v2/service/m2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfm2 "github.com/hashicorp/terraform-provider-aws/internal/service/m2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccM2Deployment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", "1"),
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

func TestAccM2Deployment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfm2.ResourceDeployment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccM2Deployment_nostart(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", "1"),
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

func TestAccM2Deployment_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var deployment m2.GetDeploymentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_m2_deployment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.M2EndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.M2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 1, 1, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", "1"),
				),
			},
			{
				Config: testAccDeploymentConfig_basic(rName, "bluage", 2, 2, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentExists(ctx, resourceName, &deployment),
					resource.TestCheckResourceAttr(resourceName, "application_version", "2"),
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

func testAccCheckDeploymentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_m2_deployment" {
				continue
			}

			applicationId, deploymentId, err := tfm2.DeploymentParseResourceId(rs.Primary.ID)
			if err != nil {
				return create.Error(names.M2, create.ErrActionCheckingDestroyed, tfm2.ResNameDeployment, rs.Primary.ID, err)
			}

			_, err = conn.GetDeployment(ctx, &m2.GetDeploymentInput{
				ApplicationId: aws.String(applicationId),
				DeploymentId:  aws.String(deploymentId),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.M2, create.ErrActionCheckingDestroyed, tfm2.ResNameDeployment, rs.Primary.ID, err)
			}

			return create.Error(names.M2, create.ErrActionCheckingDestroyed, tfm2.ResNameDeployment, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDeploymentExists(ctx context.Context, name string, deployment *m2.GetDeploymentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.M2, create.ErrActionCheckingExistence, tfm2.ResNameDeployment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.M2, create.ErrActionCheckingExistence, tfm2.ResNameDeployment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).M2Client(ctx)

		applicationId, deploymentId, err := tfm2.DeploymentParseResourceId(rs.Primary.ID)
		if err != nil {
			return create.Error(names.M2, create.ErrActionCheckingDestroyed, tfm2.ResNameDeployment, rs.Primary.ID, err)
		}

		resp, err := conn.GetDeployment(ctx, &m2.GetDeploymentInput{
			ApplicationId: aws.String(applicationId),
			DeploymentId:  aws.String(deploymentId),
		})

		if err != nil {
			return create.Error(names.M2, create.ErrActionCheckingExistence, tfm2.ResNameDeployment, rs.Primary.ID, err)
		}

		*deployment = *resp

		return nil
	}
}

func testAccDeploymentConfig_basic(rName, engineType string, appVersion, deployVersion int, start string) string {
	return acctest.ConfigCompose(testAccEnvironmentConfig_basic(rName, engineType),
		testAccApplicationConfig_versioned(rName, engineType, appVersion, 2),
		testAccDeploymentConfig_secretsManagerEndpoint(rName),
		fmt.Sprintf(`
resource "aws_m2_deployment" "test" {
  environment_id      = aws_m2_environment.test.id
  application_id      = aws_m2_application.test.id
  application_version = %[2]d
  start               = %[3]q
  depends_on          = [aws_vpc_endpoint.secretsmanager]
}
`, rName, deployVersion, start))
}

func testAccDeploymentConfig_secretsManagerEndpoint(rName string) string {
	return fmt.Sprintf(`

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "secretsmanager" {
  vpc_id            = aws_vpc.test.id
  service_name      = "com.amazonaws.${data.aws_region.current.name}.secretsmanager"
  vpc_endpoint_type = "Interface"

  security_group_ids = [
    aws_security_group.test.id,
  ]
  subnet_ids = aws_subnet.test[*].id

  private_dns_enabled = true

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
