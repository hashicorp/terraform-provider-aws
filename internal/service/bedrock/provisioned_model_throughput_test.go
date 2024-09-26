// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockProvisionedModelThroughput_basic(t *testing.T) {
	acctest.Skip(t, "Bedrock Provisioned Model Throughput has a minimum 1 month commitment and costs > $10K/month")

	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrock_provisioned_model_throughput.test"
	var v bedrock.GetProvisionedModelThroughputOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProvisionedModelThroughputDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProvisionedModelThroughputConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProvisionedModelThroughputExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "commitment_duration", "OneMonth"),
					resource.TestCheckResourceAttrSet(resourceName, "model_arn"),
					resource.TestCheckResourceAttr(resourceName, "model_units", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "provisioned_model_arn"),
					resource.TestCheckResourceAttr(resourceName, "provisioned_model_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

// TODO TestAccBedrockProvisionedModelThroughput_disappears
// TODO TestAccBedrockProvisionedModelThroughput_tags

func testAccCheckProvisionedModelThroughputDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_custom_model" {
				continue
			}

			_, err := tfbedrock.FindProvisionedModelThroughputByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Provisioned Model Throughput %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProvisionedModelThroughputExists(ctx context.Context, n string, v *bedrock.GetProvisionedModelThroughputOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockClient(ctx)

		output, err := tfbedrock.FindProvisionedModelThroughputByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccProvisionedModelThroughputConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_bedrock_foundation_model" "test" {
  model_id = "amazon.titan-text-express-v1:0:8k"
}

resource "aws_bedrock_provisioned_model_throughput" "test" {
  provisioned_model_name = %[1]q
  model_arn              = data.aws_bedrock_foundation_model.test.model_arn
  commitment_duration    = "OneMonth"
  model_units            = 1
}
`, rName)
}
