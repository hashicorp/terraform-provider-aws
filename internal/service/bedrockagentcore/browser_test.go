// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreBrowser_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_basic(rName, "test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.network_mode", "PUBLIC"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "bedrock-agentcore", regexache.MustCompile(`browser-custom/.+$`)),
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

func TestAccBedrockAgentCoreBrowser_role_recording(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_iamRole(rName),
				Check: resource.ComposeTestCheckFunc(
					// Wait for IAM role and policy to propagate
					acctest.CheckSleep(t, 5*time.Second),
				),
			},
			{
				Config: testAccBrowserConfig_role_recording(rName, "bucket.test.com", "the_prefix"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "recording.s3_location.bucket", "bucket.test.com"),
					resource.TestCheckResourceAttr(resourceName, "recording.s3_location.prefix", "the_prefix"),
					resource.TestCheckResourceAttr(resourceName, "recording.enabled", "true"),
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

func TestAccBedrockAgentCoreBrowser_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var browser bedrockagentcorecontrol.GetBrowserOutput
	rName := strings.ReplaceAll(sdkacctest.RandomWithPrefix(acctest.ResourcePrefix), "-", "_")
	resourceName := "aws_bedrockagentcore_browser.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckBrowser(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBrowserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBrowserConfig_basic(rName, "test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBrowserExists(ctx, resourceName, &browser),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceBrowser, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccCheckBrowserDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_browser" {
				continue
			}
			_, err := tfbedrockagentcore.FindBrowserByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameBrowser, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameBrowser, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBrowserExists(ctx context.Context, name string, browser *bedrockagentcorecontrol.GetBrowserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameBrowser, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameBrowser, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindBrowserByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameBrowser, rs.Primary.ID, err)
		}

		*browser = *resp

		return nil
	}
}

func testAccPreCheckBrowser(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

	input := &bedrockagentcorecontrol.ListBrowsersInput{}

	_, err := conn.ListBrowsers(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckBrowserNotRecreated(before, after *bedrockagentcorecontrol.GetBrowserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.BrowserId), aws.ToString(after.BrowserId); before != after {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameBrowser, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccBrowserConfig_basic(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name        = %[1]q
  description = %[2]q

  network_configuration = {
    network_mode = "PUBLIC"
  }
}
`, rName, description)
}

func testAccBrowserConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
}

data "aws_iam_policy_document" "test_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccBrowserConfig_role_recording(rName, bucket, prefix string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_browser" "test" {
  name               = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  network_configuration = {
    network_mode = "PUBLIC"
  }

  recording = {
    s3_location = {
      bucket = %[2]q
      prefix = %[3]q
    }
    enabled = true
  }
}
`, rName, bucket, prefix))
}
