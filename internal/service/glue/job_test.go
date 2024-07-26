// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueJob_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("job/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "execution_class", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, "2880"),
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

func TestAccGlueJob_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceJob(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlueJob_basicStreaming(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_requiredStreaming(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "glue", fmt.Sprintf("job/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "gluestreaming"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, roleResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, acctest.Ct0),
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

func TestAccGlueJob_command(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_command(rName, "testscriptlocation1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation1"),
				),
			},
			{
				Config: testAccJobConfig_command(rName, "testscriptlocation2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_defaultArguments(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccJobConfig_defaultArguments(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", acctest.Ct2),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_nonOverridableArguments(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccJobConfig_nonOverridableArguments(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", acctest.Ct2),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "First Description"),
				),
			},
			{
				Config: testAccJobConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_versionMaxCapacity(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
				),
			},
			{
				Config: testAccJobConfig_versionMaxCapacity(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
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
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "2.0"),
				),
			},
		},
	})
}

func TestAccGlueJob_executionClass(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_executionClass(rName, "FLEX"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_class", "FLEX"),
				),
			},
			{
				Config: testAccJobConfig_executionClass(rName, "STANDARD"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_executionProperty(rName, 0),
				ExpectError: regexache.MustCompile(`expected execution_property.0.max_concurrent_runs to be at least`),
			},
			{
				Config: testAccJobConfig_executionProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_property.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "execution_property.0.max_concurrent_runs", acctest.Ct1),
				),
			},
			{
				Config: testAccJobConfig_executionProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_property.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "execution_property.0.max_concurrent_runs", acctest.Ct2),
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

func TestAccGlueJob_maintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"
	maintenanceWindow := "Sun:23"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_maintenanceWindow(rName, maintenanceWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_maxRetries(rName, 11),
				ExpectError: regexache.MustCompile(`expected max_retries to be in the range`),
			},
			{
				Config: testAccJobConfig_maxRetries(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_retries", acctest.Ct0),
				),
			},
			{
				Config: testAccJobConfig_maxRetries(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_retries", acctest.Ct10),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_notificationProperty(rName, 0),
				ExpectError: regexache.MustCompile(`expected notification_property.0.notify_delay_after to be at least`),
			},
			{
				Config: testAccJobConfig_notificationProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "notification_property.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "notification_property.0.notify_delay_after", acctest.Ct1),
				),
			},
			{
				Config: testAccJobConfig_notificationProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "notification_property.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "notification_property.0.notify_delay_after", acctest.Ct2),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
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
				Config: testAccJobConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccJobConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlueJob_streamingTimeout(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_timeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, acctest.Ct1),
				),
			},
			{
				Config: testAccJobConfig_timeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, acctest.Ct2),
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

func TestAccGlueJob_timeout(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_timeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, acctest.Ct1),
				),
			},
			{
				Config: testAccJobConfig_timeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrTimeout, acctest.Ct2),
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

func TestAccGlueJob_security(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_securityConfiguration(rName, "default_encryption"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "default_encryption"),
				),
			},
			{
				Config: testAccJobConfig_securityConfiguration(rName, "custom_encryption2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_workerType(rName, "Standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "Standard"),
				),
			},
			{
				Config: testAccJobConfig_workerType(rName, "G.1X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.1X"),
				),
			},
			{
				Config: testAccJobConfig_workerType(rName, "G.2X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.2X"),
				),
			},
			{
				Config: testAccJobConfig_workerType(rName, "G.4X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.4X"),
				),
			},
			{
				Config: testAccJobConfig_workerType(rName, "G.8X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.8X"),
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

func TestAccGlueJob_pythonShell(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_pythonShell(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, "0.0625"),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
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
				Config: testAccJobConfig_pythonShellVersion(rName, acctest.Ct2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_pythonShellVersion(rName, acctest.Ct3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
			{
				Config: testAccJobConfig_pythonShellVersion(rName, "3.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_rayJob(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "glueray"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", "3.9"),
					resource.TestCheckResourceAttr(resourceName, "command.0.runtime", "Ray2.4"),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "Z.2X"),
				),
			},
		},
	})
}

func TestAccGlueJob_maxCapacity(t *testing.T) {
	ctx := acctest.Context(t)
	var job awstypes.Job
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_maxCapacity(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, names.AttrMaxCapacity, acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "command.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "glueetl"),
				),
			},
			{
				Config: testAccJobConfig_maxCapacity(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(ctx, resourceName, &job),
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

func testAccCheckJobExists(ctx context.Context, n string, v *awstypes.Job) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Job ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		output, err := tfglue.FindJobByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckJobDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_job" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

			_, err := tfglue.FindJobByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
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
