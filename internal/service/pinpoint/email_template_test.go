// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointEmailTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpoint_email_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template pinpoint.GetEmailTemplateOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_resourceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_name", rName),
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

func TestAccPinpointEmailTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpoint_email_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template pinpoint.GetEmailTemplateOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_resourceBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEmailTemplateConfig_resourceBasic2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "description", "description"),
					resource.TestCheckResourceAttr(resourceName, "text", "text"),
				),
			},
			{
				Config: testAccEmailTemplateConfig_resourceBasic3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html update"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "description", "description update"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
		},
	})
}

func testAccCheckEmailTemplateExists(ctx context.Context, name string, template *pinpoint.GetEmailTemplateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Pinpoint, create.ErrActionCheckingExistence, tfpinpoint.ResNameEmailTemplate, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Pinpoint, create.ErrActionCheckingExistence, tfpinpoint.ResNameEmailTemplate, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointClient(ctx)

		out, err := tfpinpoint.FindEmailTemplateByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.Pinpoint, create.ErrActionCheckingExistence, tfpinpoint.ResNameEmailTemplate, rs.Primary.ID, err)
		}

		*template = *out

		return nil
	}
}

func testAccCheckEmailTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_email_template" {
				continue
			}

			_, err := tfpinpoint.FindEmailTemplateByName(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.NotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.Pinpoint, create.ErrActionCheckingDestroyed, tfpinpoint.ResNameEmailTemplate, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccEmailTemplateConfig_resourceBasic(name string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_email_template" "test" {
  template_name        = "%s"
  request {
    suject = "testing"
    header {
      name = "testingname"
      value = "testingvalue"
    }
  }
}
`, name)
}

func testAccEmailTemplateConfig_resourceBasic2(name string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_email_template" "test" {
  name        = "%s"
  subject     = "subject"
  html        = "html"
  text        = "text"
  description = "description"
}
`, name)
}

func testAccEmailTemplateConfig_resourceBasic3(name string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_email_template" "test" {
  name        = "%s"
  subject     = "subject"
  html        = "html update"
  description = "description update"
}
`, name)
}
