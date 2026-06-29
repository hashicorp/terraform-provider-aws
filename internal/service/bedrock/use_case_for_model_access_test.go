// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccBedrockUseCaseForModelAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_bedrock_use_case_for_model_access.test"
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFoundationModelUseCaseAlreadyExists(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccUseCaseForModelAccessConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUseCaseForModelAccessExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "form_data"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
				ImportStateVerifyIgnore:              []string{"form_data"},
			},
		},
	})
}

func testAccBedrockUseCaseForModelAccess_createImport(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_bedrock_use_case_for_model_access.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckFoundationModelUseCase(ctx, t)
		},

		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,

		Steps: []resource.TestStep{
			{
				Config: testAccUseCaseForModelAccessConfig_createImport(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUseCaseForModelAccessExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "form_data"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAccountID),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAccountID,
				ImportStateVerifyIgnore:              []string{"form_data"},
			},
		},
	})
}

func testAccCheckUseCaseForModelAccessExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
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
		_, err := conn.GetUseCaseForModelAccess(ctx, &input)
		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameUseCaseForModelAccess, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccUseCaseForModelAccessConfig_basic() string {
	return `
resource "aws_bedrock_use_case_for_model_access" "test" {
  form_data = jsonencode({
    "companyName"         = "AWS Provider",
    "companyWebsite"      = "https://www.bedrock.test",
    "intendedUsers"       = "0",
    "industryOption"      = "Energy",
    "otherIndustryOption" = "",
    "useCases"            = ". - Testing"
  })
}
`
}

func testAccUseCaseForModelAccessConfig_createImport() string {
	return `
resource "aws_bedrock_use_case_for_model_access" "test" {
  form_data = data.aws_bedrock_use_case_for_model_access.test.form_data
}

data "aws_bedrock_use_case_for_model_access" "test" {
}
`
}

func testAccPreCheckFoundationModelUseCaseAlreadyExists(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

	input := bedrock.GetUseCaseForModelAccessInput{}
	_, err := conn.GetUseCaseForModelAccess(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err == nil {
		t.Skipf("skipping acceptance testing due to already existing use case for model access resource in the account")
	}

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		// expected error, continue with the test
		return
	}

	t.Fatalf("unexpected PreCheck error: %s", err)
}
