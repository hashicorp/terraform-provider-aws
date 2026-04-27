// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfpinpoint "github.com/hashicorp/terraform-provider-aws/internal/service/pinpoint"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPinpointEmailTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpoint_email_template.test"
	var template pinpoint.GetEmailTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.#"),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.0.%"),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.0.subject"),
					resource.TestCheckResourceAttr(resourceName, "email_template.0.subject", "testing"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "template_name"),
				ImportStateVerifyIdentifierAttribute: "template_name",
				ImportStateVerify:                    true,
			},
		},
	})
}

func TestAccPinpointEmailTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpoint_email_template.test"
	var template pinpoint.GetEmailTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.#"),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.0.%"),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.0.subject"),
					resource.TestCheckResourceAttr(resourceName, "email_template.0.subject", "testing"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "template_name"),
				ImportStateVerifyIdentifierAttribute: "template_name",
			},
			{
				Config: testAccEmailTemplateConfig_update(rName, "update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.#"),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.0.%"),
					resource.TestCheckResourceAttrSet(resourceName, "email_template.0.subject"),
					resource.TestCheckResourceAttr(resourceName, "email_template.0.subject", "update"),
				),
			},
		},
	})
}

func TestAccPinpointEmailTemplate_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpoint_email_template.test"
	var template pinpoint.GetEmailTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.PinpointServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEmailTemplateExists(ctx, t, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "template_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "template_name"),
				ImportStateVerifyIdentifierAttribute: "template_name",
			},
		},
	})
}

func testAccCheckEmailTemplateExists(ctx context.Context, t *testing.T, name string, template *pinpoint.GetEmailTemplateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Pinpoint, create.ErrActionCheckingExistence, tfpinpoint.ResNameEmailTemplate, name, errors.New("not found"))
		}

		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		out, err := tfpinpoint.FindEmailTemplateByName(ctx, conn, rs.Primary.Attributes["template_name"])
		if err != nil {
			return create.Error(names.Pinpoint, create.ErrActionCheckingExistence, tfpinpoint.ResNameEmailTemplate, rs.Primary.Attributes["template_name"], err)
		}

		*template = *out

		return nil
	}
}

func testAccCheckEmailTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).PinpointClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_email_template" {
				continue
			}

			_, err := tfpinpoint.FindEmailTemplateByName(ctx, conn, rs.Primary.Attributes["template_name"])
			if errs.IsA[*types.NotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.Pinpoint, create.ErrActionCheckingDestroyed, tfpinpoint.ResNameEmailTemplate, rs.Primary.Attributes["template_name"], errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccEmailTemplateConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_email_template" "test" {
  template_name = %[1]q
  email_template {
    subject   = "testing"
    text_part = "we are testing template text part"
    header {
      name  = "testingname"
      value = "testingvalue"
    }
  }
}
`, rName)
}

func testAccEmailTemplateConfig_update(name, subject string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_email_template" "test" {
  template_name = %[1]q
  email_template {
    subject   = %[2]q
    text_part = "we are testing template text part"
    header {
      name  = "testingname"
      value = "testingvalue"
    }
  }
}
`, name, subject)
}

func testAccEmailTemplateConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_email_template" "test" {
  template_name = %[1]q
  email_template {
    subject   = "testing"
    text_part = "we are testing template text part"
    header {
      name  = "testingname"
      value = "testingvalue"
    }
  }
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}
