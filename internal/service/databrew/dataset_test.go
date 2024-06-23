// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package databrew_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/databrew"
	"github.com/aws/aws-sdk-go-v2/service/databrew/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfdatabrew "github.com/hashicorp/terraform-provider-aws/internal/service/databrew"
)

// Acceptance test access AWS and cost money to run.
func TestAccDataBrewDataset_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dataset databrew.DescribeDatasetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_databrew_dataset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataBrew)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatasetExists(ctx, resourceName, &dataset),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "name",
			},
		},
	})
}

func TestAccDataBrewDataset_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dataset databrew.DescribeDatasetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_databrew_dataset.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.DataBrewServiceID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.DataBrewServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatasetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatasetExists(ctx, resourceName, &dataset),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfdatabrew.ResourceDataset, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDatasetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataBrewClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_databrew_dataset" {
				continue
			}

			_, err := conn.DescribeDataset(ctx, &databrew.DescribeDatasetInput{
				Name: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.DataBrew, create.ErrActionCheckingDestroyed, tfdatabrew.ResNameDataset, rs.Primary.ID, err)
			}

			return create.Error(names.DataBrew, create.ErrActionCheckingDestroyed, tfdatabrew.ResNameDataset, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckDatasetExists(ctx context.Context, name string, dataset *databrew.DescribeDatasetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return create.Error(names.DataBrew, create.ErrActionCheckingExistence, tfdatabrew.ResNameDataset, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.DataBrew, create.ErrActionCheckingExistence, tfdatabrew.ResNameDataset, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataBrewClient(ctx)
		resp, err := conn.DescribeDataset(ctx, &databrew.DescribeDatasetInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.DataBrew, create.ErrActionCheckingExistence, tfdatabrew.ResNameDataset, rs.Primary.ID, err)
		}

		*dataset = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataBrewClient(ctx)

	input := &databrew.ListDatasetsInput{}
	_, err := conn.ListDatasets(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckDatasetNotRecreated(before, after *databrew.DescribeDatasetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.Name), aws.ToString(after.Name); before != after {
			return create.Error(names.DataBrew, create.ErrActionCheckingNotRecreated, tfdatabrew.ResNameDataset, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccDatasetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
	bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket         = aws_s3_bucket.test.bucket
  key            = %[1]q
  content_base64 = "dGVzdAo="
}

resource "aws_databrew_dataset" "test" {
  name         = %[1]q
  input {
	s3_input_definition {
		bucket = aws_s3_bucket.test.id
	}
  }
}
`, rName)
}
