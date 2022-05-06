package emr_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/emr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfemr "github.com/hashicorp/terraform-provider-aws/internal/service/emr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEMRCluster_basic(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "elasticmapreduce", regexp.MustCompile("cluster/.+$")),
					resource.TestCheckResourceAttr(resourceName, "release_label", "emr-4.6.0"),
					resource.TestCheckResourceAttr(resourceName, "applications.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "applications.*", "Spark"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.subnet_id", "aws_subnet.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.emr_managed_master_security_group", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.emr_managed_slave_security_group", "aws_security_group.test", "id"),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.instance_profile", "aws_iam_instance_profile.emr_instance_profile", "arn"),
					resource.TestCheckResourceAttr(resourceName, "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr(resourceName, "ebs_root_volume_size", "21"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", "aws_iam_role.emr_service", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "autoscaling_role", "aws_iam_role.emr_autoscaling_role", "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "additional_info"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "step.#", "0"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionNot(t, endpoints.AwsUsGovPartitionID) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterAutoTerminationConfig(rName, 10000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", "1"),
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
				Config: testAccClusterAutoTerminationConfig(rName, 20000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.0.idle_timeout", "20000"),
				),
			},
			{
				Config: testAccClusterNoAutoTerminationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", "0"),
				),
			},
			{
				Config: testAccClusterAutoTerminationConfig(rName, 20000),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_termination_policy.0.idle_timeout", "20000"),
				),
			},
		},
	})
}

func TestAccEMRCluster_additionalInfo(t *testing.T) {
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterAdditionalInfoConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "scale_down_behavior", "TERMINATE_AT_TASK_COMPLETION"),
					resource.TestCheckResourceAttr(resourceName, "step.#", "0"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					acctest.CheckResourceDisappears(acctest.Provider, tfemr.ResourceCluster(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfemr.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEMRCluster_sJSON(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfigurationsJSONConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestMatchResourceAttr(resourceName, "configurations_json",
						regexp.MustCompile("{\"JAVA_HOME\":\"/usr/lib/jvm/java-1.8.0\".+")),
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
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCoreInstanceGroupAutoScalingPolicyConfig(rName, autoscalingPolicy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
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
				Config: testAccClusterCoreInstanceGroupAutoScalingPolicyConfig(rName, autoscalingPolicy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					acctest.CheckResourceAttrEquivalentJSON(resourceName, "core_instance_group.0.autoscaling_policy", autoscalingPolicy2),
				),
			},
			{
				Config: testAccClusterCoreInstanceGroupAutoScalingPolicyRemovedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.autoscaling_policy", ""),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_bidPrice(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCoreInstanceGroupBidPriceConfig(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
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
				Config: testAccClusterCoreInstanceGroupBidPriceConfig(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_instanceCount(t *testing.T) {
	var cluster1, cluster2, cluster3 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCoreInstanceGroupInstanceCountConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "2"),
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
				Config: testAccClusterCoreInstanceGroupInstanceCountConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterNotRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "1"),
				),
			},
			{
				Config: testAccClusterCoreInstanceGroupInstanceCountConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster3),
					testAccCheckClusterNotRecreated(&cluster2, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_count", "2"),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_instanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCoreInstanceGroupInstanceTypeConfig(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
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
				Config: testAccClusterCoreInstanceGroupInstanceTypeConfig(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccEMRCluster_CoreInstanceGroup_name(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCoreInstanceGroupNameConfig(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
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
				Config: testAccClusterCoreInstanceGroupNameConfig(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccEMRCluster_EC2Attributes_defaultManagedSecurityGroups(t *testing.T) {
	var cluster emr.Cluster
	var vpc ec2.Vpc

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterEC2AttributesDefaultManagedSecurityGroupsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					acctest.CheckVPCExists(vpcResourceName, &vpc),
				),
			},
			{
				Config:      testAccClusterEC2AttributesDefaultManagedSecurityGroupsConfig(rName),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`DependencyViolation`),
			},
			{
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

					err := testAccDeleteManagedSecurityGroups(conn, &vpc)

					if err != nil {
						t.Fatal(err)
					}
				},
				Config:  testAccClusterEC2AttributesDefaultManagedSecurityGroupsConfig(rName),
				Destroy: true,
			},
		},
	})
}

func TestAccEMRCluster_Kerberos_clusterDedicatedKdc(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	password := fmt.Sprintf("NeverKeepPasswordsInPlainText%s!", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Kerberos_ClusterDedicatedKdc(rName, password),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "kerberos_attributes.#", "1"),
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
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterMasterInstanceGroupBidPriceConfig(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
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
				Config: testAccClusterMasterInstanceGroupBidPriceConfig(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.bid_price", "0.51"),
				),
			},
		},
	})
}

func TestAccEMRCluster_MasterInstanceGroup_instanceCount(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterMasterInstanceGroupInstanceCountConfig(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_count", "3"),
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
				Config: testAccClusterMasterInstanceGroupInstanceCountConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_count", "1"),
				),
			},
		},
	})
}

func TestAccEMRCluster_MasterInstanceGroup_instanceType(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterMasterInstanceGroupInstanceTypeConfig(rName, "m4.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
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
				Config: testAccClusterMasterInstanceGroupInstanceTypeConfig(rName, "m4.xlarge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.instance_type", "m4.xlarge"),
				),
			},
		},
	})
}

func TestAccEMRCluster_MasterInstanceGroup_name(t *testing.T) {
	var cluster1, cluster2 emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterMasterInstanceGroupNameConfig(rName, "name1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
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
				Config: testAccClusterMasterInstanceGroupNameConfig(rName, "name2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.name", "name2"),
				),
			},
		},
	})
}

func TestAccEMRCluster_security(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_SecurityConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttrPair(resourceName, "security_configuration", "aws_emr_security_configuration.test", "name"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Step_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.main_class", ""),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.properties.%", "0"),
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
	var cluster1, cluster2, cluster3 emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Step_Single(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
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
				Config: testAccClusterConfig_Step_NoBlocks(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "step.#", "1"),
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
				Config: testAccClusterConfig_Step_Zeroed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "step.#", "0"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Step_Multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
					resource.TestCheckResourceAttr(resourceName, "step.1.action_on_failure", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.0", "spark-example"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.1", "SparkPi"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.2", "10"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_Step_Multiple_ListStates(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "step.0.action_on_failure", "TERMINATE_CLUSTER"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.args.0", "state-pusher-script"),
					resource.TestCheckResourceAttr(resourceName, "step.0.hadoop_jar_step.0.jar", "command-runner.jar"),
					resource.TestCheckResourceAttr(resourceName, "step.0.name", "Setup Hadoop Debugging"),
					resource.TestCheckResourceAttr(resourceName, "step.1.action_on_failure", "CONTINUE"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.0", "spark-example"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.1", "SparkPi"),
					resource.TestCheckResourceAttr(resourceName, "step.1.hadoop_jar_step.0.args.2", "10"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix("tf-emr-bootstrap")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_bootstrap(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "10"),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "10"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.name", "runif-2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.#", "2"),
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
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.name", "runif"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.0", "instance.isMaster=true"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.0.args.1", "echo running on master node"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.path", fmt.Sprintf("s3://%s/testscript.sh", rName)),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.2.args.#", "10"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.name", "runif-2"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.path", "s3://elasticmapreduce/bootstrap-actions/run-if"),
					resource.TestCheckResourceAttr(resourceName, "bootstrap_action.1.args.#", "2"),
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

func TestAccEMRCluster_terminationProtected(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterTerminationPolicyConfig(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
				),
			},
			{
				Config: testAccClusterTerminationPolicyConfig(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "true"),
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
				Config: testAccClusterTerminationPolicyConfig(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "termination_protection", "false"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_keepJob(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "keep_job_flow_alive_when_no_steps", "false"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "visible_to_all_users", "true"),
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
				Config: testAccClusterVisibleToAllUsersUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "visible_to_all_users", "false"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := fmt.Sprintf("s3n://%s/", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterS3LoggingConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	bucketName := fmt.Sprintf("s3n://%s/", rName)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterS3EncryptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "log_uri", bucketName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "log_encryption_kms_key_id", "kms", regexp.MustCompile(`key/.+`)),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.role", "rolename"),
					resource.TestCheckResourceAttr(resourceName, "tags.dns_zone", "env_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "env"),
					resource.TestCheckResourceAttr(resourceName, "tags.name", "name-env")),
			},
			{
				Config: testAccClusterUpdatedTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "ebs_root_volume_size", "21"),
				),
			},
			{
				Config: testAccClusterUpdatedRootVolumeSizeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
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
	var cluster emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterStepConcurrencyLevelConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step_concurrency_level", "2"),
				),
			},
			{
				Config: testAccClusterStepConcurrencyLevelConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "step_concurrency_level", "1"),
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
	var cluster emr.Cluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_emr_cluster.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.0.ebs_config.0.volumes_per_instance", "2"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.0.ebs_config.0.volumes_per_instance", "2"),
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
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterCustomAMIIDConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
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
	var cluster1, cluster2 emr.Cluster

	resourceName := "aws_emr_cluster.test"
	subnetResourceName := "aws_subnet.test"
	subnet2ResourceName := "aws_subnet.test2"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_InstanceFleets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_attributes.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.0.subnet_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ec2_attributes.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "0"),
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
				Config: testAccClusterConfig_InstanceFleet_MultipleSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster2),
					testAccCheckClusterRecreated(&cluster1, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ec2_attributes.0.subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ec2_attributes.0.subnet_ids.*", subnetResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "ec2_attributes.0.subnet_ids.*", subnet2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "master_instance_group.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_group.#", "0"),
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

func TestAccEMRCluster_InstanceFleetMaster_only(t *testing.T) {
	var cluster emr.Cluster

	resourceName := "aws_emr_cluster.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterInstanceFleetsMasterOnlyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "master_instance_fleet.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "core_instance_fleet.#", "0"),
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

func testAccCheckDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_emr_cluster" {
			continue
		}

		_, err := tfemr.FindClusterByID(conn, rs.Primary.ID)

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

func testAccCheckClusterExists(n string, v *emr.Cluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EMR Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EMRConn

		output, err := tfemr.FindClusterByID(conn, rs.Primary.ID)

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

func testAccDeleteManagedSecurityGroups(conn *ec2.EC2, vpc *ec2.Vpc) error {
	// Reference: https://docs.aws.amazon.com/emr/latest/ManagementGuide/emr-man-sec-groups.html
	managedSecurityGroups := map[string]*ec2.SecurityGroup{
		"ElasticMapReduce-master": nil,
		"ElasticMapReduce-slave":  nil,
	}

	for groupName := range managedSecurityGroups {
		securityGroup, err := tfec2.FindSecurityGroupByNameAndVPCID(conn, groupName, aws.StringValue(vpc.VpcId))

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

		err := testAccRevokeManagedSecurityGroup(conn, securityGroup)

		if err != nil {
			return fmt.Errorf("error revoking EMR Managed Security Group (%s): %w", groupName, err)
		}
	}

	for groupName, securityGroup := range managedSecurityGroups {
		if securityGroup == nil {
			continue
		}

		err := testAccDeleteManagedSecurityGroup(conn, securityGroup)

		if err != nil {
			return fmt.Errorf("error deleting EMR Managed Security Group (%s): %w", groupName, err)
		}
	}

	return nil
}

func testAccRevokeManagedSecurityGroup(conn *ec2.EC2, securityGroup *ec2.SecurityGroup) error {
	input := &ec2.RevokeSecurityGroupIngressInput{
		GroupId:       securityGroup.GroupId,
		IpPermissions: securityGroup.IpPermissions,
	}

	_, err := conn.RevokeSecurityGroupIngress(input)

	return err
}

func testAccDeleteManagedSecurityGroup(conn *ec2.EC2, securityGroup *ec2.SecurityGroup) error {
	input := &ec2.DeleteSecurityGroupInput{
		GroupId: securityGroup.GroupId,
	}

	_, err := conn.DeleteSecurityGroup(input)

	return err
}

// Sub-configs (used by other configs)

func testAccClusterBaseVPCConfig(rName string, mapPublicIPOnLaunch bool) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # Many instance types are not available in this availability zone
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

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
    Name = %[1]q
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
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName, mapPublicIPOnLaunch)
}

func testAccClusterIAMInstanceProfileBaseConfig(rName string) string {
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

func testAccClusterIAMServiceRoleBaseConfig(rName string) string {
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

func testAccClusterIAMAutoScalingRoleConfig(rName string) string {
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

func testAccClusterBootstrapActionBucketConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "tester" {
  bucket = %[1]q
}

resource "aws_s3_bucket_acl" "tester" {
  bucket = aws_s3_bucket.tester.id
  acl    = "public-read"
}

resource "aws_s3_object" "testobject" {
  bucket  = aws_s3_bucket.tester.bucket
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
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  log_uri                           = "s3://${aws_s3_bucket.test.bucket}/"
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"
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
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  %[2]s

  depends_on = [aws_route_table_association.test]
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

func testAccClusterConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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

func testAccClusterAdditionalInfoConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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

func testAccClusterConfigurationsJSONConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
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

func testAccClusterCoreInstanceGroupAutoScalingPolicyConfig(rName, autoscalingPolicy string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
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
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
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
  ]
}
`, rName, autoscalingPolicy))
}

func testAccClusterCoreInstanceGroupAutoScalingPolicyRemovedConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
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
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
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
  ]
}
`, rName))
}

func testAccClusterCoreInstanceGroupBidPriceConfig(rName, bidPrice string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, bidPrice))
}

func testAccClusterCoreInstanceGroupInstanceCountConfig(rName string, instanceCount int) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_count = %[2]d
    instance_type  = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceCount))
}

func testAccClusterCoreInstanceGroupInstanceTypeConfig(rName, instanceType string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceType))
}

func testAccClusterCoreInstanceGroupNameConfig(rName, instanceGroupName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  core_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceGroupName))
}

func testAccClusterEC2AttributesDefaultManagedSecurityGroupsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.33.1"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    instance_profile = "EMR_EC2_DefaultRole"
    subnet_id        = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName))
}

func testAccClusterConfig_Kerberos_ClusterDedicatedKdc(rName string, password string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
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
  service_role                      = "EMR_DefaultRole"
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
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  kerberos_attributes {
    kdc_admin_password = "%[2]s"
    realm              = "EC2.INTERNAL"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, password))
}

func testAccClusterMasterInstanceGroupBidPriceConfig(rName, bidPrice string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    bid_price     = %[2]q
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, bidPrice))
}

func testAccClusterMasterInstanceGroupInstanceCountConfig(rName string, instanceCount int) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, true),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.24.1"
  service_role                      = "EMR_DefaultRole"

  # Termination protection is automatically enabled for multiple master clusters
  termination_protection = false

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
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

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceCount))
}

func testAccClusterMasterInstanceGroupInstanceTypeConfig(rName, instanceType string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceType))
}

func testAccClusterMasterInstanceGroupNameConfig(rName, instanceGroupName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.12.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
    name          = %[2]q
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, instanceGroupName))
}

func testAccClusterConfig_SecurityConfiguration(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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

func testAccClusterConfig_Step_Single(rName string) string {
	return testAccClusterConfig_Step(rName, testAccClusterConfig_Step_DebugLoggingStep)
}

func testAccClusterConfig_Step_NoBlocks(rName string) string {
	return testAccClusterConfig_Step(rName, "")
}

func testAccClusterConfig_Step_Zeroed(rName string) string {
	return testAccClusterConfig_Step(rName, "step = []")
}

func testAccClusterConfig_Step_Multiple(rName string) string {
	stepConfig := acctest.ConfigCompose(testAccClusterConfig_Step_DebugLoggingStep, testAccClusterConfig_Step_SparkStep)
	return testAccClusterConfig_Step(rName, stepConfig)
}

func testAccClusterConfig_Step_Multiple_ListStates(rName string) string {
	stepConfig := acctest.ConfigCompose(
		testAccClusterConfig_Step_DebugLoggingStep,
		testAccClusterConfig_Step_SparkStep,
		"\n", `list_steps_states = ["PENDING", "RUNNING", "COMPLETED"]`,
	)
	return testAccClusterConfig_Step(rName, stepConfig)
}

func testAccClusterConfig_bootstrap(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterBootstrapActionBucketConfig(rName),
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

    args = ["1",
      "2",
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
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterBootstrapActionBucketConfig(rName),
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

    args = ["1",
      "2",
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
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterBootstrapActionBucketConfig(rName),
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

    args = ["1",
      "2",
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

func testAccClusterTerminationPolicyConfig(rName string, term string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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

func testAccClusterVisibleToAllUsersUpdatedConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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

func testAccClusterS3LoggingConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleCustomAMIIDConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
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

func testAccClusterS3EncryptionConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleCustomAMIIDConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
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

func testAccClusterUpdatedTagsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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

func testAccClusterUpdatedRootVolumeSizeConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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

func testAccClusterStepConcurrencyLevelConfig(rName string, stepConcurrencyLevel int) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.28.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  step_concurrency_level = %[2]d

  depends_on = [aws_route_table_association.test]
}
`, rName, stepConcurrencyLevel))
}

func testAccEBSConfig(rName string, volumesPerInstance int) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.28.0"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
    ebs_config {
      size                 = 32
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
    ebs_config {
      size                 = 50
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
  }
  core_instance_group {
    instance_count = 1
    instance_type  = "m4.large"
    ebs_config {
      size                 = 32
      type                 = "gp2"
      volumes_per_instance = %[2]d
    }
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, volumesPerInstance))
}

func testAccClusterCustomAMIIDConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleCustomAMIIDConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterIAMAutoScalingRoleConfig(rName),
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
  custom_ami_id        = data.aws_ami.emr-custom-ami.id
}

data "aws_ami" "emr-custom-ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}
`, rName))
}

func testAccClusterConfig_InstanceFleets(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterBootstrapActionBucketConfig(rName),
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

func testAccClusterConfig_InstanceFleet_MultipleSubnets(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterBootstrapActionBucketConfig(rName),
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

func testAccClusterInstanceFleetsMasterOnlyConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		testAccClusterIAMServiceRoleBaseConfig(rName),
		testAccClusterIAMInstanceProfileBaseConfig(rName),
		testAccClusterBootstrapActionBucketConfig(rName),
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

func testAccClusterAutoTerminationConfig(rName string, timeout int) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
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
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName, timeout))
}

func testAccClusterNoAutoTerminationConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterBaseVPCConfig(rName, false),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_emr_cluster" "test" {
  applications                      = ["Spark"]
  keep_job_flow_alive_when_no_steps = true
  name                              = %[1]q
  release_label                     = "emr-5.33.1"
  service_role                      = "EMR_DefaultRole"

  ec2_attributes {
    instance_profile                  = "EMR_EC2_DefaultRole"
    subnet_id                         = aws_subnet.test.id
    emr_managed_master_security_group = aws_security_group.test.id
    emr_managed_slave_security_group  = aws_security_group.test.id
  }

  master_instance_group {
    instance_type = "m4.large"
  }

  depends_on = [aws_route_table_association.test]
}
`, rName))
}
