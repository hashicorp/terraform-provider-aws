package datapipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatapipeline "github.com/hashicorp/terraform-provider-aws/internal/service/datapipeline"
)

func TestAccDataPipelinePipelineDefinition_basic(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_pipeline_definition.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineDefinitionExists(resourceName, &pipelineOutput),
					resource.TestCheckResourceAttr(resourceName, "pipeline_object.0.id", "Default"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_object.0.name", "Default"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_object.0.field.0.key", "workerGroup"),
					resource.TestCheckResourceAttr(resourceName, "pipeline_object.0.field.0.string_value", "workerGroup"),
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

func TestAccDataPipelinePipelineDefinition_disappears(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_pipeline_definition.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDefinitionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineDefinitionExists(resourceName, &pipelineOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatapipeline.ResourcePipelineDefinition(), resourceName),
				),
			},
		},
	})
}

func TestAccDataPipelinePipelineDefinition_complete(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_pipeline_definition.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineDefinitionConfig_complete(rName, "myAWSCLICmd"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineDefinitionExists(resourceName, &pipelineOutput),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.id", "myAWSCLICmd"),
				),
			},
			{
				Config: testAccPipelineDefinitionConfig_complete(rName, "myAWSCLICmd2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineDefinitionExists(resourceName, &pipelineOutput),
					resource.TestCheckResourceAttr(resourceName, "parameter_object.0.id", "myAWSCLICmd2"),
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

func testAccCheckPipelineDefinitionExists(resourceName string, datapipelineOutput *datapipeline.GetPipelineDefinitionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn
		resp, err := conn.GetPipelineDefinitionWithContext(context.Background(), &datapipeline.GetPipelineDefinitionInput{PipelineId: aws.String(rs.Primary.ID)})
		if err != nil {
			return fmt.Errorf("problem checking for DataPipeline Pipeline Definition existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("DataPipeline Pipeline Definition %q does not exist", rs.Primary.ID)
		}

		*datapipelineOutput = *resp

		return nil
	}
}

func testAccCheckPipelineDefinitionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datapipeline_pipeline_definition" {
			continue
		}

		resp, err := conn.GetPipelineDefinitionWithContext(context.Background(), &datapipeline.GetPipelineDefinitionInput{PipelineId: aws.String(rs.Primary.ID)})

		if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) ||
			tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking if DataPipeline Pipeline Definition was destroyed: %w", err)
		}

		if resp != nil {
			return fmt.Errorf("DataPipeline Pipeline Definition %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccPipelineDefinitionConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = %[1]q
}

resource "aws_datapipeline_pipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id
  pipeline_object {
    id   = "Default"
    name = "Default"
    field {
      key          = "workerGroup"
      string_value = "workerGroup"
    }
  }
}
`, name)
}

func testAccPipelineDefinitionConfig_complete(name string, parameterObjectID string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "datapipeline.amazonaws.com",
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

}

resource "aws_datapipeline_pipeline" "default" {
  name = %[1]q
}

resource "aws_datapipeline_pipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id

  parameter_object {
    id = %[2]q

    attribute {
      key          = "description"
      string_value = "AWS CLI command"
    }
    attribute {
      key          = "type"
      string_value = "String"
    }
    attribute {
      key          = "watermark"
      string_value = "aws [options] <command> <subcommand> [parameters]"
    }
  }

  parameter_value {
    id           = %[2]q
    string_value = "aws sts get-caller-identity"
  }

  pipeline_object {
    id   = "CliActivity"
    name = "CliActivity"

    field {
      key          = "command"
      string_value = "(sudo yum -y update aws-cli) && (#{%[2]s})"
    }
    field {
      key       = "runsOn"
      ref_value = "Ec2Instance"
    }
    field {
      key          = "type"
      string_value = "ShellCommandActivity"
    }
  }
  pipeline_object {
    id   = "Default"
    name = "Default"

    field {
      key          = "failureAndRerunMode"
      string_value = "CASCADE"
    }
    field {
      key          = "resourceRole"
      string_value = aws_iam_role.test.name
    }
    field {
      key          = "role"
      string_value = aws_iam_role.test.name
    }
    field {
      key       = "schedule"
      ref_value = "DefaultSchedule"
    }
    field {
      key          = "scheduleType"
      string_value = "cron"
    }
  }
  pipeline_object {
    id   = "Ec2Instance"
    name = "Ec2Instance"

    field {
      key          = "instanceType"
      string_value = "t1.micro"
    }
    field {
      key          = "terminateAfter"
      string_value = "50 minutes"
    }
    field {
      key          = "type"
      string_value = "Ec2Resource"
    }
  }
  pipeline_object {
    id   = "DefaultSchedule"
    name = "Every 2 day"

    field {
      key          = "period"
      string_value = "1 days"
    }
    field {
      key          = "startAt"
      string_value = "FIRST_ACTIVATION_DATE_TIME"
    }
    field {
      key          = "type"
      string_value = "Schedule"
    }
  }
}
`, name, parameterObjectID)
}
