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

func TestAccQBusinessDataAccessor_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var dataaccessor qbusiness.GetDataAccessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_dataaccessor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDataAccessor(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataAccessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataAccessorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataAccessorExists(ctx, resourceName, &dataaccessor),
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

func TestAccQBusinessDataAccessor_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var dataaccessor qbusiness.GetDataAccessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_dataaccessor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDataAccessor(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataAccessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataAccessorConfig_tags(rName, acctest.CtKey1, acctest.CtValue1, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataAccessorExists(ctx, resourceName, &dataaccessor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataAccessorConfig_tags(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, "value2updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataAccessorExists(ctx, resourceName, &dataaccessor),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, "value2updated"),
				),
			},
		},
	})
}

func TestAccQBusinessDataAccessor_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var dataaccessor qbusiness.GetDataAccessorOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_qbusiness_dataaccessor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckDataAccessor(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, "qbusiness"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataAccessorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataAccessorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataAccessorExists(ctx, resourceName, &dataaccessor),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfqbusiness.ResourceDataAccessor, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreCheckDataAccessor(ctx context.Context, t *testing.T) {
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

func testAccCheckDataAccessorDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_qbusiness_dataaccessor" {
				continue
			}

			_, err := tfqbusiness.FindDataAccessorByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Amazon Q Dataaccessor %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDataAccessorExists(ctx context.Context, n string, v *qbusiness.GetDataAccessorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QBusinessClient(ctx)

		output, err := tfqbusiness.FindDataAccessorByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDataAccessorConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_dataaccessor" "test" {
  application_id = aws_qbusiness_app.test.id
  display_name   = %[1]q
  principal      = "arn:aws:iam::359246571101:role/zoom-ai-companion"

  action_configuration {
    action = "qbusiness:SearchRelevantContent"
  }

}
`, rName))
}

func testAccDataAccessorConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccAppConfig_basic(rName), fmt.Sprintf(`
resource "aws_qbusiness_dataaccessor" "test" {
  application_id = aws_qbusiness_app.test.id
  display_name   = %[1]q
  principal      = "arn:aws:iam::359246571101:role/zoom-ai-companion"

  action_configuration {
    action = "qbusiness:SearchRelevantContent"
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
