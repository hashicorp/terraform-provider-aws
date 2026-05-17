// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderWorkflow_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "imagebuilder", regexache.MustCompile(fmt.Sprintf("workflow/test/%s/1.0.0/[1-9][0-9]*", rName))),
					resource.TestCheckResourceAttr(resourceName, "change_description", ""),
					resource.TestMatchResourceAttr(resourceName, "data", regexache.MustCompile(`schemaVersion`)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrKMSKeyID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.CheckResourceAttrAccountID(ctx, resourceName, names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.WorkflowTypeTest)),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1.0.0"),
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

func TestAccImageBuilderWorkflow_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_name(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfimagebuilder.ResourceWorkflow(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderWorkflow_changeDescription(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_changeDescription(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "change_description", "description1"),
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

func TestAccImageBuilderWorkflow_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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

func TestAccImageBuilderWorkflow_kmsKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_workflow.test"
	kmsKeyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_kmsKeyID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrKMSKeyID, kmsKeyResourceName, names.AttrARN),
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

func TestAccImageBuilderWorkflow_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkflowConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWorkflowConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccImageBuilderWorkflow_uri(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_workflow.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWorkflowDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWorkflowConfig_uri(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWorkflowExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "data"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrURI),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrURI},
			},
		},
	})
}

func testAccCheckWorkflowDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ImageBuilderClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_imagebuilder_workflow" {
				continue
			}

			_, err := tfimagebuilder.FindWorkflowByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Image Builder Workflow %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWorkflowExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ImageBuilderClient(ctx)

		_, err := tfimagebuilder.FindWorkflowByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccWorkflowConfig_name(rName string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}
`, rName)
}

func testAccWorkflowConfig_changeDescription(rName, changeDescription string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  change_description = %[2]q

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}
`, rName, changeDescription)
}

func testAccWorkflowConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  description = %[2]q

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}
`, rName, description)
}

func testAccWorkflowConfig_kmsKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  kms_key_id = aws_kms_key.test.arn

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}
`, rName)
}

func testAccWorkflowConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  tags = {
    %[2]q = %[3]q
  }

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}
`, rName, tagKey1, tagValue1)
}

func testAccWorkflowConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  data = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccWorkflowConfig_uri(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.bucket
  key    = "test.yml"

  content = <<-EOT
  name: test-image
  description: Workflow to test an image
  schemaVersion: 1.0

  parameters:
    - name: waitForActionAtEnd
      type: boolean

  steps:
    - name: LaunchTestInstance
      action: LaunchInstance
      onFailure: Abort
      inputs:
        waitFor: "ssmAgent"

    - name: TerminateTestInstance
      action: TerminateInstance
      onFailure: Continue
      inputs:
        instanceId.$: "$.stepOutputs.LaunchTestInstance.instanceId"

    - name: WaitForActionAtEnd
      action: WaitForAction
      if:
        booleanEquals: true
        value: "$.parameters.waitForActionAtEnd"
  EOT
}

resource "aws_imagebuilder_workflow" "test" {
  name    = %[1]q
  version = "1.0.0"
  type    = "TEST"

  uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
}
`, rName)
}
