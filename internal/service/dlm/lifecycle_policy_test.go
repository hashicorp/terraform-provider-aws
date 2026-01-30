// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dlm_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/dlm"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdlm "github.com/hashicorp/terraform-provider-aws/internal/service/dlm"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDLMLifecyclePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_type", "EBS_SNAPSHOT_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "10"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.deprecate_rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", acctest.CtBasic),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_event(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dlm", regexache.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_type", "EVENT_BASED_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.target", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.encryption_configuration.0.encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.0.interval", "15"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.0.interval_unit", "MONTHS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.type", "MANAGED_CWE"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.0.description_regex", "^.*Created for policy: policy-1234567890abcdef0.*$"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.0.event_type", "shareSnapshot"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.event_source.0.parameters.0.snapshot_owner.0", "data.aws_caller_identity.current", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_cron(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
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

func TestAccDLMLifecyclePolicy_scriptsAlias(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_scriptsAlias(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.execution_handler", "AWS_VSS_BACKUP"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.execute_operation_on_script_failure", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.maximum_retry_count", "3"),
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

func TestAccDLMLifecyclePolicy_scriptsSSMDocument(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_scriptsSSMDocument(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.execution_handler"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.execute_operation_on_script_failure", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.execution_timeout", "60"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.maximum_retry_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.scripts.0.stages.0", "PRE"),
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

func TestAccDLMLifecyclePolicy_archiveRuleCount(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_archiveRuleCount(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.archive_rule.0.archive_retain_rule.0.retention_archive_tier.0.count", "10"),
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

func TestAccDLMLifecyclePolicy_archiveRuleInterval(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_archiveRuleInterval(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.archive_rule.0.archive_retain_rule.0.retention_archive_tier.0.interval", "6"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.archive_rule.0.archive_retain_rule.0.retention_archive_tier.0.interval_unit", "MONTHS"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_retainInterval(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.interval", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_deprecate(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.deprecate_rule.0.count", "10"),
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

func TestAccDLMLifecyclePolicy_defaultPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_defaultPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dlm", regexache.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "default_policy", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.copy_tags", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.create_interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.extend_deletion", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_type", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.retain_interval", "7"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_language", "SIMPLIFIED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.#", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_policy"},
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_defaultPolicyExclusions(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_defaultPolicyExclusions(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "dlm", regexache.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "default_policy", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.copy_tags", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.create_interval", "5"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.extend_deletion", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_type", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.retain_interval", "7"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_language", "SIMPLIFIED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusions.0.exclude_boot_volumes", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusions.0.exclude_tags.test", "exclude"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.exclusions.0.exclude_volume_types.0", "gp2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_policy"},
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_fastRestore(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_fastRestore(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.fast_restore_rule.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.schedule.0.fast_restore_rule.0.availability_zones.#", "data.aws_availability_zones.available", "names.#"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.fast_restore_rule.0.count", "10"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_shareRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.share_rule.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_parametersInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_parametersVolume(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", "1"),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_variableTags(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "tf-acc-full"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrExecutionRoleARN),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "21:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "10"),
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
					checkLifecyclePolicyExists(ctx, t, resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dlm_lifecycle_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_crossRegionCopyRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", "1"),
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
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", "1"),
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
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", "0"),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_crossRegionCopyRuleImageManagement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dlm_lifecycle_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_crossRegionCopyRuleImageManagement(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_type", "IMAGE_MANAGEMENT"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.encrypted", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.retain_rule.0.interval", "15"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.retain_rule.0.interval_unit", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.target_region", acctest.AlternateRegion()),
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

func TestAccDLMLifecyclePolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_dlm_lifecycle_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
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
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccLifecyclePolicyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DLMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLifecyclePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLifecyclePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					checkLifecyclePolicyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdlm.ResourceLifecyclePolicy(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdlm.ResourceLifecyclePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckLifecyclePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DLMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dlm_lifecycle_policy" {
				continue
			}

			_, err := tfdlm.FindLifecyclePolicyByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
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

func checkLifecyclePolicyExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.DLM, create.ErrActionCheckingExistence, tfdlm.ResNameLifecyclePolicy, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).DLMClient(ctx)

		_, err := tfdlm.FindLifecyclePolicyByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return create.Error(names.DLM, create.ErrActionCheckingExistence, tfdlm.ResNameLifecyclePolicy, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DLMClient(ctx)

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

func testAccLifecyclePolicyConfig_defaultPolicy(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn
  default_policy     = "VOLUME"

  policy_details {
    create_interval = 5
    resource_type   = "VOLUME"
    policy_language = "SIMPLIFIED"
  }
}
`)
}

func testAccLifecyclePolicyConfig_defaultPolicyExclusions(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn
  default_policy     = "VOLUME"

  policy_details {
    create_interval = 5
    resource_type   = "VOLUME"
    policy_language = "SIMPLIFIED"

    exclusions {
      exclude_boot_volumes = false
      exclude_tags = {
        test = "exclude"
      }
      exclude_volume_types = ["gp2"]
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

func testAccLifecyclePolicyConfig_archiveRuleCount(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        cron_expression = "cron(5 14 3 * ? *)"
      }

      archive_rule {
        archive_retain_rule {
          retention_archive_tier {
            count = 10
          }
        }
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

func testAccLifecyclePolicyConfig_archiveRuleInterval(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        cron_expression = "cron(5 14 3 * ? *)"
      }

      archive_rule {
        archive_retain_rule {
          retention_archive_tier {
            interval      = 6
            interval_unit = "MONTHS"
          }
        }
      }

      retain_rule {
        interval      = 12
        interval_unit = "MONTHS"
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`)
}

func testAccLifecyclePolicyConfig_scriptsAlias(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
data "aws_iam_policy" "test" {
  name = "AWSDataLifecycleManagerSSMFullAccess"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.id
  policy_arn = data.aws_iam_policy.test.arn
}

resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["INSTANCE"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
        scripts {
          execute_operation_on_script_failure = false
          execution_handler                   = "AWS_VSS_BACKUP"
          maximum_retry_count                 = 3
        }
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

func testAccLifecyclePolicyConfig_scriptsSSMDocument(rName string) string {
	return acctest.ConfigCompose(lifecyclePolicyBaseConfig(rName), `
data "aws_iam_policy" "test" {
  name = "AWSDataLifecycleManagerSSMFullAccess"
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.id
  policy_arn = data.aws_iam_policy.test.arn
}

resource "aws_ssm_document" "test" {
  name          = "tf-acc-basic"
  document_type = "Command"

  tags = {
    DLMScriptsAccess = "true"
  }

  content = <<DOC
  {
    "schemaVersion": "2.2",
    "description": "SSM Document Template for Amazon Data Lifecycle Manager Pre/Post script feature",
    "parameters": {
		"executionId": {
			"type": "String",
			"default": "None",
			"description": "(Required) Specifies the unique identifier associated with a pre and/or post execution",
			"allowedPattern": "^(None|[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12})$"
			},
		"command": {
			"type": "String",
			"default": "dry-run",
			"description": "(Required) Specifies whether pre-script and/or post-script should be executed.",
			"allowedValues": [
				"pre-script",
				"post-script",
				"dry-run"
				]
			}
  	},
   "mainSteps": [
      {
      "action": "aws:runShellScript",
      "description": "Run Database freeze/thaw commands",
      "name": "run_pre_post_scripts",
      "inputs": {
        "runCommand": [ "#!/bin/bash\n\n"]
	    }
      }
    ]
  }
DOC
}

resource "aws_dlm_lifecycle_policy" "test" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["INSTANCE"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
        scripts {
          execute_operation_on_script_failure = false
          execution_handler                   = aws_ssm_document.test.arn
          execution_timeout                   = 60
          maximum_retry_count                 = 3
          stages                              = ["PRE"]
        }
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

func testAccLifecyclePolicyConfig_crossRegionCopyRuleImageManagement(rName string) string {
	return acctest.ConfigCompose(
		lifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    policy_type    = "IMAGE_MANAGEMENT"
    resource_types = ["INSTANCE"]

    schedule {
      name = %[1]q

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }

      cross_region_copy_rule {
        target_region = %[2]q
        encrypted     = false
        retain_rule {
          interval      = 15
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

func testAccLifecyclePolicyConfig_updateCrossRegionCopyRule(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		lifecyclePolicyBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider            = "awsalternate"
  description         = %[1]q
  enable_key_rotation = true

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
