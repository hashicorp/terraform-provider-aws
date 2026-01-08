// Copyright (c) Altos Labs, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func TestAccSageMakerCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster sagemaker.DescribeClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				Source:            "hashicorp/http",
				VersionConstraint: "~> 3.0",
			},
		},
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "cluster_name", rName),
					resource.TestCheckResourceAttr(resourceName, "cluster_status", string(awstypes.ClusterStatusInservice)),
					resource.TestCheckResourceAttr(resourceName, "instance_group.#", "2"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", regexache.MustCompile(`cluster/.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        rName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "cluster_name",
			},
		},
	})
}

func TestAccSageMakerCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster sagemaker.DescribeClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				Source:            "hashicorp/http",
				VersionConstraint: "~> 3.0",
			},
		},
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceCluster, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster sagemaker.DescribeClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				Source:            "hashicorp/http",
				VersionConstraint: "~> 3.0",
			},
		},
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateId:                        rName,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "cluster_name",
			},
			{
				Config: testAccClusterConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerCluster_updateInstanceGroups(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster1, cluster2 sagemaker.DescribeClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				Source:            "hashicorp/http",
				VersionConstraint: "~> 3.0",
			},
		},
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "instance_group.#", "2"),
				),
			},
			{
				Config: testAccClusterConfig_updateInstanceGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "instance_group.#", "2"),
					// Check that worker group instance count was updated
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "instance_group.*", map[string]string{
						"instance_group_name": "worker-group",
						"instance_count":      "2",
					}),
				),
			},
		},
	})
}

func TestAccSageMakerCluster_forceReplacement(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster1, cluster2 sagemaker.DescribeClusterOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				Source:            "hashicorp/http",
				VersionConstraint: "~> 3.0",
			},
		},
		CheckDestroy: testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "instance_group.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "instance_group.*", map[string]string{
						"instance_group_name": "worker-group",
						"instance_type":       "ml.t3.medium",
					}),
				),
			},
			{
				Config: testAccClusterConfig_forceReplacement(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					func(s *terraform.State) error {
						// Expect the cluster to be recreated
						if err := testAccCheckClusterNotRecreated(&cluster1, &cluster2)(s); err == nil {
							return create.Error(names.SageMaker, create.ErrActionCheckingRecreated, tfsagemaker.ResNameCluster, "unknown", errors.New("expected cluster to be recreated but it was not"))
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "instance_group.#", "2"),
					// Check that worker group instance type was updated
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "instance_group.*", map[string]string{
						"instance_group_name": "worker-group",
						"instance_type":       "ml.t3.large",
					}),
				),
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_cluster" {
				continue
			}

			if rs.Primary == nil {
				continue
			}

			clusterName := rs.Primary.Attributes["cluster_name"]
			if clusterName == "" {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameCluster, clusterName, errors.New("name not set"))
			}
			_, err := tfsagemaker.FindClusterByName(ctx, conn, clusterName)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameCluster, clusterName, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameCluster, clusterName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, name string, cluster *sagemaker.DescribeClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameCluster, name, errors.New("not found"))
		}

		if rs.Primary == nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameCluster, name, errors.New("primary instance not set"))
		}

		clusterName := rs.Primary.Attributes["cluster_name"]
		if clusterName == "" {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameCluster, name, errors.New("name not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		resp, err := tfsagemaker.FindClusterByName(ctx, conn, clusterName)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameCluster, clusterName, err)
		}

		if resp == nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameCluster, clusterName, errors.New("response is nil"))
		}

		*cluster = *resp

		return nil
	}
}

func testAccCheckClusterNotRecreated(before, after *sagemaker.DescribeClusterOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before == nil || after == nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingNotRecreated, tfsagemaker.ResNameCluster, "unknown", errors.New("cluster output is nil"))
		}

		if before, after := aws.ToString(before.ClusterArn), aws.ToString(after.ClusterArn); before != after {
			return create.Error(names.SageMaker, create.ErrActionCheckingNotRecreated, tfsagemaker.ResNameCluster, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_cluster" "test" {
  cluster_name = %[1]q

  instance_group {
    instance_group_name = "controller-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  instance_group {
    instance_group_name = "worker-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
  }
  
  depends_on = [
    aws_vpc_endpoint.s3,
    aws_s3_object.on_create_script,
    aws_s3_object.add_users_script,
    aws_s3_object.apply_hotfix_script,
    aws_s3_object.config_py,
    aws_s3_object.lifecycle_script_py,
    aws_s3_object.mount_fsx_script,
    aws_s3_object.mount_fsx_openzfs_script,
    aws_s3_object.setup_mariadb_accounting_script,
    aws_s3_object.setup_rds_accounting_script,
    aws_s3_object.setup_sssd_py,
    aws_s3_object.shared_users_sample,
    aws_s3_object.start_slurm_script,
    aws_s3_object.hold_lustre_client_script,
    aws_s3_object.mock_gpu_driver_script,
    aws_s3_object.enroot_conf,
    aws_s3_object.fsx_ubuntu_script,
    aws_s3_object.install_docker_script,
    aws_s3_object.install_enroot_pyxis_script,
    aws_s3_object.motd_script,
    aws_s3_object.motd_txt,
    aws_s3_object.gen_keypair_ubuntu_script,
    aws_s3_object.ssh_to_compute_script,
    aws_s3_object.provisioning_parameters,
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.test-ec2,
    aws_iam_role_policy_attachment.test-s3
  ]
}
`, rName))
}

func testAccClusterConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_cluster" "test" {
  cluster_name = %[1]q

  instance_group {
    instance_group_name = "controller-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  instance_group {
    instance_group_name = "worker-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
  }

  tags = {
    %[2]q = %[3]q
  }
  
  depends_on = [
    aws_vpc_endpoint.s3,
    aws_s3_object.on_create_script,
    aws_s3_object.add_users_script,
    aws_s3_object.apply_hotfix_script,
    aws_s3_object.config_py,
    aws_s3_object.lifecycle_script_py,
    aws_s3_object.mount_fsx_script,
    aws_s3_object.mount_fsx_openzfs_script,
    aws_s3_object.setup_mariadb_accounting_script,
    aws_s3_object.setup_rds_accounting_script,
    aws_s3_object.setup_sssd_py,
    aws_s3_object.shared_users_sample,
    aws_s3_object.start_slurm_script,
    aws_s3_object.hold_lustre_client_script,
    aws_s3_object.mock_gpu_driver_script,
    aws_s3_object.enroot_conf,
    aws_s3_object.fsx_ubuntu_script,
    aws_s3_object.install_docker_script,
    aws_s3_object.install_enroot_pyxis_script,
    aws_s3_object.motd_script,
    aws_s3_object.motd_txt,
    aws_s3_object.gen_keypair_ubuntu_script,
    aws_s3_object.ssh_to_compute_script,
    aws_s3_object.provisioning_parameters,
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.test-ec2,
    aws_iam_role_policy_attachment.test-s3
  ]
}
`, rName, tagKey1, tagValue1))
}

func testAccClusterConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_cluster" "test" {
  cluster_name = %[1]q

  instance_group {
    instance_group_name = "controller-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  instance_group {
    instance_group_name = "worker-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
  
  depends_on = [
    aws_vpc_endpoint.s3,
    aws_s3_object.on_create_script,
    aws_s3_object.add_users_script,
    aws_s3_object.apply_hotfix_script,
    aws_s3_object.config_py,
    aws_s3_object.lifecycle_script_py,
    aws_s3_object.mount_fsx_script,
    aws_s3_object.mount_fsx_openzfs_script,
    aws_s3_object.setup_mariadb_accounting_script,
    aws_s3_object.setup_rds_accounting_script,
    aws_s3_object.setup_sssd_py,
    aws_s3_object.shared_users_sample,
    aws_s3_object.start_slurm_script,
    aws_s3_object.hold_lustre_client_script,
    aws_s3_object.mock_gpu_driver_script,
    aws_s3_object.enroot_conf,
    aws_s3_object.fsx_ubuntu_script,
    aws_s3_object.install_docker_script,
    aws_s3_object.install_enroot_pyxis_script,
    aws_s3_object.motd_script,
    aws_s3_object.motd_txt,
    aws_s3_object.gen_keypair_ubuntu_script,
    aws_s3_object.ssh_to_compute_script,
    aws_s3_object.provisioning_parameters,
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.test-ec2,
    aws_iam_role_policy_attachment.test-s3
  ]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccClusterConfig_updateInstanceGroups(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_cluster" "test" {
  cluster_name = %[1]q

  instance_group {
    instance_group_name = "controller-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  instance_group {
    instance_group_name = "worker-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 2
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
  }
  
  depends_on = [
    aws_vpc_endpoint.s3,
    aws_s3_object.on_create_script,
    aws_s3_object.add_users_script,
    aws_s3_object.apply_hotfix_script,
    aws_s3_object.config_py,
    aws_s3_object.lifecycle_script_py,
    aws_s3_object.mount_fsx_script,
    aws_s3_object.mount_fsx_openzfs_script,
    aws_s3_object.setup_mariadb_accounting_script,
    aws_s3_object.setup_rds_accounting_script,
    aws_s3_object.setup_sssd_py,
    aws_s3_object.shared_users_sample,
    aws_s3_object.start_slurm_script,
    aws_s3_object.hold_lustre_client_script,
    aws_s3_object.mock_gpu_driver_script,
    aws_s3_object.enroot_conf,
    aws_s3_object.fsx_ubuntu_script,
    aws_s3_object.install_docker_script,
    aws_s3_object.install_enroot_pyxis_script,
    aws_s3_object.motd_script,
    aws_s3_object.motd_txt,
    aws_s3_object.gen_keypair_ubuntu_script,
    aws_s3_object.ssh_to_compute_script,
    aws_s3_object.provisioning_parameters,
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.test-ec2,
    aws_iam_role_policy_attachment.test-s3
  ]
}
`, rName))
}

func testAccClusterConfig_forceReplacement(rName string) string {
	return acctest.ConfigCompose(testAccClusterConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_cluster" "test" {
  cluster_name = %[1]q

  instance_group {
    instance_group_name = "controller-group"
    instance_type       = "ml.t3.medium"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  instance_group {
    instance_group_name = "worker-group"
    instance_type       = "ml.t3.large"
    instance_count      = 1
    execution_role      = aws_iam_role.test.arn

    lifecycle_config {
      source_s3_uri = "s3://${aws_s3_bucket.test.id}"
      on_create     = "on_create.sh"
    }
  }

  vpc_config {
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
  }
  
  depends_on = [
    aws_vpc_endpoint.s3,
    aws_s3_object.on_create_script,
    aws_s3_object.add_users_script,
    aws_s3_object.apply_hotfix_script,
    aws_s3_object.config_py,
    aws_s3_object.lifecycle_script_py,
    aws_s3_object.mount_fsx_script,
    aws_s3_object.mount_fsx_openzfs_script,
    aws_s3_object.setup_mariadb_accounting_script,
    aws_s3_object.setup_rds_accounting_script,
    aws_s3_object.setup_sssd_py,
    aws_s3_object.shared_users_sample,
    aws_s3_object.start_slurm_script,
    aws_s3_object.hold_lustre_client_script,
    aws_s3_object.mock_gpu_driver_script,
    aws_s3_object.enroot_conf,
    aws_s3_object.fsx_ubuntu_script,
    aws_s3_object.install_docker_script,
    aws_s3_object.install_enroot_pyxis_script,
    aws_s3_object.motd_script,
    aws_s3_object.motd_txt,
    aws_s3_object.gen_keypair_ubuntu_script,
    aws_s3_object.ssh_to_compute_script,
    aws_s3_object.provisioning_parameters,
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.test-ec2,
    aws_iam_role_policy_attachment.test-s3
  ]
}
`, rName))
}

func testAccClusterConfig_base(rName string) string {
	return testAccClusterConfig_baseWithParams(rName, "controller-group", "worker-group")
}

func testAccClusterConfig_baseWithParams(rName, controllerGroup, workerGroup string) string {
	return fmt.Sprintf(`
locals {
  provisioning_parameters_content = jsonencode({
    version = "1.0.0"
    workload_manager = "slurm"
    controller_group = %[2]q
    worker_groups = [{
      instance_group_name = %[3]q
      partition_name = "normal"
    }]
  })
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

data "aws_iam_policy_document" "s3_endpoint" {
  statement {
    effect = "Allow"
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_vpc_endpoint" "s3" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]
  policy          = data.aws_iam_policy_document.s3_endpoint.json
  
  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "http" "add_users_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/add_users.sh"
}

data "http" "apply_hotfix_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/apply_hotfix.sh"
}

data "http" "config_py" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/config.py"
}

data "http" "lifecycle_script_py" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/lifecycle_script.py"
}

data "http" "mount_fsx_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/mount_fsx.sh"
}

data "http" "mount_fsx_openzfs_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/mount_fsx_openzfs.sh"
}

data "http" "setup_mariadb_accounting_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/setup_mariadb_accounting.sh"
}

data "http" "setup_rds_accounting_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/setup_rds_accounting.sh"
}

data "http" "setup_sssd_py" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/setup_sssd.py"
}

data "http" "shared_users_sample" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/shared_users_sample.txt"
}

data "http" "start_slurm_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/start_slurm.sh"
}

# Hotfix directory files
data "http" "hold_lustre_client_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/hotfix/hold-lustre-client.sh"
}

data "http" "mock_gpu_driver_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/hotfix/mock-gpu-driver-deb.sh"
}

# Utils directory files
data "http" "enroot_conf" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/enroot.conf"
}

data "http" "fsx_ubuntu_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/fsx_ubuntu.sh"
}

data "http" "install_docker_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/install_docker.sh"
}

data "http" "install_enroot_pyxis_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/install_enroot_pyxis.sh"
}

data "http" "motd_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/motd.sh"
}

data "http" "motd_txt" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/motd.txt"
}

# Additional utils directory files that are referenced by lifecycle_script.py
data "http" "gen_keypair_ubuntu_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/gen-keypair-ubuntu.sh"
}

data "http" "ssh_to_compute_script" {
  url = "https://raw.githubusercontent.com/aws-samples/awsome-distributed-training/main/1.architectures/5.sagemaker-hyperpod/LifecycleScripts/base-config/utils/ssh-to-compute.sh"
}

resource "aws_s3_object" "on_create_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "on_create.sh"
  content = file("~/dev/terraform-hpcet/examples/lifecycle-scripts/on_create.sh")
}

resource "aws_s3_object" "add_users_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "add_users.sh"
  content = data.http.add_users_script.response_body
}

resource "aws_s3_object" "apply_hotfix_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "apply_hotfix.sh"
  content = data.http.apply_hotfix_script.response_body
}

resource "aws_s3_object" "config_py" {
  bucket  = aws_s3_bucket.test.id
  key     = "config.py"
  content = data.http.config_py.response_body
}

resource "aws_s3_object" "lifecycle_script_py" {
  bucket  = aws_s3_bucket.test.id
  key     = "lifecycle_script.py"
  content = data.http.lifecycle_script_py.response_body
}

resource "aws_s3_object" "mount_fsx_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "mount_fsx.sh"
  content = data.http.mount_fsx_script.response_body
}

resource "aws_s3_object" "mount_fsx_openzfs_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "mount_fsx_openzfs.sh"
  content = data.http.mount_fsx_openzfs_script.response_body
}

resource "aws_s3_object" "setup_mariadb_accounting_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "setup_mariadb_accounting.sh"
  content = data.http.setup_mariadb_accounting_script.response_body
}

resource "aws_s3_object" "setup_rds_accounting_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "setup_rds_accounting.sh"
  content = data.http.setup_rds_accounting_script.response_body
}

resource "aws_s3_object" "setup_sssd_py" {
  bucket  = aws_s3_bucket.test.id
  key     = "setup_sssd.py"
  content = data.http.setup_sssd_py.response_body
}

resource "aws_s3_object" "shared_users_sample" {
  bucket  = aws_s3_bucket.test.id
  key     = "shared_users_sample.txt"
  content = data.http.shared_users_sample.response_body
}

resource "aws_s3_object" "start_slurm_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "start_slurm.sh"
  content = data.http.start_slurm_script.response_body
}

# Hotfix directory S3 objects
resource "aws_s3_object" "hold_lustre_client_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "hotfix/hold-lustre-client.sh"
  content = data.http.hold_lustre_client_script.response_body
}

resource "aws_s3_object" "mock_gpu_driver_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "hotfix/mock-gpu-driver-deb.sh"
  content = data.http.mock_gpu_driver_script.response_body
}

# Utils directory S3 objects
resource "aws_s3_object" "enroot_conf" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/enroot.conf"
  content = data.http.enroot_conf.response_body
}

resource "aws_s3_object" "fsx_ubuntu_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/fsx_ubuntu.sh"
  content = data.http.fsx_ubuntu_script.response_body
}

resource "aws_s3_object" "install_docker_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/install_docker.sh"
  content = data.http.install_docker_script.response_body
}

resource "aws_s3_object" "install_enroot_pyxis_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/install_enroot_pyxis.sh"
  content = data.http.install_enroot_pyxis_script.response_body
}

resource "aws_s3_object" "motd_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/motd.sh"
  content = data.http.motd_script.response_body
}

resource "aws_s3_object" "motd_txt" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/motd.txt"
  content = data.http.motd_txt.response_body
}

resource "aws_s3_object" "provisioning_parameters" {
  bucket  = aws_s3_bucket.test.id
  key     = "provisioning_parameters.json"
  content = local.provisioning_parameters_content
}

# Additional utils directory S3 objects
resource "aws_s3_object" "gen_keypair_ubuntu_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/gen-keypair-ubuntu.sh"
  content = data.http.gen_keypair_ubuntu_script.response_body
}

resource "aws_s3_object" "ssh_to_compute_script" {
  bucket  = aws_s3_bucket.test.id
  key     = "utils/ssh-to-compute.sh"
  content = data.http.ssh_to_compute_script.response_body
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test-ec2" {
  statement {
    actions = [
      "ec2:CreateNetworkInterface",
      "ec2:CreateNetworkInterfacePermission",
      "ec2:DeleteNetworkInterface",
      "ec2:DeleteNetworkInterfacePermission",
      "ec2:DescribeNetworkInterfaces",
      "ec2:DescribeVpcs",
      "ec2:DescribeDhcpOptions",
      "ec2:DescribeSubnets",
      "ec2:DescribeSecurityGroups",
      "ec2:DetachNetworkInterface"
    ]
    resources = ["*"]
  }

  statement {
    actions = ["ec2:CreateTags"]
    resources = [
      "arn:aws:ec2:*:*:network-interface/*"
    ]
  }
}

resource "aws_iam_policy" "test-ec2" {
  name        = "%[1]s-ec2-policy"
  description = "Custom policy to be assumed by SageMaker Hyperpod"
  policy      = data.aws_iam_policy_document.test-ec2.json
}

resource "aws_iam_role_policy_attachment" "test-ec2" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test-ec2.arn
}

data "aws_iam_policy_document" "test-s3" {
  statement {
    actions = [
      "s3:ListBucket",
    ]
    resources = [aws_s3_bucket.test.arn]
  }
  statement {
    actions = [
      "s3:GetObject",
    ]
    resources = ["${aws_s3_bucket.test.arn}/*"]
  }
}

resource "aws_iam_policy" "test-s3" {
  name        = "%[1]s-s3-policy"
  description = "Custom S3 policy to be assumed by SageMaker Hyperpod"
  policy      = data.aws_iam_policy_document.test-s3.json
}

resource "aws_iam_role_policy_attachment" "test-s3" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test-s3.arn
}

data "aws_iam_policy" "test" {
  name = "AmazonSageMakerClusterInstanceRolePolicy"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = data.aws_iam_policy.test.arn
}

resource "aws_security_group" "test" {
  name_prefix = %[1]q
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

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
`, rName, controllerGroup, workerGroup)
}
