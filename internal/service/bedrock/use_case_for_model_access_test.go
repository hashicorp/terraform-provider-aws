// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockUseCaseForModelAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var usecaseformodelaccess bedrock.GetUseCaseForModelAccessOutput
	resourceName := "aws_bedrock_use_case_for_model_access.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUseCaseForModelAccessConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUseCaseForModelAccessExists(ctx, t, resourceName, &usecaseformodelaccess),
					resource.TestCheckResourceAttrSet(resourceName, "form_data"),
				),
			},
		},
	})
}

func testAccCheckUseCaseForModelAccessExists(ctx context.Context, t *testing.T, name string, usecaseformodelaccess *bedrock.GetUseCaseForModelAccessOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameUseCaseForModelAccess, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameUseCaseForModelAccess, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		input := bedrock.GetUseCaseForModelAccessInput{}
		resp, err := conn.GetUseCaseForModelAccess(ctx, &input)
		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameUseCaseForModelAccess, rs.Primary.ID, err)
		}

		*usecaseformodelaccess = *resp

		return nil
	}
}

func testAccUseCaseForModelAccessConfig_basic() string {
	return fmt.Sprintf(`
resource "aws_bedrock_use_case_for_model_access" "test" {
  form_data = jsonencode({
    "companyName"         = "AWS Provider",
    "companyWebsite"      = "https://www.test.com",
    "intendedUsers"       = "0",
    "industryOption"      = "Energy",
    "otherIndustryOption" = "",
    "useCases"            = ". - Generating developer documentation\n- Code generation/refactoring\n- Summarization of issues / documents"
  })
}
`)
}
