// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/fis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fis/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tffis "github.com/hashicorp/terraform-provider-aws/internal/service/fis"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFISExperimentTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "test-action-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:terminate-instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "An experiment template for testing"),
					resource.TestCheckResourceAttr(resourceName, "experiment_options.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "COUNT(1)"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", acctest.Ct1),
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

func TestAccFISExperimentTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffis.ResourceExperimentTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFISExperimentTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "An experiment template for testing"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "test-action-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", ""),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:terminate-instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "to-terminate-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "COUNT(1)"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
				),
			},
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "Artic Lake", "test-action-2", "Lane 8", "aws:ec2:stop-instances", "Instances", "to-stop-1", "aws:ec2:instance", "ALL", "env2", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Artic Lake"),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "test-action-2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", "Lane 8"),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:stop-instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Instances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "to-stop-1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "to-stop-1"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", acctest.Ct1),
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

func TestAccFISExperimentTemplate_spot(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_actionParameter(rName, "Send Spot Instance Interruptions", "Send-Spot-Instance-Interruptions", "Send Spot Instance Interruptions", "aws:ec2:send-spot-instance-interruptions", "SpotInstances", "send-spot-instance-interruptions-target", "durationBeforeInterruption", "PT2M", "aws:ec2:spot-instance", "PERCENT(25)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Send Spot Instance Interruptions"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "Send-Spot-Instance-Interruptions"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", "Send Spot Instance Interruptions"),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ec2:send-spot-instance-interruptions"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.key", "durationBeforeInterruption"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.value", "PT2M"),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "SpotInstances"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "send-spot-instance-interruptions-target"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "send-spot-instance-interruptions-target"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:spot-instance"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "PERCENT(25)"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_arns.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "env"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.value", "test"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFISExperimentTemplate_eks(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_eks(rName, "kubernetes custom resource creation", "k8s-pod-delete", "k8s pod delete", "aws:eks:inject-kubernetes-custom-resource", "Cluster", "kubernetes-custom-resource-creation-target", "kubernetesApiVersion", "litmuschaos.io/v1alpha1", "kubernetesKind", "ChaosEngine", "kubernetesNamespace", "observability", "kubernetesSpec", "{\"engineState\":\"active\",\"appinfo\":{\"appns\":\"observability\",\"applabel\":\"app=nginx\",\"appkind\":\"deployment\"},\"chaosServiceAccount\":\"pod-delete-sa\",\"experiments\":[{\"name\":\"pod-delete\",\"spec\":{\"components\":{\"env\":[{\"name\":\"TOTAL_CHAOS_DURATION\",\"value\":\"60\"},{\"name\":\"CHAOS_INTERVAL\",\"value\":\"60\"},{\"name\":\"PODS_AFFECTED_PERC\",\"value\":\"30\"}]},\"probe\":[]}}],\"annotationCheck\":\"false\"}", "maxDuration", "PT2M", "aws:eks:cluster", "ALL", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "kubernetes custom resource creation"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_fis", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "k8s-pod-delete"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", "k8s pod delete"),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:eks:inject-kubernetes-custom-resource"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", "5"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.key", "kubernetesApiVersion"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.value", "litmuschaos.io/v1alpha1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.1.key", "kubernetesKind"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.1.value", "ChaosEngine"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.2.key", "kubernetesNamespace"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.2.value", "observability"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.3.key", "kubernetesSpec"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.3.value", "{\"engineState\":\"active\",\"appinfo\":{\"appns\":\"observability\",\"applabel\":\"app=nginx\",\"appkind\":\"deployment\"},\"chaosServiceAccount\":\"pod-delete-sa\",\"experiments\":[{\"name\":\"pod-delete\",\"spec\":{\"components\":{\"env\":[{\"name\":\"TOTAL_CHAOS_DURATION\",\"value\":\"60\"},{\"name\":\"CHAOS_INTERVAL\",\"value\":\"60\"},{\"name\":\"PODS_AFFECTED_PERC\",\"value\":\"30\"}]},\"probe\":[]}}],\"annotationCheck\":\"false\"}"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.4.key", "maxDuration"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.4.value", "PT2M"),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Cluster"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "kubernetes-custom-resource-creation-target"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "kubernetes-custom-resource-creation-target"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:eks:cluster"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.resource_arns.0", "aws_eks_cluster.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFISExperimentTemplate_ebs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_ebsVolume(rName, "EBS Volume Pause I/O Experiment", "ebs-paused-io-action", "EBS Volume Pause I/O", "aws:ebs:pause-volume-io", "Volumes", "ebs-volume-to-pause-io", names.AttrDuration, "PT6M", "aws:ec2:ebs-volume", "ALL", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "EBS Volume Pause I/O Experiment"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_fis", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "ebs-paused-io-action"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", "EBS Volume Pause I/O"),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ebs:pause-volume-io"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.key", names.AttrDuration),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.value", "PT6M"),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Volumes"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "ebs-volume-to-pause-io"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "ebs-volume-to-pause-io"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:ebs-volume"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.resource_arns.0", "aws_ebs_volume.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFISExperimentTemplate_ebsParameters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_ebsVolumeParameters(rName, "EBS Volume Pause I/O Experiment", "ebs-paused-io-action", "EBS Volume Pause I/O", "aws:ebs:pause-volume-io", "Volumes", "ebs-volume-to-pause-io", names.AttrDuration, "PT6M", "aws:ec2:ebs-volume", "ALL", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "EBS Volume Pause I/O Experiment"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, "aws_iam_role.test_fis", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.source", "none"),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.0.value", ""),
					resource.TestCheckResourceAttr(resourceName, "stop_condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.name", "ebs-paused-io-action"),
					resource.TestCheckResourceAttr(resourceName, "action.0.description", "EBS Volume Pause I/O"),
					resource.TestCheckResourceAttr(resourceName, "action.0.action_id", "aws:ebs:pause-volume-io"),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.key", names.AttrDuration),
					resource.TestCheckResourceAttr(resourceName, "action.0.parameter.0.value", "PT6M"),
					resource.TestCheckResourceAttr(resourceName, "action.0.start_after.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.key", "Volumes"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.0.value", "ebs-volume-to-pause-io"),
					resource.TestCheckResourceAttr(resourceName, "action.0.target.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.name", "ebs-volume-to-pause-io"),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_type", "aws:ec2:ebs-volume"),
					resource.TestCheckResourceAttr(resourceName, "target.0.selection_mode", "ALL"),
					resource.TestCheckResourceAttr(resourceName, "target.0.parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.parameters.availabilityZoneIdentifier", "aws_ebs_volume.test", names.AttrAvailabilityZone),
					resource.TestCheckResourceAttr(resourceName, "target.0.filter.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "target.0.resource_tag.0.key", "Name"),
					resource.TestCheckResourceAttrPair(resourceName, "target.0.resource_tag.0.value", "aws_ebs_volume.test", "tags.Name"),
					resource.TestCheckResourceAttr(resourceName, "target.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccFISExperimentTemplate_loggingConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			// Cloudwatch Logging
			{
				Config: testAccExperimentTemplateConfig_logConfigCloudWatch(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.log_schema_version", acctest.Ct2),
					acctest.CheckResourceAttrRegionalARN(resourceName, "log_configuration.0.cloudwatch_logs_configuration.0.log_group_arn", "logs", fmt.Sprintf("log-group:%s:*", rName)),
				),
			},
			// Delete Logging
			{
				Config: testAccExperimentTemplateConfig_basic(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
				),
			},
			// S3 Logging
			{
				Config: testAccExperimentTemplateConfig_logConfigS3(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.log_schema_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3_configuration.0.prefix", ""),
				),
			},
			{
				Config: testAccExperimentTemplateConfig_logConfigS3Prefix(rName, "An experiment template for testing", "test-action-1", "", "aws:ec2:terminate-instances", "Instances", "to-terminate-1", "aws:ec2:instance", "COUNT(1)", "env", "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.log_schema_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3_configuration.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "log_configuration.0.s3_configuration.0.prefix", "test"),
				),
			},
		},
	})
}

func TestAccFISExperimentTemplate_updateExperimentOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_fis_experiment_template.test"
	var conf awstypes.ExperimentTemplate

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fis.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExperimentTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExperimentTemplateConfig_ExperimentOptions(rName, "skip"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "experiment_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "experiment_options.0.account_targeting", "single-account"),
					resource.TestCheckResourceAttr(resourceName, "experiment_options.0.empty_target_resolution_mode", "skip"),
				),
			},
			{
				Config: testAccExperimentTemplateConfig_ExperimentOptions(rName, "fail"),
				Check: resource.ComposeTestCheckFunc(
					testAccExperimentTemplateExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "experiment_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "experiment_options.0.empty_target_resolution_mode", "fail"),
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

func testAccExperimentTemplateExists(ctx context.Context, n string, v *awstypes.ExperimentTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)

		output, err := tffis.FindExperimentTemplateByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckExperimentTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).FISClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_fis_experiment_template" {
				continue
			}

			_, err := tffis.FindExperimentTemplateByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("FIS Experiment Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccExperimentTemplateConfig_basic(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[8]q
    selection_mode = %[9]q

    resource_tag {
      key   = %[10]q
      value = %[11]q
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV)
}

func testAccExperimentTemplateConfig_actionParameter(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK, paramV, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }

    parameter {
      key   = %[8]q
      value = %[9]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[10]q
    selection_mode = %[11]q

    resource_tag {
      key   = %[12]q
      value = %[13]q
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK, paramV, targetResType, targetSelectMode, targetResTagK, targetResTagV)
}
func testAccExperimentTemplateConfig_baseEKSCluster(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test-AmazonEKSClusterPolicy" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)
  assign_ipv6_address_on_creation = true

  tags = {
    Name                          = %[1]q
    "kubernetes.io/cluster/%[1]s" = "shared"
  }
}

resource "aws_eks_cluster" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }

  depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
}
`, rName))
}

func testAccExperimentTemplateConfig_eks(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK1, paramV1, paramK2, paramV2, paramK3, paramV3, paramK4, paramV4, paramK5, paramV5, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return acctest.ConfigCompose(testAccExperimentTemplateConfig_baseEKSCluster(rName), fmt.Sprintf(`
resource "aws_iam_role" "test_fis" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test_fis.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }

    parameter {
      key   = %[8]q
      value = %[9]q
    }

    parameter {
      key   = %[10]q
      value = %[11]q
    }

    parameter {
      key   = %[12]q
      value = %[13]q
    }

    parameter {
      key   = %[14]q
      value = %[15]q
    }

    parameter {
      key   = %[16]q
      value = %[17]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[18]q
    selection_mode = %[19]q

    resource_arns = tolist([aws_eks_cluster.test.arn])
  }

  tags = {
    Name = %[1]q
  }
}
`, rName+"-fis", desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK1, paramV1, paramK2, paramV2, paramK3, paramV3, paramK4, paramV4, paramK5, paramV5, targetResType, targetSelectMode, targetResTagK, targetResTagV))
}

func testAccExperimentTemplateConfig_baseEBSVolume(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 40

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccExperimentTemplateConfig_ebsVolume(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK1, paramV1, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return acctest.ConfigCompose(testAccExperimentTemplateConfig_baseEBSVolume(rName), fmt.Sprintf(`
resource "aws_iam_role" "test_fis" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test_fis.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }

    parameter {
      key   = %[8]q
      value = %[9]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[10]q
    selection_mode = %[11]q

    resource_arns = tolist([aws_ebs_volume.test.arn])
  }

  tags = {
    Name = %[1]q
  }
}
`, rName+"-fis", desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK1, paramV1, targetResType, targetSelectMode, targetResTagK, targetResTagV))
}

func testAccExperimentTemplateConfig_ebsVolumeParameters(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK1, paramV1, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return acctest.ConfigCompose(testAccExperimentTemplateConfig_baseEBSVolume(rName), fmt.Sprintf(`
resource "aws_iam_role" "test_fis" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}
resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test_fis.arn
  stop_condition {
    source = "none"
  }
  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q
    target {
      key   = %[6]q
      value = %[7]q
    }
    parameter {
      key   = %[8]q
      value = %[9]q
    }
  }
  target {
    name           = %[7]q
    resource_type  = %[10]q
    selection_mode = %[11]q
    resource_tag {
      key   = "Name"
      value = aws_ebs_volume.test.tags.Name
    }
    parameters = {
      availabilityZoneIdentifier = aws_ebs_volume.test.availability_zone
    }
  }
  tags = {
    Name = %[1]q
  }
}
`, rName+"-fis", desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, paramK1, paramV1, targetResType, targetSelectMode, targetResTagK, targetResTagV))
}

func testAccExperimentTemplateConfig_logConfigCloudWatch(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[8]q
    selection_mode = %[9]q

    resource_tag {
      key   = %[10]q
      value = %[11]q
    }
  }

  log_configuration {
    log_schema_version = 2

    cloudwatch_logs_configuration {
      log_group_arn = "${aws_cloudwatch_log_group.test.arn}:*"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV)
}

func testAccExperimentTemplateConfig_logConfigS3(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[8]q
    selection_mode = %[9]q

    resource_tag {
      key   = %[10]q
      value = %[11]q
    }
  }

  log_configuration {
    log_schema_version = 2

    s3_configuration {
      bucket_name = aws_s3_bucket.test.bucket
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV)
}

func testAccExperimentTemplateConfig_logConfigS3Prefix(rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_fis_experiment_template" "test" {
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  action {
    name        = %[3]q
    description = %[4]q
    action_id   = %[5]q

    target {
      key   = %[6]q
      value = %[7]q
    }
  }

  target {
    name           = %[7]q
    resource_type  = %[8]q
    selection_mode = %[9]q

    resource_tag {
      key   = %[10]q
      value = %[11]q
    }
  }

  log_configuration {
    log_schema_version = 2

    s3_configuration {
      bucket_name = aws_s3_bucket.test.bucket
      prefix      = "test"
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, desc, actionName, actionDesc, actionID, actionTargetK, actionTargetV, targetResType, targetSelectMode, targetResTagK, targetResTagV)
}

func testAccExperimentTemplateConfig_ExperimentOptions(rName, mode string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = [
          "fis.${data.aws_partition.current.dns_suffix}",
        ]
      }
    }]
    Version = "2012-10-17"
  })
}

resource "aws_fis_experiment_template" "test" {
  description = "An experiment template for testing"
  role_arn    = aws_iam_role.test.arn

  stop_condition {
    source = "none"
  }

  experiment_options {
    account_targeting            = "single-account"
    empty_target_resolution_mode = %[2]q
  }

  action {
    name        = "test-action-1"
    description = ""
    action_id   = "aws:ec2:terminate-instances"

    target {
      key   = "Instances"
      value = "to-terminate-1"
    }
  }

  target {
    name           = "to-terminate-1"
    resource_type  = "aws:ec2:instance"
    selection_mode = "ALL"

    resource_tag {
      key   = "env2"
      value = "test2"
    }
  }
}
`, rName, mode)
}
