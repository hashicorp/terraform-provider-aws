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
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "sample_propmpts_control_mode", "DISABLED"),
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
				Config: testAccWebexperienceConfig_tags(rName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
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
				Config: testAccWebexperienceConfig_tags(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebexperienceExists(ctx, resourceName, &webex),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_webexperience" "test" {
  application_id               = aws_qbusiness_app.test.application_id
  sample_propmpts_control_mode = "DISABLED"
  title                        = %[1]q
}
`, rName))
}

func testAccWebexperienceConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_webexperience" "test" {
  application_id               = aws_qbusiness_app.test.application_id
  sample_propmpts_control_mode = "DISABLED"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
