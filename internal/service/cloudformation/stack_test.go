// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCloudFormationStack_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckNoResourceAttr(resourceName, "on_failure"),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct2),
					resource.TestCheckResourceAttrSet(resourceName, "outputs.DefaultSgId"),
					resource.TestCheckResourceAttrSet(resourceName, "outputs.VpcID"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct0),
					resource.TestCheckNoResourceAttr(resourceName, "template_url"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccStackConfig_basic(rName),
				PlanOnly: true,
			},
		},
	})
}

func TestAccCloudFormationStack_CreationFailure_doNothing(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackConfig_creationFailure(rName, string(awstypes.OnFailureDoNothing)),
				ExpectError: regexache.MustCompile(`(?s)stack status \(CREATE_FAILED\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_CreationFailure_delete(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackConfig_creationFailure(rName, string(awstypes.OnFailureDelete)),
				ExpectError: regexache.MustCompile(`(?s)stack status \(DELETE_COMPLETE\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_CreationFailure_rollback(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccStackConfig_creationFailure(rName, string(awstypes.OnFailureRollback)),
				ExpectError: regexache.MustCompile(`(?s)stack status \(ROLLBACK_COMPLETE\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_updateFailure(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	vpcCidrInitial := "10.0.0.0/16"
	vpcCidrInvalid := "1000.0.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_params(rName, vpcCidrInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
			{
				Config:      testAccStackConfig_params(rName, vpcCidrInvalid),
				ExpectError: regexache.MustCompile(`stack status \(UPDATE_ROLLBACK_COMPLETE\).*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccCloudFormationStack_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcloudformation.ResourceStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStack_yaml(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_yaml(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
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

func TestAccCloudFormationStack_defaultParams(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_defaultParams(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
		},
	})
}

func TestAccCloudFormationStack_allAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"
	expectedPolicyBody := "{\"Statement\":[{\"Action\":\"Update:*\",\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"LogicalResourceId/StaticVPC\"},{\"Action\":\"Update:*\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"*\"}]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_allAttributesBodies(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.First", "Mickey"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", acctest.Ct10),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", names.AttrParameters, "policy_body"},
			},
			{
				Config: testAccStackConfig_allAttributesBodiesModified(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags.First", "Mickey"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", acctest.Ct10),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4332
func TestAccCloudFormationStack_withParams(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	vpcCidrInitial := "10.0.0.0/16"
	vpcCidrUpdated := "12.0.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_params(rName, vpcCidrInitial),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", vpcCidrInitial),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", names.AttrParameters},
			},
			{
				Config: testAccStackConfig_params(rName, vpcCidrUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", vpcCidrUpdated),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct0),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4534
func TestAccCloudFormationStack_WithURL_withParams(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_templateURLParams(rName, "tf-cf-stack.json", "11.0.0.0/16"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", names.AttrParameters, "template_url"},
			},
			{
				Config: testAccStackConfig_templateURLParams(rName, "tf-cf-stack.json", "13.0.0.0/16"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
		},
	})
}

func TestAccCloudFormationStack_WithURLWithParams_withYAML(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_templateURLParamsYAML(rName, "tf-cf-stack.test", "13.0.0.0/16"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", names.AttrParameters, "template_url"},
			},
		},
	})
}

// Test for https://github.com/hashicorp/terraform/issues/5653
func TestAccCloudFormationStack_WithURLWithParams_noUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_templateURLParams(rName, "tf-cf-stack-1.json", "11.0.0.0/16"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", names.AttrParameters, "template_url"},
			},
			{
				Config: testAccStackConfig_templateURLParams(rName, "tf-cf-stack-2.json", "11.0.0.0/16"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
		},
	})
}

func TestAccCloudFormationStack_withTransform(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_transform(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
			{
				PlanOnly: true,
				Config:   testAccStackConfig_transform(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
				),
			},
		},
	})
}

// TestAccCloudFormationStack_onFailure verifies https://github.com/hashicorp/terraform-provider-aws/issues/5204
func TestAccCloudFormationStack_onFailure(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_onFailure(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "on_failure", string(awstypes.OnFailureDoNothing)),
				),
			},
		},
	})
}

func TestAccCloudFormationStack_outputsUpdated(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_parametersAndOutputs(rName, "in1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.DataIn", "in1"),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outputs.DataOut", "pre-in1-post"),
					resource.TestCheckOutput("stack_DataOut", "pre-in1-post"),
				),
			},
			{
				Config: testAccStackConfig_parametersAndOutputs(rName, "in2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("outputs")),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "parameters.DataIn", "in2"),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outputs.DataOut", "pre-in2-post"),
					resource.TestCheckOutput("stack_DataOut", "pre-in2-post"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameters},
			},
		},
	})
}

func TestAccCloudFormationStack_templateUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var stack awstypes.Stack
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CloudFormationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckStackDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccStackConfig_templateUpdate(rName, "out1", acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outputs.out1", acctest.CtValue1),
					resource.TestCheckOutput("stack_output", "out1:value1"),
				),
			},
			{
				Config: testAccStackConfig_templateUpdate(rName, "out2", acctest.CtValue2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue(resourceName, tfjsonpath.New("outputs")),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckStackExists(ctx, resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "outputs.out2", acctest.CtValue2),
					resource.TestCheckOutput("stack_output", "out2:value2"),
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

func testAccCheckStackExists(ctx context.Context, n string, v *awstypes.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		output, err := tfcloudformation.FindStackByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStackDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudformation_stack" {
				continue
			}

			_, err := tfcloudformation.FindStackByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudFormation Stack %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStackConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "10.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
}
`, rName)
}

func testAccStackConfig_creationFailure(rName, onFailure string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name       = %[1]q
  on_failure = %[2]q

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "1000.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
}
`, rName, onFailure)
}

func testAccStackConfig_yaml(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
Resources:
  MyVPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        -
          Key: Name
          Value: Primary_CF_VPC

Outputs:
  DefaultSgId:
    Description: The ID of default security group
    Value: !GetAtt MyVPC.DefaultSecurityGroup
  VpcID:
    Description: The VPC ID
    Value: !Ref MyVPC
STACK
}
`, rName)
}

func testAccStackConfig_defaultParams(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<BODY
{
    "Parameters": {
        "TopicName": {
            "Type": "String"
        },
        "VPCCIDR": {
            "Type": "String",
            "Default": "10.10.0.0/16"
        }
    },
    "Resources": {
        "NotificationTopic": {
            "Type": "AWS::SNS::Topic",
            "Properties": {
                "TopicName": {
                    "Ref": "TopicName"
                }
            }
        },
        "MyVPC": {
            "Type": "AWS::EC2::VPC",
            "Properties": {
                "CidrBlock": {
                    "Ref": "VPCCIDR"
                },
                "Tags": [
                    {
                        "Key": "Name",
                        "Value": "Primary_CF_VPC"
                    }
                ]
            }
        }
    },
    "Outputs": {
        "VPCCIDR": {
            "Value": {
                "Ref": "VPCCIDR"
            }
        }
    }
}
BODY


  parameters = {
    TopicName = %[1]q
  }
}
`, rName)
}

var testAccStackConfig_allAttributesWithBodies_tpl = `
data "aws_partition" "current" {}

resource "aws_cloudformation_stack" "test" {
  name          = %[1]q
  template_body = <<STACK
{
  "Parameters" : {
    "VpcCIDR" : {
      "Description" : "CIDR to be used for the VPC",
      "Type" : "String"
    }
  },
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": %[1]q}
        ]
      }
    },
    "StaticVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": "%[1]s-2"}
        ]
      }
    },
    "InstanceRole" : {
      "Type" : "AWS::IAM::Role",
      "Properties" : {
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [ {
            "Effect": "Allow",
            "Principal": { "Service": "ec2.${data.aws_partition.current.dns_suffix}" },
            "Action": "sts:AssumeRole"
          } ]
        },
        "Path" : "/",
        "Policies" : [ {
          "PolicyName": "terraformtest",
          "PolicyDocument": {
            "Version": "2012-10-17",
            "Statement": [ {
              "Effect": "Allow",
              "Action": [ "ec2:DescribeSnapshots" ],
              "Resource": [ "*" ]
            } ]
          }
        } ]
      }
    }
  }
}
STACK
  parameters = {
    VpcCIDR = "10.0.0.0/16"
  }

  policy_body        = <<POLICY
%[2]s
POLICY
  capabilities       = ["CAPABILITY_IAM"]
  notification_arns  = [aws_sns_topic.test.arn]
  on_failure         = "DELETE"
  timeout_in_minutes = 10
  tags = {
    First  = "Mickey"
    Second = "Mouse"
  }
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`

var policyBody = `
{
  "Statement" : [
    {
      "Effect" : "Deny",
      "Action" : "Update:*",
      "Principal": "*",
      "Resource" : "LogicalResourceId/StaticVPC"
    },
    {
      "Effect" : "Allow",
      "Action" : "Update:*",
      "Principal": "*",
      "Resource" : "*"
    }
  ]
}
`

func testAccStackConfig_allAttributesBodies(rName string) string {
	return fmt.Sprintf(
		testAccStackConfig_allAttributesWithBodies_tpl,
		rName,
		policyBody)
}

func testAccStackConfig_allAttributesBodiesModified(rName string) string {
	return fmt.Sprintf(
		testAccStackConfig_allAttributesWithBodies_tpl,
		rName,
		policyBody)
}

func testAccStackConfig_params(rName, cidr string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q
  parameters = {
    VpcCIDR = %[2]q
  }
  template_body = <<STACK
{
  "Parameters" : {
    "VpcCIDR" : {
      "Description" : "CIDR to be used for the VPC",
      "Type" : "String"
    }
  },
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  }
}
STACK

  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, cidr)
}

func testAccStackConfig_baseTemplateURL(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_s3_bucket" "b" {
  bucket = %[1]q
}

resource "aws_s3_bucket_public_access_block" "b" {
  bucket = aws_s3_bucket.b.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_ownership_controls" "b" {
  bucket = aws_s3_bucket.b.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_acl" "b" {
  bucket = aws_s3_bucket_policy.b.bucket
  acl    = "public-read"
}

resource "aws_s3_bucket_policy" "b" {
  depends_on = [
    aws_s3_bucket_public_access_block.b,
    aws_s3_bucket_ownership_controls.b,
  ]

  bucket = aws_s3_bucket.b.id
  policy = <<POLICY
{
  "Version":"2008-10-17",
  "Statement": [
    {
      "Sid":"AllowPublicRead",
      "Effect":"Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket_website_configuration" "b" {
  bucket = aws_s3_bucket_acl.b.bucket
  index_document {
    suffix = "index.html"
  }
  error_document {
    key = "error.html"
  }
}
`, rName)
}

func testAccStackConfig_templateURLParams(rName, bucketKey, vpcCidr string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseTemplateURL(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket_acl.b.bucket
  key    = %[2]q
  source = "test-fixtures/cloudformation-template.json"
}

resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  parameters = {
    VpcCIDR = %[3]q
  }

  template_url       = "https://${aws_s3_bucket.b.id}.s3-${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${aws_s3_object.object.key}"
  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr))
}

func testAccStackConfig_templateURLParamsYAML(rName, bucketKey, vpcCidr string) string {
	return acctest.ConfigCompose(testAccStackConfig_baseTemplateURL(rName), fmt.Sprintf(`
resource "aws_s3_object" "object" {
  bucket = aws_s3_bucket_acl.b.bucket
  key    = %[2]q
  source = "test-fixtures/cloudformation-template.yaml"
}

resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  parameters = {
    VpcCIDR = %[3]q
  }

  template_url       = "https://${aws_s3_bucket.b.id}.s3-${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${aws_s3_object.object.key}"
  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr))
}

func testAccStackConfig_transform(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Transform": "AWS::Serverless-2016-10-31",
  "Resources": {
    "Api": {
      "Type": "AWS::Serverless::Api",
      "Properties": {
        "StageName": "Prod",
        "EndpointConfiguration": "REGIONAL",
        "DefinitionBody": {
          "swagger": "2.0",
          "paths": {
            "/": {
              "get": {
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "responses": {
                  "200": {
                    "description": "200 response"
                  }
                },
                "x-amazon-apigateway-integration": {
                  "responses": {
                    "default": {
                      "statusCode": "200"
                    }
                  },
                  "requestTemplates": {
                    "application/json": "{\"statusCode\": 200}"
                  },
                  "passthroughBehavior": "when_no_match",
                  "type": "mock"
                }
              }
            }
          }
        }
      }
    }
  }
}
STACK

  capabilities       = ["CAPABILITY_AUTO_EXPAND"]
  on_failure         = "DELETE"
  timeout_in_minutes = 10
}
`, rName)
}

func testAccStackConfig_onFailure(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_cloudformation_stack" "test" {
  name       = %[1]q
  on_failure = "DO_NOTHING"

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"
    Resources = {
      S3Bucket = {
        Type = "AWS::S3::Bucket"
      }
    }
  })
}
`, rName)
}

func testAccStackConfig_parametersAndOutputs(rName, dataIn string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = <<STACK
  {
    "AWSTemplateFormatVersion": "2010-09-09",
    "Description": "AWS CloudFormation template that a transformed copy of its input parameter",
    "Parameters": {
      "DataIn": {
        "Type": "String",
        "Description": "Input data"
      }
    },
    "Resources": {
      "NullResource": {
        "Type": "AWS::CloudFormation::WaitConditionHandle"
      }
    },
    "Outputs": {
      "DataOut": {
        "Description": "Output data",
        "Value": {
          "Fn::Join": [
            "",
            [
              "pre-",
              {
                "Ref": "DataIn"
              },
              "-post"
            ]
          ]
        }
      }
    }
}
STACK

  parameters = {
    DataIn = %[2]q
  }
}

output "stack_DataOut" {
  value = aws_cloudformation_stack.test.outputs["DataOut"]
}
`, rName, dataIn)
}

func testAccStackConfig_templateUpdate(rName, name, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = %[1]q

  template_body = jsonencode(
    merge(
      jsondecode(local.template),
      { "Outputs" = local.outputs },
    )
  )
}

locals {
  template = <<STACK
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "AWS CloudFormation template that returns a constant value",
  "Conditions": {
	  "NeverTrue": {"Fn::Equals": ["true","false"]}
  },
  "Resources": {
	  "nullRes": {
		  "Type": "Custom::NullResource",
		  "Condition": "NeverTrue",
		  "Properties": {
			  "ServiceToken": ""
		  }
	  }
  }
}
STACK
  outputs = {
    %[2]s = {
      Value = %[3]q
    }
  }
}

output "stack_output" {
  value = format("%%s:%%s", %[2]q, aws_cloudformation_stack.test.outputs[%[2]q])
}
`, rName, name, value)
}
