package datasync_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDataSyncTask_basic(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSyncDestinationLocationResourceName := "aws_datasync_location_s3.destination"
	dataSyncSourceLocationResourceName := "aws_datasync_location_nfs.source"
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`task/task-.+`)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_log_group_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_location_arn", dataSyncDestinationLocationResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "BEST_EFFORT"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "-1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "OFF"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "ALWAYS"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "PRESERVE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "CHANGED"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "INT_VALUE"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "POINT_IN_TIME_CONSISTENT"),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "source_location_arn", dataSyncSourceLocationResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "0"),
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
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatasync.ResourceTask(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncTask_schedule(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskScheduleConfig(rName, "cron(0 12 ? * SUN,WED *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(0 12 ? * SUN,WED *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskScheduleConfig(rName, "cron(0 12 ? * SUN,MON *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "schedule.0.schedule_expression", "cron(0 12 ? * SUN,MON *)"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_cloudWatchLogGroupARN(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskCloudWatchLogGroupARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", "aws_cloudwatch_log_group.test1", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskCloudWatchLogGroupARN2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", "aws_cloudwatch_log_group.test2", "arn")),
			},
		},
	})
}

func TestAccDataSyncTask_excludes(t *testing.T) {
	var task1 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskExcludesConfig(rName, "/folder1|/folder2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "excludes.#", "1"),
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
				Config: testAccTaskExcludesConfig(rName, "/test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "excludes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.filter_type", "SIMPLE_PATTERN"),
					resource.TestCheckResourceAttr(resourceName, "excludes.0.value", "/test"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_atimeMtime(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsAtimeMtimeConfig(rName, "NONE", "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
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
				Config: testAccTaskDefaultSyncOptionsAtimeMtimeConfig(rName, "BEST_EFFORT", "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.atime", "BEST_EFFORT"),
					resource.TestCheckResourceAttr(resourceName, "options.0.mtime", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_bytesPerSecond(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsBytesPerSecondConfig(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsBytesPerSecondConfig(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.bytes_per_second", "2"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_gid(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsGidConfig(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsGidConfig(rName, "INT_VALUE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.gid", "INT_VALUE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_logLevel(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsLogLevelConfig(rName, "OFF"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "OFF"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsLogLevelConfig(rName, "BASIC"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.log_level", "BASIC"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_overwriteMode(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsOverwriteModeConfig(rName, "NEVER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "NEVER"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsOverwriteModeConfig(rName, "ALWAYS"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.overwrite_mode", "ALWAYS"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_posixPermissions(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsPOSIXPermissionsConfig(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsPOSIXPermissionsConfig(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.posix_permissions", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_preserveDeletedFiles(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsPreserveDeletedFilesConfig(rName, "REMOVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "REMOVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsPreserveDeletedFilesConfig(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_deleted_files", "PRESERVE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_preserveDevices(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsPreserveDevicesConfig(rName, "PRESERVE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "PRESERVE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsPreserveDevicesConfig(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.preserve_devices", "NONE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_taskQueueing(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsTaskQueueingConfig(rName, "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "ENABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsTaskQueueingConfig(rName, "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.task_queueing", "DISABLED"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_transferMode(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsTransferModeConfig(rName, "CHANGED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "CHANGED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsTransferModeConfig(rName, "ALL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.transfer_mode", "ALL"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_uid(t *testing.T) {
	var task1, task2 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsUIDConfig(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsUIDConfig(rName, "INT_VALUE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.uid", "INT_VALUE"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_DefaultSyncOptions_verifyMode(t *testing.T) {
	var task1, task2, task3 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefaultSyncOptionsVerifyModeConfig(rName, "NONE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "NONE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTaskDefaultSyncOptionsVerifyModeConfig(rName, "POINT_IN_TIME_CONSISTENT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "POINT_IN_TIME_CONSISTENT"),
				),
			},
			{
				Config: testAccTaskDefaultSyncOptionsVerifyModeConfig(rName, "ONLY_FILES_TRANSFERRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task3),
					testAccCheckTaskNotRecreated(&task2, &task3),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.verify_mode", "ONLY_FILES_TRANSFERRED"),
				),
			},
		},
	})
}

func TestAccDataSyncTask_tags(t *testing.T) {
	acctest.Skip(t, "Tagging on creation is inconsistent")
	var task1, task2, task3 datasync.DescribeTaskOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_datasync_task.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, datasync.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task1),
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
				Config: testAccTaskTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task2),
					testAccCheckTaskNotRecreated(&task1, &task2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTaskTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskExists(resourceName, &task3),
					testAccCheckTaskNotRecreated(&task2, &task3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckTaskDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_task" {
			continue
		}

		_, err := tfdatasync.FindTaskByARN(conn, rs.Primary.ID)

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

func testAccCheckTaskExists(resourceName string, task *datasync.DescribeTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

		output, err := tfdatasync.FindTaskByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if aws.StringValue(output.Status) != datasync.TaskStatusAvailable && aws.StringValue(output.Status) != datasync.TaskStatusRunning {
			return fmt.Errorf("Task %q not available or running: last status (%s), error code (%s), error detail: %s",
				rs.Primary.ID, aws.StringValue(output.Status), aws.StringValue(output.ErrorCode), aws.StringValue(output.ErrorDetail))
		}

		*task = *output

		return nil
	}
}

func testAccCheckTaskNotRecreated(i, j *datasync.DescribeTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TaskArn) != aws.StringValue(j.TaskArn) {
			return errors.New("DataSync Task was recreated")
		}

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	input := &datasync.ListTasksInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListTasks(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTaskDestinationLocationS3BaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "destination" {
  name = "%[1]s-destination"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "datasync.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
POLICY
}

resource "aws_s3_bucket" "destination" {
  bucket        = "%[1]s-destination"
  force_destroy = true
}

resource "aws_iam_role_policy" "destination" {
  role   = aws_iam_role.destination.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": [
      "s3:*"
    ],
    "Effect": "Allow",
    "Resource": [
      "${aws_s3_bucket.destination.arn}",
      "${aws_s3_bucket.destination.arn}/*"
    ]
  }]
}
POLICY
}

resource "aws_datasync_location_s3" "destination" {
  s3_bucket_arn = aws_s3_bucket.destination.arn
  subdirectory  = "/destination"

  s3_config {
    bucket_access_role_arn = aws_iam_role.destination.arn
  }

  depends_on = [aws_iam_role_policy.destination]
}
`, rName)
}

func testAccTaskSourceLocationNFSBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.source.availability_zone", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_vpc" "source" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "source" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "source" {
  vpc_id = aws_vpc.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "source" {
  default_route_table_id = aws_vpc.source.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.source.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "source" {
  name   = %[1]q
  vpc_id = aws_vpc.source.id

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

# EFS as our NFS server
resource "aws_efs_file_system" "source" {}

resource "aws_efs_mount_target" "source" {
  file_system_id  = aws_efs_file_system.source.id
  security_groups = [aws_security_group.source.id]
  subnet_id       = aws_subnet.source.id
}

resource "aws_instance" "source" {
  depends_on = [
    aws_default_route_table.source,
  ]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.source.id]
  subnet_id                   = aws_subnet.source.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_agent" "source" {
  ip_address = aws_instance.source.public_ip
  name       = %[1]q
}

resource "aws_datasync_location_nfs" "source" {
  server_hostname = aws_efs_mount_target.source.dns_name
  subdirectory    = "/"

  on_prem_config {
    agent_arns = [aws_datasync_agent.source.arn]
  }
}
`, rName))
}

func testAccTaskConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn
}
`, rName))
}

func testAccTaskScheduleConfig(rName, cron string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  schedule {
    schedule_expression = %[2]q
  }
}
`, rName, cron))
}

func testAccTaskCloudWatchLogGroupARNConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test1.arn
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn
}
`, rName))
}

func testAccTaskCloudWatchLogGroupARN2Config(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test2.arn
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn
}
`, rName))
}

func testAccTaskExcludesConfig(rName, value string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  excludes {
    filter_type = "SIMPLE_PATTERN"
    value       = %[2]q
  }
}
`, rName, value))
}

func testAccTaskDefaultSyncOptionsAtimeMtimeConfig(rName, atime, mtime string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    atime = %[2]q
    mtime = %[3]q
  }
}
`, rName, atime, mtime))
}

func testAccTaskDefaultSyncOptionsBytesPerSecondConfig(rName string, bytesPerSecond int) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    bytes_per_second = %[2]d
  }
}
`, rName, bytesPerSecond))
}

func testAccTaskDefaultSyncOptionsGidConfig(rName, gid string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    gid = %[2]q
  }
}
`, rName, gid))
}

func testAccTaskDefaultSyncOptionsLogLevelConfig(rName, logLevel string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_datasync_task" "test" {
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    log_level = %[2]q
  }
}
`, rName, logLevel))
}

func testAccTaskDefaultSyncOptionsOverwriteModeConfig(rName, overwriteMode string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    overwrite_mode = %[2]q
  }
}
`, rName, overwriteMode))
}

func testAccTaskDefaultSyncOptionsPOSIXPermissionsConfig(rName, posixPermissions string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    posix_permissions = %[2]q
  }
}
`, rName, posixPermissions))
}

func testAccTaskDefaultSyncOptionsPreserveDeletedFilesConfig(rName, preserveDeletedFiles string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    preserve_deleted_files = %[2]q
  }
}
`, rName, preserveDeletedFiles))
}

func testAccTaskDefaultSyncOptionsPreserveDevicesConfig(rName, preserveDevices string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    preserve_devices = %[2]q
  }
}
`, rName, preserveDevices))
}

func testAccTaskDefaultSyncOptionsTaskQueueingConfig(rName, taskQueueing string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    task_queueing = %[2]q
  }
}
`, rName, taskQueueing))
}

func testAccTaskDefaultSyncOptionsTransferModeConfig(rName, transferMode string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    transfer_mode = %[2]q
  }
}
`, rName, transferMode))
}

func testAccTaskDefaultSyncOptionsUIDConfig(rName, uid string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    uid = %[2]q
  }
}
`, rName, uid))
}

func testAccTaskDefaultSyncOptionsVerifyModeConfig(rName, verifyMode string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  options {
    verify_mode = %[2]q
  }
}
`, rName, verifyMode))
}

func testAccTaskTags1Config(rName, key1, value1 string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1))
}

func testAccTaskTags2Config(rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccTaskDestinationLocationS3BaseConfig(rName),
		testAccTaskSourceLocationNFSBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_datasync_task" "test" {
  destination_location_arn = aws_datasync_location_s3.destination.arn
  name                     = %[1]q
  source_location_arn      = aws_datasync_location_nfs.source.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2))
}
