// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rekognition_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfrekognition "github.com/hashicorp/terraform-provider-aws/internal/service/rekognition"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRekognitionProject_WithImport(t *testing.T) {
	ctx := acctest.Context(t)

	rProjectId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CONTENT_MODERATION"
	autoUpdate := "ENABLED"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, feature, rProjectId),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_contentModeration(rProjectId, autoUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", rProjectId),
					resource.TestCheckResourceAttr(resourceName, "name", rProjectId),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_update", autoUpdate),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"arn"},
			},
		},
	})
}

func TestAccRekognitionProject_ContentModeration(t *testing.T) {
	ctx := acctest.Context(t)

	rProjectId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CONTENT_MODERATION"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_contentModeration(rProjectId+"-1", "ENABLED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", rProjectId+"-1"),
					resource.TestCheckResourceAttr(resourceName, "name", rProjectId+"-1"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_update", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
				),
			},
			{
				Config: testAccProjectConfig_contentModeration(rProjectId+"-2", "DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", rProjectId+"-2"),
					resource.TestCheckResourceAttr(resourceName, "name", rProjectId+"-2"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_update", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
				),
			},
		},
	})
}

func TestAccRekognitionProject_CustomLabels(t *testing.T) {
	ctx := acctest.Context(t)

	rProjectId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_rekognition_project.test"
	feature := "CUSTOM_LABELS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.Rekognition)
			testAccProjectPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.Rekognition),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProjectDestroy(ctx, feature, rProjectId),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfig_customLabels(rProjectId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "id", rProjectId),
					resource.TestCheckResourceAttr(resourceName, "name", rProjectId),
					resource.TestCheckResourceAttr(resourceName, "feature", feature),
					resource.TestCheckNoResourceAttr(resourceName, "auto_update"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"arn"},
			},
		},
	})
}

func testAccCheckProjectDestroy(ctx context.Context, feature string, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_rekognition_project" {
				continue
			}

			_, err := tfrekognition.FindProjectByName(ctx, conn, name, awstypes.CustomizationFeature(feature))
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.Rekognition, create.ErrActionCheckingDestroyed, tfrekognition.ResNameProject, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccProjectPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionClient(ctx)

	input := &rekognition.DescribeProjectsInput{}
	_, err := conn.DescribeProjects(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccProjectConfig_contentModeration(rProjectId string, autoUpdate string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_project" "test" {
  name        = "%s"
	auto_update = "%s" 
	feature     = "CONTENT_MODERATION"
}
`, rProjectId, autoUpdate)
}

// auto-update not supported for custom_labels
func testAccProjectConfig_customLabels(rProjectId string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_project" "test" {
  name        = "%s"
	feature     = "CUSTOM_LABELS"
}
`, rProjectId)
}
