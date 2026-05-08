// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("glue", fmt.Sprintf("job/%s", rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("command"), knownvalue.ListExact([]knownvalue.Check{knownvalue.MapExact(map[string]knownvalue.Check{
						names.AttrName:    knownvalue.StringExact("glueetl"),
						"python_version":  knownvalue.NotNull(),
						"runtime":         knownvalue.NotNull(),
						"script_location": knownvalue.StringExact("testscriptlocation"),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("glue_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("job_mode"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(2880)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("worker_type"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGlueJob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglue.ResourceJob(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueJob_basicStreaming(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_requiredStreaming(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNExact("glue", fmt.Sprintf("job/%s", rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("command"), knownvalue.ListExact([]knownvalue.Check{knownvalue.MapExact(map[string]knownvalue.Check{
						names.AttrName:    knownvalue.StringExact("gluestreaming"),
						"python_version":  knownvalue.NotNull(),
						"runtime":         knownvalue.NotNull(),
						"script_location": knownvalue.StringExact("testscriptlocation"),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("glue_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("job_mode"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("worker_type"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGlueJob_command(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_command(rName, "testscriptlocation1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation1"),
				),
			},
			{
				Config: testAccJobConfig_command(rName, "testscriptlocation2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation2"),
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

func TestAccGlueJob_defaultArguments(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_defaultArguments(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccJobConfig_defaultArguments(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-bookmark-option", "job-bookmark-enable"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-language", "scala"),
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

func TestAccGlueJob_nonOverridableArguments(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_nonOverridableArguments(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccJobConfig_nonOverridableArguments(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-bookmark-option", "job-bookmark-enable"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-language", "scala"),
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

func TestAccGlueJob_description(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "First Description"),
				),
			},
			{
				Config: testAccJobConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Second Description"),
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

func TestAccGlueJob_glueVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_versionMaxCapacity(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
				),
			},
			{
				Config: testAccJobConfig_versionMaxCapacity(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "1.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_versionNumberOfWorkers(rName, "2.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "2.0"),
				),
			},
		},
	})
}

func TestAccGlueJob_executionClass(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_executionClass(rName, "FLEX"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_class", "FLEX"),
				),
			},
			{
				Config: testAccJobConfig_executionClass(rName, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_class", "STANDARD"),
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

func TestAccGlueJob_executionProperty(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_executionProperty(rName, 0),
				ExpectError: regexache.MustCompile(`expected execution_property.0.max_concurrent_runs to be at least`),
			},
			{
				Config: testAccJobConfig_executionProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "execution_property.0.max_concurrent_runs", "1"),
				),
			},
			{
				Config: testAccJobConfig_executionProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "execution_property.0.max_concurrent_runs", "2"),
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

func TestAccGlueJob_jobRunQueuingEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_jobRunQueuingEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "job_run_queuing_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_jobRunQueuingEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "job_run_queuing_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccGlueJob_maintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"
	maintenanceWindow := "Sun:23"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_maintenanceWindow(rName, maintenanceWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "Sun:23"),
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

func TestAccGlueJob_maxRetries(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_maxRetries(rName, 11),
				ExpectError: regexache.MustCompile(`expected max_retries to be in the range`),
			},
			{
				Config: testAccJobConfig_maxRetries(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_retries", "0"),
				),
			},
			{
				Config: testAccJobConfig_maxRetries(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_retries", "10"),
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

func TestAccGlueJob_notificationProperty(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_notificationProperty(rName, 0),
				ExpectError: regexache.MustCompile(`expected notification_property.0.notify_delay_after to be at least`),
			},
			{
				Config: testAccJobConfig_notificationProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_property.0.notify_delay_after", "1"),
				),
			},
			{
				Config: testAccJobConfig_notificationProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_property.0.notify_delay_after", "2"),
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

func TestAccGlueJob_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
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
				Config: testAccJobConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccJobConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueJob_StreamingTimeout_createNonNull(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_streamingTimeout(rName, "500"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(500)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_streamingTimeout(rName, "1500"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(1500)),
				},
			},
			{
				Config: testAccJobConfig_streamingTimeout(rName, "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(0)),
				},
			},
		},
	})
}

func TestAccGlueJob_StreamingTimeout_createNull(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_streamingTimeout(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(0)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_streamingTimeout(rName, "500"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(500)),
				},
			},
			{
				Config: testAccJobConfig_streamingTimeout(rName, "1500"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(1500)),
				},
			},
			{
				Config: testAccJobConfig_streamingTimeout(rName, "0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(0)),
				},
			},
		},
	})
}

func TestAccGlueJob_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_timeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(1)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_timeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(2)),
				},
			},
		},
	})
}

func TestAccGlueJob_security(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_securityConfiguration(rName, "default_encryption"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "default_encryption"),
				),
			},
			{
				Config: testAccJobConfig_securityConfiguration(rName, "custom_encryption2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "custom_encryption2"),
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

func TestAccGlueJob_workerType(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_workerType(rName, "Standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("worker_type"), knownvalue.StringExact("Standard")),
				},
			},
			{
				Config: testAccJobConfig_workerType(rName, "G.1X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("worker_type"), knownvalue.StringExact("G.1X")),
				},
			},
			{
				Config: testAccJobConfig_workerType(rName, "G.2X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("worker_type"), knownvalue.StringExact("G.2X")),
				},
			},
			{
				Config: testAccJobConfig_workerType(rName, "G.4X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("worker_type"), knownvalue.StringExact("G.4X")),
				},
			},
			{
				Config: testAccJobConfig_workerType(rName, "R.1X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("worker_type"), knownvalue.StringExact("R.1X")),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGlueJob_pythonShell(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_pythonShell(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "0.0625"),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_pythonShellVersion(rName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_pythonShellVersion(rName, "3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", "3"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
			{
				Config: testAccJobConfig_pythonShellVersion(rName, "3.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", "3.9"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
		},
	})
}

func TestAccGlueJob_rayJob(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_rayJob(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(480)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccGlueJob_rayJobUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_rayJobWithDescription(rName, "Initial job"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Initial job")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(480)),
				},
			},
			{
				Config: testAccJobConfig_rayJobWithDescription(rName, "Updated job"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Updated job")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTimeout), knownvalue.Int64Exact(480)),
				},
			},
		},
	})
}

func TestAccGlueJob_maxCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_maxCapacity(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "10"),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "glueetl"),
				),
			},
			{
				Config: testAccJobConfig_maxCapacity(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "15"),
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

func TestAccGlueJob_sourceControlDetails(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_sourceControlDetails(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.repository", rName),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.provider", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.branch", "test-branch"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.last_commit_id", "test-commit-id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_sourceControlDetails(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.repository", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.provider", "GITHUB"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.branch", "test-branch"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.0.last_commit_id", "test-commit-id"),
				),
			},
		},
	})
}

func TestAccGlueJob_jobMode(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_jobMode(rName, "NOTEBOOK"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "glue", fmt.Sprintf("job/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "execution_class", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, "2880"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "job_mode", "NOTEBOOK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_jobMode(rName, "VISUAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, t, resourceName, &job),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "glue", fmt.Sprintf("job/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "execution_class", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, "2880"),
					resource.TestCheckResourceAttr(resourceName, "source_control_details.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "job_mode", "VISUAL"),
				),
			},
		},
	})
}

func testAccCheckJobExists(ctx context.Context, t *testing.T, n string, v *awstypes.Job) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

		output, err := tfglue.FindJobByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckJobDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_job" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

			_, err := tfglue.FindJobByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glue Job %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccJobConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy" "AWSGlueServiceRole" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "glue.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = data.aws_iam_policy.AWSGlueServiceRole.arn
  role       = aws_iam_role.test.name
}
`, rName)
}

func testAccJobConfig_command(rName, scriptLocation string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = %[2]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, scriptLocation))
}

func testAccJobConfig_defaultArguments(rName, jobBookmarkOption, jobLanguage string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  default_arguments = {
    "--job-bookmark-option" = %[2]q
    "--job-language"        = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, jobBookmarkOption, jobLanguage))
}

func testAccJobConfig_nonOverridableArguments(rName, jobBookmarkOption, jobLanguage string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  non_overridable_arguments = {
    "--job-bookmark-option" = %[2]q
    "--job-language"        = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, jobBookmarkOption, jobLanguage))
}

func testAccJobConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  description  = %[1]q
  max_capacity = 10
  name         = %[2]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, description, rName))
}

func testAccJobConfig_versionMaxCapacity(rName, glueVersion string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  glue_version = %[1]q
  max_capacity = 10
  name         = %[2]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, glueVersion, rName))
}

func testAccJobConfig_versionNumberOfWorkers(rName, glueVersion string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  glue_version      = %[1]q
  name              = %[2]q
  number_of_workers = 1
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, glueVersion, rName))
}

func testAccJobConfig_executionClass(rName, executionClass string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  execution_class   = %[2]q
  name              = %[1]q
  number_of_workers = 2
  role_arn          = aws_iam_role.test.arn
  worker_type       = "G.1X"
  glue_version      = "3.0"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, executionClass))
}

func testAccJobConfig_executionProperty(rName string, maxConcurrentRuns int) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  execution_property {
    max_concurrent_runs = %[2]d
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maxConcurrentRuns))
}

func testAccJobConfig_maintenanceWindow(rName, maintenanceWindow string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name               = %[2]q
  role_arn           = aws_iam_role.test.arn
  maintenance_window = %[2]q

  command {
    name            = "gluestreaming"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maintenanceWindow))
}

func testAccJobConfig_maxRetries(rName string, maxRetries int) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  max_retries  = %[1]d
  name         = %[2]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, maxRetries, rName))
}

func testAccJobConfig_notificationProperty(rName string, notifyDelayAfter int) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  notification_property {
    notify_delay_after = %[2]d
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, notifyDelayAfter))
}

func testAccJobConfig_required(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccJobConfig_requiredStreaming(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn

  command {
    name            = "gluestreaming"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccJobConfig_jobRunQueuingEnabled(rName string, jobRunQueuingEnabled bool) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity            = 10
  name                    = %[1]q
  role_arn                = aws_iam_role.test.arn
  job_run_queuing_enabled = %[2]t

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, jobRunQueuingEnabled))
}

func testAccJobConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1))
}

func testAccJobConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name              = %[1]q
  number_of_workers = 1
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"

  command {
    script_location = "testscriptlocation"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccJobConfig_timeout(rName string, timeout int) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  timeout      = %[2]d

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, timeout))
}

func testAccJobConfig_streamingTimeout(rName, timeout string) string {
	if timeout == "" {
		timeout = "null"
	}

	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  timeout      = %[2]s

  command {
    name            = "gluestreaming"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, timeout))
}

func testAccJobConfig_securityConfiguration(rName string, securityConfiguration string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity           = 10
  name                   = %[1]q
  role_arn               = aws_iam_role.test.arn
  security_configuration = %[2]q

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, securityConfiguration))
}

func testAccJobConfig_workerType(rName string, workerType string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name              = %[1]q
  role_arn          = aws_iam_role.test.arn
  worker_type       = %[2]q
  number_of_workers = 10

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, workerType))
}

func testAccJobConfig_pythonShell(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  max_capacity = 0.0625

  command {
    name            = "pythonshell"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccJobConfig_pythonShellVersion(rName string, pythonVersion string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  max_capacity = 0.0625

  command {
    name            = "pythonshell"
    script_location = "testscriptlocation"
    python_version  = %[2]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, pythonVersion))
}

func testAccJobConfig_rayJob(rName string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  glue_version      = "4.0"
  name              = %[1]q
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Z.2X"
  number_of_workers = 10

  command {
    name            = "glueray"
    python_version  = "3.9"
    runtime         = "Ray2.4"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccJobConfig_rayJobWithDescription(rName, description string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  glue_version      = "4.0"
  name              = %[1]q
  description       = %[2]q
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Z.2X"
  number_of_workers = 10

  command {
    name            = "glueray"
    python_version  = "3.9"
    runtime         = "Ray2.4"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, description))
}

func testAccJobConfig_maxCapacity(rName string, maxCapacity float64) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  max_capacity = %[2]g

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, maxCapacity))
}

func testAccJobConfig_sourceControlDetails(rName, repo string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  source_control_details {
    provider       = "GITHUB"
    repository     = %[2]q
    branch         = "test-branch"
    owner          = "test-owner"
    last_commit_id = "test-commit-id"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, repo))
}

func testAccJobConfig_jobMode(rName, jobMode string) string {
	return acctest.ConfigCompose(testAccJobConfig_base(rName), fmt.Sprintf(`
resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = %[1]q
  role_arn     = aws_iam_role.test.arn
  job_mode     = %[2]q

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, jobMode))
}
