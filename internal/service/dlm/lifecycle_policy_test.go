// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dlm_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dlm"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfdlm "github.com/hashicorp/terraform-provider-aws/internal/service/dlm"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDLMLifecyclePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "dlm", regexache.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_type", "EBS_SNAPSHOT_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.deprecate_rule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", acctest.CtBasic),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccDLMLifecyclePolicy_event(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_event(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "dlm", regexache.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_type", "EVENT_BASED_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.target", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.encryption_configuration.0.encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.0.interval", "15"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.0.interval_unit", "MONTHS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.type", "MANAGED_CWE"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.0.description_regex", "^.*Created for policy: policy-1234567890abcdef0.*$"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.0.event_type", "shareSnapshot"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.event_source.0.parameters.0.snapshot_owner.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccDLMLifecyclePolicy_cron(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_cron(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.cron_expression", "cron(0 18 ? * WED *)"),
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

func TestAccDLMLifecyclePolicy_retainInterval(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_retainInterval(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.interval", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.interval_unit", "DAYS"),
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

func TestAccDLMLifecyclePolicy_deprecate(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_deprecate(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.deprecate_rule.0.count", acctest.Ct10),
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

func TestAccDLMLifecyclePolicy_fastRestore(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_fastRestore(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.fast_restore_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.schedule.0.fast_restore_rule.0.availability_zones.#", "data.aws_availability_zones.available", "names.#"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.fast_restore_rule.0.count", acctest.Ct10),
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

func TestAccDLMLifecyclePolicy_shareRule(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_shareRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.share_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.schedule.0.share_rule.0.target_accounts.0", "data.aws_caller_identity.current", names.AttrAccountID),
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

func TestAccDLMLifecyclePolicy_parameters_instance(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_parametersInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.exclude_boot_volume", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.no_reboot", acctest.CtFalse),
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

func TestAccDLMLifecyclePolicy_parameters_volume(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_parametersVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.exclude_boot_volume", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.no_reboot", acctest.CtFalse),
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

func TestAccDLMLifecyclePolicy_variableTags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_variableTags(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.variable_tags.instance_id", "$(instance-id)"),
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

func TestAccDLMLifecyclePolicy_full(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-full"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "21:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.copy_tags", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLifecyclePolicyConfig_fullUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-full-updated"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "24"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "09:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "100"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.copy_tags", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full-updated"),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_crossRegionCopyRule(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dlm_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_crossRegionCopyRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.retain_rule.0.interval", "15"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.retain_rule.0.interval_unit", "MONTHS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.target", acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLifecyclePolicyConfig_updateCrossRegionCopyRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.cmk_arn", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.copy_tags", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.encrypted", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.retain_rule.0.interval", "30"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.retain_rule.0.interval_unit", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.target", acctest.AlternateRegion()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLifecyclePolicyConfig_noCrossRegionCopyRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dlm_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccLifecyclePolicyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLifecyclePolicyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdlm.ResourceLifecyclePolicy(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdlm.ResourceLifecyclePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLifecyclePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DLMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dlm_lifecycle_policy" {
				continue
			}

			_, err := tfdlm.FindLifecyclePolicyByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.DLM, create.ErrActionCheckingDestroyed, tfdlm.ResNameLifecyclePolicy, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func checkLifecyclePolicyExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DLM, create.ErrActionCheckingExistence, tfdlm.ResNameLifecyclePolicy, name, errors.New("not found"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DLMClient(ctx)

		_, err := tfdlm.FindLifecyclePolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.DLM, create.ErrActionCheckingExistence, tfdlm.ResNameLifecyclePolicy, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DLMClient(ctx)

	input := &dlm.GetLifecyclePoliciesInput{}

	_, err := conn.GetLifecyclePolicies(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func lifecyclePolicyBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}
`, rName)
}

func testAccLifecyclePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_event(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_policy" "test" {
  name = "AWSDataLifecycleManagerServiceRole"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.id
  policy_arn = data.aws_iam_policy.test.arn
}

resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    policy_type = "EVENT_BASED_POLICY"

    action {
      name = "tf-acc-basic"
      cross_region_copy {
        encryption_configuration {}
        retain_rule {
          interval      = 15
          interval_unit = "MONTHS"
        }

        target = %[1]q
      }
    }

    event_source {
      type = "MANAGED_CWE"
      parameters {
        description_regex = "^.*Created for policy: policy-1234567890abcdef0.*$"
        event_type        = "shareSnapshot"
        snapshot_owner    = [data.aws_caller_identity.current.account_id]
      }
    }
  }
}
`, acctest.AlternateRegion()))
}

func testAccLifecyclePolicyConfig_cron(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        cron_expression = "cron(0 18 ? * WED *)"
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_retainInterval(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        interval_unit = "DAYS"
        interval      = 1
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_deprecate(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["INSTANCE"]
    policy_type    = "IMAGE_MANAGEMENT"

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }

      deprecate_rule {
        count = 10
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_fastRestore(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]
    policy_type    = "EBS_SNAPSHOT_MANAGEMENT"

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }

      fast_restore_rule {
        availability_zones = data.aws_availability_zones.available.names
        count              = 10
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_shareRule(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
data "aws_caller_identity" "current" {}

resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]
    policy_type    = "EBS_SNAPSHOT_MANAGEMENT"

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }

      share_rule {
        target_accounts = [data.aws_caller_identity.current.account_id]
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_parametersInstance(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["INSTANCE"]
    policy_type    = "IMAGE_MANAGEMENT"

    parameters {
      no_reboot = false
    }

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_parametersVolume(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["INSTANCE"]
    policy_type    = "EBS_SNAPSHOT_MANAGEMENT"

    parameters {
      exclude_boot_volume = true
    }

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_variableTags(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["INSTANCE"]
    policy_type    = "IMAGE_MANAGEMENT"

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }

      variable_tags = {
        instance_id = "$(instance-id)"
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_full(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-full"
  execution_role_arn = aws_iam_role.test.arn
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-full"

      create_rule {
        interval      = 12
        interval_unit = "HOURS"
        times         = ["21:42"]
      }

      retain_rule {
        count = 10
      }

      tags_to_add = {
        tf-acc-test-added = "full"
      }

      copy_tags = false
    }

    target_tags = {
      tf-acc-test = "full"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_fullUpdate(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-full-updated"
  execution_role_arn = "${aws_iam_role.test.arn}-doesnt-exist"
  state              = "DISABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-full-updated"

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["09:42"]
      }

      retain_rule {
        count = 100
      }

      tags_to_add = {
        tf-acc-test-added = "full-updated"
      }

      copy_tags = true
    }

    target_tags = {
      tf-acc-test = "full-updated"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_crossRegionCopyRule(rName string) string {
	return acctest.ConfigCompose(
		lifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = %[1]q

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }

      cross_region_copy_rule {
        target    = %[2]q
        encrypted = false
        retain_rule {
          interval      = 15
          interval_unit = "MONTHS"
        }
      }
    }

    target_tags = {
      Name = %[1]q
    }
  }
}
`, rName, acctest.AlternateRegion()))
}

func testAccLifecyclePolicyConfig_updateCrossRegionCopyRule(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		lifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider    = "awsalternate"
  description = %[1]q

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": %[1]q,
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

resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = %[1]q

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }

      cross_region_copy_rule {
        target    = %[2]q
        encrypted = true
        cmk_arn   = aws_kms_key.test.arn
        copy_tags = true
        retain_rule {
          interval      = 30
          interval_unit = "DAYS"
        }
      }
    }

    target_tags = {
      Name = %[1]q
    }
  }
}
`, rName, acctest.AlternateRegion()))
}

func testAccLifecyclePolicyConfig_noCrossRegionCopyRule(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		lifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = %[1]q

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      Name = %[1]q
    }
  }
}
`, rName))
}

func testAccLifecyclePolicyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), fmt.Sprintf(`
resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "test"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      test = "true"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccLifecyclePolicyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), fmt.Sprintf(`
resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "test"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      test = "true"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
