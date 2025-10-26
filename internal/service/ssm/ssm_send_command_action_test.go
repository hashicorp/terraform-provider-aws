// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSSMSendCommandAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSM)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccSendCommandActionConfig_basic(rName),
			},
		},
	})
}

func TestAccSSMSendCommandAction_withParameters(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSM)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccSendCommandActionConfig_withParameters(rName),
			},
		},
	})
}

func TestAccSSMSendCommandAction_trigger(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.SSM)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.SSMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccSendCommandActionConfig_trigger(rName),
			},
		},
	})
}

func testAccSendCommandActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSendCommandActionConfig_base(rName),
		`
action "aws_ssm_send_command" "test" {
  config {
    instance_ids  = [aws_instance.test.id]
    document_name = "AWS-RunShellScript"
    parameters = {
      commands = ["echo 'Hello World'"]
    }
  }
}
`)
}

func testAccSendCommandActionConfig_withParameters(rName string) string {
	return acctest.ConfigCompose(
		testAccSendCommandActionConfig_base(rName),
		`
action "aws_ssm_send_command" "test" {
  config {
    instance_ids  = [aws_instance.test.id]
    document_name = "AWS-RunShellScript"
    parameters = {
      commands = [
        "echo 'Test command'",
        "uptime"
      ]
    }
    timeout = 600
  }
}
`)
}

func testAccSendCommandActionConfig_trigger(rName string) string {
	return acctest.ConfigCompose(
		testAccSendCommandActionConfig_base(rName),
		`
action "aws_ssm_send_command" "test" {
  config {
    instance_ids  = [aws_instance.test.id]
    document_name = "AWS-RunShellScript"
    parameters = {
      commands = ["echo 'Triggered command'"]
    }
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.aws_ssm_send_command.test]
    }
  }
}
`)
}

func testAccSendCommandActionConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

resource "aws_instance" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  iam_instance_profile = aws_iam_instance_profile.test.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

// providerWithActions gets the AWS provider as a ProviderServerWithActions
func providerWithActions(ctx context.Context, t *testing.T) tfprotov5.ProviderServerWithActions { //nolint:staticcheck // SA1019: Working in alpha situation
	t.Helper()

	factories := acctest.ProtoV5ProviderFactories
	providerFactory, exists := factories["aws"]
	if !exists {
		t.Fatal("AWS provider factory not found in ProtoV5ProviderFactories")
	}

	providerServer, err := providerFactory()
	if err != nil {
		t.Fatalf("Failed to create provider server: %v", err)
	}

	providerWithActions, ok := providerServer.(tfprotov5.ProviderServerWithActions) //nolint:staticcheck // SA1019: Working in alpha situation
	if !ok {
		t.Fatal("Provider does not implement ProviderServerWithActions")
	}

	schemaResp, err := providerWithActions.GetProviderSchema(ctx, &tfprotov5.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("Failed to get provider schema: %v", err)
	}

	if len(schemaResp.ActionSchemas) == 0 {
		t.Fatal("Expected to find action schemas but didn't find any!")
	}

	providerConfigValue, err := buildProviderConfiguration(t, schemaResp.Provider)
	if err != nil {
		t.Fatalf("Failed to build provider configuration: %v", err)
	}

	configureResp, err := providerWithActions.ConfigureProvider(ctx, &tfprotov5.ConfigureProviderRequest{
		TerraformVersion: "1.0.0",
		Config:           providerConfigValue,
	})
	if err != nil {
		t.Fatalf("Failed to configure provider: %v", err)
	}

	if len(configureResp.Diagnostics) > 0 {
		var diagMessages []string
		for _, diag := range configureResp.Diagnostics {
			diagMessages = append(diagMessages, fmt.Sprintf("Severity: %s, Summary: %s, Detail: %s", diag.Severity, diag.Summary, diag.Detail))
		}
		t.Fatalf("Provider configuration failed: %v", diagMessages)
	}

	return providerWithActions
}

// buildProviderConfiguration creates a minimal provider configuration from the schema
func buildProviderConfiguration(t *testing.T, providerSchema *tfprotov5.Schema) (*tfprotov5.DynamicValue, error) {
	t.Helper()

	providerType := providerSchema.Block.ValueType()
	configMap := make(map[string]tftypes.Value)

	if objType, ok := providerType.(tftypes.Object); ok {
		for attrName, attrType := range objType.AttributeTypes {
			configMap[attrName] = tftypes.NewValue(attrType, nil)
		}
	}

	configValue, err := tfprotov5.NewDynamicValue(
		providerType,
		tftypes.NewValue(providerType, configMap),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	return &configValue, nil
}
