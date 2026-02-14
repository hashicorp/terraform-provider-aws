// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datapipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datapipeline"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datapipeline/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfdatapipeline "github.com/hashicorp/terraform-provider-aws/internal/service/datapipeline"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataPipelinePipeline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf1, conf2 awstypes.PipelineDescription
	rName1 := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	rName2 := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataPipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, t, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName1),
				),
			},
			{
				Config: testAccPipelineConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, t, resourceName, &conf2),
					testAccCheckPipelineNotEqual(&conf1, &conf2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName2),
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

func TestAccDataPipelinePipeline_description(t *testing.T) {
	ctx := acctest.Context(t)
	var conf1, conf2 awstypes.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataPipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_description(rName, "test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, t, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
				),
			},
			{
				Config: testAccPipelineConfig_description(rName, "update description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, t, resourceName, &conf2),
					testAccCheckPipelineNotEqual(&conf1, &conf2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "update description"),
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

func TestAccDataPipelinePipeline_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var conf awstypes.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DataPipelineServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, t, resourceName, &conf),
					testAccCheckPipelineDisappears(ctx, t, &conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPipelineDisappears(ctx context.Context, t *testing.T, conf *awstypes.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataPipelineClient(ctx)
		params := &datapipeline.DeletePipelineInput{
			PipelineId: conf.PipelineId,
		}

		_, err := conn.DeletePipeline(ctx, params)
		if err != nil {
			return err
		}
		return tfdatapipeline.WaitForDeletion(ctx, conn, *conf.PipelineId)
	}
}

func testAccCheckPipelineDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DataPipelineClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datapipeline_pipeline" {
				continue
			}
			// Try to find the Pipeline
			pipelineDescription, err := tfdatapipeline.FindPipeline(ctx, conn, rs.Primary.ID)
			if errs.IsA[*awstypes.PipelineNotFoundException](err) {
				continue
			} else if errs.IsA[*awstypes.PipelineDeletedException](err) {
				continue
			}

			if err != nil {
				return err
			}
			if pipelineDescription != nil {
				return fmt.Errorf("Pipeline still exists")
			}
		}

		return nil
	}
}

func testAccCheckPipelineExists(ctx context.Context, t *testing.T, n string, v *awstypes.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataPipeline ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).DataPipelineClient(ctx)

		pipelineDescription, err := tfdatapipeline.FindPipeline(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}
		if pipelineDescription == nil {
			return fmt.Errorf("DataPipeline not found")
		}

		*v = *pipelineDescription
		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).DataPipelineClient(ctx)

	input := &datapipeline.ListPipelinesInput{}

	_, err := conn.ListPipelines(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckPipelineNotEqual(pipeline1, pipeline2 *awstypes.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(pipeline1.PipelineId) == aws.ToString(pipeline2.PipelineId) {
			return fmt.Errorf("Pipeline IDs are equal")
		}

		return nil
	}
}

func testAccPipelineConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = "%[1]s"
}`, rName)
}

func testAccPipelineConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name        = "%[1]s"
  description = %[2]q
}`, rName, description)
}
