// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerHub_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeHubOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHubConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "hub_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_description", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_search_keywords.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "s3_storage_config.#", "0"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("hub/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHubConfig_basic(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "hub_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_description", rNameUpdated),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("hub/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func TestAccSageMakerHub_searchKeywords(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeHubOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHubConfig_searchKeywords(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "hub_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_description", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_search_keywords.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hub_search_keywords.*", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccHubConfig_searchKeywordsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "hub_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_description", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_search_keywords.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hub_search_keywords.*", rName),
					resource.TestCheckTypeSetElemAttr(resourceName, "hub_search_keywords.*", fmt.Sprintf("%s-1", rName)),
				),
			},
			{
				Config: testAccHubConfig_searchKeywords(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "hub_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_description", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_search_keywords.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "hub_search_keywords.*", rName),
				),
			},
		},
	})
}

func TestAccSageMakerHub_s3(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeHubOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHubConfig_s3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "hub_name", rName),
					resource.TestCheckResourceAttr(resourceName, "s3_storage_config.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "s3_storage_config.0.s3_output_path"),
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

func TestAccSageMakerHub_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeHubOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHubConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
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
				Config: testAccHubConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccHubConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerHub_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var mpg sagemaker.DescribeHubOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccHubConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubExists(ctx, resourceName, &mpg),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceHub(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckHubDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_hub" {
				continue
			}

			_, err := tfsagemaker.FindHubByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker AI Hub (%s): %w", rs.Primary.ID, err)
			}

			return fmt.Errorf("sagemaker Hub %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckHubExists(ctx context.Context, n string, mpg *sagemaker.DescribeHubOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Hub ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)
		resp, err := tfsagemaker.FindHubByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccHubConfig_basic(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_hub" "test" {
  hub_name        = %[1]q
  hub_description = %[2]q
}
`, rName, desc)
}

func testAccHubConfig_searchKeywords(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_hub" "test" {
  hub_name        = %[1]q
  hub_description = %[1]q

  hub_search_keywords = ["%[1]s"]
}
`, rName)
}

func testAccHubConfig_searchKeywordsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_hub" "test" {
  hub_name        = %[1]q
  hub_description = %[1]q

  hub_search_keywords = ["%[1]s", "%[1]s-1"]
}
`, rName)
}

func testAccHubConfig_s3(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_sagemaker_hub" "test" {
  hub_name        = %[1]q
  hub_description = %[1]q

  s3_storage_config {
    s3_output_path = "s3://${aws_s3_bucket.test.bucket}/"
  }
}
`, rName)
}

func testAccHubConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_hub" "test" {
  hub_name        = %[1]q
  hub_description = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccHubConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_hub" "test" {
  hub_name        = %[1]q
  hub_description = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
