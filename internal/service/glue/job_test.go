package glue_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
)

func TestAccGlueJob_basic(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("job/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "2880"),
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

func TestAccGlueJob_basicStreaming(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"
	roleResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_RequiredStreaming(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("job/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "gluestreaming"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout", "0"),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_Command(rName, "testscriptlocation1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation1"),
				),
			},
			{
				Config: testAccJobConfig_Command(rName, "testscriptlocation2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_DefaultArguments(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "default_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccJobConfig_DefaultArguments(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobnonOverridableArgumentsConfig(rName, "job-bookmark-disable", "python"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-bookmark-option", "job-bookmark-disable"),
					resource.TestCheckResourceAttr(resourceName, "non_overridable_arguments.--job-language", "python"),
				),
			},
			{
				Config: testAccJobnonOverridableArgumentsConfig(rName, "job-bookmark-enable", "scala"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_Description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccJobConfig_Description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "description", "Second Description"),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_Version_maxCapacity(rName, "0.9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "0.9"),
				),
			},
			{
				Config: testAccJobConfig_Version_maxCapacity(rName, "1.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "1.0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobConfig_Version_numberOfWorkers(rName, "2.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "glue_version", "2.0"),
				),
			},
		},
	})
}

func TestAccGlueJob_executionProperty(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_ExecutionProperty(rName, 0),
				ExpectError: regexp.MustCompile(`expected execution_property.0.max_concurrent_runs to be at least`),
			},
			{
				Config: testAccJobConfig_ExecutionProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "execution_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "execution_property.0.max_concurrent_runs", "1"),
				),
			},
			{
				Config: testAccJobConfig_ExecutionProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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

func TestAccGlueJob_maxRetries(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_MaxRetries(rName, 11),
				ExpectError: regexp.MustCompile(`expected max_retries to be in the range`),
			},
			{
				Config: testAccJobConfig_MaxRetries(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_retries", "0"),
				),
			},
			{
				Config: testAccJobConfig_MaxRetries(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccJobConfig_NotificationProperty(rName, 0),
				ExpectError: regexp.MustCompile(`expected notification_property.0.notify_delay_after to be at least`),
			},
			{
				Config: testAccJobConfig_NotificationProperty(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "notification_property.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "notification_property.0.notify_delay_after", "1"),
				),
			},
			{
				Config: testAccJobConfig_NotificationProperty(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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
	var job1, job2, job3 glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job1),
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
				Config: testAccJobTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccJobTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueJob_streamingTimeout(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_Timeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "timeout", "1"),
				),
			},
			{
				Config: testAccJobConfig_Timeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "timeout", "2"),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_Timeout(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "timeout", "1"),
				),
			},
			{
				Config: testAccJobConfig_Timeout(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "timeout", "2"),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_SecurityConfiguration(rName, "default_encryption"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "security_configuration", "default_encryption"),
				),
			},
			{
				Config: testAccJobConfig_SecurityConfiguration(rName, "custom_encryption2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_WorkerType(rName, "Standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "Standard"),
				),
			},
			{
				Config: testAccJobConfig_WorkerType(rName, "G.1X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.1X"),
				),
			},
			{
				Config: testAccJobConfig_WorkerType(rName, "G.2X"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "worker_type", "G.2X"),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_PythonShell(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "0.0625"),
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
				Config: testAccJobConfig_PythonShellWithVersion(rName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
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
				Config: testAccJobConfig_PythonShellWithVersion(rName, "3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.python_version", "3"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "pythonshell"),
				),
			},
		},
	})
}

func TestAccGlueJob_maxCapacity(t *testing.T) {
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_MaxCapacity(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "10"),
					resource.TestCheckResourceAttr(resourceName, "command.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "command.0.script_location", "testscriptlocation"),
					resource.TestCheckResourceAttr(resourceName, "command.0.name", "glueetl"),
				),
			},
			{
				Config: testAccJobConfig_MaxCapacity(rName, 15),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					resource.TestCheckResourceAttr(resourceName, "max_capacity", "15"),
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
	var job glue.Job

	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))
	resourceName := "aws_glue_job.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckJobDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccJobConfig_Required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobExists(resourceName, &job),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceJob(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckJobExists(resourceName string, job *glue.Job) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Job ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetJob(&glue.GetJobInput{
			JobName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.Job == nil {
			return fmt.Errorf("Glue Job (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Job.Name) == rs.Primary.ID {
			*job = *output.Job
			return nil
		}

		return fmt.Errorf("Glue Job (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckJobDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_job" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetJob(&glue.GetJobInput{
			JobName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				return nil
			}

		}

		job := output.Job
		if job != nil && aws.StringValue(job.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Job %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccJobConfig_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy" "AWSGlueServiceRole" {
  arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSGlueServiceRole"
}

resource "aws_iam_role" "test" {
  name = "%s"

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

func testAccJobConfig_Command(rName, scriptLocation string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, scriptLocation)
}

func testAccJobConfig_DefaultArguments(rName, jobBookmarkOption, jobLanguage string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  default_arguments = {
    "--job-bookmark-option" = "%s"
    "--job-language"        = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, jobBookmarkOption, jobLanguage)
}

func testAccJobnonOverridableArgumentsConfig(rName, jobBookmarkOption, jobLanguage string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  non_overridable_arguments = {
    "--job-bookmark-option" = "%s"
    "--job-language"        = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, jobBookmarkOption, jobLanguage)
}

func testAccJobConfig_Description(rName, description string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  description  = "%s"
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), description, rName)
}

func testAccJobConfig_Version_maxCapacity(rName, glueVersion string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  glue_version = "%s"
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), glueVersion, rName)
}

func testAccJobConfig_Version_numberOfWorkers(rName, glueVersion string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  glue_version      = "%s"
  name              = "%s"
  number_of_workers = 2
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), glueVersion, rName)
}

func testAccJobConfig_ExecutionProperty(rName string, maxConcurrentRuns int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  execution_property {
    max_concurrent_runs = %d
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, maxConcurrentRuns)
}

func testAccJobConfig_MaxRetries(rName string, maxRetries int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  max_retries  = %d
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), maxRetries, rName)
}

func testAccJobConfig_NotificationProperty(rName string, notifyDelayAfter int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  notification_property {
    notify_delay_after = %d
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, notifyDelayAfter)
}

func testAccJobConfig_Required(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName)
}

func testAccJobConfig_RequiredStreaming(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn

  command {
    name            = "gluestreaming"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName)
}

func testAccJobTags1Config(rName, tagKey1, tagValue1 string) string {
	return testAccJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name              = %[1]q
  number_of_workers = 2
  role_arn          = aws_iam_role.test.arn
  worker_type       = "Standard"

  command {
    script_location = "testscriptlocation"
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, tagKey1, tagValue1)
}

func testAccJobTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccJobConfig_Base(rName) + fmt.Sprintf(`
resource "aws_glue_job" "test" {
  name              = %[1]q
  number_of_workers = 2
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
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccJobConfig_Timeout(rName string, timeout int) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity = 10
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  timeout      = %d

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, timeout)
}

func testAccJobConfig_SecurityConfiguration(rName string, securityConfiguration string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  max_capacity           = 10
  name                   = "%s"
  role_arn               = aws_iam_role.test.arn
  security_configuration = "%s"

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, securityConfiguration)
}

func testAccJobConfig_WorkerType(rName string, workerType string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name              = "%s"
  role_arn          = aws_iam_role.test.arn
  worker_type       = "%s"
  number_of_workers = 10

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, workerType)
}

func testAccJobConfig_PythonShell(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  max_capacity = 0.0625

  command {
    name            = "pythonshell"
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName)
}

func testAccJobConfig_PythonShellWithVersion(rName string, pythonVersion string) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  max_capacity = 0.0625

  command {
    name            = "pythonshell"
    script_location = "testscriptlocation"
    python_version  = "%s"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, pythonVersion)
}

func testAccJobConfig_MaxCapacity(rName string, maxCapacity float64) string {
	return fmt.Sprintf(`
%s

resource "aws_glue_job" "test" {
  name         = "%s"
  role_arn     = aws_iam_role.test.arn
  max_capacity = %g

  command {
    script_location = "testscriptlocation"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, testAccJobConfig_Base(rName), rName, maxCapacity)
}
