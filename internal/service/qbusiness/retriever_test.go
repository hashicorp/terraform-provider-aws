// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/qbusiness"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfqbusiness "github.com/hashicorp/terraform-provider-aws/internal/service/qbusiness"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccQBusinessRetriever_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var retriever qbusiness.GetRetrieverOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_retriever.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRetriever(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRetrieverDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRetrieverConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRetrieverExists(ctx, resourceName, &retriever),
					resource.TestCheckResourceAttrSet(resourceName, "retriever_id"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "type"),
					resource.TestCheckResourceAttr(resourceName, "native_index_configuration.string_boost_override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
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

func TestAccQBusinessRetriever_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var retriever qbusiness.GetRetrieverOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_retriever.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRetriever(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRetrieverDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRetrieverConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRetrieverExists(ctx, resourceName, &retriever),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceRetriever, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessRetriever_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var retriever qbusiness.GetRetrieverOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_retriever.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckRetriever(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRetrieverDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRetrieverConfig_tags(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRetrieverExists(ctx, resourceName, &retriever),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRetrieverConfig_tags(rName, "key1", "value1new", "key2", "value2new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRetrieverExists(ctx, resourceName, &retriever),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1new"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2new"),
				),
			},
		},
	})
}

func testAccPreCheckRetriever(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

	input := &qbusiness.ListApplicationsInput{}

	_, err := conn.ListApplications(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckRetrieverDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_retriever" {
				continue
			}

			_, err := tfqbusiness.FindRetrieverByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amazon Q Retriever %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRetrieverExists(ctx context.Context, n string, v *qbusiness.GetRetrieverOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		output, err := tfqbusiness.FindRetrieverByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRetrieverConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_qbusiness_retriever" "test" {
  application_id = "3f8d08f8-5729-4096-94e8-142378ae83c0"
  display_name   = %[1]q
  type           = "NATIVE_INDEX"

  native_index_configuration {
    index_id = "61934168-bdba-4b28-9869-5564ee794e89"
  }
}
`, rName)
}

func testAccRetrieverConfig_boostOverrides(rName string) string {
	return fmt.Sprintf(`
resource "aws_qbusiness_retriever" "test" {
  application_id = "3f8d08f8-5729-4096-94e8-142378ae83c0"
  display_name   = %[1]q
  type           = "NATIVE_INDEX"

  native_index_configuration {
    index_id = "61934168-bdba-4b28-9869-5564ee794e89"

    string_boost_override {
      boost_key      = "_category"
      boosting_level = "HIGH"

      attribute_value_boosting = {
        "key1" = "VERY_HIGH"
        "key2" = "VERY_HIGH"
      }
    }

    string_list_boost_override {
      boost_key      = "string_list"
      boosting_level = "HIGH"
    }

    date_boost_override {
      boost_key         = "date"
      boosting_level    = "HIGH"
      boosting_duration = 100
    }

    number_boost_override {
      boost_key      = "view_count"
      boosting_level = "HIGH"
      boosting_type  = "PRIORITIZE_LARGER_VALUES"
    }
  }
}
`, rName)
}

func testAccRetrieverConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_qbusiness_retriever" "test" {
  application_id = aws_qbusiness_app.test.application_id
  display_name   = %[1]q

  native_index_configuration {
    index_id = aws_qbusiness_index.test.index_id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

resource "aws_qbusiness_app" "test" {
  display_name         = %[1]q
  iam_service_role_arn = aws_iam_role.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Action": "sts:AssumeRole",
	"Principal": {
		"Service": "qbusiness.${data.aws_partition.current.dns_suffix}"
	},
	"Effect": "Allow",
	"Sid": ""
	}
	]
}
EOF
}

resource "aws_qbusiness_index" "test" {
  application_id = aws_qbusiness_app.test.application_id
  display_name   = %[1]q
  capacity_configuration {
    units = 1
  }
  description = %[1]q
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
