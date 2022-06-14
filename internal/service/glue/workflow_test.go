package glue_test

import (
	"fmt"
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

func TestAccGlueWorkflow_basic(t *testing.T) {
	var workflow glue.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWorkflow(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "glue", fmt.Sprintf("workflow/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccGlueWorkflow_maxConcurrentRuns(t *testing.T) {
	var workflow glue.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWorkflow(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "2"),
				),
			},
			{
				Config: testAccWorkflowConfig_maxConcurrentRuns(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "max_concurrent_runs", "1"),
				),
			},
		},
	})
}

func TestAccGlueWorkflow_defaultRunProperties(t *testing.T) {
	var workflow glue.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWorkflow(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_defaultRunProperties(rName, "firstPropValue", "secondPropValue"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.--run-prop1", "firstPropValue"),
					resource.TestCheckResourceAttr(resourceName, "default_run_properties.--run-prop2", "secondPropValue"),
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

func TestAccGlueWorkflow_description(t *testing.T) {
	var workflow glue.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWorkflow(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_description(rName, "First Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "description", "First Description"),
				),
			},
			{
				Config: testAccWorkflowConfig_description(rName, "Second Description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
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

func TestAccGlueWorkflow_tags(t *testing.T) {
	var workflow glue.Workflow
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWorkflow(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
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
				Config: testAccWorkflowConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkflowConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccGlueWorkflow_disappears(t *testing.T) {
	var workflow glue.Workflow

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_workflow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckWorkflow(t) },
		ErrorCheck:        acctest.ErrorCheck(t, glue.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWorkflowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(resourceName, &workflow),
					acctest.CheckResourceDisappears(acctest.Provider, tfglue.ResourceWorkflow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckWorkflow(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

	_, err := conn.ListWorkflows(&glue.ListWorkflowsInput{})

	// Some endpoints that do not support Glue Workflows return InternalFailure
	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckWorkflowExists(resourceName string, workflow *glue.Workflow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Workflow ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetWorkflow(&glue.GetWorkflowInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.Workflow == nil {
			return fmt.Errorf("Glue Workflow (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Workflow.Name) == rs.Primary.ID {
			*workflow = *output.Workflow
			return nil
		}

		return fmt.Errorf("Glue Workflow (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckWorkflowDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_workflow" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetWorkflow(&glue.GetWorkflowInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
				return nil
			}

		}

		workflow := output.Workflow
		if workflow != nil && aws.StringValue(workflow.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Workflow %s still exists", rs.Primary.ID)
		}

		return err
	}

	return nil
}

func testAccWorkflowConfig_defaultRunProperties(rName, firstPropValue, secondPropValue string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = "%s"

  default_run_properties = {
    "--run-prop1" = "%s"
    "--run-prop2" = "%s"
  }
}
`, rName, firstPropValue, secondPropValue)
}

func testAccWorkflowConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  description = "%s"
  name        = "%s"
}
`, description, rName)
}

func testAccWorkflowConfig_required(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = "%s"
}
`, rName)
}

func testAccWorkflowConfig_maxConcurrentRuns(rName string, runs int) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name                = %[1]q
  max_concurrent_runs = %[2]d
}
`, rName, runs)
}

func testAccWorkflowConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccWorkflowConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_workflow" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
