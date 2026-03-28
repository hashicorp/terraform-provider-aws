// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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

func TestAccEC2DeleteDefaultVPCAction_serial(t *testing.T) { // nosemgrep: ci.vpc-in-test-name
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccEC2DeleteDefaultVPCAction_basic,
		"noDefaultVPC":  testAccEC2DeleteDefaultVPCAction_noDefaultVPC,
		"trigger":       testAccEC2DeleteDefaultVPCAction_trigger,
	}

	acctest.RunSerialTests1Level(t, testCases, 5*time.Second)
}

func testAccEC2DeleteDefaultVPCAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var hadDefaultVPC bool
	var initialStateCaptured bool

	t.Cleanup(func() {
		restoreDefaultVPC(t, &hadDefaultVPC)
	})

	resource.Test(t, resource.TestCase{
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
				PreConfig: func() {
					// Capture initial state before any modifications (only once)
					if !initialStateCaptured {
						hadDefaultVPC = testAccDefaultVPCExists(ctx, t)
						initialStateCaptured = true
					}
					// Ensure default VPC exists before action
					testAccDeleteDefaultVPCPreCheckDefaultVPCExists(ctx, t)
				},
				Config: testAccDeleteDefaultVPCActionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultVPCExists(ctx),
				),
			},
			{
				PreConfig: func() {
					if err := invokeDeleteDefaultVPCAction(ctx, t); err != nil {
						t.Fatalf("Failed to invoke delete default VPC action: %v", err)
					}
				},
				Config: testAccDeleteDefaultVPCActionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultVPCDeleted(ctx),
				),
			},
		},
	})
}

func testAccEC2DeleteDefaultVPCAction_noDefaultVPC(t *testing.T) {
	ctx := acctest.Context(t)
	var hadDefaultVPC bool
	var initialStateCaptured bool

	t.Cleanup(func() {
		restoreDefaultVPC(t, &hadDefaultVPC)
	})

	resource.Test(t, resource.TestCase{
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
				PreConfig: func() {
					// Capture initial state before any modifications (only once)
					if !initialStateCaptured {
						hadDefaultVPC = testAccDefaultVPCExists(ctx, t)
						initialStateCaptured = true
					}
					testAccDeleteDefaultVPCIfExists(ctx, t)
				},
				Config: testAccDeleteDefaultVPCActionConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultVPCDeleted(ctx),
				),
			},
		},
	})
}

func testAccEC2DeleteDefaultVPCAction_trigger(t *testing.T) {
	ctx := acctest.Context(t)
	var hadDefaultVPC bool
	var initialStateCaptured bool

	t.Cleanup(func() {
		restoreDefaultVPC(t, &hadDefaultVPC)
	})

	resource.Test(t, resource.TestCase{
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
				PreConfig: func() {
					// Capture initial state before any modifications (only once)
					if !initialStateCaptured {
						hadDefaultVPC = testAccDefaultVPCExists(ctx, t)
						initialStateCaptured = true
					}
					testAccDeleteDefaultVPCPreCheckDefaultVPCExists(ctx, t)
				},
				Config: testAccDeleteDefaultVPCActionConfig_trigger(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDefaultVPCDeleted(ctx),
				),
			},
		},
	})
}

// testAccDefaultVPCExists checks if a default VPC exists in the current region
func testAccDefaultVPCExists(ctx context.Context, t *testing.T) bool {
	t.Helper()

	// Safety check for provider availability
	if acctest.Provider == nil {
		t.Log("WARNING: acctest.Provider is nil, cannot check for default VPC")
		return false
	}
	if acctest.Provider.Meta() == nil {
		t.Log("WARNING: acctest.Provider.Meta() is nil, cannot check for default VPC")
		return false
	}

	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVpcsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{acctest.CtTrue},
			},
		},
	}

	output, err := conn.DescribeVpcs(ctx, input)
	if err != nil {
		t.Logf("WARNING: error checking for default VPC: %s", err)
		return false
	}

	exists := len(output.Vpcs) > 0
	if exists {
		t.Logf("Default VPC already exists")
	} else {
		t.Logf("Default VPC does not exists")
	}

	return exists
}

// testAccDeleteDefaultVPCPreCheckDefaultVPCExists ensures a default VPC exists, creating one if needed
func testAccDeleteDefaultVPCPreCheckDefaultVPCExists(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVpcsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{acctest.CtTrue},
			},
		},
	}

	output, err := conn.DescribeVpcs(ctx, input)
	if err != nil {
		t.Fatalf("Error checking for default VPC: %s", err)
	}

	if len(output.Vpcs) == 0 {
		// Create default VPC
		createInput := ec2.CreateDefaultVpcInput{}
		createOutput, err := conn.CreateDefaultVpc(ctx, &createInput)
		if err != nil {
			t.Fatalf("Error creating default VPC for test: %s", err)
		}
		vpcID := aws.ToString(createOutput.Vpc.VpcId)
		t.Logf("Created default VPC for test: %s", vpcID)

		// Wait for VPC to become available
		if _, err := tfec2.WaitVPCCreated(ctx, conn, vpcID); err != nil {
			t.Logf("Warning: error waiting for default VPC to become available: %s", err)
		}
	}
}

// testAccDeleteDefaultVPCIfExists deletes the default VPC if it exists
func testAccDeleteDefaultVPCIfExists(ctx context.Context, t *testing.T) {
	t.Helper()

	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVpcsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{acctest.CtTrue},
			},
		},
	}

	output, err := conn.DescribeVpcs(ctx, input)
	if err != nil {
		t.Fatalf("Error checking for default VPC: %s", err)
	}

	if len(output.Vpcs) > 0 {
		vpcID := aws.ToString(output.Vpcs[0].VpcId)

		// Delete VPC dependencies using the action implementation
		progressFn := func(msg string) {
			t.Logf("Test progress: %s", msg)
		}

		if err := tfec2.DeleteDefaultVPCDependencies(ctx, conn, vpcID, progressFn); err != nil {
			t.Logf("Warning: failed to delete default VPC dependencies for %s: %v", vpcID, err)
		}

		// Delete the VPC itself
		deleteInput := ec2.DeleteVpcInput{
			VpcId: aws.String(vpcID),
		}
		_, err := conn.DeleteVpc(ctx, &deleteInput)
		if err != nil {
			t.Logf("Warning: failed to delete default VPC %s: %v", vpcID, err)
		}
	}
}

// testAccCheckDefaultVPCExists verifies a default VPC exists
func testAccCheckDefaultVPCExists(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		input := &ec2.DescribeVpcsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("isDefault"),
					Values: []string{acctest.CtTrue},
				},
			},
		}

		output, err := conn.DescribeVpcs(ctx, input)
		if err != nil {
			return fmt.Errorf("error describing VPCs: %w", err)
		}

		if len(output.Vpcs) == 0 {
			return fmt.Errorf("default VPC does not exist")
		}

		return nil
	}
}

// testAccCheckDefaultVPCDeleted verifies no default VPC exists
func testAccCheckDefaultVPCDeleted(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		input := &ec2.DescribeVpcsInput{
			Filters: []awstypes.Filter{
				{
					Name:   aws.String("isDefault"),
					Values: []string{acctest.CtTrue},
				},
			},
		}

		output, err := conn.DescribeVpcs(ctx, input)
		if err != nil {
			return fmt.Errorf("error describing VPCs: %w", err)
		}

		if len(output.Vpcs) > 0 {
			return fmt.Errorf("default VPC still exists: %s", aws.ToString(output.Vpcs[0].VpcId))
		}

		return nil
	}
}

// restoreDefaultVPC creates a new default VPC if one existed before the test
func restoreDefaultVPC(t *testing.T, hadDefaultVPC *bool) {
	t.Helper()

	// Check if we should restore
	if hadDefaultVPC == nil || !*hadDefaultVPC {
		t.Log("No default VPC existed before test, skipping restore")
		return
	}

	// Use a background context to ensure restore completes even if test context is cancelled
	ctx := context.Background()
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	// Check if default VPC already exists
	describeInput := &ec2.DescribeVpcsInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("isDefault"),
				Values: []string{acctest.CtTrue},
			},
		},
	}

	output, err := conn.DescribeVpcs(ctx, describeInput)
	if err != nil {
		t.Errorf("ERROR: Failed to check for default VPC during restore: %v", err)
		return
	}

	if len(output.Vpcs) > 0 {
		t.Log("Default VPC already exists, skipping restore")
		return
	}

	// Create default VPC
	t.Log("Restoring default VPC after test...")
	createInput := ec2.CreateDefaultVpcInput{}
	createOutput, err := conn.CreateDefaultVpc(ctx, &createInput)
	if err != nil {
		t.Errorf("ERROR: Failed to restore default VPC after test: %v", err)
		return
	}

	vpcID := aws.ToString(createOutput.Vpc.VpcId)
	t.Logf("Created default VPC: %s", vpcID)

	// Wait for VPC to become available
	if _, err := tfec2.WaitVPCCreated(ctx, conn, vpcID); err != nil {
		t.Errorf("ERROR: Default VPC %s was created but failed to become available: %v", vpcID, err)
		return
	}

	t.Logf("Successfully restored default VPC: %s", vpcID)
}

// deleteDefaultVPCProviderWithActions gets the provider as ProviderServerWithActions
func deleteDefaultVPCProviderWithActions(ctx context.Context, t *testing.T) tfprotov5.ProviderServerWithActions { //nolint:staticcheck // SA1019: Working in alpha situation
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

	providerConfigValue, err := deleteDefaultVPCBuildProviderConfiguration(t, schemaResp.Provider)
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

// deleteDefaultVPCBuildProviderConfiguration creates a minimal provider configuration from the schema
func deleteDefaultVPCBuildProviderConfiguration(t *testing.T, providerSchema *tfprotov5.Schema) (*tfprotov5.DynamicValue, error) {
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

// buildDeleteDefaultVPCActionConfig builds the action configuration
func buildDeleteDefaultVPCActionConfig() (tftypes.Type, map[string]tftypes.Value) {
	configType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			names.AttrTimeout: tftypes.Number,
			names.AttrRegion:  tftypes.String,
		},
	}

	config := map[string]tftypes.Value{
		names.AttrTimeout: tftypes.NewValue(tftypes.Number, nil),
		names.AttrRegion:  tftypes.NewValue(tftypes.String, nil),
	}

	return configType, config
}

// invokeDeleteDefaultVPCAction invokes the action programmatically
func invokeDeleteDefaultVPCAction(ctx context.Context, t *testing.T) error {
	t.Helper()

	p := deleteDefaultVPCProviderWithActions(ctx, t)
	configType, configMap := buildDeleteDefaultVPCActionConfig()
	actionTypeName := "aws_vpc_delete_default_vpc"

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
			t.Logf("Received event type: %T", eventType)
		}
	}

	return nil
}

func testAccDeleteDefaultVPCActionConfig_basic() string {
	return `
action "aws_vpc_delete_default_vpc" "test" {
}
`
}

func testAccDeleteDefaultVPCActionConfig_trigger() string {
	return `
action "aws_vpc_delete_default_vpc" "test" {
  config {
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.aws_vpc_delete_default_vpc.test]
    }
  }
}
`
}
