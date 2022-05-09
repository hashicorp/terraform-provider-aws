package dlm_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdlm "github.com/hashicorp/terraform-provider-aws/internal/service/dlm"
)

func TestAccDLMLifecyclePolicy_basic(t *testing.T) {
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dlm", regexp.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
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
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "basic"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyEventConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dlm", regexp.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.policy_type", "EVENT_BASED_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.target", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.encryption_configuration.0.encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.0.interval", "15"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.action.0.cross_region_copy.0.retain_rule.0.interval_unit", "MONTHS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.type", "MANAGED_CWE"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.0.description_regex", "^.*Created for policy: policy-1234567890abcdef0.*$"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.event_source.0.parameters.0.event_type", "shareSnapshot"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.event_source.0.parameters.0.snapshot_owner.0", "data.aws_caller_identity.current", "account_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyCronConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyRetainIntervalConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyDeprecateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
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

func TestAccDLMLifecyclePolicy_fastRestore(t *testing.T) {
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyFastRestoreConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyShareRuleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.share_rule.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.schedule.0.share_rule.0.target_accounts.0", "data.aws_caller_identity.current", "account_id"),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyParametersInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.exclude_boot_volume", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.no_reboot", "false"),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyParametersVolumeConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.exclude_boot_volume", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.parameters.0.no_reboot", "false"),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyVariableTagsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
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
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyFullConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-full"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "21:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "10"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.copy_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: dlmLifecyclePolicyFullUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-full-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "24"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "09:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "100"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.copy_tags", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full-updated"),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_crossRegionCopyRule(t *testing.T) {
	var providers []*schema.Provider

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dlm_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyConfigCrossRegionCopyRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.encrypted", "false"),
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
				Config: dlmLifecyclePolicyConfigUpdateCrossRegionCopyRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.cmk_arn", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.copy_tags", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.0.encrypted", "true"),
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
				Config: dlmLifecyclePolicyConfigNoCrossRegionCopyRule(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.cross_region_copy_rule.#", "0"),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dlm_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
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
				Config: dlmLifecyclePolicyConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: dlmLifecyclePolicyConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDLMLifecyclePolicy_disappears(t *testing.T) {
	resourceName := "aws_dlm_lifecycle_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, dlm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdlm.ResourceLifecyclePolicy(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdlm.ResourceLifecyclePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func dlmLifecyclePolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DLMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dlm_lifecycle_policy" {
			continue
		}

		input := dlm.GetLifecyclePolicyInput{
			PolicyId: aws.String(rs.Primary.ID),
		}

		out, err := conn.GetLifecyclePolicy(&input)

		if tfawserr.ErrCodeEquals(err, dlm.ErrCodeResourceNotFoundException) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error getting DLM Lifecycle Policy (%s): %s", rs.Primary.ID, err)
		}

		if out.Policy != nil {
			return fmt.Errorf("DLM lifecycle policy still exists: %#v", out)
		}
	}

	return nil
}

func checkDlmLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DLMConn

		input := dlm.GetLifecyclePolicyInput{
			PolicyId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetLifecyclePolicy(&input)

		if err != nil {
			return fmt.Errorf("error getting DLM Lifecycle Policy (%s): %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DLMConn

	input := &dlm.GetLifecyclePoliciesInput{}

	_, err := conn.GetLifecyclePolicies(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func dlmLifecyclePolicyBaseConfig(rName string) string {
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

func dlmLifecyclePolicyBasicConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyEventConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), fmt.Sprintf(`
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

func dlmLifecyclePolicyCronConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyRetainIntervalConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyDeprecateConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyFastRestoreConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyShareRuleConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyParametersInstanceConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyParametersVolumeConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyVariableTagsConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyFullConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyFullUpdateConfig(rName string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), `
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

func dlmLifecyclePolicyConfigCrossRegionCopyRule(rName string) string {
	return acctest.ConfigCompose(
		dlmLifecyclePolicyBaseConfig(rName),
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

func dlmLifecyclePolicyConfigUpdateCrossRegionCopyRule(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		dlmLifecyclePolicyBaseConfig(rName),
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

func dlmLifecyclePolicyConfigNoCrossRegionCopyRule(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		dlmLifecyclePolicyBaseConfig(rName),
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

func dlmLifecyclePolicyConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), fmt.Sprintf(`
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

func dlmLifecyclePolicyConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(dlmLifecyclePolicyBaseConfig(rName), fmt.Sprintf(`
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
