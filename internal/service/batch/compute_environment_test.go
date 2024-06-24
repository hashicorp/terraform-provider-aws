// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/google/go-cmp/cmp"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestExpandEC2ConfigurationsUpdate(t *testing.T) {
	t.Parallel()

	//lintignore:AWSAT002
	testCases := []struct {
		flattened []interface{}
		expected  []*batch.Ec2Configuration
	}{
		{
			flattened: []interface{}{},
			expected: []*batch.Ec2Configuration{
				{
					ImageType: aws.String("default"),
				},
			},
		},
		{
			flattened: []interface{}{
				map[string]interface{}{
					"image_type": "ECS_AL1",
				},
			},
			expected: []*batch.Ec2Configuration{
				{
					ImageType: aws.String("ECS_AL1"),
				},
			},
		},
		{
			flattened: []interface{}{
				map[string]interface{}{
					"image_id_override": "ami-deadbeef",
				},
			},
			expected: []*batch.Ec2Configuration{
				{
					ImageIdOverride: aws.String("ami-deadbeef"),
				},
			},
		},
		{
			flattened: []interface{}{
				map[string]interface{}{
					"image_id_override": "ami-deadbeef",
					"image_type":        "ECS_AL1",
				},
			},
			expected: []*batch.Ec2Configuration{
				{
					ImageIdOverride: aws.String("ami-deadbeef"),
					ImageType:       aws.String("ECS_AL1"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		expanded := tfbatch.ExpandEC2ConfigurationsUpdate(testCase.flattened, "default")
		if diff := cmp.Diff(expanded, testCase.expected); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
	}
}

func TestExpandLaunchTemplateSpecificationUpdate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		flattened []interface{}
		expected  *batch.LaunchTemplateSpecification
	}{
		{
			flattened: []interface{}{},
			expected: &batch.LaunchTemplateSpecification{
				LaunchTemplateId: aws.String(""),
			},
		},
		{
			flattened: []interface{}{
				map[string]interface{}{
					"launch_template_id": "lt-123456",
				},
			},
			expected: &batch.LaunchTemplateSpecification{
				LaunchTemplateId: aws.String("lt-123456"),
				Version:          aws.String(""),
			},
		},
		{
			flattened: []interface{}{
				map[string]interface{}{
					"launch_template_name": "my-launch-template",
				},
			},
			expected: &batch.LaunchTemplateSpecification{
				LaunchTemplateName: aws.String("my-launch-template"),
				Version:            aws.String(""),
			},
		},
		{
			flattened: []interface{}{
				map[string]interface{}{
					"launch_template_id": "lt-123456",
					names.AttrVersion:    "$LATEST",
				},
			},
			expected: &batch.LaunchTemplateSpecification{
				LaunchTemplateId: aws.String("lt-123456"),
				Version:          aws.String("$LATEST"),
			},
		},
		{
			flattened: []interface{}{
				map[string]interface{}{
					"launch_template_name": "my-launch-template",
					names.AttrVersion:      "$LATEST",
				},
			},
			expected: &batch.LaunchTemplateSpecification{
				LaunchTemplateName: aws.String("my-launch-template"),
				Version:            aws.String("$LATEST"),
			},
		},
	}

	for _, testCase := range testCases {
		expanded := tfbatch.ExpandLaunchTemplateSpecificationUpdate(testCase.flattened)
		if diff := cmp.Diff(expanded, testCase.expected); diff != "" {
			t.Errorf("unexpected diff (+wanted, -got): %s", diff)
		}
	}
}

func TestAccBatchComputeEnvironment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttr(resourceName, "eks_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UNMANAGED"),
				),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfbatch.ResourceComputeEnvironment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBatchComputeEnvironment_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_nameGenerated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrNameGenerated(resourceName, "compute_environment_name"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_namePrefix(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, "compute_environment_name", rName),
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

func TestAccBatchComputeEnvironment_eksConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	eksClusterResourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"kubernetes": {
				Source:            "hashicorp/kubernetes",
				VersionConstraint: "~> 2.15",
			},
		},
		CheckDestroy: testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_eksConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					resource.TestCheckResourceAttr(resourceName, "eks_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "eks_configuration.0.eks_cluster_arn", eksClusterResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "eks_configuration.0.kubernetes_namespace", "test"),
				),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_createEC2(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_ec2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.placement_group", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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

func TestAccBatchComputeEnvironment_updatePolicyCreate(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_ec2UpdatePolicyCreate(rName, 30, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT_PROGRESSIVE"),
					resource.TestCheckResourceAttr(resourceName, "update_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.terminate_jobs_on_update", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.job_execution_timeout_minutes", "30"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccComputeEnvironmentConfig_ec2UpdatePolicyCreate(rName, 60, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT_PROGRESSIVE"),
					resource.TestCheckResourceAttr(resourceName, "update_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.terminate_jobs_on_update", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.job_execution_timeout_minutes", "60"),
				),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_updatePolicyUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_ec2UpdatePolicyOmitted(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT_PROGRESSIVE"),
					resource.TestCheckResourceAttr(resourceName, "update_policy.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccComputeEnvironmentConfig_ec2UpdatePolicyCreate(rName, 60, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT_PROGRESSIVE"),
					resource.TestCheckResourceAttr(resourceName, "update_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.terminate_jobs_on_update", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "update_policy.0.job_execution_timeout_minutes", "60"),
				),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_CreateEC2DesiredVCPUsEC2KeyPairImageID_computeResourcesTags(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	resourceName := "aws_batch_compute_environment.test"
	amiDatasourceName := "data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64"
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_ec2DesiredVCPUsEC2KeyPairImageIDAndResourcesTags(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", "8"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.ec2_key_pair", keyPairResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.image_id", amiDatasourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.key1", acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_spot(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_spotAllocationStrategyAndBidPercentage(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "60"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_fargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_fargateSpot(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE_SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_state(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UNMANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_state(rName, "disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "DISABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "UNMANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	serviceRoleResourceName1 := "aws_iam_role.batch_service"
	serviceRoleResourceName2 := "aws_iam_role.batch_service_2"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_fargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName1, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_fargateUpdatedServiceRole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName2, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/batch")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_fargateDefaultServiceRole(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrServiceRole, "iam", regexache.MustCompile(`role/aws-service-role/batch`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_resourcesMaxVCPUsMinVCPUs(rName, 4, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_resourcesMaxVCPUsMinVCPUs(rName, 4, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_resourcesMaxVCPUsMinVCPUs(rName, 4, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_resourcesMaxVCPUsMinVCPUs(rName, 4, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_resourcesMaxVCPUsMinVCPUs(rName, 8, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "8"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_resourcesMaxVCPUsMinVCPUs(rName, 2, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_ec2Configuration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "compute_resources.0.ec2_configuration.0.image_id_override"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttrSet(resourceName, "compute_resources.0.ec2_configuration.1.image_id_override"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.1.image_type", "ECS_AL2_NVIDIA"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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

func TestAccBatchComputeEnvironment_ec2ConfigurationPlacementGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	securityGroupResourceName := "aws_security_group.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_ec2ConfigurationPlacementGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "compute_resources.0.ec2_configuration.0.image_id_override"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttrSet(resourceName, "compute_resources.0.ec2_configuration.1.image_id_override"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.1.image_type", "ECS_AL2_NVIDIA"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.placement_group", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	launchTemplateResourceName := "aws_launch_template.test"
	serviceRoleResourceName := "aws_iam_role.batch_service"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_launchTemplate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_name", launchTemplateResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_updateLaunchTemplateInExisting(rName, "$Default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_name", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_updateLaunchTemplateInExisting(rName, "$Latest"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.launch_template_name", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Latest"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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
	ctx := acctest.Context(t)
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
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentConfig_fargate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName1, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName1, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeEnvironmentConfig_fargateUpdatedSecurityGroupsAndSubnets(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName2, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName3, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.spot_iam_fleet_role", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName2, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "FARGATE"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRole, serviceRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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

func TestAccBatchComputeEnvironment_createUnmanagedWithComputeResources(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccComputeEnvironmentConfig_unmanagedResources(rName),
				ExpectError: regexache.MustCompile("no `compute_resources` can be specified when `type` is \"UNMANAGED\""),
			},
		},
	})
}

func TestAccBatchComputeEnvironment_updateEC2(t *testing.T) {
	ctx := acctest.Context(t)
	var ce batch.ComputeEnvironmentDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_batch_compute_environment.test"
	instanceProfileResourceName := "aws_iam_instance_profile.ecs_instance"
	updatedInstanceProfileResourceName := "aws_iam_instance_profile.ecs_instance_2"
	securityGroupResourceName := "aws_security_group.test"
	updatedSecurityGroupResourceName := "aws_security_group.test_2"
	subnetResourceName := "aws_subnet.test"
	updatedSubnetResourceName := "aws_subnet.test_2"
	ec2KeyPairResourceName := "aws_key_pair.test"
	launchTemplateResourceName := "aws_launch_template.test"
	spotFleetRoleResourceName := "aws_iam_role.ec2_spot_fleet"
	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccComputeenvironmentConfig_ec2PreUpdate(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT_PROGRESSIVE"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeenvironmentConfig_ec2Update(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "SPOT_CAPACITY_OPTIMIZED"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", "100"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "compute_resources.0.ec2_configuration.0.image_id_override"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.ec2_key_pair", ec2KeyPairResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", updatedInstanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.launch_template.0.launch_template_id", launchTemplateResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.0.version", "$Latest"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", updatedSecurityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", updatedSubnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.updated", "yes"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "SPOT"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
				),
			},
			{
				Config: testAccComputeenvironmentConfig_ec2PreUpdate(rName, publicKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckComputeEnvironmentExists(ctx, resourceName, &ce),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "batch", fmt.Sprintf("compute-environment/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_name_prefix", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.allocation_strategy", "BEST_FIT_PROGRESSIVE"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.bid_percentage", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.desired_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_configuration.0.image_type", "ECS_AL2"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.ec2_key_pair", ""),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.image_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.instance_role", instanceProfileResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.instance_type.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "compute_resources.0.instance_type.*", "optimal"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.launch_template.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.max_vcpus", "16"),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.min_vcpus", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.security_group_ids.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.security_group_ids.*", securityGroupResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "compute_resources.0.spot_iam_fleet_role", spotFleetRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.subnets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "compute_resources.0.subnets.*", subnetResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.tags.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_resources.0.type", "EC2"),
					resource.TestCheckResourceAttrSet(resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatusReason),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "MANAGED"),
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

// Test plan time errors...

func TestAccBatchComputeEnvironment_createEC2WithoutComputeResources(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckComputeEnvironmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccComputeEnvironmentConfig_ec2NoResources(rName),
				ExpectError: regexache.MustCompile(`computeResources must be provided for a MANAGED compute environment`),
			},
		},
	})
}

func testAccCheckComputeEnvironmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_compute_environment" {
				continue
			}

			_, err := tfbatch.FindComputeEnvironmentDetailByName(ctx, conn, rs.Primary.ID)

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
}

func testAccCheckComputeEnvironmentExists(ctx context.Context, n string, v *batch.ComputeEnvironmentDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Batch Compute Environment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

		output, err := tfbatch.FindComputeEnvironmentDetailByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BatchConn(ctx)

	input := &batch.DescribeComputeEnvironmentsInput{}

	_, err := conn.DescribeComputeEnvironmentsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccComputeEnvironmentConfig_base(rName string) string {
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

func testAccComputeEnvironmentConfig_baseDefaultSLR(rName string) string {
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

func testAccComputeEnvironmentConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentConfig_nameGenerated(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), `
resource "aws_batch_compute_environment" "test" {
  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`)
}

func testAccComputeEnvironmentConfig_namePrefix(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name_prefix = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentConfig_eksConfiguration(rName string) string {
	//lintignore:AT004
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "cluster" {
  name = "%[1]s-cluster"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "eks.${data.aws_partition.current.dns_suffix}",
          "eks-nodegroup.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.cluster.name
}

resource "aws_iam_role" "node" {
  name = "%[1]s-node"

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_iam_role_policy_attachment" "node-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.node.name
}

resource "aws_iam_role_policy_attachment" "node-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.node.name
}

resource "aws_iam_instance_profile" "node" {
  name = "%[1]s-node"
  role = aws_iam_role.node.name
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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

resource "aws_main_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  vpc_id         = aws_vpc.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 0
    protocol    = -1
    to_port     = 0
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [
    "aws_iam_role_policy_attachment.cluster-AmazonEKSClusterPolicy",
    "aws_main_route_table_association.test",
  ]
}

data "aws_eks_cluster_auth" "cluster" {
  name = aws_eks_cluster.test.id
}

provider "kubernetes" {
  host                   = aws_eks_cluster.test.endpoint
  cluster_ca_certificate = base64decode(aws_eks_cluster.test.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.cluster.token
}

resource "kubernetes_namespace" "test" {
  metadata {
    name = "test"
  }
}

resource "kubernetes_cluster_role" "test" {
  metadata {
    name = "aws-batch-cluster-role"
  }

  rule {
    api_groups = [""]
    resources  = ["namespaces"]
    verbs      = ["get"]
  }

  rule {
    api_groups = [""]
    resources  = ["nodes", "pods", "configmaps"]
    verbs      = ["get", "list", "watch"]
  }

  rule {
    api_groups = ["apps"]
    resources  = ["daemonsets", "deployments", "statefulsets", "replicasets"]
    verbs      = ["get", "list", "watch"]
  }

  rule {
    api_groups = ["rbac.authorization.k8s.io"]
    resources  = ["clusterroles", "clusterrolebindings"]
    verbs      = ["get", "list"]
  }
}

resource "kubernetes_cluster_role_binding" "test" {
  metadata {
    name = "aws-batch-cluster-role-binding"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.test.metadata[0].name
  }

  subject {
    kind      = "User"
    name      = "aws-batch"
    api_group = "rbac.authorization.k8s.io"
  }
}

resource "kubernetes_role" "test" {
  metadata {
    name      = "aws-batch-compute-environment-role"
    namespace = kubernetes_namespace.test.metadata[0].name
  }

  rule {
    api_groups = [""]
    resources  = ["pods"]
    verbs      = ["create", "get", "list", "watch", "delete", "patch"]
  }

  rule {
    api_groups = [""]
    resources  = ["serviceaccounts"]
    verbs      = ["get", "list"]
  }

  rule {
    api_groups = ["rbac.authorization.k8s.io"]
    resources  = ["roles", "rolebindings"]
    verbs      = ["get", "list"]
  }
}

resource "kubernetes_role_binding" "test" {
  metadata {
    name      = "aws-batch-compute-environment-role-binding"
    namespace = kubernetes_namespace.test.metadata[0].name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "Role"
    name      = kubernetes_role.test.metadata[0].name
  }

  subject {
    kind      = "User"
    name      = "aws-batch"
    api_group = "rbac.authorization.k8s.io"
  }
}

resource "kubernetes_config_map" "aws_auth" {
  metadata {
    name      = "aws-auth"
    namespace = "kube-system"
  }

  data = {
    mapRoles = <<EOF
- rolearn: arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/AWSServiceRoleForBatch
  username: ${kubernetes_role_binding.test.subject[0].name}
EOF
  }
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  eks_configuration {
    eks_cluster_arn      = aws_eks_cluster.test.arn
    kubernetes_namespace = kubernetes_namespace.test.metadata[0].name
  }

  type = "MANAGED"

  compute_resources {
    type                = "EC2"
    allocation_strategy = "BEST_FIT_PROGRESSIVE"
    min_vcpus           = 0
    max_vcpus           = 128

    instance_type = ["m5.large"]

    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = aws_subnet.test[*].id

    instance_role = aws_iam_instance_profile.node.arn
  }

  depends_on = [
    kubernetes_config_map.aws_auth,
    aws_iam_role_policy_attachment.node-AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.node-AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.node-AmazonEC2ContainerRegistryReadOnly,
  ]
}
`, rName))
}

func testAccComputeEnvironmentConfig_ec2(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_ec2UpdatePolicyCreate(rName string, timeout int, terminate bool) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_baseDefaultSLR(rName), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    allocation_strategy = "BEST_FIT_PROGRESSIVE"
    instance_role       = aws_iam_instance_profile.ecs_instance.arn
    instance_type       = ["optimal"]
    max_vcpus           = 4
    min_vcpus           = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  update_policy {
    job_execution_timeout_minutes = %[2]d
    terminate_jobs_on_update      = %[3]t
  }

  type = "MANAGED"
}
`, rName, timeout, terminate))
}

func testAccComputeEnvironmentConfig_ec2UpdatePolicyOmitted(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_baseDefaultSLR(rName), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    allocation_strategy = "BEST_FIT_PROGRESSIVE"
    instance_role       = aws_iam_instance_profile.ecs_instance.arn
    instance_type       = ["optimal"]
    max_vcpus           = 4
    min_vcpus           = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  type = "MANAGED"
}
`, rName))
}

func testAccComputeEnvironmentConfig_ec2DesiredVCPUsEC2KeyPairImageIDAndResourcesTags(rName, publicKey string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
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
    image_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccComputeEnvironmentConfig_fargate(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_fargateDefaultServiceRole(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_fargateUpdatedServiceRole(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_fargateSpot(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_spot(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_spotAllocationStrategyAndBidPercentage(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_state(rName string, state string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "UNMANAGED"
  state        = %[2]q
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName, state))
}

func testAccComputeEnvironmentConfig_resourcesMaxVCPUsMinVCPUs(rName string, maxVcpus int, minVcpus int) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_fargateUpdatedSecurityGroupsAndSubnets(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_ec2NoResources(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  service_role = aws_iam_role.batch_service.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.batch_service]
}
`, rName))
}

func testAccComputeEnvironmentConfig_unmanagedResources(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_launchTemplate(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  vpc_security_group_ids = [
    aws_security_group.test.id
  ]
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

    max_vcpus           = 16
    min_vcpus           = 0
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

func testAccComputeEnvironmentConfig_updateLaunchTemplateInExisting(rName string, version string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), fmt.Sprintf(`
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

func testAccComputeEnvironmentConfig_ec2Configuration(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = ["optimal"]

    ec2_configuration {
      image_id_override = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
      image_type        = "ECS_AL2"
    }

    ec2_configuration {
      image_id_override = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
      image_type        = "ECS_AL2_NVIDIA"
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

func testAccComputeEnvironmentConfig_ec2ConfigurationPlacementGroup(rName string) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_base(rName), acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance.arn
    instance_type = ["optimal"]

    ec2_configuration {
      image_id_override = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
      image_type        = "ECS_AL2"
    }

    ec2_configuration {
      image_id_override = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
      image_type        = "ECS_AL2_NVIDIA"
    }

    max_vcpus = 16
    min_vcpus = 0

    placement_group = aws_placement_group.test.name

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

func testAccComputeEnvironmentConfig_baseForUpdates(rName string, publicKey string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "ecs_instance_2" {
  name = "%[1]s_ecs_instance_2"

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

resource "aws_iam_role_policy_attachment" "ecs_instance_2" {
  role       = aws_iam_role.ecs_instance_2.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_2" {
  name = "%[1]s_2"
  role = aws_iam_role_policy_attachment.ecs_instance_2.role
}

resource "aws_security_group" "test_2" {
  name   = "%[1]s_2"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "%[1]s_2"
  }
}

resource "aws_subnet" "test_2" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.2.0/24"

  tags = {
    Name = "%[1]s_2"
  }
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q
}

resource "aws_launch_template" "test" {
  name = %[1]q
}
`, rName, publicKey)
}

func testAccComputeenvironmentConfig_ec2PreUpdate(rName string, publicKey string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentConfig_base(rName),
		testAccComputeEnvironmentConfig_baseForUpdates(rName, publicKey),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    allocation_strategy = "BEST_FIT_PROGRESSIVE"
    instance_role       = aws_iam_instance_profile.ecs_instance.arn
    instance_type = [
      "optimal",
    ]
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }

  type = "MANAGED"
}
`, rName))
}

func testAccComputeenvironmentConfig_ec2Update(rName string, publicKey string) string {
	return acctest.ConfigCompose(
		testAccComputeEnvironmentConfig_base(rName),
		testAccComputeEnvironmentConfig_baseForUpdates(rName, publicKey),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    allocation_strategy = "SPOT_CAPACITY_OPTIMIZED"
    bid_percentage      = 100
    ec2_key_pair        = aws_key_pair.test.id
    instance_role       = aws_iam_instance_profile.ecs_instance_2.arn
    ec2_configuration {
      image_id_override = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
      image_type        = "ECS_AL2"
    }
    launch_template {
      launch_template_id = aws_launch_template.test.id
      version            = "$Latest"
    }
    instance_type = [
      "c4.large",
    ]
    max_vcpus = 16
    security_group_ids = [
      aws_security_group.test_2.id
    ]
    spot_iam_fleet_role = aws_iam_role.ec2_spot_fleet.arn
    subnets = [
      aws_subnet.test_2.id
    ]
    type = "SPOT"
    tags = {
      updated = "yes"
    }
  }

  type = "MANAGED"
}
`, rName))
}
