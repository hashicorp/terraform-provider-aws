// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRInstanceGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "autoscaling_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
		},
	})
}

func TestAccEMRInstanceGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tfemr.ResourceInstanceGroup(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfemr.ResourceInstanceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/1355
func TestAccEMRInstanceGroup_Disappears_emrCluster(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster awstypes.Cluster
	var ig awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"
	emrClusterResourceName := "aws_emr_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, t, emrClusterResourceName, &cluster),
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &ig),
					acctest.CheckSDKResourceDisappears(ctx, t, tfemr.ResourceCluster(), emrClusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRInstanceGroup_bidPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccInstanceGroupConfig_bidPrice(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "bid_price", "0.30"),
					testAccInstanceGroupRecreated(t, &v1, &v2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
					testAccInstanceGroupRecreated(t, &v2, &v1),
				),
			},
		},
	})
}

func TestAccEMRInstanceGroup_sJSON(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_configurationsJSON(rName, "partitionName1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "configurations_json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations_json",
					names.AttrStatus,
				},
			},
			{
				Config: testAccInstanceGroupConfig_configurationsJSON(rName, "partitionName2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "configurations_json"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"configurations_json",
					names.AttrStatus,
				},
			},
		},
	})
}

func TestAccEMRInstanceGroup_autoScalingPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_autoScalingPolicy(rName, 1, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "autoscaling_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrInstanceCount,
					names.AttrStatus,
				},
			},
			{
				Config: testAccInstanceGroupConfig_autoScalingPolicy(rName, 2, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "autoscaling_policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrInstanceCount,
					names.AttrStatus,
				},
			},
		},
	})
}

// Confirm we can scale down the instance count.
// See https://github.com/hashicorp/terraform-provider-aws/issues/1264.
func TestAccEMRInstanceGroup_instanceCountDecrease(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_instanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccInstanceGroupConfig_instanceCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, "0"),
				),
			},
		},
	})
}

// Confirm we can create with a 0 instance count.
// See https://github.com/hashicorp/terraform-provider-aws/issues/38837.
func TestAccEMRInstanceGroup_instanceCountCreateZero(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_instanceCount(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceCount, "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
		},
	})
}

func TestAccEMRInstanceGroup_EBS_ebsOptimized(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.InstanceGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_ebs(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"ebs_config.0.iops",
					names.AttrStatus,
				},
			},
			{
				Config: testAccInstanceGroupConfig_ebs(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckInstanceGroupExists(ctx context.Context, t *testing.T, n string, v *awstypes.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).EMRClient(ctx)

		output, err := tfemr.FindInstanceGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["cluster_id"], rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccInstanceGroupResourceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["cluster_id"], rs.Primary.ID), nil
	}
}

func testAccInstanceGroupRecreated(t *testing.T, before, after *awstypes.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(before.Id) == aws.ToString(after.Id) {
			t.Fatalf("Expected change of Instance Group Ids, but both were %v", aws.ToString(before.Id))
		}

		return nil
	}
}

func testAccInstanceGroupConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  autoscaling_role                  = aws_iam_role.emr_autoscaling_role.arn
  configurations                    = "test-fixtures/emr_configurations.json"
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.26.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]
}
`, rName))
}

func testAccInstanceGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccInstanceGroupConfig_base(rName), `
resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  instance_count = 1
  instance_type  = "c4.large"
}
`)
}

func testAccInstanceGroupConfig_bidPrice(rName string) string {
	return acctest.ConfigCompose(testAccInstanceGroupConfig_base(rName), `
resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  bid_price      = "0.30"
  instance_count = 1
  instance_type  = "c4.large"
}
`)
}

func testAccInstanceGroupConfig_configurationsJSON(rName, name string) string {
	return acctest.ConfigCompose(testAccInstanceGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_group" "test" {
  cluster_id          = aws_emr_cluster.test.id
  instance_count      = 1
  instance_type       = "c4.large"
  configurations_json = <<EOF
    [
      {
        "Classification": "yarn-site",
        "Properties": {
          "yarn.nodemanager.node-labels.provider": "config",
          "yarn.nodemanager.node-labels.provider.configured-node-partition": %[1]q
        }
      }
    ]
EOF
}
`, name))
}

func testAccInstanceGroupConfig_autoScalingPolicy(rName string, min, max int) string {
	return acctest.ConfigCompose(testAccInstanceGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_group" "test" {
  cluster_id         = aws_emr_cluster.test.id
  instance_type      = "c4.large"
  autoscaling_policy = <<EOT
{
  "Constraints": {
    "MinCapacity": %[1]d,
    "MaxCapacity": %[2]d
  },
  "Rules": [
    {
      "Name": "ScaleOutMemoryPercentage",
      "Description": "Scale out if YARNMemoryAvailablePercentage is less than 15",
      "Action": {
        "SimpleScalingPolicyConfiguration": {
          "AdjustmentType": "CHANGE_IN_CAPACITY",
          "ScalingAdjustment": 1,
          "CoolDown": 300
        }
      },
      "Trigger": {
        "CloudWatchAlarmDefinition": {
          "ComparisonOperator": "LESS_THAN",
          "EvaluationPeriods": 1,
          "MetricName": "YARNMemoryAvailablePercentage",
          "Namespace": "AWS/ElasticMapReduce",
          "Period": 300,
          "Statistic": "AVERAGE",
          "Threshold": 15.0,
          "Unit": "PERCENT"
        }
      }
    }
  ]
}
EOT
}
`, min, max))
}

func testAccInstanceGroupConfig_ebs(rName string, o bool) string {
	return acctest.ConfigCompose(testAccInstanceGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  instance_count = 1
  instance_type  = "c4.large"
  ebs_optimized  = %[1]t

  ebs_config {
    size = 10
    type = "gp2"
  }
}
`, o))
}

func testAccInstanceGroupConfig_instanceCount(rName string, count int) string {
	return acctest.ConfigCompose(testAccInstanceGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  instance_count = %[1]d
  instance_type  = "c4.large"
}
`, count))
}
