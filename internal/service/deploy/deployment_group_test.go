// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodedeploy "github.com/hashicorp/terraform-provider-aws/internal/service/deploy"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDeployDeploymentGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_basic(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codedeploy", fmt.Sprintf(`deploymentgroup:%s/%s`, rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config_name", "CodeDeployDefault.OneAtATime"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRoleARN, "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_set.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_tag_filter.*", map[string]string{
						names.AttrKey:   "filterkey",
						names.AttrType:  "KEY_AND_VALUE",
						names.AttrValue: "filtervalue",
					}),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "trigger_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "compute_platform", "Server"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_group_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_modified(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codedeploy", fmt.Sprintf(`deploymentgroup:%s/%s-updated`, rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_group_name", rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "deployment_config_name", "CodeDeployDefault.OneAtATime"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrServiceRoleARN, "aws_iam_role.test_updated", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_set.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_tag_filter.*", map[string]string{
						names.AttrKey:   "filterkey",
						names.AttrType:  "KEY_AND_VALUE",
						names.AttrValue: "anotherfiltervalue",
					}),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "trigger_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Basic_tagSet(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_set.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_tag_set.*", map[string]string{
						"ec2_tag_filter.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_tag_set.*.ec2_tag_filter.*", map[string]string{
						names.AttrKey:   "filterkey",
						names.AttrType:  "KEY_AND_VALUE",
						names.AttrValue: "filtervalue",
					}),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_filter.#", acctest.Ct0),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_modified(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_set.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_tag_set.*", map[string]string{
						"ec2_tag_filter.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ec2_tag_set.*.ec2_tag_filter.*", map[string]string{
						names.AttrKey:   "filterkey",
						names.AttrType:  "KEY_AND_VALUE",
						names.AttrValue: "anotherfiltervalue",
					}),
					resource.TestCheckResourceAttr(resourceName, "ec2_tag_filter.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_onPremiseTag(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_onPremiseTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "deployment_config_name", "CodeDeployDefault.OneAtATime"),
					resource.TestCheckResourceAttr(resourceName, "on_premises_instance_tag_filter.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "on_premises_instance_tag_filter.*", map[string]string{
						names.AttrKey:   "filterkey",
						names.AttrType:  "KEY_AND_VALUE",
						names.AttrValue: "filtervalue",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodedeploy.ResourceDeploymentGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Disappears_app(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodedeploy.ResourceApp(), "aws_codedeploy_app.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDeploymentGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Trigger_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_triggerConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					testAccCheckDeploymentGroupTriggerEvents(&group, "test-trigger", []string{
						"DeploymentFailure",
					}),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_triggerConfigurationUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					testAccCheckDeploymentGroupTriggerEvents(&group, "test-trigger", []string{
						"DeploymentFailure",
						"DeploymentSuccess",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Trigger_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_triggerConfigurationCreateMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					testAccCheckDeploymentGroupTriggerEvents(&group, "test-trigger-1", []string{
						"DeploymentFailure",
					}),
					testAccCheckDeploymentGroupTriggerEvents(&group, "test-trigger-2", []string{
						"InstanceFailure",
					}),
					testAccCheckDeploymentGroupTriggerTargetARN(&group, "test-trigger-2",
						regexache.MustCompile(fmt.Sprintf("^arn:%s:sns:[^:]+:[0-9]{12}:%s-2$", acctest.Partition(), rName))),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_triggerConfigurationUpdateMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					testAccCheckDeploymentGroupTriggerEvents(&group, "test-trigger-1", []string{
						"DeploymentFailure",
						"DeploymentStart",
						"DeploymentStop",
						"DeploymentSuccess",
					}),
					testAccCheckDeploymentGroupTriggerEvents(&group, "test-trigger-2", []string{
						"InstanceFailure",
					}),
					testAccCheckDeploymentGroupTriggerTargetARN(&group, "test-trigger-2",
						regexache.MustCompile(fmt.Sprintf("^arn:%s:sns:[^:]+:[0-9]{12}:%s-3$", acctest.Partition(), rName))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_AutoRollback_create(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_autoRollbackConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_rollback_configuration.0.events.*", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_AutoRollback_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_autoRollbackConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_rollback_configuration.0.events.*", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_autoRollbackConfigurationUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.events.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_rollback_configuration.0.events.*", "DEPLOYMENT_STOP_ON_ALARM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_rollback_configuration.0.events.*", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_AutoRollback_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_autoRollbackConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_rollback_configuration.0.events.*", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_autoRollbackConfigurationNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_AutoRollback_disable(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_autoRollbackConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_rollback_configuration.0.events.*", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_autoRollbackConfigurationDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "auto_rollback_configuration.0.events.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "auto_rollback_configuration.0.events.*", "DEPLOYMENT_FAILURE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Alarm_create(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_alarmConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.alarms.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "alarm_configuration.0.alarms.*", "test-alarm"),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.ignore_poll_alarm_failure", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Alarm_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_alarmConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.alarms.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "alarm_configuration.0.alarms.*", "test-alarm"),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.ignore_poll_alarm_failure", acctest.CtFalse),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_alarmConfigurationUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.alarms.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "alarm_configuration.0.alarms.*", "test-alarm"),
					resource.TestCheckTypeSetElemAttr(resourceName, "alarm_configuration.0.alarms.*", "test-alarm-2"),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.ignore_poll_alarm_failure", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Alarm_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_alarmConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.alarms.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "alarm_configuration.0.alarms.*", "test-alarm"),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.ignore_poll_alarm_failure", acctest.CtFalse),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_alarmConfigurationNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_Alarm_disable(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_alarmConfigurationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.alarms.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "alarm_configuration.0.alarms.*", "test-alarm"),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.ignore_poll_alarm_failure", acctest.CtFalse),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_alarmConfigurationDisable(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.alarms.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "alarm_configuration.0.alarms.*", "test-alarm"),
					resource.TestCheckResourceAttr(resourceName, "alarm_configuration.0.ignore_poll_alarm_failure", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// When no configuration is provided, a deploymentStyle object with default values is computed
func TestAccDeployDeploymentGroup_DeploymentStyle_default(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_styleDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_style.0.deployment_option"),
					resource.TestCheckResourceAttrSet(resourceName, "deployment_style.0.deployment_type"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_DeploymentStyle_create(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_styleCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_DeploymentStyle_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_styleCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_styleUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITHOUT_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "IN_PLACE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// Delete reverts to default configuration. It does not remove the deployment_style block
func TestAccDeployDeploymentGroup_DeploymentStyle_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_styleCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_styleDefault(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITHOUT_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "IN_PLACE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_LoadBalancerInfo_create(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_LoadBalancerInfo_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group-2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_LoadBalancerInfo_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoNone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_LoadBalancerInfoTargetGroupInfo_create(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.target_group_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_LoadBalancerInfoTargetGroupInfo_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.target_group_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.target_group_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group-2",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_LoadBalancerInfoTargetGroupInfo_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.target_group_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoDelete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_InPlaceDeploymentWithTrafficControl_create(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_inPlaceTrafficControlCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "IN_PLACE"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_InPlaceDeploymentWithTrafficControl_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_inPlaceTrafficControlCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "IN_PLACE"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_inPlaceTrafficControlUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "CONTINUE_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_BlueGreenDeployment_create(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_blueGreenConfigCreateASG(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "COPY_AUTO_SCALING_GROUP"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_BlueGreenDeployment_updateWithASG(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_blueGreenConfigCreateASG(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "COPY_AUTO_SCALING_GROUP"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_blueGreenConfigUpdateASG(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "COPY_AUTO_SCALING_GROUP"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
				),
			},
		},
	})
}

func TestAccDeployDeploymentGroup_BlueGreenDeployment_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_blueGreenConfigCreateNoASG(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_blueGreenConfigUpdateNoASG(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "CONTINUE_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
				),
			},
		},
	})
}

// Without "Computed: true" on blue_green_deployment_config, removing the resource
// from configuration causes an error, because the remote resource still exists.
func TestAccDeployDeploymentGroup_BlueGreenDeployment_delete(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_blueGreenConfigCreateNoASG(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "120"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_blueGreenConfigDelete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "IN_PLACE"),
					// The state is preserved, but AWS ignores it
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_BlueGreenDeployment_complete(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_blueGreenComplete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "60"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", acctest.Ct0),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_blueGreenCompleteUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_option", "WITH_TRAFFIC_CONTROL"),
					resource.TestCheckResourceAttr(resourceName, "deployment_style.0.deployment_type", "BLUE_GREEN"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.elb_info.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "load_balancer_info.0.elb_info.*", map[string]string{
						names.AttrName: "acc-test-codedeploy-dep-group",
					}),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "CONTINUE_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.green_fleet_provisioning_option.0.action", "DISCOVER_EXISTING"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "KEEP_ALIVE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_ECS_blueGreen(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ecsClusterResourceName := "aws_ecs_cluster.test"
	ecsServiceResourceName := "aws_ecs_service.test"
	lbTargetGroupBlueResourceName := "aws_lb_target_group.blue"
	lbTargetGroupGreenResourceName := "aws_lb_target_group.green"
	resourceName := "aws_codedeploy_deployment_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_ecsBlueGreen(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ecs_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_service.0.cluster_name", ecsClusterResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_service.0.service_name", ecsServiceResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.prod_traffic_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.prod_traffic_route.0.listener_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.0.name", lbTargetGroupBlueResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.1.name", lbTargetGroupGreenResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.test_traffic_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "CONTINUE_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "5"),
				),
			},
			{
				Config: testAccDeploymentGroupConfig_ecsBlueGreenUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "ecs_service.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_service.0.cluster_name", ecsClusterResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "ecs_service.0.service_name", ecsServiceResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.prod_traffic_route.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.prod_traffic_route.0.listener_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.0.name", lbTargetGroupBlueResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer_info.0.target_group_pair_info.0.target_group.1.name", lbTargetGroupGreenResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer_info.0.target_group_pair_info.0.test_traffic_route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.action_on_timeout", "STOP_DEPLOYMENT"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.deployment_ready_option.0.wait_time_in_minutes", "30"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.action", "TERMINATE"),
					resource.TestCheckResourceAttr(resourceName, "blue_green_deployment_config.0.terminate_blue_instances_on_deployment_success.0.termination_wait_time_in_minutes", "60"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_OutdatedInstancesStrategy_update(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_outdatedInstancesStrategy(rName, "UPDATE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "outdated_instances_strategy", "UPDATE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDeployDeploymentGroup_OutdatedInstancesStrategy_ignore(t *testing.T) {
	ctx := acctest.Context(t)
	var group types.DeploymentGroupInfo
	resourceName := "aws_codedeploy_deployment_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DeployServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeploymentGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDeploymentGroupConfig_outdatedInstancesStrategy(rName, "IGNORE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDeploymentGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "outdated_instances_strategy", "IGNORE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDeploymentGroupImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDeploymentGroupTriggerEvents(group *types.DeploymentGroupInfo, triggerName string, expectedEvents []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found := false
		for _, actual := range group.TriggerConfigurations {
			if *actual.TriggerName == triggerName {
				found = true

				numberOfEvents := len(actual.TriggerEvents)
				if numberOfEvents != len(expectedEvents) {
					return fmt.Errorf("Trigger events do not match. Expected: %d. Got: %d.",
						len(expectedEvents), numberOfEvents)
				}

				actualEvents := make([]string, 0, numberOfEvents)
				for _, event := range actual.TriggerEvents {
					actualEvents = append(actualEvents, string(event))
				}
				sort.Strings(actualEvents)

				if !reflect.DeepEqual(actualEvents, expectedEvents) {
					return fmt.Errorf("Trigger events do not match.\nExpected: %v\nGot: %v\n",
						expectedEvents, actualEvents)
				}
				break
			}
		}
		if found {
			return nil
		} else {
			return fmt.Errorf("trigger configuration %q not found", triggerName)
		}
	}
}

func testAccCheckDeploymentGroupTriggerTargetARN(group *types.DeploymentGroupInfo, triggerName string, r *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		found := false
		for _, actual := range group.TriggerConfigurations {
			if *actual.TriggerName == triggerName {
				found = true
				if !r.MatchString(*actual.TriggerTargetArn) {
					return fmt.Errorf("Trigger target arn does not match regular expression.\nRegex: %v\nTriggerTargetArn: %v\n",
						r, *actual.TriggerTargetArn)
				}
				break
			}
		}
		if found {
			return nil
		} else {
			return fmt.Errorf("trigger configuration %q not found", triggerName)
		}
	}
}

func testAccCheckDeploymentGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DeployClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codedeploy_deployment_group" {
				continue
			}

			_, err := tfcodedeploy.FindDeploymentGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_name"], rs.Primary.Attributes["deployment_group_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeDeploy Deployment Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDeploymentGroupExists(ctx context.Context, n string, v *types.DeploymentGroupInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DeployClient(ctx)

		output, err := tfcodedeploy.FindDeploymentGroupByTwoPartKey(ctx, conn, rs.Primary.Attributes["app_name"], rs.Primary.Attributes["deployment_group_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDeploymentGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["app_name"], rs.Primary.Attributes["deployment_group_name"]), nil
	}
}

func testAccDeploymentGroupConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_codedeploy_app" "test" {
  name = %[1]q
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:CompleteLifecycleAction",
        "autoscaling:DeleteLifecycleHook",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLifecycleHooks",
        "autoscaling:PutLifecycleHook",
        "autoscaling:RecordLifecycleActionHeartbeat",
        "codedeploy:*",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus",
        "tag:GetTags",
        "tag:GetResources",
        "sns:Publish"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "codedeploy.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccDeploymentGroupConfig_basic(rName string, tagGroup bool) string {
	var tagGroupOrFilter string
	if tagGroup {
		tagGroupOrFilter = `
ec2_tag_set {
  ec2_tag_filter {
    key   = "filterkey"
    type  = "KEY_AND_VALUE"
    value = "filtervalue"
  }
}
`
	} else {
		tagGroupOrFilter = `
ec2_tag_filter {
  key   = "filterkey"
  type  = "KEY_AND_VALUE"
  value = "filtervalue"
}
`
	}

	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn
  %[2]s
}
`, rName, tagGroupOrFilter))
}

func testAccDeploymentGroupConfig_modified(rName string, tagGroup bool) string {
	var tagGroupOrFilter string
	if tagGroup {
		tagGroupOrFilter = `
ec2_tag_set {
  ec2_tag_filter {
    key   = "filterkey"
    type  = "KEY_AND_VALUE"
    value = "anotherfiltervalue"
  }
}
`
	} else {
		tagGroupOrFilter = `
ec2_tag_filter {
  key   = "filterkey"
  type  = "KEY_AND_VALUE"
  value = "anotherfiltervalue"
}
`
	}

	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = "%[1]s-updated"
  service_role_arn      = aws_iam_role.test_updated.arn
  %[2]s
}

resource "aws_iam_role_policy" "test_updated" {
  name = "%[1]s-updated"
  role = aws_iam_role.test_updated.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "autoscaling:CompleteLifecycleAction",
        "autoscaling:DeleteLifecycleHook",
        "autoscaling:DescribeAutoScalingGroups",
        "autoscaling:DescribeLifecycleHooks",
        "autoscaling:PutLifecycleHook",
        "autoscaling:RecordLifecycleActionHeartbeat",
        "ec2:DescribeInstances",
        "ec2:DescribeInstanceStatus",
        "tag:GetTags",
        "tag:GetResources"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_role" "test_updated" {
  name = "%[1]s-updated"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "codedeploy.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}
`, rName, tagGroupOrFilter))
}

func testAccDeploymentGroupConfig_onPremiseTags(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  on_premises_instance_tag_filter {
    key   = "filterkey"
    type  = "KEY_AND_VALUE"
    value = "filtervalue"
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_triggerConfigurationCreate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  trigger_configuration {
    trigger_events     = ["DeploymentFailure"]
    trigger_name       = "test-trigger"
    trigger_target_arn = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_triggerConfigurationUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  trigger_configuration {
    trigger_events     = ["DeploymentSuccess", "DeploymentFailure"]
    trigger_name       = "test-trigger"
    trigger_target_arn = aws_sns_topic.test.arn
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_triggerConfigurationCreateMultiple(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  trigger_configuration {
    trigger_events     = ["DeploymentFailure"]
    trigger_name       = "test-trigger-1"
    trigger_target_arn = aws_sns_topic.test.arn
  }

  trigger_configuration {
    trigger_events     = ["InstanceFailure"]
    trigger_name       = "test-trigger-2"
    trigger_target_arn = aws_sns_topic.test_2.arn
  }
}

resource "aws_sns_topic" "test_2" {
  name = "%[1]s-2"
}
`, rName))
}

func testAccDeploymentGroupConfig_triggerConfigurationUpdateMultiple(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  trigger_configuration {
    trigger_events     = ["DeploymentStart", "DeploymentSuccess", "DeploymentFailure", "DeploymentStop"]
    trigger_name       = "test-trigger-1"
    trigger_target_arn = aws_sns_topic.test.arn
  }

  trigger_configuration {
    trigger_events     = ["InstanceFailure"]
    trigger_name       = "test-trigger-2"
    trigger_target_arn = aws_sns_topic.test_3.arn
  }
}

resource "aws_sns_topic" "test_2" {
  name = "%[1]s-2"
}

resource "aws_sns_topic" "test_3" {
  name = "%[1]s-3"
}
`, rName))
}

func testAccDeploymentGroupConfig_autoRollbackConfigurationCreate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_autoRollbackConfigurationUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE", "DEPLOYMENT_STOP_ON_ALARM"]
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_autoRollbackConfigurationNone(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn
}
`, rName))
}

func testAccDeploymentGroupConfig_autoRollbackConfigurationDisable(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  auto_rollback_configuration {
    enabled = false
    events  = ["DEPLOYMENT_FAILURE"]
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_alarmConfigurationCreate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  alarm_configuration {
    alarms  = ["test-alarm"]
    enabled = true
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_alarmConfigurationUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  alarm_configuration {
    alarms                    = ["test-alarm", "test-alarm-2"]
    enabled                   = true
    ignore_poll_alarm_failure = true
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_alarmConfigurationNone(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn
}
`, rName))
}

func testAccDeploymentGroupConfig_alarmConfigurationDisable(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  alarm_configuration {
    alarms  = ["test-alarm"]
    enabled = false
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_styleDefault(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn
}
`, rName))
}

func testAccDeploymentGroupConfig_styleCreate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_styleUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITHOUT_TRAFFIC_CONTROL"
    deployment_type   = "IN_PLACE"
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_loadBalancerInfoNone(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn
}
`, rName))
}

func testAccDeploymentGroupConfig_loadBalancerInfoCreate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_loadBalancerInfoUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group-2"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoCreate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  load_balancer_info {
    target_group_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  load_balancer_info {
    target_group_info {
      name = "acc-test-codedeploy-dep-group-2"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_loadBalancerInfoTargetInfoDelete(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn
}
`, rName))
}

func testAccDeploymentGroupConfig_inPlaceTrafficControlCreate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "IN_PLACE"
  }

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_inPlaceTrafficControlUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_baseBlueGreenConfigASG(rName string) string {
	return acctest.ConfigCompose(
		testAccDeploymentGroupConfig_base(rName),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  default_for_az    = "true"
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = "t2.micro"
  name          = %[1]q

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "test" {
  name             = %[1]q
  max_size         = 2
  min_size         = 0
  desired_capacity = 1

  vpc_zone_identifier = [data.aws_subnet.test.id]

  launch_configuration = aws_launch_configuration.test.name

  lifecycle {
    create_before_destroy = true
  }

  tag {
    key                 = "Name"
    value               = %[1]q
    propagate_at_launch = true
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_blueGreenConfigDelete(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn
}
`, rName))
}

func testAccDeploymentGroupConfig_blueGreenConfigCreateASG(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_baseBlueGreenConfigASG(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  autoscaling_groups = [aws_autoscaling_group.test.name]

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout    = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "COPY_AUTO_SCALING_GROUP"
    }

    terminate_blue_instances_on_deployment_success {
      action                           = "TERMINATE"
      termination_wait_time_in_minutes = 120
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_blueGreenConfigUpdateASG(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_baseBlueGreenConfigASG(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  autoscaling_groups = [aws_autoscaling_group.test.name]

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout    = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "COPY_AUTO_SCALING_GROUP"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_blueGreenConfigCreateNoASG(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout    = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action                           = "TERMINATE"
      termination_wait_time_in_minutes = 120
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_blueGreenConfigUpdateNoASG(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_blueGreenComplete(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout    = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 60
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_blueGreenCompleteUpdated(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  load_balancer_info {
    elb_info {
      name = "acc-test-codedeploy-dep-group"
    }
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    green_fleet_provisioning_option {
      action = "DISCOVER_EXISTING"
    }

    terminate_blue_instances_on_deployment_success {
      action = "KEEP_ALIVE"
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_ecsBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    cidr_blocks = [aws_vpc.test.cidr_block]
    from_port   = 80
    protocol    = "6"
    to_port     = 8000
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "blue" {
  name        = format("%%s-blue", substr(aws_lb.test.name, 0, 26))
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.test.id
}

resource "aws_lb_target_group" "green" {
  name        = format("%%s-green", substr(aws_lb.test.name, 0, 26))
  port        = 80
  protocol    = "HTTP"
  target_type = "ip"
  vpc_id      = aws_vpc.test.id
}

resource "aws_lb" "test" {
  internal = true
  name     = %[1]q
  subnets  = aws_subnet.test[*].id
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_lb_target_group.blue.arn
    type             = "forward"
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  cpu                      = "256"
  family                   = %[1]q
  memory                   = "512"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]

  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "test",
    "networkMode": "awsvpc",
    "portMappings": [
      {
        "containerPort": 80,
        "hostPort": 80
      }
    ]
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  cluster         = aws_ecs_cluster.test.id
  desired_count   = 1
  launch_type     = "FARGATE"
  name            = %[1]q
  task_definition = aws_ecs_task_definition.test.arn

  deployment_controller {
    type = "CODE_DEPLOY"
  }

  load_balancer {
    container_name   = "test"
    container_port   = "80"
    target_group_arn = aws_lb_target_group.blue.id
  }

  network_configuration {
    assign_public_ip = true
    security_groups  = [aws_security_group.test.id]
    subnets          = aws_subnet.test[*].id
  }
}

resource "aws_codedeploy_app" "test" {
  compute_platform = "ECS"
  name             = %[1]q
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "codedeploy.${data.aws_partition.current.dns_suffix}"
        ]
      }
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "cloudwatch:DescribeAlarms",
        "ecs:CreateTaskSet",
        "ecs:DeleteTaskSet",
        "ecs:DescribeServices",
        "ecs:UpdateServicePrimaryTaskSet",
        "elasticloadbalancing:DescribeListeners",
        "elasticloadbalancing:DescribeRules",
        "elasticloadbalancing:DescribeTargetGroups",
        "elasticloadbalancing:ModifyListener",
        "elasticloadbalancing:ModifyRule",
        "lambda:InvokeFunction",
        "s3:GetObject",
        "s3:GetObjectMetadata",
        "s3:GetObjectVersion",
        "sns:Publish"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName))
}

func testAccDeploymentGroupConfig_ecsBlueGreen(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_ecsBase(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name               = aws_codedeploy_app.test.name
  deployment_config_name = "CodeDeployDefault.ECSAllAtOnce"
  deployment_group_name  = %[1]q
  service_role_arn       = aws_iam_role.test.arn

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout = "CONTINUE_DEPLOYMENT"
    }

    terminate_blue_instances_on_deployment_success {
      action                           = "TERMINATE"
      termination_wait_time_in_minutes = 5
    }
  }

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  ecs_service {
    cluster_name = aws_ecs_cluster.test.name
    service_name = aws_ecs_service.test.name
  }

  load_balancer_info {
    target_group_pair_info {
      prod_traffic_route {
        listener_arns = [aws_lb_listener.test.arn]
      }

      target_group {
        name = aws_lb_target_group.blue.name
      }

      target_group {
        name = aws_lb_target_group.green.name
      }
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_ecsBlueGreenUpdate(rName string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_ecsBase(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name               = aws_codedeploy_app.test.name
  deployment_config_name = "CodeDeployDefault.ECSAllAtOnce"
  deployment_group_name  = %[1]q
  service_role_arn       = aws_iam_role.test.arn

  auto_rollback_configuration {
    enabled = true
    events  = ["DEPLOYMENT_FAILURE"]
  }

  blue_green_deployment_config {
    deployment_ready_option {
      action_on_timeout    = "STOP_DEPLOYMENT"
      wait_time_in_minutes = 30
    }

    terminate_blue_instances_on_deployment_success {
      action                           = "TERMINATE"
      termination_wait_time_in_minutes = 60
    }
  }

  deployment_style {
    deployment_option = "WITH_TRAFFIC_CONTROL"
    deployment_type   = "BLUE_GREEN"
  }

  ecs_service {
    cluster_name = aws_ecs_cluster.test.name
    service_name = aws_ecs_service.test.name
  }

  load_balancer_info {
    target_group_pair_info {
      prod_traffic_route {
        listener_arns = [aws_lb_listener.test.arn]
      }

      target_group {
        name = aws_lb_target_group.blue.name
      }

      target_group {
        name = aws_lb_target_group.green.name
      }
    }
  }
}
`, rName))
}

func testAccDeploymentGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDeploymentGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name              = aws_codedeploy_app.test.name
  deployment_group_name = %[1]q
  service_role_arn      = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccDeploymentGroupConfig_outdatedInstancesStrategy(rName string, outdatedInstancesStrategy string) string {
	return acctest.ConfigCompose(testAccDeploymentGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_codedeploy_deployment_group" "test" {
  app_name                    = aws_codedeploy_app.test.name
  deployment_group_name       = %[1]q
  service_role_arn            = aws_iam_role.test.arn
  outdated_instances_strategy = %[2]q
}
`, rName, outdatedInstancesStrategy))
}
