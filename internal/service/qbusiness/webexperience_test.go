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
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQBusinessWebexperience_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var webex qbusiness.GetWebExperienceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_webexperience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWebexperience(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebexperienceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebexperienceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttrSet(resourceName, "webexperience_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
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

func TestAccQBusinessWebexperience_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var webex qbusiness.GetWebExperienceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_webexperience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWebexperience(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebexperienceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebexperienceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceWebexperience, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQBusinessWebexperience_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var webex qbusiness.GetWebExperienceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_webexperience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWebexperience(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebexperienceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebexperienceConfig_tags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWebexperienceConfig_tags(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, "value2updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, "value2updated"),
				),
			},
		},
	})
}

func TestAccQBusinessWebexperience_title(t *testing.T) {
	ctx := acctest.Context(t)
	var webex qbusiness.GetWebExperienceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_webexperience.test"
	title1 := "title1"
	title2 := "title2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWebexperience(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebexperienceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebexperienceConfig_title(rName, title1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttr(resourceName, "title", title1),
					resource.TestCheckResourceAttr(resourceName, "subtitle", title1),
					resource.TestCheckResourceAttr(resourceName, "welcome_message", title1),
				),
			},
			{
				Config: testAccWebexperienceConfig_title(rName, title2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttr(resourceName, "title", title2),
					resource.TestCheckResourceAttr(resourceName, "subtitle", title2),
					resource.TestCheckResourceAttr(resourceName, "welcome_message", title2),
				),
			},
		},
	})
}

func TestAccQBusinessWebexperience_samplePromptsControlMode(t *testing.T) {
	ctx := acctest.Context(t)
	var webex qbusiness.GetWebExperienceOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_webexperience.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWebexperience(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebexperienceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebexperienceConfig_samplePromptsControlMode(rName, "ENABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttr(resourceName, "sample_prompts_control_mode", "ENABLED"),
				),
			},
			{
				Config: testAccWebexperienceConfig_samplePromptsControlMode(rName, "DISABLED"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttr(resourceName, "sample_prompts_control_mode", "DISABLED"),
				),
			},
		},
	})
}

func testAccPreCheckWebexperience(ctx context.Context, t *testing.T) {
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

func testAccCheckWebexperienceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_webexperience" {
				continue
			}

			_, err := tfqbusiness.FindWebexperienceByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("amazon q webexperience %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWebexperienceExists(ctx context.Context, n string, v *qbusiness.GetWebExperienceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		output, err := tfqbusiness.FindWebexperienceByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWebexperienceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), `
resource "aws_qbusiness_webexperience" "test" {
  application_id              = aws_qbusiness_app.test.id
  sample_prompts_control_mode = "DISABLED"
}
`)
}

func testAccWebexperienceConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_webexperience" "test" {
  application_id              = aws_qbusiness_app.test.id
  sample_prompts_control_mode = "DISABLED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccWebexperienceConfig_title(rName, title string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_webexperience" "test" {
  application_id              = aws_qbusiness_app.test.id
  sample_prompts_control_mode = "DISABLED"
  title                       = %[1]q
  subtitle                    = %[1]q
  welcome_message             = %[1]q
}
`, title))
}

func testAccWebexperienceConfig_samplePromptsControlMode(rName, control string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_webexperience" "test" {
  application_id              = aws_qbusiness_app.test.id
  sample_prompts_control_mode = %[1]q
}
`, control))
}
