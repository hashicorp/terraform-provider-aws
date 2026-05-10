// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDaemonNameFromARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
		isNull   bool
	}{
		{
			name:     "valid daemon ARN",
			input:    "arn:aws:ecs:us-west-2:123456789012:daemon/my-cluster/my-daemon", // lintignore:AWSAT003,AWSAT005
			expected: "my-daemon",
		},
		{
			name:   "too few parts",
			input:  "arn:aws:ecs:us-west-2:123456789012:daemon/my-cluster", // lintignore:AWSAT003,AWSAT005
			isNull: true,
		},
		{
			name:   "too many parts",
			input:  "arn:aws:ecs:us-west-2:123456789012:daemon/a/b/c", // lintignore:AWSAT003,AWSAT005
			isNull: true,
		},
		{
			name:   "empty string",
			input:  "",
			isNull: true,
		},
		{
			name:   "no slashes",
			input:  "something",
			isNull: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := tfecs.DaemonNameFromARN(tc.input)
			if tc.isNull {
				if !got.IsNull() {
					t.Errorf("expected null, got %s", got.ValueString())
				}
			} else {
				if got.ValueString() != tc.expected {
					t.Errorf("got %s, expected %s", got.ValueString(), tc.expected)
				}
			}
		})
	}
}

func TestAccECSDaemon_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ecs", regexache.MustCompile(`daemon/.+/`+rName+`$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "cluster_arn", "ecs", regexache.MustCompile(`cluster/`+rName+`$`)),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "daemon_task_definition", "ecs", regexache.MustCompile(`daemon-task-definition/`+rName+`:\d+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_arns.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore: []string{
					"capacity_provider_arns",   // API doesn't return this
					"daemon_task_definition",   // API doesn't return this
					"deployment_configuration", // API doesn't return this
				},
			},
		},
	})
}

func TestAccECSDaemon_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceDaemon, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSDaemon_deploymentConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_deploymentConfiguration(rName, 50.0, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.drain_percent", "50"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "10"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore: []string{
					"capacity_provider_arns",
					"daemon_task_definition",
					"deployment_configuration",
				},
			},
			{
				Config: testAccDaemonConfig_deploymentConfiguration(rName, 75.0, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.drain_percent", "75"),
					resource.TestCheckResourceAttr(resourceName, "deployment_configuration.0.bake_time_in_minutes", "20"),
				),
			},
		},
	})
}

func TestAccECSDaemon_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateVerifyIgnore: []string{
					"daemon_task_definition",
					"capacity_provider_arns",
					"deployment_configuration",
				},
			},
			{
				Config: testAccDaemonConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDaemonConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECSDaemon_minimumValues(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	// Verifies the API accepts minimum values for deployment_configuration without error.
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_deploymentConfiguration(rName, 1.0, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccECSDaemon_boundaryValues(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	// Verifies the API accepts boundary values for deployment_configuration without error.
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_deploymentConfiguration(rName, 100.0, 1440),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccECSDaemon_alarms(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	// Note: deployment_configuration (including alarms) is write-only.
	// We verify the API accepts alarm configuration without error.
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_alarms(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccDaemonConfig_alarms(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccECSDaemon_enableManagedTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_enableManagedTags(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccDaemonConfig_enableManagedTags(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccECSDaemon_enableExecuteCommand(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_enableExecuteCommand(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccDaemonConfig_enableExecuteCommand(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccECSDaemon_propagateTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_propagateTags(rName, "DAEMON"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, "DAEMON"),
				),
			},
			{
				Config: testAccDaemonConfig_propagateTags(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPropagateTags, "NONE"),
				),
			},
		},
	})
}

func TestAccECSDaemon_updateTaskDefinition(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_daemon.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDaemonDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonConfig_updateTaskDefinition(rName, "nginx:latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccDaemonConfig_updateTaskDefinition(rName, "nginx:alpine"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDaemonExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccCheckDaemonDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_daemon" {
				continue
			}

			arn := rs.Primary.Attributes[names.AttrARN]

			_, err := tfecs.FindDaemonByARN(ctx, conn, arn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Daemon %s still exists", arn)
		}

		return nil
	}
}

func testAccCheckDaemonExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		arn := rs.Primary.Attributes[names.AttrARN]

		_, err := tfecs.FindDaemonByARN(ctx, conn, arn)

		return err
	}
}

func testAccDaemonConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.infra.arn

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = [aws_subnet.test[0].id]
        security_groups = [aws_security_group.test.id]
      }
    }
  }
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "infra" {
  name = "%[1]s-infra"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "infra" {
  role       = aws_iam_role.infra.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRolePolicyForManagedInstances"
}

resource "aws_iam_role" "instance" {
  name = "%[1]s-instance"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "instance" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.instance.name
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  container_definition {
    name      = "test"
    image     = "nginx:latest"
    memory    = 128
    essential = true
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}
`, rName))
}

func testAccDaemonConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
}
`, rName))
}

func testAccDaemonConfig_deploymentConfiguration(rName string, drainPercent float64, bakeTime int) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]

  deployment_configuration {
    drain_percent        = %[2]f
    bake_time_in_minutes = %[3]d
  }
}
`, rName, drainPercent, bakeTime))
}

func testAccDaemonConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDaemonConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDaemonConfig_alarms(rName string, enable bool) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]

  deployment_configuration {
    alarms {
      enable      = %[2]t
      alarm_names = ["alarm1", "alarm2"]
    }
  }
}
`, rName, enable))
}

func testAccDaemonConfig_enableManagedTags(rName string, enable bool) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                    = %[1]q
  cluster_arn             = aws_ecs_cluster.test.arn
  daemon_task_definition  = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns  = [aws_ecs_capacity_provider.test.arn]
  enable_ecs_managed_tags = %[2]t
}
`, rName, enable))
}

func testAccDaemonConfig_enableExecuteCommand(rName string, enable bool) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
  enable_execute_command = %[2]t
}
`, rName, enable))
}

func testAccDaemonConfig_propagateTags(rName string, propagate string) string {
	return acctest.ConfigCompose(testAccDaemonConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
  propagate_tags         = %[2]q
}
`, rName, propagate))
}

func testAccDaemonConfig_updateTaskDefinition(rName, image string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.infra.arn

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = [aws_subnet.test[0].id]
        security_groups = [aws_security_group.test.id]
      }
    }
  }
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "infra" {
  name = "%[1]s-infra"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "infra" {
  role       = aws_iam_role.infra.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRolePolicyForManagedInstances"
}

resource "aws_iam_role" "instance" {
  name = "%[1]s-instance"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "instance" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.instance.name
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  container_definition {
    name      = "test"
    image     = %[2]q
    memory    = 128
    essential = true
  }
}

resource "aws_ecs_daemon" "test" {
  name                   = %[1]q
  cluster_arn            = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
}
`, rName, image))
}
