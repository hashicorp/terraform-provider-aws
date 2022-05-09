package batch_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/batch"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccBatchComputeEnvironment_basic(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "UNMANAGED"),
				),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_disappears(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceDisappears(acctest.Provider, tfbatch.ResourceComputeEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBatchComputeEnvironment_nameGenerated(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentNameGeneratedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					create.TestCheckResourceAttrNameGenerated(resourceName, "compute_environment_name"),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", "terraform-"),
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

func TestAccBatchComputeEnvironment_namePrefix(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentNamePrefixConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					create.TestCheckResourceAttrNameFromPrefix(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", rName),
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

func TestAccBatchComputeEnvironment_createEC2(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentEC2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_CreateEC2DesiredVCPUsEC2KeyPairImageID_computeResourcesTags(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	resourceName := "aws_batch_compute_environment.test"
	amiDatasourceName := "data.aws_ami.amzn-ami-minimal-hvm-ebs"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	keyPairResourceName := "aws_key_pair.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentEC2WithDesiredVcpusEC2KeyPairImageIdAndComputeResourcesTagsConfig(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "8"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.ec2_key_pair", keyPairResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.image_id", amiDatasourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_createSpot(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentSpotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_CreateSpotAllocationStrategy_bidPercentage(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentSpotWithAllocationStrategyAndBidPercentageConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "60"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_createFargate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentFargateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_createFargateSpot(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentFargateSpotConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE_SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_updateState(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentStateConfig(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "UNMANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentStateConfig(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "DISABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "UNMANAGED"),
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

func TestAccBatchComputeEnvironment_updateServiceRole(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName1 := "aws_iam_role.batch_service"
	serviceRoleResourceName2 := "aws_iam_role.batch_service_2"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentFargateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName1, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentFargateUpdatedServiceRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName2, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_defaultServiceRole(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
			acctest.PreCheckIAMServiceLinkedRole(t, "/aws-service-role/batch")
		},
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentFargateDefaultServiceRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "service_role", "iam", regexp.MustCompile(`role/aws-service-role/batch`)),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_ComputeResources_minVCPUs(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentComputeResourcesMaxVcpusMinVcpusConfig(rName, 4, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentComputeResourcesMaxVcpusMinVcpusConfig(rName, 4, 4),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentComputeResourcesMaxVcpusMinVcpusConfig(rName, 4, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_ComputeResources_maxVCPUs(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentComputeResourcesMaxVcpusMinVcpusConfig(rName, 4, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "4"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentComputeResourcesMaxVcpusMinVcpusConfig(rName, 8, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "8"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentComputeResourcesMaxVcpusMinVcpusConfig(rName, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_ec2Configuration(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentEC2Configuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "compute_resources.0.ec2_configuration.0.image_id_override"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_launchTemplate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	launchTemplateResourceName := "aws_launch_template.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentLaunchTemplateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_name", launchTemplateResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_updateLaunchTemplate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	launchTemplateResourceName := "aws_launch_template.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rName, "$Default"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_name", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rName, "$Latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_id", launchTemplateResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_name", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Latest"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_UpdateSecurityGroupsAndSubnets_fargate(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName1 := "aws_security_group.test"
	securityGroupResourceName2 := "aws_security_group.test_2"
	securityGroupResourceName3 := "aws_security_group.test_3"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName1 := "aws_subnet.test"
	subnetResourceName2 := "aws_subnet.test_2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentFargateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName1, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName1, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentFargateUpdatedSecurityGroupsAndSubnetsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName2, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName3, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName2, "id"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "service_role", serviceRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "status_reason"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "MANAGED"),
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

func TestAccBatchComputeEnvironment_tags(t *testing.T) {
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccComputeEnvironmentTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccComputeEnvironmentTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(resourceName, &ce),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_createUnmanagedWithComputeResources(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccComputeEnvironmentUnmanagedWithComputeResourcesConfig(rName),
				ExpectError: regexp.MustCompile("no `compute_resources` can be specified when `type` is \"UNMANAGED\""),
			},
		},
	})
}

// Test plan time errors...

func TestAccBatchComputeEnvironment_createEC2WithoutComputeResources(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccComputeEnvironmentEC2WithoutComputeResourcesConfig(rName),
				ExpectError: regexp.MustCompile(`computeResources must be provided for a MANAGED compute environment`),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_createSpotWithoutIAMFleetRole(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, batch.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBatchComputeEnvironmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccComputeEnvironmentSpotWithoutIAMFleetRoleConfig(rName),
				ExpectError: regexp.MustCompile(`ComputeResources.spotIamFleetRole cannot not be null or empty`),
			},
		},
	})
}

func testAccCheckBatchComputeEnvironmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_batch_compute_environment" {
			continue
		}

		_, err := tfbatch.FindComputeEnvironmentDetailByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Batch Compute Environment %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckComputeEnvironmentExists(n string, v *batch.ComputeEnvironmentDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Compute Environment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn

		computeEnvironment, err := tfbatch.FindComputeEnvironmentDetailByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *computeEnvironment

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn

	input := &batch.DescribeComputeEnvironmentsInput{}

	_, err := conn.DescribeComputeEnvironments(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccComputeEnvironmentBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_instance" {
  name = "%[1]s_ecs_instance"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "ec2.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance" {
  role       = aws_iam_role.ecs_instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance" {
  name = %[1]q
  role = aws_iam_role_policy_attachment.ecs_instance.role
}

resource "aws_iam_role" "batch_service" {
  name = "%[1]s_batch_service"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "batch.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_service" {
  role       = aws_iam_role.batch_service.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ec2_spot_fleet" {
  name = "%[1]s_ec2_spot_fleet"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "spotfleet.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ec2_spot_fleet" {
  role       = aws_iam_role.ec2_spot_fleet.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2SpotFleetTaggingRole"
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.1.0/24"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccComputeEnvironmentBasicConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentNameGeneratedConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		`
resource "aws_batch_compute_environment" "test" {
  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`)
}

func testAccComputeEnvironmentNamePrefixConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name_prefix = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentEC2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentEC2WithDesiredVcpusEC2KeyPairImageIdAndComputeResourcesTagsConfig(rName, publicKey string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus     = 16
    min_vcpus     = 4
    desired_vcpus = 8
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"

    tags = {
      key1 = "value1"
    }

    ec2_key_pair = aws_key_pair.test.id
    image_id     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}
`, rName, publicKey))
}

func testAccComputeEnvironmentFargateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentFargateDefaultServiceRoleConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE"
  }

  type = "MANAGED"
}
`, rName))
}

func testAccComputeEnvironmentFargateUpdatedServiceRoleConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE"
  }

  service_role = aws_iam_role.batch_service_2.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service_2]
}

resource "aws_iam_role" "batch_service_2" {
  name = "%[1]s_batch_service_2"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Effect": "Allow",
    "Principal": {
      "Service": "batch.${data.aws_partition.current.dns_suffix}"
    }
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "batch_service_2" {
  role       = aws_iam_role.batch_service_2.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}
`, rName))
}

func testAccComputeEnvironmentFargateSpotConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "FARGATE_SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "managed"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentSpotConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 2
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "spot"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentSpotWithAllocationStrategyAndBidPercentageConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    allocation_strategy = "Best_Fit"
    bid_percentage      = 60
    instance_role       = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentStateConfig(rName string, state string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  state        = %[2]q
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName, state))
}

func testAccComputeEnvironmentComputeResourcesMaxVcpusMinVcpusConfig(rName string, maxVcpus int, minVcpus int) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = ["optimal"]
    max_vcpus     = %[2]d
    min_vcpus     = %[3]d
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName, maxVcpus, minVcpus))
}

func testAccComputeEnvironmentFargateUpdatedSecurityGroupsAndSubnetsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test_2.id,
      aws_security_group.test_3.id,
    ]
    subnets = [
      aws_subnet.test_2.id
    ]
    type = "FARGATE"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}

resource "aws_security_group" "test_2" {
  name   = "%[1]s_2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s_2"
  }
}

resource "aws_security_group" "test_3" {
  name   = "%[1]s_3"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s_3"
  }
}

resource "aws_subnet" "test_2" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.2.0/24"

  tags = {
    Name = "%[1]s_2"
  }
}
`, rName))
}

func testAccComputeEnvironmentEC2WithoutComputeResourcesConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentUnmanagedWithComputeResourcesConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentSpotWithoutIAMFleetRoleConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentLaunchTemplateConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]

    launch_template {
      launch_template_name = aws_launch_template.test.name
    }

    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentUpdateLaunchTemplateInExistingComputeEnvironment(rName string, version string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "c4.large",
    ]

    launch_template {
      launch_template_id = aws_launch_template.test.id
      version            = %[2]q
    }

    max_vcpus = 16
    min_vcpus = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName, version))
}

func testAccComputeEnvironmentTags1Config(rName string, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  depends_on = [aws_iam_role_policy_attachment.batch_service]

  compute_environment_name = %[1]q
  service_role             = aws_iam_role.batch_service.arn
  type                     = "UNMANAGED"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccComputeEnvironmentTags2Config(rName string, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  depends_on = [aws_iam_role_policy_attachment.batch_service]

  compute_environment_name = %[1]q
  service_role             = aws_iam_role.batch_service.arn
  type                     = "UNMANAGED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccComputeEnvironmentEC2Configuration(rName string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentBaseConfig(rName),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = ["optimal"]
    ec2_configuration {
      image_id_override = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
      image_type        = "ECS_AL2"
    }

    max_vcpus = 16
    min_vcpus = 0

    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "SPOT"
  }

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}
