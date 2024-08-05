// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/emr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEMRCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticmapreduce", regexache.MustCompile("cluster/.+$")),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-4.6.0"),
					resource.TestCheckResourceAttr(resourceName, "applications.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "applications.*", "Spark"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.subnet_id", "aws_subnet.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.emr_managed_master_security_group", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.emr_managed_slave_security_group", "aws_security_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.instance_profile", "aws_iam_instance_profile.emr_instance_profile", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr(resourceName, "ebs_root_volume_size", "21"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, "aws_iam_role.emr_service", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "autoscaling_role", "aws_iam_role.emr_autoscaling_role", names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "additional_info"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_autoTerminationPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_autoTermination(rName, 10000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.0.idle_timeout", "10000"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_autoTermination(rName, 20000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.0.idle_timeout", "20000"),
				),
			},
			{
				Config: testAccClusterConfig_noAutoTermination(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", acctest.Ct0),
				),
			},
			{
				Config: testAccClusterConfig_autoTermination(rName, 20000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.0.idle_timeout", "20000"),
				),
			},
		},
	})
}

func TestAccEMRCluster_additionalInfo(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster
	expectedJSON := `
{
  "instanceAwsClientConfiguration": {
    "proxyPort": 8099,
    "proxyHost": "myproxy.example.com"
  }
}`

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_additionalInfo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct0),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "additional_info", expectedJSON),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"additional_info",
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceCluster(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfemr.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRCluster_sJSON(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_configurationsJSON(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestMatchResourceAttr(resourceName, "configurations_json",
						regexache.MustCompile("{\"JAVA_HOME\":\"/usr/lib/jvm/java-1.8.0\".+")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_autoScalingPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2, cluster3 emr.Cluster
	autoscalingPolicy1 := `
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 2
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
`
	autoscalingPolicy2 := `
{
  "Constraints": {
    "MinCapacity": 1,
    "MaxCapacity": 3
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
`
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_coreInstanceGroupAutoScalingPolicy(rName, autoscalingPolicy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "core_instance_group.0.autoscaling_policy", autoscalingPolicy1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_coreInstanceGroupAutoScalingPolicy(rName, autoscalingPolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "core_instance_group.0.autoscaling_policy", autoscalingPolicy2),
				),
			},
			{
				Config: testAccClusterConfig_coreInstanceGroupAutoScalingPolicyRemoved(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.autoscaling_policy", ""),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_bidPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_coreInstanceGroupBidPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.bid_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_coreInstanceGroupBidPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_instanceCount(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2, cluster3 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_coreInstanceGroupInstanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_coreInstanceGroupInstanceCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", acctest.Ct1),
				),
			},
			{
				Config: testAccClusterConfig_coreInstanceGroupInstanceCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_instanceType(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_coreInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_coreInstanceGroupInstanceType(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_name(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_coreInstanceGroupName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_coreInstanceGroupName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccEMRCluster_EC2Attributes_defaultManagedSecurityGroups(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster
	var vpc ec2types.Vpc

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ec2AttributesDefaultManagedSecurityGroups(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					acctest.CheckVPCExists(ctx, vpcResourceName, &vpc),
				),
			},
			{
				Config:      testAccClusterConfig_ec2AttributesDefaultManagedSecurityGroups(rName),
				Destroy:     true,
				ExpectError: regexache.MustCompile(`DependencyViolation`),
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

					err := testAccDeleteManagedSecurityGroups(ctx, conn, &vpc)

					if err != nil {
						t.Fatal(err)
					}
				},
				Config:  testAccClusterConfig_ec2AttributesDefaultManagedSecurityGroups(rName),
				Destroy: true,
			},
		},
	})
}

func TestAccEMRCluster_Kerberos_clusterDedicatedKdc(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	password := fmt.Sprintf("NeverKeepPasswordsInPlainText%s!", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_kerberosDedicatedKdc(rName, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.0.kdc_admin_password", password),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.0.realm", "EC2.INTERNAL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
					"kerberos_attributes.0.kdc_admin_password",
				},
			},
		},
	})
}

func TestAccEMRCluster_MasterInstanceGroup_bidPrice(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_masterInstanceGroupBidPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.bid_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_masterInstanceGroupBidPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccEMRCluster_MasterInstanceGroup_instanceCount(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_masterInstanceGroupInstanceCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_count", acctest.Ct3),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_masterInstanceGroupInstanceCount(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_count", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccEMRCluster_MasterInstanceGroup_instanceType(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_masterInstanceGroupInstanceType(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.large"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_masterInstanceGroupInstanceType(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccEMRCluster_MasterInstanceGroup_name(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_masterInstanceGroupName(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.name", "name1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_masterInstanceGroupName(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccEMRCluster_security(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_securityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttrPair(resourceName, "security_configuration", "aws_emr_security_configuration.test", names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_Step_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_stepSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.main_class", ""),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.properties.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_Step_mode(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2, cluster3 emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_stepSingle(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_stepNoBlocks(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_stepZeroed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_Step_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_stepMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
					resource.TestCheckResourceAttr(resourceName, "step.1.action_on_failure", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.0", "spark-example"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.1", "SparkPi"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.2", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.1.name", "Spark Step"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_Step_multiple_listStates(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_stepMultipleListStates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
					resource.TestCheckResourceAttr(resourceName, "step.1.action_on_failure", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.0", "spark-example"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.1", "SparkPi"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.2", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.1.name", "Spark Step"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
					"list_steps_states",
				},
			},
		},
	})
}

func TestAccEMRCluster_Bootstrap_ordering(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_bootstrap(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "11"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.0", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.1", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.2", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.3", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.4", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.5", "5"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.6", "6"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.7", "7"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.8", "8"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.9", "9"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.10", acctest.Ct10),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_bootstrapAdd(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "11"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.0", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.1", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.2", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.3", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.4", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.5", "5"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.6", "6"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.7", "7"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.8", "8"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.9", "9"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.10", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.name", "runif-2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.1", "echo also running on master node"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_bootstrapReorder(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.#", "11"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.0", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.1", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.2", ""),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.3", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.4", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.5", "5"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.6", "6"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.7", "7"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.8", "8"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.9", "9"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.10", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "runif-2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.1", "echo also running on master node"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_PlacementGroupConfigs(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_PlacementGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "placement_group_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "placement_group_config.0.instance_role", "MASTER"),
					resource.TestCheckResourceAttr(resourceName, "placement_group_config.0.placement_strategy", "SPREAD"),
				),
			},
			{
				Config: testAccClusterConfig_PlacementGroupWithOptionalUnset(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "placement_group_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "placement_group_config.0.instance_role", "MASTER"),
					resource.TestCheckResourceAttr(resourceName, "placement_group_config.0.placement_strategy", "SPREAD"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_terminationProtected(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_terminationPolicy(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", acctest.CtFalse),
				),
			},
			{
				Config: testAccClusterConfig_terminationPolicy(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				//Need to turn off termination_protection to allow the job to be deleted
				Config: testAccClusterConfig_terminationPolicy(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_keepJob(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_keepJob(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "keep_job_flow_alive_when_no_steps", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_visibleToAllUsers(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "visible_to_all_users", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_visibleToAllUsersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "visible_to_all_users", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_s3Logging(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := fmt.Sprintf("s3n://%s/", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_s3Logging(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "log_uri", bucketName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_s3LogEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := fmt.Sprintf("s3n://%s/", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_s3Encryption(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "log_uri", bucketName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "log_encryption_kms_key_id", "kms", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "tags.role", "rolename"),
					resource.TestCheckResourceAttr(resourceName, "tags.dns_zone", "env_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "env"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", "name-env")),
			},
			{
				Config: testAccClusterConfig_updatedTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "tags.dns_zone", "new_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.Env", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", "name-env"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_RootVolume_size(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "ebs_root_volume_size", "21"),
				),
			},
			{
				Config: testAccClusterConfig_updatedRootVolumeSize(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "ebs_root_volume_size", "48"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_StepConcurrency_level(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_stepConcurrencyLevel(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step_concurrency_level", acctest.Ct2),
				),
			},
			{
				Config: testAccClusterConfig_stepConcurrencyLevel(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step_concurrency_level", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_ebs(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ebs(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.ebs_config.0.volumes_per_instance", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.ebs_config.0.volumes_per_instance", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_CustomAMI_id(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_customAMIID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttrSet(resourceName, "custom_ami_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_InstanceFleet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster1, cluster2 emr.Cluster

	resourceName := "aws_emr_cluster.test"
	subnetResourceName := "aws_subnet.test"
	subnet2ResourceName := "aws_subnet.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_instanceFleets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.subnet_id", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.0.subnet_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ec2_attributes.0.subnet_ids.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.0.launch_specifications.0.spot_specification.0.allocation_strategy", "capacity-optimized"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"core_instance_fleet",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_instanceFleetMultipleSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.0.subnet_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ec2_attributes.0.subnet_ids.*", subnetResourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ec2_attributes.0.subnet_ids.*", subnet2ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.0.launch_specifications.0.spot_specification.0.allocation_strategy", "price-capacity-optimized"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"core_instance_fleet",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_InstanceFleetMaster_only(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_instanceFleetsMasterOnly(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func TestAccEMRCluster_unhealthyNodeReplacement(t *testing.T) {
	ctx := acctest.Context(t)
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EMRServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_unhealthyNodeReplacement(rName, acctest.CtTrue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "unhealthy_node_replacement", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
			{
				Config: testAccClusterConfig_unhealthyNodeReplacement(rName, acctest.CtFalse),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "unhealthy_node_replacement", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"cluster_state", // Ignore RUNNING versus WAITING changes
					"configurations",
					"keep_job_flow_alive_when_no_steps",
				},
			},
		},
	})
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_emr_cluster" {
				continue
			}

			_, err := tfemr.FindClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EMR Cluster %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, n string, v *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn(ctx)

		output, err := tfemr.FindClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Id) != aws.StringValue(j.Id) {
			return fmt.Errorf("EMR Cluster recreated: %s -> %s", aws.StringValue(i.Id), aws.StringValue(j.Id))
		}

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Id) == aws.StringValue(j.Id) {
			return fmt.Errorf("EMR Cluster not recreated: %s", aws.StringValue(i.Id))
		}

		return nil
	}
}

func testAccDeleteManagedSecurityGroups(ctx context.Context, conn *ec2.Client, vpc *ec2types.Vpc) error {
	// Reference: https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-man-sec-groups.html
	managedSecurityGroups := map[string]*ec2types.SecurityGroup{
		"ElasticMapReduce-master": nil,
		"ElasticMapReduce-slave":  nil,
	}

	for groupName := range managedSecurityGroups {
		securityGroup, err := tfec2.FindSecurityGroupByNameAndVPCIDAndOwnerID(ctx, conn, groupName, aws.StringValue(vpc.VpcId), aws.StringValue(vpc.OwnerId))

		if err != nil {
			return fmt.Errorf("error describing EMR Managed Security Group (%s): %w", groupName, err)
		}

		managedSecurityGroups[groupName] = securityGroup
	}

	// EMR Managed Security Groups rules reference each other, so rules from all
	// groups must be revoked first.
	for groupName, securityGroup := range managedSecurityGroups {
		if securityGroup == nil {
			continue
		}

		err := testAccRevokeManagedSecurityGroup(ctx, conn, securityGroup)

		if err != nil {
			return fmt.Errorf("error revoking EMR Managed Security Group (%s): %w", groupName, err)
		}
	}

	for groupName, securityGroup := range managedSecurityGroups {
		if securityGroup == nil {
			continue
		}

		err := testAccDeleteManagedSecurityGroup(ctx, conn, securityGroup)

		if err != nil {
			return fmt.Errorf("error deleting EMR Managed Security Group (%s): %w", groupName, err)
		}
	}

	return nil
}

func testAccRevokeManagedSecurityGroup(ctx context.Context, conn *ec2.Client, securityGroup *ec2types.SecurityGroup) error {
	input := &ec2.RevokeSecurityGroupIngressInput{
		GroupId:       securityGroup.GroupId,
		IpPermissions: securityGroup.IpPermissions,
	}

	_, err := conn.RevokeSecurityGroupIngress(ctx, input)

	return err
}

func testAccDeleteManagedSecurityGroup(ctx context.Context, conn *ec2.Client, securityGroup *ec2types.SecurityGroup) error {
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: securityGroup.GroupId,
	}

	_, err := conn.DeleteSecurityGroup(ctx, input)

	return err
}

// Sub-configs (used by other configs)

func testAccClusterConfig_baseVPC(rName string, mapPublicIPOnLaunch bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

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

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port = 0
    protocol  = "-1"
    self      = true
    to_port   = 0
  }

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  tags = {
    Name                                     = %[1]q
    for-use-with-amazon-emr-managed-policies = true
  }

  # EMR will modify ingress rules
  lifecycle {
    ignore_changes = [ingress]
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = %[2]t
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name                                     = %[1]q
    for-use-with-amazon-emr-managed-policies = true
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

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName, mapPublicIPOnLaunch))
}

func testAccClusterConfig_baseIAMInstanceProfile(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_instance_profile" "emr_instance_profile" {
  name = "%[1]s_profile"
  role = aws_iam_role.emr_instance_profile.name
}

resource "aws_iam_role" "emr_instance_profile" {
  name = "%[1]s_profile_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_instance_profile" {
  role       = aws_iam_role.emr_instance_profile.id
  policy_arn = aws_iam_policy.emr_instance_profile.arn
}

resource "aws_iam_policy" "emr_instance_profile" {
  name = "%[1]s_profile"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "cloudwatch:*",
            "dynamodb:*",
            "ec2:Describe*",
            "elasticmapreduce:Describe*",
            "elasticmapreduce:ListBootstrapActions",
            "elasticmapreduce:ListClusters",
            "elasticmapreduce:ListInstanceGroups",
            "elasticmapreduce:ListInstances",
            "elasticmapreduce:ListSteps",
            "kinesis:CreateStream",
            "kinesis:DeleteStream",
            "kinesis:DescribeStream",
            "kinesis:GetRecords",
            "kinesis:GetShardIterator",
            "kinesis:MergeShards",
            "kinesis:PutRecord",
            "kinesis:SplitShard",
            "rds:Describe*",
            "s3:*",
            "sdb:*",
            "sns:*",
            "sqs:*"
        ]
    }]
}
EOT
}
`, rName)
}

func testAccClusterConfig_baseIAMServiceRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "emr_service" {
  name = "%[1]s_default_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceRole"
}

`, rName)
}

func testAccClusterConfig_baseIAMServiceRolev2(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "emr_service" {
  name = "%[1]s_default_role"

  assume_role_policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEMRServicePolicy_v2"
}

`, rName)
}

func testAccClusterConfig_baseIAMAutoScalingRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "emr_autoscaling_role" {
  name               = "%[1]s_autoscaling_role"
  assume_role_policy = data.aws_iam_policy_document.emr_autoscaling_role.json
}

data "aws_iam_policy_document" "emr_autoscaling_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["elasticmapreduce.${data.aws_partition.current.dns_suffix}", "application-autoscaling.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "emr_autoscaling_role" {
  role       = aws_iam_role.emr_autoscaling_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}
`, rName)
}

func testAccClusterConfig_baseBootstrapActionBucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "tester" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "tester" {
  bucket = aws_s3_bucket.tester.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "tester" {
  bucket = aws_s3_bucket.tester.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "tester" {
  depends_on = [
    aws_s3_bucket_public_access_block.tester,
    aws_s3_bucket_ownership_controls.tester,
  ]

  bucket = aws_s3_bucket.tester.id
  acl    = "public-read"
}

resource "aws_s3_object" "testobject" {
  bucket  = aws_s3_bucket_acl.tester.bucket
  key     = "testscript.sh"
  content = <<EOF
#!/bin/bash
echo $@
EOF

  acl = "public-read"
}
`, rName)
}

func testAccClusterConfig_Step(rName string, stepConfig string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  log_uri                           = "s3://${aws_s3_bucket.test.bucket}/"
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn
  termination_protection            = false

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  %[2]s

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}
`, rName, stepConfig))
}

func testAccClusterIAMServiceRoleCustomAMIIDConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "emr_service" {
  name = "%[1]s_default_role"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "elasticmapreduce.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy_attachment" "emr_service" {
  role       = aws_iam_role.emr_service.id
  policy_arn = aws_iam_policy.emr_service.arn
}

resource "aws_iam_policy" "emr_service" {
  name = "%[1]s_emr"

  policy = <<EOT
{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Resource": "*",
        "Action": [
            "ec2:AuthorizeSecurityGroupEgress",
            "ec2:AuthorizeSecurityGroupIngress",
            "ec2:CancelSpotInstanceRequests",
            "ec2:CreateNetworkInterface",
            "ec2:CreateSecurityGroup",
            "ec2:CreateTags",
            "ec2:DeleteNetworkInterface",
            "ec2:DeleteSecurityGroup",
            "ec2:DeleteTags",
            "ec2:DescribeAvailabilityZones",
            "ec2:DescribeAccountAttributes",
            "ec2:DescribeDhcpOptions",
            "ec2:DescribeImages",
            "ec2:DescribeInstanceStatus",
            "ec2:DescribeInstances",
            "ec2:DescribeKeyPairs",
            "ec2:DescribeNetworkAcls",
            "ec2:DescribeNetworkInterfaces",
            "ec2:DescribePrefixLists",
            "ec2:DescribeRouteTables",
            "ec2:DescribeSecurityGroups",
            "ec2:DescribeSpotInstanceRequests",
            "ec2:DescribeSpotPriceHistory",
            "ec2:DescribeSubnets",
            "ec2:DescribeVpcAttribute",
            "ec2:DescribeVpcEndpoints",
            "ec2:DescribeVpcEndpointServices",
            "ec2:DescribeVpcs",
            "ec2:DetachNetworkInterface",
            "ec2:ModifyImageAttribute",
            "ec2:ModifyInstanceAttribute",
            "ec2:RequestSpotInstances",
            "ec2:RevokeSecurityGroupEgress",
            "ec2:RunInstances",
            "ec2:TerminateInstances",
            "ec2:DeleteVolume",
            "ec2:DescribeVolumeStatus",
            "ec2:DescribeVolumes",
            "ec2:DetachVolume",
            "iam:GetRole",
            "iam:GetRolePolicy",
            "iam:ListInstanceProfiles",
            "iam:ListRolePolicies",
            "iam:PassRole",
            "s3:CreateBucket",
            "s3:Get*",
            "s3:List*",
            "sdb:BatchPutAttributes",
            "sdb:Select",
            "sqs:CreateQueue",
            "sqs:Delete*",
            "sqs:GetQueue*",
            "sqs:PurgeQueue",
            "sqs:ReceiveMessage"
        ]
    }]
}
EOT
}
`, rName)
}

const testAccClusterConfig_Step_DebugLoggingStep = `
  # Example from: https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-plan-debugging.html
  step {
    action_on_failure = "TERMINATE_CLUSTER"
    name              = "Setup Hadoop Debugging"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["state-pusher-script"]
    }
  }
`

const testAccClusterConfig_Step_SparkStep = `
  # Example from: https://docs.aws.amazon.com/emr/latest/ReleaseGuide/emr-spark-submit-step.html
  step {
    action_on_failure = "CONTINUE"
    name              = "Spark Step"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["spark-example", "SparkPi", "10"]
    }
  }
`

// Configs

func testAccClusterConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

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

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 21
}
`, rName))
}

func testAccClusterConfig_additionalInfo(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  additional_info = <<EOF
{
	"instanceAwsClientConfiguration": {
		"proxyPort": 8099,
		"proxyHost": "myproxy.example.com"
	}
}
EOF

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

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 21
}
`, rName))
}

func testAccClusterConfig_configurationsJSON(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Hadoop", "Spark"]

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

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations_json = <<EOF
[
   {
     "Classification": "hadoop-env",
     "Configurations": [
       {
         "Classification": "export",
         "Properties": {
           "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
         }
       }
     ],
     "Properties": {}
   },
   {
     "Classification": "spark-env",
     "Configurations": [
       {
         "Classification": "export",
         "Properties": {
           "JAVA_HOME": "/usr/lib/jvm/java-1.8.0"
         }
       }
     ],
     "Properties": {}
   }
]
EOF

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  service_role         = aws_iam_role.emr_service.arn
  ebs_root_volume_size = 21
}
`, rName))
}

func testAccClusterConfig_coreInstanceGroupAutoScalingPolicy(rName, autoscalingPolicy string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = [
        "application-autoscaling.${data.aws_partition.current.dns_suffix}",
        "elasticmapreduce.${data.aws_partition.current.dns_suffix}",
      ]
      type = "Service"
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  autoscaling_role                  = aws_iam_role.test.arn
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    autoscaling_policy = <<POLICY
%[2]s
POLICY
    instance_type      = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, autoscalingPolicy))
}

func testAccClusterConfig_coreInstanceGroupAutoScalingPolicyRemoved(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      identifiers = [
        "application-autoscaling.${data.aws_partition.current.dns_suffix}",
        "elasticmapreduce.${data.aws_partition.current.dns_suffix}",
      ]
      type = "Service"
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonElasticMapReduceforAutoScalingRole"
}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  autoscaling_role                  = aws_iam_role.test.arn
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName))
}

func testAccClusterConfig_coreInstanceGroupBidPrice(rName, bidPrice string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, bidPrice))
}

func testAccClusterConfig_coreInstanceGroupInstanceCount(rName string, instanceCount int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_count = %[2]d
    instance_type  = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, instanceCount))
}

func testAccClusterConfig_coreInstanceGroupInstanceType(rName, instanceType string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = %[2]q
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, instanceType))
}

func testAccClusterConfig_coreInstanceGroupName(rName, instanceGroupName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, instanceGroupName))
}

func testAccClusterConfig_ec2AttributesDefaultManagedSecurityGroups(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.33.1"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    instance_profile = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id        = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName))
}

func testAccClusterConfig_kerberosDedicatedKdc(rName string, password string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_security_configuration" "test" {
  configuration = <<EOF
{
  "AuthenticationConfiguration": {
    "KerberosConfiguration": {
      "Provider": "ClusterDedicatedKdc",
      "ClusterDedicatedKdcConfiguration": {
        "TicketLifetimeInHours": 24
      }
    }
  }
}
EOF
}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  security_configuration            = aws_emr_security_configuration.test.name
  service_role                      = aws_iam_role.emr_service.arn
  termination_protection            = false

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  kerberos_attributes {
    kdc_admin_password = "%[2]s"
    realm              = "EC2.INTERNAL"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, password))
}

func testAccClusterConfig_masterInstanceGroupBidPrice(rName, bidPrice string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, bidPrice))
}

func testAccClusterConfig_masterInstanceGroupInstanceCount(rName string, instanceCount int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, true),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.24.1"
  service_role                      = aws_iam_role.emr_service.arn

  # Termination protection is automatically enabled for multiple master clusters
  termination_protection = false

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_count = %[2]d
    instance_type  = "m4.large"
  }

  # core_instance_group is required with multiple masters
  core_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, instanceCount))
}

func testAccClusterConfig_masterInstanceGroupInstanceType(rName, instanceType string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = %[2]q
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, instanceType))
}

func testAccClusterConfig_masterInstanceGroupName(rName, instanceGroupName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, instanceGroupName))
}

func testAccClusterConfig_securityConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.5.0"
  applications  = ["Spark"]

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

  security_configuration = aws_emr_security_configuration.test.name

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}

resource "aws_emr_security_configuration" "test" {
  configuration = <<EOF
{
  "EncryptionConfiguration": {
    "AtRestEncryptionConfiguration": {
      "S3EncryptionConfiguration": {
        "EncryptionMode": "SSE-S3"
      },
      "LocalDiskEncryptionConfiguration": {
        "EncryptionKeyProviderType": "AwsKms",
        "AwsKmsKey": "${aws_kms_key.test.arn}"
      }
    },
    "EnableInTransitEncryption": false,
    "EnableAtRestEncryption": true
  }
}
EOF
}

resource "aws_kms_key" "test" {
  description             = "Terraform %[1]s"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName))
}

func testAccClusterConfig_stepSingle(rName string) string {
	return testAccClusterConfig_Step(rName, testAccClusterConfig_Step_DebugLoggingStep)
}

func testAccClusterConfig_stepNoBlocks(rName string) string {
	return testAccClusterConfig_Step(rName, "")
}

func testAccClusterConfig_stepZeroed(rName string) string {
	return testAccClusterConfig_Step(rName, "step = []")
}

func testAccClusterConfig_stepMultiple(rName string) string {
	stepConfig := acctest.ConfigCompose(testAccClusterConfig_Step_DebugLoggingStep, testAccClusterConfig_Step_SparkStep)
	return testAccClusterConfig_Step(rName, stepConfig)
}

func testAccClusterConfig_stepMultipleListStates(rName string) string {
	stepConfig := acctest.ConfigCompose(
		testAccClusterConfig_Step_DebugLoggingStep,
		testAccClusterConfig_Step_SparkStep,
		"\n", `list_steps_states = ["PENDING", "RUNNING", "COMPLETED"]`,
	)
	return testAccClusterConfig_Step(rName, stepConfig)
}

func testAccClusterConfig_bootstrap(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseBootstrapActionBucket(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.0.0"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  bootstrap_action {
    path = "s3://${aws_s3_object.testobject.bucket}/${aws_s3_object.testobject.key}"
    name = "test"

    args = [
      "0",
      "1",
      "",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
    ]
  }
}
`, rName))
}

func testAccClusterConfig_bootstrapAdd(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseBootstrapActionBucket(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.0.0"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  bootstrap_action {
    path = "s3://${aws_s3_object.testobject.bucket}/${aws_s3_object.testobject.key}"
    name = "test"

    args = [
      "0",
      "1",
      "",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
    ]
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif-2"
    args = ["instance.isMaster=true", "echo also running on master node"]
  }
}
`, rName))
}

func testAccClusterConfig_bootstrapReorder(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseBootstrapActionBucket(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.0.0"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif-2"
    args = ["instance.isMaster=true", "echo also running on master node"]
  }

  bootstrap_action {
    path = "s3://${aws_s3_object.testobject.bucket}/${aws_s3_object.testobject.key}"
    name = "test"

    args = [
      "0",
      "1",
      "",
      "3",
      "4",
      "5",
      "6",
      "7",
      "8",
      "9",
      "10",
    ]
  }
}
`, rName))
}

func testAccClusterConfig_terminationPolicy(rName string, term string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

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

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = %[2]s

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, rName, term))
}

func testAccClusterConfig_keepJob(rName string, keepJob bool) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

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

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = %[2]t
  termination_protection            = false

  step {
    action_on_failure = "CONTINUE"
    name              = "Sleep Step"

    hadoop_jar_step {
      jar  = "command-runner.jar"
      args = ["/bin/sleep", "60"]
    }
  }

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, rName, keepJob))
}

func testAccClusterConfig_visibleToAllUsersUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

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

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  visible_to_all_users              = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, rName))
}

func testAccClusterConfig_s3Logging(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterIAMServiceRoleCustomAMIIDConfig(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

  termination_protection            = false
  keep_job_flow_alive_when_no_steps = true

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  log_uri = "s3://${aws_s3_bucket.test.bucket}/"

  ec2_attributes {
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    subnet_id                         = aws_subnet.test.id
  }

  service_role = aws_iam_role.emr_service.arn
}

data "aws_caller_identity" "current" {}
`, rName))
}

func testAccClusterConfig_s3Encryption(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterIAMServiceRoleCustomAMIIDConfig(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kms_key" "test" {
  description = "Terraform acc test %[1]s"

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.32.0"
  applications  = ["Spark"]

  termination_protection            = false
  keep_job_flow_alive_when_no_steps = true

  master_instance_group {
    instance_type = "c4.large"
  }

  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }

  log_encryption_kms_key_id = aws_kms_key.test.arn
  log_uri                   = "s3://${aws_s3_bucket.test.bucket}/"

  ec2_attributes {
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    subnet_id                         = aws_subnet.test.id
  }

  service_role = aws_iam_role.emr_service.arn
}

data "aws_caller_identity" "current" {}
`, rName))
}

func testAccClusterConfig_updatedTags(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

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

  tags = {
    dns_zone = "new_zone"
    Env      = "production"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role     = aws_iam_role.emr_service.arn
  autoscaling_role = aws_iam_role.emr_autoscaling_role.arn
}
`, rName))
}

func testAccClusterConfig_updatedRootVolumeSize(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-4.6.0"
  applications  = ["Spark"]

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

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 48
}
`, rName))
}

func testAccClusterConfig_stepConcurrencyLevel(rName string, stepConcurrencyLevel int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.28.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  step_concurrency_level = %[2]d

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, stepConcurrencyLevel))
}

func testAccClusterConfig_ebs(rName string, volumesPerInstance int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.28.0"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m5.xlarge"
    ebs_config {
      size                 = 32
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
    ebs_config {
      size                 = 50
      throughput           = 500
      type                 = "gp3"
      volumes_per_instance = %[2]d
    }
  }
  core_instance_group {
    instance_count = 1
    instance_type  = "m5.xlarge"
    ebs_config {
      size                 = 32
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
    ebs_config {
      size                 = 125
      type                 = "sc1"
      volumes_per_instance = %[2]d
    }
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, volumesPerInstance))
}

func testAccClusterConfig_customAMIID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterIAMServiceRoleCustomAMIIDConfig(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.7.0"
  applications  = ["Spark"]

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

  tags = {
    role     = "rolename"
    dns_zone = "env_zone"
    env      = "env"
    name     = "name-env"
  }

  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  configurations = "test-fixtures/emr_configurations.json"

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
  ]

  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 48
  custom_ami_id        = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
}
`, rName))
}

func testAccClusterConfig_instanceFleets(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseBootstrapActionBucket(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.30.1"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_fleet {
    instance_type_configs {
      instance_type = "m3.xlarge"
    }

    target_on_demand_capacity = 1
  }
  core_instance_fleet {
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 80
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m3.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.2xlarge"
      weighted_capacity = 2
    }
    launch_specifications {
      spot_specification {
        allocation_strategy      = "capacity-optimized"
        block_duration_minutes   = 0
        timeout_action           = "SWITCH_TO_ON_DEMAND"
        timeout_duration_minutes = 10
      }
    }
    name                      = "core fleet"
    target_on_demand_capacity = 0
    target_spot_capacity      = 2
  }
  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }
}
`, rName))
}

func testAccClusterConfig_instanceFleetMultipleSubnets(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseBootstrapActionBucket(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.30.1"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_fleet {
    instance_type_configs {
      instance_type = "m3.xlarge"
    }

    target_on_demand_capacity = 1
  }
  core_instance_fleet {
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 80
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m3.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.xlarge"
      weighted_capacity = 1
    }
    instance_type_configs {
      bid_price_as_percentage_of_on_demand_price = 100
      ebs_config {
        size                 = 100
        type                 = "gp2"
        volumes_per_instance = 1
      }
      instance_type     = "m4.2xlarge"
      weighted_capacity = 2
    }
    launch_specifications {
      spot_specification {
        allocation_strategy      = "price-capacity-optimized"
        block_duration_minutes   = 0
        timeout_action           = "SWITCH_TO_ON_DEMAND"
        timeout_duration_minutes = 10
      }
    }
    name                      = "core fleet"
    target_on_demand_capacity = 0
    target_spot_capacity      = 2
  }
  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_ids                        = [aws_subnet.test.id, aws_subnet.test2.id]
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }
}

resource "aws_subnet" "test2" {
  availability_zone       = data.aws_availability_zones.available.names[1]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  map_public_ip_on_launch = false
  vpc_id                  = aws_vpc.test.id
}

resource "aws_route_table_association" "test2" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test2.id
}
`, rName))
}

func testAccClusterConfig_instanceFleetsMasterOnly(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseBootstrapActionBucket(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.30.1"
  applications  = ["Hadoop", "Hive"]
  log_uri       = "s3n://terraform/testlog/"

  master_instance_fleet {
    instance_type_configs {
      instance_type = "m3.xlarge"
    }

    target_on_demand_capacity = 1
  }
  service_role = aws_iam_role.emr_service.arn
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]

  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }

  bootstrap_action {
    path = "s3://elasticmapreduce/bootstrap-actions/run-if"
    name = "runif"
    args = ["instance.isMaster=true", "echo running on master node"]
  }
}
`, rName))
}

func testAccClusterConfig_autoTermination(rName string, timeout int) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  auto_termination_policy {
    idle_timeout = %[2]d
  }

  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.33.1"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, timeout))
}

func testAccClusterConfig_noAutoTermination(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.33.1"
  service_role                      = aws_iam_role.emr_service.arn

  ec2_attributes {
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName))
}

func testAccClusterConfig_IAMServiceRoleWithPlacementGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseIAMServiceRolev2(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy_attachment" "emr_placementgroup" {
  role       = aws_iam_role.emr_service.id
  policy_arn = aws_iam_policy.emr_placementgroup.arn
}

resource "aws_iam_policy" "emr_placementgroup" {
  name = "%[1]s_placementgroup_profile"

  policy = <<EOT
{
  "Version": "2012-10-17",
  "Statement": [{
      "Effect": "Allow",
      "Resource": "arn:${data.aws_partition.current.partition}:ec2:*:*:placement-group/pg-*",
      "Action": [
          "ec2:CreatePlacementGroup",
          "ec2:CreateTags",
          "ec2:DeleteTags"
      ]
  },
  {
      "Sid": "PassRoleForEC2",
      "Effect": "Allow",
      "Action": "iam:PassRole",
      "Resource": "${aws_iam_role.emr_instance_profile.arn}",
      "Condition": {
          "StringLike": {
              "iam:PassedToService": "ec2.${data.aws_partition.current.dns_suffix}*"
          }
      }
  }]
}
EOT
}
`, rName))
}

func testAccClusterConfig_PlacementGroup(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, true),
		testAccClusterConfig_IAMServiceRoleWithPlacementGroup(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.23.0"
  applications  = ["Spark"]
  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }
  master_instance_group {
    instance_type  = "c4.large"
    instance_count = 3
  }
  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }
  tags = {
    role                                     = "rolename"
    dns_zone                                 = "env_zone"
    env                                      = "env"
    name                                     = "name-env"
    for-use-with-amazon-emr-managed-policies = true
  }
  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"
  configurations      = "test-fixtures/emr_configurations.json"

  placement_group_config {
    instance_role      = "MASTER"
    placement_strategy = "SPREAD"
  }
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_placementgroup
  ]
  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 21
}
`, rName))
}

func testAccClusterConfig_PlacementGroupWithOptionalUnset(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, true),
		testAccClusterConfig_IAMServiceRoleWithPlacementGroup(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		testAccClusterConfig_baseIAMAutoScalingRole(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  name          = %[1]q
  release_label = "emr-5.23.0"
  applications  = ["Spark"]
  ec2_attributes {
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
  }
  master_instance_group {
    instance_type  = "c4.large"
    instance_count = 3
  }
  core_instance_group {
    instance_count = 1
    instance_type  = "c4.large"
  }
  tags = {
    role                                     = "rolename"
    dns_zone                                 = "env_zone"
    env                                      = "env"
    name                                     = "name-env"
    for-use-with-amazon-emr-managed-policies = true
  }
  keep_job_flow_alive_when_no_steps = true
  termination_protection            = false

  scale_down_behavior = "TERMINATE_AT_TASK_COMPLETION"
  configurations      = "test-fixtures/emr_configurations.json"

  placement_group_config {
    instance_role = "MASTER"
  }
  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_instance_profile,
    aws_iam_role_policy_attachment.emr_autoscaling_role,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_placementgroup
  ]
  service_role         = aws_iam_role.emr_service.arn
  autoscaling_role     = aws_iam_role.emr_autoscaling_role.arn
  ebs_root_volume_size = 21
}
`, rName))
}

func testAccClusterConfig_unhealthyNodeReplacement(rName, unhealthyNodeReplacement string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_baseVPC(rName, false),
		testAccClusterConfig_baseIAMServiceRole(rName),
		testAccClusterConfig_baseIAMInstanceProfile(rName),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.33.1"
  service_role                      = aws_iam_role.emr_service.arn
  termination_protection            = false
  unhealthy_node_replacement        = %[2]s

  ec2_attributes {
    instance_profile                  = aws_iam_instance_profile.emr_instance_profile.arn
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [
    aws_route_table_association.test,
    aws_iam_role_policy_attachment.emr_service,
    aws_iam_role_policy_attachment.emr_instance_profile,
  ]
}
`, rName, unhealthyNodeReplacement))
}
