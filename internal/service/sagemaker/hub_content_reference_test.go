// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerHubContentReference_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var ref sagemaker.DescribeHubContentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub_content_reference.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubContentReferenceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHubContentReferenceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, &ref),
					resource.TestCheckResourceAttr(resourceName, "hub_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_content_name", rName),
					resource.TestCheckResourceAttr(resourceName, "hub_content_status", string(awstypes.HubContentStatusAvailable)),
					resource.TestCheckResourceAttrSet(resourceName, "hub_content_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "hub_content_version"),
					resource.TestCheckResourceAttrSet(resourceName, "hub_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "sagemaker_public_hub_content_arn"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccHubContentReferenceImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "hub_name",
			},
		},
	})
}

func TestAccSageMakerHubContentReference_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub_content_reference.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubContentReferenceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHubContentReferenceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, nil),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfsagemaker.ResourceHubContentReference, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccSageMakerHubContentReference_minVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var ref sagemaker.DescribeHubContentOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub_content_reference.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubContentReferenceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHubContentReferenceConfig_minVersion(rName, "1.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, &ref),
					resource.TestCheckResourceAttr(resourceName, "min_version", "1.0.0"),
				),
			},
			{
				Config: testAccHubContentReferenceConfig_minVersion(rName, "2.0.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, &ref),
					resource.TestCheckResourceAttr(resourceName, "min_version", "2.0.0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccHubContentReferenceImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "hub_name",
			},
			{
				Config: testAccHubContentReferenceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, &ref),
					resource.TestCheckNoResourceAttr(resourceName, "min_version"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccSageMakerHubContentReference_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_hub_content_reference.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubContentReferenceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHubContentReferenceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccHubContentReferenceImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "hub_name",
			},
			{
				Config: testAccHubContentReferenceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccHubContentReferenceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHubContentReferenceExists(ctx, t, resourceName, nil),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testAccCheckHubContentReferenceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_hub_content_reference" {
				continue
			}

			hubName := rs.Primary.Attributes["hub_name"]
			hubContentName := rs.Primary.Attributes["hub_content_name"]

			_, err := tfsagemaker.FindHubContentByName(ctx, conn, hubName, hubContentName, awstypes.HubContentTypeModelReference)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameHubContentReference, hubName, err)
			}

			return create.Error(names.SageMaker, create.ErrActionCheckingDestroyed, tfsagemaker.ResNameHubContentReference, hubName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckHubContentReferenceExists(ctx context.Context, t *testing.T, n string, ref *sagemaker.DescribeHubContentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHubContentReference, n, errors.New("not found"))
		}

		hubName := rs.Primary.Attributes["hub_name"]
		hubContentName := rs.Primary.Attributes["hub_content_name"]

		conn := acctest.ProviderMeta(ctx, t).SageMakerClient(ctx)

		output, err := tfsagemaker.FindHubContentByName(ctx, conn, hubName, hubContentName, awstypes.HubContentTypeModelReference)
		if err != nil {
			return create.Error(names.SageMaker, create.ErrActionCheckingExistence, tfsagemaker.ResNameHubContentReference, hubName, err)
		}

		if ref != nil {
			*ref = *output
		}

		return nil
	}
}

func testAccHubContentReferenceImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "hub_name", "hub_content_name")
}

func testAccHubContentReferenceConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_sagemaker_hub" "test" {
  hub_name        = %[1]q
  hub_description = %[1]q
}
`, rName)
}

func testAccHubContentReferenceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccHubContentReferenceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hub_content_reference" "test" {
  hub_name                         = aws_sagemaker_hub.test.hub_name
  hub_content_name                 = %[1]q
  sagemaker_public_hub_content_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"
}
`, rName))
}

func testAccHubContentReferenceConfig_minVersion(rName, minVersion string) string {
	return acctest.ConfigCompose(testAccHubContentReferenceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hub_content_reference" "test" {
  hub_name                         = aws_sagemaker_hub.test.hub_name
  hub_content_name                 = %[1]q
  sagemaker_public_hub_content_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"
  min_version                      = %[2]q
}
`, rName, minVersion))
}

func testAccHubContentReferenceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccHubContentReferenceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hub_content_reference" "test" {
  hub_name                         = aws_sagemaker_hub.test.hub_name
  hub_content_name                 = %[1]q
  sagemaker_public_hub_content_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccHubContentReferenceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccHubContentReferenceConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_hub_content_reference" "test" {
  hub_name                         = aws_sagemaker_hub.test.hub_name
  hub_content_name                 = %[1]q
  sagemaker_public_hub_content_arn = "arn:${data.aws_partition.current.partition}:sagemaker:${data.aws_region.current.name}:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func TestStripARNVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arn  string
		want string
	}{
		{
			name: "ARN with version suffix is stripped",
			arn:  "arn:aws:sagemaker:us-east-1:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct/1.0.0", //lintignore:AWSAT003,AWSAT005
			want: "arn:aws:sagemaker:us-east-1:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct",       //lintignore:AWSAT003,AWSAT005
		},
		{
			name: "ARN without version suffix is returned unchanged",
			arn:  "arn:aws:sagemaker:us-east-1:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct", //lintignore:AWSAT003,AWSAT005
			want: "arn:aws:sagemaker:us-east-1:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-3-1-8b-instruct", //lintignore:AWSAT003,AWSAT005
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tfsagemaker.StripARNVersion(&tt.arn); *got != tt.want {
				t.Errorf("StripARNVersion(%q) = %q, want %q", tt.arn, *got, tt.want)
			}
		})
	}
}
