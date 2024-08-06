// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataSyncTask_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSyncDestinationLocationResourceName := "aws_datasync_location_s3.test"
	dataSyncSourceLocationResourceName := "aws_datasync_location_nfs.test"
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "datasync", regexache.MustCompile(`task/task-.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrCloudWatchLogGroupARN, ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_location_arn", dataSyncDestinationLocationResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "excludes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "includes.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "BEST_EFFORT"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "-1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.object_tags", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "ALWAYS"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.security_descriptor_copy_flags", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "POINT_IN_TIME_CONSISTENT"),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, "source_location_arn", dataSyncSourceLocationResourceName, names.AttrARN),
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

func TestAccDataSyncTask_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceTask(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncTask_schedule(t *testing.T) {
	ctx := acctest.Context(t)
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_schedule(rName, "cron(0 12 ? * SUN,WED *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(0 12 ? * SUN,WED *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_schedule(rName, "cron(0 12 ? * SUN,MON *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(0 12 ? * SUN,MON *)"),
				),
			},
			{
				Config: testAccTaskConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccDataSyncTask_cloudWatchLogGroupARN(t *testing.T) {
	ctx := acctest.Context(t)
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_cloudWatchLogGroupARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCloudWatchLogGroupARN, "aws_cloudwatch_log_group.test1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_cloudWatchLogGroupARN2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCloudWatchLogGroupARN, "aws_cloudwatch_log_group.test2", names.AttrARN)),
			},
		},
	})
}

func TestAccDataSyncTask_excludes(t *testing.T) {
	ctx := acctest.Context(t)
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_excludes(rName, "/folder1|/folder2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "excludes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.filter_type", "SIMPLE_PATTERN"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.value", "/folder1|/folder2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_excludes(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "excludes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.filter_type", "SIMPLE_PATTERN"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.value", "/test"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_includes(t *testing.T) {
	ctx := acctest.Context(t)
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_includes(rName, "/folder1|/folder2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "includes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "includes.0.filter_type", "SIMPLE_PATTERN"),
					resource.TestCheckResourceAttr(resourceName, "includes.0.value", "/folder1|/folder2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_includes(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "includes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "includes.0.filter_type", "SIMPLE_PATTERN"),
					resource.TestCheckResourceAttr(resourceName, "includes.0.value", "/test"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_atimeMtime(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsAtimeMtime(rName, "NONE", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsAtimeMtime(rName, "BEST_EFFORT", "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "BEST_EFFORT"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_bytesPerSecond(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsBytesPerSecond(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsBytesPerSecond(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_gid(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsGID(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsGID(rName, "INT_VALUE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "INT_VALUE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_logLevel(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsLogLevel(rName, "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsLogLevel(rName, "BASIC"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "BASIC"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_objectTags(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsObjectTags(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.object_tags", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsObjectTags(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.object_tags", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_overwriteMode(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsOverwriteMode(rName, "NEVER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "NEVER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsOverwriteMode(rName, "ALWAYS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "ALWAYS"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_posixPermissions(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsPOSIXPermissions(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsPOSIXPermissions(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_preserveDeletedFiles(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsPreserveDeletedFiles(rName, "REMOVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "REMOVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsPreserveDeletedFiles(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_preserveDevices(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsPreserveDevices(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "PRESERVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsPreserveDevices(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "NONE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_securityDescriptorCopyFlags(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsSecurityDescriptorCopyFlags(rName, domainName, "OWNER_DACL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.security_descriptor_copy_flags", "OWNER_DACL"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsSecurityDescriptorCopyFlags(rName, domainName, "OWNER_DACL_SACL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.security_descriptor_copy_flags", "OWNER_DACL_SACL"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "NONE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_taskQueueing(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsQueueing(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsQueueing(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "DISABLED"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_transferMode(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsTransferMode(rName, "CHANGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "CHANGED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsTransferMode(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "ALL"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_uid(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsUID(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsUID(rName, "INT_VALUE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "INT_VALUE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_verifyMode(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2, task3 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_defaultSyncOptionsVerifyMode(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsVerifyMode(rName, "POINT_IN_TIME_CONSISTENT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "POINT_IN_TIME_CONSISTENT"),
				),
			},
			{
				Config: testAccTaskConfig_defaultSyncOptionsVerifyMode(rName, "ONLY_FILES_TRANSFERRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task3),
					testAccCheckTaskNotRecreated(&task2, &task3),
					resource.TestCheckResourceAttr(resourceName, "options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "ONLY_FILES_TRANSFERRED"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_taskReportConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_taskReportConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.output_type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.report_level", "SUCCESSES_AND_ERRORS"),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.s3_destination.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.s3_object_versioning", "INCLUDE"),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.s3_destination.0.subdirectory", "test/"),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.report_overrides.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.report_overrides.0.deleted_override", "ERRORS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.report_overrides.0.skipped_override", "ERRORS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.report_overrides.0.transferred_override", "ERRORS_ONLY"),
					resource.TestCheckResourceAttr(resourceName, "task_report_config.0.report_overrides.0.verified_override", "ERRORS_ONLY"),
					resource.TestCheckResourceAttrPair(resourceName, "task_report_config.0.s3_destination.0.bucket_access_role_arn", "aws_iam_role.report_test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "task_report_config.0.s3_destination.0.s3_bucket_arn", "aws_s3_bucket.report_test", names.AttrARN),
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

func TestAccDataSyncTask_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var task1, task2, task3 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task1),
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
				Config: testAccTaskConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTaskConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(ctx, resourceName, &task3),
					testAccCheckTaskNotRecreated(&task2, &task3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccCheckTaskDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_task" {
				continue
			}

			_, err := tfdatasync.FindTaskByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Task %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTaskExists(ctx context.Context, resourceName string, task *datasync.DescribeTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

		output, err := tfdatasync.FindTaskByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if output.Status != awstypes.TaskStatusAvailable && output.Status != awstypes.TaskStatusRunning {
			return fmt.Errorf("Task %q not available or running: last status (%s), error code (%s), error detail: %s",
				rs.Primary.ID, string(output.Status), aws.ToString(output.ErrorCode), aws.ToString(output.ErrorDetail))
		}

		*task = *output

		return nil
	}
}

func testAccCheckTaskNotRecreated(i, j *datasync.DescribeTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.TaskArn) != aws.ToString(j.TaskArn) {
			return errors.New("DataSync Task was recreated")
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.ListTasksInput{
		MaxResults: aws.Int32(1),
	}

	_, err := conn.ListTasks(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTaskConfig_baseLocationS3(rName string) string {
	return acctest.ConfigCompose(testAccLocationS3Config_base(rName), `
resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = aws_s3_bucket.test.arn
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = aws_iam_role.test.arn
  }

  depends_on = [aws_iam_role_policy.test]
}
`)
}

func testAccTaskConfig_baseLocationNFS(rName string) string {
	return acctest.ConfigCompose(testAccLocationNFSConfig_base(rName), `
# EFS as our NFS server
resource "aws_efs_file_system" "test" {}

resource "aws_efs_mount_target" "test" {
  file_system_id  = aws_efs_file_system.test.id
  security_groups = [aws_security_group.test.id]
  subnet_id       = aws_subnet.test[0].id
}

resource "aws_datasync_location_nfs" "test" {
  server_hostname = aws_efs_mount_target.test.dns_name
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.test.arn]
  }
}
`)
}

func testAccTaskConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn
}
`, rName))
}

func testAccTaskConfig_schedule(rName, cron string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  schedule {
    schedule_expression = %[2]q
  }
}
`, rName, cron))
}

func testAccTaskConfig_cloudWatchLogGroupARN(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test1.arn
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn
}
`, rName))
}

func testAccTaskConfig_cloudWatchLogGroupARN2(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test2.arn
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn
}
`, rName))
}

func testAccTaskConfig_excludes(rName, value string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  excludes {
    filter_type = "SIMPLE_PATTERN"
    value       = %[2]q
  }
}
`, rName, value))
}

func testAccTaskConfig_includes(rName, value string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  includes {
    filter_type = "SIMPLE_PATTERN"
    value       = %[2]q
  }
}
`, rName, value))
}

func testAccTaskConfig_defaultSyncOptionsAtimeMtime(rName, atime, mtime string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    atime = %[2]q
    mtime = %[3]q
  }
}
`, rName, atime, mtime))
}

func testAccTaskConfig_defaultSyncOptionsBytesPerSecond(rName string, bytesPerSecond int) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    bytes_per_second = %[2]d
  }
}
`, rName, bytesPerSecond))
}

func testAccTaskConfig_defaultSyncOptionsGID(rName, gid string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    gid = %[2]q
  }
}
`, rName, gid))
}

func testAccTaskConfig_defaultSyncOptionsLogLevel(rName, logLevel string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    log_level = %[2]q
  }
}
`, rName, logLevel))
}

func testAccTaskConfig_defaultSyncOptionsObjectTags(rName, objectTags string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    object_tags = %[2]q
  }
}
`, rName, objectTags))
}

func testAccTaskConfig_defaultSyncOptionsOverwriteMode(rName, overwriteMode string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    overwrite_mode = %[2]q
  }
}
`, rName, overwriteMode))
}

func testAccTaskConfig_defaultSyncOptionsPOSIXPermissions(rName, posixPermissions string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    posix_permissions = %[2]q
  }
}
`, rName, posixPermissions))
}

func testAccTaskConfig_defaultSyncOptionsPreserveDeletedFiles(rName, preserveDeletedFiles string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    preserve_deleted_files = %[2]q
  }
}
`, rName, preserveDeletedFiles))
}

func testAccTaskConfig_defaultSyncOptionsPreserveDevices(rName, preserveDevices string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    preserve_devices = %[2]q
  }
}
`, rName, preserveDevices))
}

// https://docs.aws.amazon.com/datasync/latest/userguide/API_Options.html#DataSync-Type-Options-SecurityDescriptorCopyFlags:
// This value is only used for transfers between SMB and Amazon FSx for Windows File Server locations, or between two Amazon FSx for Windows File Server locations.
func testAccTaskConfig_defaultSyncOptionsSecurityDescriptorCopyFlags(rName, domain, securityDescriptorCopyFlags string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "test" {
  name = "/aws/service/datasync/ami"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_default_route_table.test]

  ami                         = data.aws_ssm_parameter.test.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}

resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = "/test/"
  user            = "Guest"
}

resource "aws_directory_service_directory" "test" {
  edition  = "Standard"
  name     = %[2]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }
}

resource "aws_fsx_windows_file_system" "test" {
  active_directory_id = aws_directory_service_directory.test.id
  security_group_ids  = [aws_security_group.test.id]
  skip_final_backup   = true
  storage_capacity    = 32
  subnet_ids          = [aws_subnet.test[0].id]
  throughput_capacity = 8

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_location_fsx_windows_file_system" "test" {
  fsx_filesystem_arn  = aws_fsx_windows_file_system.test.arn
  user                = "SomeUser"
  password            = "SuperSecretPassw0rd"
  security_group_arns = [aws_security_group.test.arn]
}

resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_fsx_windows_file_system.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_smb.test.arn

  options {
    gid                            = "NONE"
    posix_permissions              = "NONE"
    security_descriptor_copy_flags = %[3]q
    uid                            = "NONE"
  }
}
`, rName, domain, securityDescriptorCopyFlags))
}

func testAccTaskConfig_defaultSyncOptionsQueueing(rName, taskQueueing string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    task_queueing = %[2]q
  }
}
`, rName, taskQueueing))
}

func testAccTaskConfig_defaultSyncOptionsTransferMode(rName, transferMode string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    transfer_mode = %[2]q
  }
}
`, rName, transferMode))
}

func testAccTaskConfig_defaultSyncOptionsUID(rName, uid string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    uid = %[2]q
  }
}
`, rName, uid))
}

func testAccTaskConfig_defaultSyncOptionsVerifyMode(rName, verifyMode string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  options {
    verify_mode = %[2]q
  }
}
`, rName, verifyMode))
}

func testAccTaskConfig_tags1(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccTaskConfig_tags2(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}

func testAccTaskConfig_taskReportConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskConfig_baseLocationS3(rName),
		testAccTaskConfig_baseLocationNFS(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "report_test" {
  bucket        = "%[1]s-report-test"
  force_destroy = true
}

resource "aws_iam_role" "report_test" {
  name               = "%[1]s-report-test"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "datasync.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "report_test" {
  role   = aws_iam_role.report_test.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
	"Action": [
	  "s3:*"
	],
	"Effect": "Allow",
	"Resource": [
	  "${aws_s3_bucket.report_test.arn}",
	  "${aws_s3_bucket.report_test.arn}/*"
	]
  }]
}
POLICY
}

resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.test.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.test.arn

  task_report_config {
    s3_destination {
      bucket_access_role_arn = aws_iam_role.report_test.arn
      s3_bucket_arn          = aws_s3_bucket.report_test.arn
      subdirectory           = "test/"
    }
    report_overrides {
      deleted_override     = "ERRORS_ONLY"
      skipped_override     = "ERRORS_ONLY"
      transferred_override = "ERRORS_ONLY"
      verified_override    = "ERRORS_ONLY"
    }
    s3_object_versioning = "INCLUDE"
    output_type          = "STANDARD"
    report_level         = "SUCCESSES_AND_ERRORS"
  }
}
`, rName))
}
