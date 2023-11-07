// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datapipeline_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatapipeline "github.com/hashicorp/terraform-provider-aws/internal/service/datapipeline"
)

func TestAccDataPipelinePipeline_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var conf1, conf2 datapipeline.PipelineDescription
	rName1 := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	rName2 := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datapipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccPipelineConfig_basic(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf2),
					testAccCheckPipelineNotEqual(&conf1, &conf2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
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
	var conf1, conf2 datapipeline.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datapipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_description(rName, "test description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf1),
					resource.TestCheckResourceAttr(resourceName, "description", "test description"),
				),
			},
			{
				Config: testAccPipelineConfig_description(rName, "update description"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf2),
					testAccCheckPipelineNotEqual(&conf1, &conf2),
					resource.TestCheckResourceAttr(resourceName, "description", "update description"),
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
	var conf datapipeline.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datapipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf),
					testAccCheckPipelineDisappears(ctx, &conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataPipelinePipeline_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var conf datapipeline.PipelineDescription
	rName := fmt.Sprintf("tf-datapipeline-%s", sdkacctest.RandString(5))
	resourceName := "aws_datapipeline_pipeline.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datapipeline.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPipelineDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPipelineConfig_tags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccPipelineConfig_tags(rName, "foo", "bar2", "fizz2", "buzz2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.foo", "bar2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.fizz2", "buzz2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPipelineConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists(ctx, resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckPipelineDisappears(ctx context.Context, conf *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn(ctx)
		params := &datapipeline.DeletePipelineInput{
			PipelineId: conf.PipelineId,
		}

		_, err := conn.DeletePipelineWithContext(ctx, params)
		if err != nil {
			return err
		}
		return tfdatapipeline.WaitForDeletion(ctx, conn, *conf.PipelineId)
	}
}

func testAccCheckPipelineDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datapipeline_pipeline" {
				continue
			}
			// Try to find the Pipeline
			pipelineDescription, err := tfdatapipeline.PipelineRetrieve(ctx, rs.Primary.ID, conn)
			if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) {
				continue
			} else if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
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

func testAccCheckPipelineExists(ctx context.Context, n string, v *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DataPipeline ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn(ctx)

		pipelineDescription, err := tfdatapipeline.PipelineRetrieve(ctx, rs.Primary.ID, conn)

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
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataPipelineConn(ctx)

	input := &datapipeline.ListPipelinesInput{}

	_, err := conn.ListPipelinesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckPipelineNotEqual(pipeline1, pipeline2 *datapipeline.PipelineDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(pipeline1.PipelineId) == aws.StringValue(pipeline2.PipelineId) {
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

func testAccPipelineConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_datapipeline_pipeline" "default" {
  name = "%[1]s"

  tags = {
    %[2]s = %[3]q
    %[4]s = %[5]q
  }
}`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
