// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRInstanceGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "autoscaling_policy", ""),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
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
	var v emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceInstanceGroup(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceInstanceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/1355
func TestAccEMRInstanceGroup_Disappears_emrCluster(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster
	var ig emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"
	emrClusterResourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, emrClusterResourceName, &cluster),
					testAccCheckInstanceGroupExists(ctx, resourceName, &ig),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceCluster(), emrClusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRInstanceGroup_bidPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v1),
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
					testAccCheckInstanceGroupExists(ctx, resourceName, &v2),
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
					testAccCheckInstanceGroupExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "bid_price", ""),
					testAccInstanceGroupRecreated(t, &v2, &v1),
				),
			},
		},
	})
}

func TestAccEMRInstanceGroup_sJSON(t *testing.T) {
	ctx := acctest.Context(t)
	var v emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_configurationsJSON(rName, "partitionName1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "configurations_json"),
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
				Config: testAccInstanceGroupConfig_configurationsJSON(rName, "partitionName2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "configurations_json"),
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

func TestAccEMRInstanceGroup_autoScalingPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var v emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_autoScalingPolicy(rName, 1, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "autoscaling_policy"),
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
				Config: testAccInstanceGroupConfig_autoScalingPolicy(rName, 2, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "autoscaling_policy"),
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

// Confirm we can scale down the instance count.
// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/1264
func TestAccEMRInstanceGroup_instanceCount(t *testing.T) {
	ctx := acctest.Context(t)
	var v emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_basic(rName),
				Check:  testAccCheckInstanceGroupExists(ctx, resourceName, &v),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateIdFunc:       testAccInstanceGroupResourceImportStateIdFunc(resourceName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrStatus},
			},
			{
				Config: testAccInstanceGroupConfig_zeroCount(rName),
				Check:  testAccCheckInstanceGroupExists(ctx, resourceName, &v),
			},
		},
	})
}

func TestAccEMRInstanceGroup_EBS_ebsOptimized(t *testing.T) {
	ctx := acctest.Context(t)
	var v emr.InstanceGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_instance_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceGroupConfig_ebs(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtTrue),
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
				Config: testAccInstanceGroupConfig_ebs(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "ebs_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", acctest.CtFalse),
				),
			},
		},
	})
}

func testAccCheckInstanceGroupExists(ctx context.Context, name string, ig *emr.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No task group id set")
		}

		meta := acctest.Provider.Meta()
		conn := meta.(*conns.AWSClient).EMRConn(ctx)
		group, err := tfemr.FetchInstanceGroup(ctx, conn, rs.Primary.Attributes["cluster_id"], rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("EMR error: %v", err)
		}

		if group == nil {
			return fmt.Errorf("No match found for (%s)", name)
		}
		*ig = *group

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

func testAccInstanceGroupRecreated(t *testing.T, before, after *emr.InstanceGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(before.Id) == aws.StringValue(after.Id) {
			t.Fatalf("Expected change of Instance Group Ids, but both were %v", aws.StringValue(before.Id))
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

func testAccInstanceGroupConfig_zeroCount(rName string) string {
	return acctest.ConfigCompose(testAccInstanceGroupConfig_base(rName), `
resource "aws_emr_instance_group" "test" {
  cluster_id     = aws_emr_cluster.test.id
  instance_count = 0
  instance_type  = "c4.large"
}
`)
}
