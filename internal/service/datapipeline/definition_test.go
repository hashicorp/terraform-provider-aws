package datapipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatapipeline "github.com/hashicorp/terraform-provider-aws/internal/service/datapipeline"
)

func TestAccDataPipelineDefinition_basic(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_definition.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataPipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDataPipelineDefinitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataPipelineDefinitionExists(resourceName, &pipelineOutput),
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

func TestAccDataPipelineDefinition_disappears(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_definition.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataPipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDataPipelineDefinitionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataPipelineDefinitionExists(resourceName, &pipelineOutput),
					acctest.CheckResourceDisappears(acctest.Provider, tfdatapipeline.ResourceDefinition(), resourceName),
				),
			},
		},
	})
}

func TestAccDataPipelineDefinition_complete(t *testing.T) {
	var pipelineOutput datapipeline.GetPipelineDefinitionOutput
	resourceName := "aws_datapipeline_definition.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataPipelineDefinitionDestroy,
		ErrorCheck:        acctest.ErrorCheck(t, datapipeline.EndpointsID),
		Steps: []resource.TestStep{
			{
				Config: testAccDataPipelineDefinitionConfigComplete(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataPipelineDefinitionExists(resourceName, &pipelineOutput),
				),
			},
			{
				Config: testAccDataPipelineDefinitionConfigCompleteUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataPipelineDefinitionExists(resourceName, &pipelineOutput),
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

func testAccCheckDataPipelineDefinitionExists(resourceName string, datapipelineOutput *datapipeline.GetPipelineDefinitionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn
		resp, err := conn.GetPipelineDefinitionWithContext(context.Background(), &datapipeline.GetPipelineDefinitionInput{PipelineId: aws.String(rs.Primary.ID)})
		if err != nil {
			return fmt.Errorf("problem checking for DataPipeline Definition existence: %w", err)
		}

		if resp == nil {
			return fmt.Errorf("datapipeline definition %q does not exist", rs.Primary.ID)
		}

		*datapipelineOutput = *resp

		return nil
	}
}

func testAccCheckDataPipelineDefinitionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datapipeline_definition" {
			continue
		}

		resp, err := conn.GetPipelineDefinitionWithContext(context.Background(), &datapipeline.GetPipelineDefinitionInput{PipelineId: aws.String(rs.Primary.ID)})

		if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) ||
			tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("problem while checking DataPipeline Definition was destroyed: %w", err)
		}

		if resp != nil {
			return fmt.Errorf("datapipeline definition %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccDataPipelineDefinitionConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = %[1]q
}

resource "aws_datapipeline_definition" "test" {
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

func testAccDataPipelineDefinitionConfigComplete(name string) string {
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

resource "aws_datapipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id

  parameter_object {
    id = "myAWSCLICmd"

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
    id           = "myAWSCLICmd"
    string_value = "aws sts get-caller-identity"
  }

  pipeline_object {
    id   = "CliActivity"
    name = "CliActivity"

    field {
      key          = "command"
      string_value = "(sudo yum -y update aws-cli) && (#{myAWSCLICmd})"
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
`, name)
}

func testAccDataPipelineDefinitionConfigCompleteUpdate(name string) string {
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

resource "aws_datapipeline_definition" "test" {
  pipeline_id = aws_datapipeline_pipeline.default.id

  parameter_object {
    id = "myAWSCLICmd"

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
    id           = "myAWSCLICmd"
    string_value = "aws sts get-caller-identity"
  }

  pipeline_object {
    id   = "CliActivity"
    name = "CliActivity"

    field {
      key          = "command"
      string_value = "(sudo yum -y update aws-cli) && (#{myAWSCLICmd})"
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
`, name)
}
