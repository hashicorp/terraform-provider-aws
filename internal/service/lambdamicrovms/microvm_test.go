// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms_test

import (
	"context"
	"errors"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambdamicrovms "github.com/hashicorp/terraform-provider-aws/internal/service/lambdamicrovms"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaMicrovmsMicrovm_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambdamicrovms.GetMicrovmOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdamicrovms_microvm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicrovmsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMicrovmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMicrovmConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMicrovmExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.MicrovmStateRunning)),
					resource.TestCheckResourceAttrSet(resourceName, "microvm_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEndpoint),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, "image_arn", "lambda", regexache.MustCompile(`microvm-image:.+$`)),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "microvm_id"),
				ImportStateVerifyIdentifierAttribute: "microvm_id",
				ImportStateVerifyIgnore: []string{
					names.AttrExecutionRoleARN,
					"image_identifier",
					"logging",
					"run_hook_payload",
				},
			},
		},
	})
}

func TestAccLambdaMicrovmsMicrovm_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambdamicrovms.GetMicrovmOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdamicrovms_microvm.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicrovmsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMicrovmDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMicrovmConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMicrovmExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdamicrovms.ResourceMicrovm, resourceName),
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

func testAccCheckMicrovmDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaMicrovmsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdamicrovms_microvm" {
				continue
			}

			_, err := tflambdamicrovms.FindMicrovmByID(ctx, conn, rs.Primary.Attributes["microvm_id"])
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.LambdaMicrovms, create.ErrActionCheckingDestroyed, tflambdamicrovms.ResNameMicrovm, rs.Primary.Attributes["microvm_id"], err)
			}

			return create.Error(names.LambdaMicrovms, create.ErrActionCheckingDestroyed, tflambdamicrovms.ResNameMicrovm, rs.Primary.Attributes["microvm_id"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMicrovmExists(ctx context.Context, t *testing.T, name string, v *lambdamicrovms.GetMicrovmOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LambdaMicrovms, create.ErrActionCheckingExistence, tflambdamicrovms.ResNameMicrovm, name, errors.New("not found"))
		}

		id := rs.Primary.Attributes["microvm_id"]
		if id == "" {
			return create.Error(names.LambdaMicrovms, create.ErrActionCheckingExistence, tflambdamicrovms.ResNameMicrovm, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaMicrovmsClient(ctx)

		out, err := tflambdamicrovms.FindMicrovmByID(ctx, conn, id)
		if err != nil {
			return create.Error(names.LambdaMicrovms, create.ErrActionCheckingExistence, tflambdamicrovms.ResNameMicrovm, id, err)
		}
		*v = *out

		return nil
	}
}

func testAccMicrovmConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccImageConfig_basic(rName), `
resource "aws_lambdamicrovms_microvm" "test" {
  image_identifier = aws_lambdamicrovms_image.test.arn
}
`)
}
