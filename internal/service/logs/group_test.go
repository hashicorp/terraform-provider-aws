// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsLogGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "logs", "log-group:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, "log_group_class", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
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

func TestAccLogsLogGroup_nameGenerate(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
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

func TestAccLogsLogGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccLogsLogGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					acctest.CheckSDKResourceDisappears(ctx, t, tflogs.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLogsLogGroup_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"
	kmsKey1ResourceName := "aws_kms_key.test.0"
	kmsKey2ResourceName := "aws_kms_key.test.1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_kmsKey(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKey1ResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_kmsKey(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKey2ResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_logGroupClass(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// CloudWatch Logs IA is available in all AWS Commercial regions.
			acctest.PreCheckPartition(t, endpoints.AwsPartitionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_logGroupClass(rName, "INFREQUENT_ACCESS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "log_group_class", "INFREQUENT_ACCESS"),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_retentionPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 1096),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "1096"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupConfig_retentionPolicy(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resource1Name := "aws_cloudwatch_log_group.test.0"
	resource2Name := "aws_cloudwatch_log_group.test.1"
	resource3Name := "aws_cloudwatch_log_group.test.2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resource1Name, &v1),
					testAccCheckLogGroupExists(ctx, t, resource2Name, &v2),
					testAccCheckLogGroupExists(ctx, t, resource3Name, &v3),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_skipDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupNoDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_skipDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_skipDestroyInconsistentPlan(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
				),
			},
			{
				Config: testAccGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrSkipDestroy, acctest.CtFalse),
				),
			},
		},
	})
}

// Test whether the log group is successfully created with the DELIVERY log group class when retention_in_days is set.
// Even if retention_in_days is changed in the configuration, the diff should be suppressed and the plan should be empty.
func TestAccLogsLogGroup_logGroupClassDELIVERY1(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_logGroupClassDEVIVERYWithRetentionInDays(rName, 30),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					// AWS API forces retention_in_days to 2 for DELIVERY log group class
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "2"),
				),
			},
			{
				Config: testAccGroupConfig_logGroupClassDEVIVERYWithRetentionInDays(rName, 60),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: testAccGroupConfig_logGroupClassDEVIVERYWithoutRetentionInDays(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// Test whether the log group is successfully created with the DELIVERY log group class when retention_in_days is not set.
func TestAccLogsLogGroup_logGroupClassDELIVERY2(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_logGroupClassDEVIVERYWithoutRetentionInDays(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					// AWS API forces retention_in_days to 2 for DELIVERY log group class
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "2"),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_requiredTags(t *testing.T) {
	ctx := acctest.Context(t)
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
			// Creation with required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccGroupConfig_tags1(rName, tagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
			// Updates which remove required tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccGroupConfig_tags1(rName, nonRequiredTagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
		},
	})
}

func TestAccLogsLogGroup_requiredTags_defaultTags(t *testing.T) {
	ctx := acctest.Context(t)
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("error"),
					testAccGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
			// Creation with required tags in default_tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyComplianceAndDefaultTags1("error", tagKey, acctest.CtValue1),
					testAccGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
			// Updates which remove required tags from default_tags fail
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyComplianceAndDefaultTags1("error", nonRequiredTagKey, acctest.CtValue1),
					testAccGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				ExpectError: regexache.MustCompile("Missing Required Tags"),
			},
		},
	})
}

func TestAccLogsLogGroup_requiredTags_warning(t *testing.T) {
	ctx := acctest.Context(t)
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("warning"),
					testAccGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
			// Updates adding required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("warning"),
					testAccGroupConfig_tags1(rName, tagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
			// Updates which remove required tags also succeed
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("warning"),
					testAccGroupConfig_tags1(rName, nonRequiredTagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_requiredTags_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	tagKey := acctest.SkipIfEnvVarNotSet(t, "TF_ACC_REQUIRED_TAG_KEY")
	nonRequiredTagKey := "NotARequiredKey"

	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckResourceGroupsTaggingAPIRequiredTags(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			// New resources missing required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("disabled"),
					testAccGroupConfig_basic(rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
			// Updates adding required tags succeeds
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("disabled"),
					testAccGroupConfig_tags1(rName, tagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							tagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						tagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
			// Updates which remove required tags also succeed
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigTagPolicyCompliance("disabled"),
					testAccGroupConfig_tags1(rName, nonRequiredTagKey, acctest.CtValue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
						})),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
						nonRequiredTagKey: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func TestAccLogsLogGroup_deletionProtectionEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig_deletionProtectionEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection_enabled", acctest.CtTrue),
				),
			},
			{
				// Disable deletion protection
				Config: testAccGroupConfig_deletionProtectionEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

// A smoke test to verify setting provider_meta does not trigger unexpected
// behavior for Plugin SDK V2 based resources
func TestAccLogsLogGroup_providerMeta(t *testing.T) {
	ctx := acctest.Context(t)
	var v types.LogGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_cloudwatch_log_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLogGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					acctest.ConfigProviderMeta(),
					testAccGroupConfig_basic(rName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLogGroupExists(ctx, t, resourceName, &v),
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

func testAccCheckLogGroupExists(ctx context.Context, t *testing.T, n string, v *types.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		output, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckLogGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_group" {
				continue
			}

			_, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Log Group still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGroupNoDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_group" {
				continue
			}

			_, err := tflogs.FindLogGroupByName(ctx, conn, rs.Primary.ID)

			return err
		}

		return nil
	}
}

func testAccGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccGroupConfig_nameGenerated() string {
	return `
resource "aws_cloudwatch_log_group" "test" {}
`
}

func testAccGroupConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name_prefix = %[1]q
}
`, namePrefix)
}

func testAccGroupConfig_kmsKey(rName string, idx int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  description             = "%[1]s-${count.index}"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": "kms:*",
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test[%[2]d].arn
}
`, rName, idx)
}

func testAccGroupConfig_logGroupClass(rName string, val string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name            = %[1]q
  log_group_class = %[2]q
}
`, rName, val)
}

func testAccGroupConfig_retentionPolicy(rName string, val int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  retention_in_days = %[2]d
}
`, rName, val)
}

func testAccGroupConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  count = 3

  name = "%[1]s-${count.index}"
}
`, rName)
}

func testAccGroupConfig_skipDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name         = %[1]q
  skip_destroy = true
}
`, rName)
}

func testAccGroupConfig_logGroupClassDEVIVERYWithRetentionInDays(rName string, retentionInDays int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = %[1]q
  log_group_class   = "DELIVERY"
  retention_in_days = %[2]d
}
`, rName, retentionInDays)
}

func testAccGroupConfig_logGroupClassDEVIVERYWithoutRetentionInDays(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name            = %[1]q
  log_group_class = "DELIVERY"
}
`, rName)
}

func testAccGroupConfig_tags1(rName, key1, value1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q

  tags = {
    %[2]s = %[3]q
  }
}
`, rName, key1, value1)
}

func testAccGroupConfig_deletionProtectionEnabled(rName string, deletionProtectionEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q

  deletion_protection_enabled = %[2]t
}
`, rName, deletionProtectionEnabled)
}
