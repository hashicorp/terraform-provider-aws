// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSClusterDataSource_ecsCluster(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "service_connect_defaults.#", resourceName, "service_connect_defaults.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccECSClusterDataSource_ecsClusterContainerInsights(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_containerInsights(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "setting.#", resourceName, "setting.#"),
				),
			},
		},
	})
}

func TestAccECSClusterDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "pending_tasks_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "registered_container_instances_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "running_tasks_count", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "service_connect_defaults.#", resourceName, "service_connect_defaults.#"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccECSClusterDataSource_configuration(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_configuration(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.execute_command_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration.0.execute_command_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.execute_command_configuration.0.logging", "OVERRIDE"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.execute_command_configuration.0.log_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration.0.execute_command_configuration.0.log_configuration.0.cloud_watch_log_group_name", "aws_cloudwatch_log_group.test", names.AttrName),
					resource.TestCheckResourceAttr(dataSourceName, "configuration.0.managed_storage_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration.0.managed_storage_configuration.0.fargate_ephemeral_storage_kms_key_id", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration.0.managed_storage_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccECSClusterDataSource_capacityProviders(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_cluster.test"
	resourceName := "aws_ecs_cluster.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_capacityProviders(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "active_services_count", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "capacity_providers.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "capacity_providers.*", "FARGATE"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "capacity_providers.*", "FARGATE_SPOT"),
					resource.TestCheckResourceAttr(dataSourceName, "default_capacity_provider_strategy.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "default_capacity_provider_strategy.*", map[string]string{
						"base":              "1",
						"capacity_provider": "FARGATE",
						names.AttrWeight:    "100",
					}),
					resource.TestCheckResourceAttrSet(dataSourceName, "statistics.#"),
				),
			},
		},
	})
}

func TestFlattenClusterAttachments(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input []awstypes.Attachment
		want  []any
	}{
		"nil": {
			input: nil,
			want:  []any{},
		},
		"attachment with details": {
			input: []awstypes.Attachment{
				{
					Id:     aws.String("a1"),
					Type:   aws.String("as_policy"),
					Status: aws.String("CREATED"),
					Details: []awstypes.KeyValuePair{
						{Name: aws.String("capacityProviderName"), Value: aws.String("cp1")},
					},
				},
			},
			want: []any{
				map[string]any{
					"details": []any{
						map[string]any{
							names.AttrName:  "capacityProviderName",
							names.AttrValue: "cp1",
						},
					},
					names.AttrID:     "a1",
					names.AttrStatus: "CREATED",
					names.AttrType:   "as_policy",
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfecs.FlattenClusterAttachments(testCase.input)

			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("unexpected diff (+want, -got): %s", diff)
			}
		})
	}
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName)
}

func testAccClusterDataSourceConfig_containerInsights(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName)
}

func testAccClusterDataSourceConfig_tags(rName, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName, tagKey, tagValue)
}

func testAccClusterDataSourceConfig_configuration(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "ECSClusterFargatePolicy"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          "AWS" : "*"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow generate data key access for Fargate tasks."
        Effect = "Allow"
        Principal = {
          Service = "fargate.amazonaws.com"
        }
        Action = [
          "kms:GenerateDataKeyWithoutPlaintext"
        ]
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:aws:ecs:clusterAccount" = [
              data.aws_caller_identity.current.account_id
            ]
            "kms:EncryptionContext:aws:ecs:clusterName" = [
              %[1]q
            ]
          }
        }
        Resource = "*"
      },
      {
        Sid    = "Allow grant creation permission for Fargate tasks."
        Effect = "Allow"
        Principal = {
          Service = "fargate.amazonaws.com"
        }
        Action = [
          "kms:CreateGrant"
        ]
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:aws:ecs:clusterAccount" = [
              data.aws_caller_identity.current.account_id
            ]
            "kms:EncryptionContext:aws:ecs:clusterName" = [
              %[1]q
            ]
          }
          "ForAllValues:StringEquals" = {
            "kms:GrantOperations" = [
              "Decrypt"
            ]
          }
        }
        Resource = "*"
      }
    ]
  })
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q

  configuration {
    execute_command_configuration {
      kms_key_id = aws_kms_key.test.arn
      logging    = "OVERRIDE"

      log_configuration {
        cloud_watch_encryption_enabled = true
        cloud_watch_log_group_name     = aws_cloudwatch_log_group.test.name
      }
    }

    managed_storage_configuration {
      fargate_ephemeral_storage_kms_key_id = aws_kms_key.test.arn
      kms_key_id                           = aws_kms_key.test.arn
    }
  }
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster.test.name
}
`, rName)
}

func testAccClusterDataSourceConfig_capacityProviders(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name       = aws_ecs_cluster.test.name
  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}

data "aws_ecs_cluster" "test" {
  cluster_name = aws_ecs_cluster_capacity_providers.test.cluster_name
}
`, rName)
}
