// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSClusterCapacityProviders_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              acctest.Ct1,
						names.AttrWeight:    "100",
						"capacity_provider": "FARGATE",
					}),
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

func TestAccECSClusterCapacityProviders_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecs.ResourceClusterCapacityProviders(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSClusterCapacityProviders_defaults(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_defaults(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrClusterName, rName),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", acctest.Ct0),
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

func TestAccECSClusterCapacityProviders_destroy(t *testing.T) {
	ctx := acctest.Context(t)

	// This test proves that https://github.com/hashicorp/terraform-provider-aws/issues/11409
	// has been addressed by aws_ecs_cluster_capacity_providers.
	//
	// If we were configuring capacity providers directly on the cluster, the
	// test would fail with a timeout error.
	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_destroyBefore(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					func(s *terraform.State) error {
						if got, want := int(cluster.RegisteredContainerInstancesCount), 2; got != want {
							return fmt.Errorf("RegisteredContainerInstancesCount = %v, want %v", got, want)
						}

						return nil
					},
				),
			},
			{
				Config: testAccClusterCapacityProvidersConfig_destroyAfter(rName),
			},
		},
	})
}

func TestAccECSClusterCapacityProviders_Update_capacityProviders(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_1(rName, "FARGATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_2(rName, "FARGATE", "FARGATE_SPOT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE_SPOT"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_0(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_1(rName, "FARGATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "capacity_providers.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "capacity_providers.*", "FARGATE"),
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

func TestAccECSClusterCapacityProviders_Update_defaultStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster awstypes.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_cluster_capacity_providers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCapacityProvidersConfig_defaultProviderStrategy1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              acctest.Ct1,
						names.AttrWeight:    "100",
						"capacity_provider": "FARGATE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_defaultProviderStrategy2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              acctest.Ct1,
						names.AttrWeight:    "50",
						"capacity_provider": "FARGATE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "",
						names.AttrWeight:    "50",
						"capacity_provider": "FARGATE_SPOT",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_defaultProviderStrategy3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              acctest.Ct2,
						names.AttrWeight:    "25",
						"capacity_provider": "FARGATE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "",
						names.AttrWeight:    "75",
						"capacity_provider": "FARGATE_SPOT",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClusterCapacityProvidersConfig_defaultProviderStrategy4(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_ecs_cluster.test", &cluster),
					resource.TestCheckResourceAttr(resourceName, "default_capacity_provider_strategy.#", acctest.Ct0),
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

func testAccClusterCapacityProvidersConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_defaults(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_0(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = []
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_1(rName, provider1 string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = [%[2]q]
}
`, rName, provider1)
}

func testAccClusterCapacityProvidersConfig_2(rName, provider1, provider2 string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = [%[2]q, %[3]q]
}
`, rName, provider1, provider2)
}

func testAccClusterCapacityProvidersConfig_defaultProviderStrategy1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_defaultProviderStrategy2(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 50
    capacity_provider = "FARGATE"
  }

  default_capacity_provider_strategy {
    weight            = 50
    capacity_provider = "FARGATE_SPOT"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_defaultProviderStrategy3(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 2
    weight            = 25
    capacity_provider = "FARGATE"
  }

  default_capacity_provider_strategy {
    weight            = 75
    capacity_provider = "FARGATE_SPOT"
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_defaultProviderStrategy4(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name = aws_ecs_cluster.test.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_destroyBefore(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-ecs-hvm-2.0.*-x86_64-ebs"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name       = aws_ecs_cluster.test.name
  capacity_providers = [aws_ecs_capacity_provider.test.name]

  default_capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.test.name
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Principal = {
        Service = "ec2.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }
  })
}

data "aws_partition" "test" {}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.test.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
  role       = aws_iam_role.test.id
}

resource "aws_iam_instance_profile" "test" {
  name       = %[1]q
  depends_on = [aws_iam_role_policy_attachment.test]
  role       = aws_iam_role.test.name
}

resource "aws_launch_template" "test" {
  name                                 = %[1]q
  image_id                             = data.aws_ami.test.id
  instance_type                        = "t3.micro"
  instance_initiated_shutdown_behavior = "terminate"
  vpc_security_group_ids               = [aws_security_group.test.id]

  iam_instance_profile {
    name = aws_iam_instance_profile.test.name
  }

  user_data = base64encode(<<EOL
#!/bin/bash
echo "ECS_CLUSTER=${aws_ecs_cluster.test.name}" >> /etc/ecs/ecs.config
EOL
  )
}

resource "aws_autoscaling_group" "test" {
  desired_capacity    = 2
  max_size            = 4
  min_size            = 2
  name                = %[1]q
  vpc_zone_identifier = [aws_subnet.test.id]

  wait_for_capacity_timeout = "5m"

  instance_refresh {
    strategy = "Rolling"
  }

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }

  tag {
    key                 = "AmazonECSManaged"
    value               = ""
    propagate_at_launch = true
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName)
}

func testAccClusterCapacityProvidersConfig_destroyAfter(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
