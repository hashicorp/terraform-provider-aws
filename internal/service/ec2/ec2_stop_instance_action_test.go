// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2StopInstanceAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccStopInstanceActionConfig_force(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExistsLocal(ctx, resourceName, &v),
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameRunning),
				),
			},
			{
				PreConfig: func() {
					if v.InstanceId == nil {
						t.Fatal("Instance ID is nil")
					}

					if err := invokeStopInstanceAction(ctx, t, *v.InstanceId, true); err != nil {
						t.Fatalf("Failed to invoke stop instance action: %v", err)
					}
				},
				Config: testAccStopInstanceActionConfig_force(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameStopped),
				),
			},
		},
	})
}

func TestAccEC2StopInstanceAction_trigger(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Instance
	resourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccStopInstanceActionConfig_trigger(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExistsLocal(ctx, resourceName, &v),
					testAccCheckInstanceState(ctx, resourceName, awstypes.InstanceStateNameStopped),
				),
			},
		},
	})
}

func testAccCheckInstanceExistsLocal(ctx context.Context, n string, v *awstypes.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		instance, err := tfec2.FindInstanceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *instance

		return nil
	}
}

func testAccCheckInstanceState(ctx context.Context, n string, expectedState awstypes.InstanceStateName) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		instance, err := tfec2.FindInstanceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if instance.State.Name != expectedState {
			return fmt.Errorf("Expected instance state %s, got %s", expectedState, instance.State.Name)
		}

		return nil
	}
}

func testAccStopInstanceActionConfig_force(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}

action "aws_ec2_stop_instance" "test" {
  config {
    instance_id = aws_instance.test.id
    force       = true
  }
}
`, rName))
}

func testAccStopInstanceActionConfig_trigger(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  tags = {
    Name = %[1]q
  }
}

action "aws_ec2_stop_instance" "test" {
  config {
    instance_id = aws_instance.test.id
    force       = true
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_ec2_stop_instance.test]
    }
  }
}
`, rName))
}

// Step 1: Get the AWS provider as a ProviderServerWithActions
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

// Step 2: Build action configuration
func buildStopInstanceActionConfig(instanceID string, force bool) (tftypes.Type, map[string]tftypes.Value) {
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			names.AttrInstanceID: tftypes.String,
			"force":              tftypes.Bool,
			names.AttrTimeout:    tftypes.Number,
			names.AttrRegion:     tftypes.String,
		},
	}

	config := map[string]tftypes.Value{
		names.AttrInstanceID: tftypes.NewValue(tftypes.String, instanceID),
		"force":              tftypes.NewValue(tftypes.Bool, force),
		names.AttrTimeout:    tftypes.NewValue(tftypes.Number, nil),
		names.AttrRegion:     tftypes.NewValue(tftypes.String, nil),
	}

	return configType, config
}

// Step 3: Programmatic action invocation
func invokeStopInstanceAction(ctx context.Context, t *testing.T, instanceID string, force bool) error {
	t.Helper()

	p := providerWithActions(ctx, t)
	configType, configMap := buildStopInstanceActionConfig(instanceID, force)
	actionTypeName := "aws_ec2_stop_instance"

	testConfig, err := tfprotov5.NewDynamicValue(
		configType,
		tftypes.NewValue(configType, configMap),
	)
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	invokeResp, err := p.InvokeAction(ctx, &tfprotov5.InvokeActionRequest{
		ActionType: actionTypeName,
		Config:     &testConfig,
	})
	if err != nil {
		return fmt.Errorf("invoke failed: %w", err)
	}

	// Process events and check for completion
	for event := range invokeResp.Events {
		switch eventType := event.Type.(type) {
		case tfprotov5.ProgressInvokeActionEventType:
			t.Logf("Progress: %s", eventType.Message)
		case tfprotov5.CompletedInvokeActionEventType:
			return nil
		default:
			// Handle any other event types or errors
			t.Logf("Received event type: %T", eventType)
		}
	}

	return nil
}
