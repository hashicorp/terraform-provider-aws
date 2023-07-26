// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffinspace "github.com/hashicorp/terraform-provider-aws/internal/service/finspace"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFinSpaceKxCluster_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxClusterStatusRunning)),
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

func TestAccFinSpaceKxCluster_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffinspace.ResourceKxCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFinSpaceKxCluster_description(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_description(rName, "cluster description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "description", "cluster description"),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_database(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_database(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxClusterStatusRunning)),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_cacheConfigurations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_cacheConfigurations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxClusterStatusRunning)),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_code(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"
	codePath := "test-fixtures/code.zip"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_code(rName, codePath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_multiAZ(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_multiAZ(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxClusterStatusRunning)),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_rdb(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_rdb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxClusterStatusRunning)),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_executionRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_executionRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxClusterStatusRunning)),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_autoScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_autoScaling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "status", string(types.KxClusterStatusRunning)),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_initializationScript(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"
	// Need to set these to the bucket/key you want to use
	codePath := "test-fixtures/code.zip"
	initScriptPath := "code/helloworld.q"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_initScript(rName, codePath, initScriptPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_commandLineArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_commandLineArgs1(rName, "arg1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "command_line_arguments.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "command_line_arguments.arg1", "value1"),
				),
			},
		},
	})
}

func TestAccFinSpaceKxCluster_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxcluster finspace.GetKxClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_finspace_kx_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccKxClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccKxClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKxClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxClusterExists(ctx, resourceName, &kxcluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckKxClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_finspace_kx_cluster" {
				continue
			}

			input := &finspace.GetKxClusterInput{
				ClusterName:   aws.String(rs.Primary.Attributes["name"]),
				EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
			}
			_, err := conn.GetKxCluster(ctx, input)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.FinSpace, create.ErrActionCheckingDestroyed, tffinspace.ResNameKxCluster, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckKxClusterExists(ctx context.Context, name string, kxcluster *finspace.GetKxClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxCluster, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxCluster, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient(ctx)
		resp, err := conn.GetKxCluster(ctx, &finspace.GetKxClusterInput{
			ClusterName:   aws.String(rs.Primary.Attributes["name"]),
			EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
		})

		if err != nil {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxCluster, rs.Primary.ID, err)
		}

		*kxcluster = *resp

		return nil
	}
}

func testAccKxClusterConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

output "account_id" {
  value = data.aws_caller_identity.current.account_id
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}

data "aws_iam_policy_document" "key_policy" {
  statement {
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    resources = [
      aws_kms_key.test.arn,
    ]

    principals {
      type        = "Service"
      identifiers = ["finspace.amazonaws.com"]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["${aws_finspace_kx_environment.test.arn}/*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    actions = [
      "kms:*",
    ]

    resources = [
      "*",
    ]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}

resource "aws_kms_key_policy" "test" {
  key_id = aws_kms_key.test.id
  policy = data.aws_iam_policy_document.key_policy.json
}

resource "aws_vpc" "test" {
  cidr_block           = "172.31.0.0/16"
  enable_dns_hostnames = true
}

resource "aws_subnet" "test" {
  vpc_id               = aws_vpc.test.id
  cidr_block           = "172.31.32.0/20"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

data "aws_route_tables" "rts" {
  vpc_id = aws_vpc.test.id
}

resource "aws_route" "r" {
  route_table_id         = tolist(data.aws_route_tables.rts.ids)[0]
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
}
`, rName)
}

func testAccKxClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName))
}

func testAccKxClusterConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  description          = %[2]q
  environment_id       = aws_finspace_kx_environment.test.id
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  type                 = "HDB"
  release_label        = "1.0"
  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName, description))
}

func testAccKxClusterConfig_commandLineArgs1(rName, arg1, val1 string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  type                 = "HDB"
  release_label        = "1.0"
  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }

  command_line_arguments = {
    %[2]q = %[3]q
  }
}
`, rName, arg1, val1))
}

func testAccKxClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  release_label        = "1.0"
  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccKxClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  release_label        = "1.0"
  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccKxClusterConfig_database(rName string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
}

resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]

  database {
    database_name = aws_finspace_kx_database.test.name
  }

  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName))
}

func testAccKxClusterConfig_cacheConfigurations(rName string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
}

resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]

  cache_storage_configurations {
    type = "CACHE_1000"
    size = 1200
  }

  database {
    database_name = aws_finspace_kx_database.test.name
    cache_configurations {
      cache_type = "CACHE_1000"
      db_paths   = ["/"]
    }
  }

  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName))
}

func testAccKxClusterConfig_code(rName, path string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_iam_policy_document" "bucket_policy" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:GetObjectTagging"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
    ]

    principals {
      type        = "Service"
      identifiers = ["finspace.amazonaws.com"]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["${aws_finspace_kx_environment.test.arn}/*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    actions = [
      "s3:ListBucket"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}",
    ]

    principals {
      type        = "Service"
      identifiers = ["finspace.amazonaws.com"]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["${aws_finspace_kx_environment.test.arn}/*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.bucket_policy.json
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.id
  key    = %[2]q
  source = %[2]q
}

resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }

  code {
    s3_bucket = aws_s3_bucket.test.id
    s3_key    = aws_s3_object.object.key
  }
}
`, rName, path))
}

func testAccKxClusterConfig_multiAZ(rName string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_subnet" "test2" {
  vpc_id               = aws_vpc.test.id
  cidr_block           = "172.31.16.0/20"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[1]
}

resource "aws_subnet" "test3" {
  vpc_id               = aws_vpc.test.id
  cidr_block           = "172.31.64.0/20"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[2]
}

resource "aws_finspace_kx_cluster" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
  type           = "HDB"
  release_label  = "1.0"
  az_mode        = "MULTI"
  capacity_configuration {
    node_count = 3
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id, aws_subnet.test2.id, aws_subnet.test3.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName))
}

func testAccKxClusterConfig_rdb(rName string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "RDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]

  savedown_storage_configuration {
    type = "SDS01"
    size = 500
  }

  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName))
}

func testAccKxClusterConfig_executionRole(rName string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %[1]q
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = ["finspace:ConnectKxCluster", "finspace:GetKxConnectionString"]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

resource "aws_iam_role" "test" {
  name                = %[1]q
  managed_policy_arns = [aws_iam_policy.test.arn]
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          "Service" : "prod.finspacekx.aws.internal",
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  execution_role       = aws_iam_role.test.arn

  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName))
}

func testAccKxClusterConfig_autoScaling(rName string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_finspace_kx_cluster" "test" {
  name                 = %[1]q
  environment_id       = aws_finspace_kx_environment.test.id
  type                 = "HDB"
  release_label        = "1.0"
  az_mode              = "SINGLE"
  availability_zone_id = aws_finspace_kx_environment.test.availability_zones[0]
  capacity_configuration {
    node_count = 3
    node_type  = "kx.s.xlarge"
  }

  auto_scaling_configuration {
    min_node_count             = 3
    max_node_count             = 5
    auto_scaling_metric        = "CPU_UTILIZATION_PERCENTAGE"
    metric_target              = 25.0
    scale_in_cooldown_seconds  = 30.0
    scale_out_cooldown_seconds = 30.0
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }
}
`, rName))
}

func testAccKxClusterConfig_initScript(rName, codePath, relPath string) string {
	return acctest.ConfigCompose(
		testAccKxClusterConfigBase(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:GetObjectTagging"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
    ]

    principals {
      type        = "Service"
      identifiers = ["finspace.amazonaws.com"]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["${aws_finspace_kx_environment.test.arn}/*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }

  statement {
    actions = [
      "s3:ListBucket"
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}",
    ]

    principals {
      type        = "Service"
      identifiers = ["finspace.amazonaws.com"]
    }

    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values   = ["${aws_finspace_kx_environment.test.arn}/*"]
    }

    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values   = [data.aws_caller_identity.current.account_id]
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket.test.id
  key    = %[2]q
  source = %[2]q
}

resource "aws_finspace_kx_database" "test" {
  name           = %[1]q
  environment_id = aws_finspace_kx_environment.test.id
}

resource "aws_finspace_kx_cluster" "test" {
  name                  = %[1]q
  environment_id        = aws_finspace_kx_environment.test.id
  type                  = "HDB"
  release_label         = "1.0"
  az_mode               = "SINGLE"
  availability_zone_id  = aws_finspace_kx_environment.test.availability_zones[0]
  initialization_script = %[3]q
  capacity_configuration {
    node_count = 2
    node_type  = "kx.s.xlarge"
  }

  vpc_configuration {
    vpc_id             = aws_vpc.test.id
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = [aws_subnet.test.id]
    ip_address_type    = "IP_V4"
  }

  cache_storage_configurations {
    type = "CACHE_1000"
    size = 1200
  }

  database {
    database_name = aws_finspace_kx_database.test.name
    cache_configurations {
      cache_type = "CACHE_1000"
      db_paths   = ["/"]
    }
  }

  code {
    s3_bucket = aws_s3_bucket.test.id
    s3_key    = aws_s3_object.object.key
  }
}
`, rName, codePath, relPath))
}
