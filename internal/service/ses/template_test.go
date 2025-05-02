// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ses_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfses "github.com/hashicorp/terraform-provider-aws/internal/service/ses"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ses_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template awstypes.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_resourceBasic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
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

func TestAccSESTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ses_template.test"
	var template awstypes.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_resourceBasic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "ses", fmt.Sprintf("template/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTemplateConfig_resourceBasic2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", "text"),
				),
			},
			{
				Config: testAccTemplateConfig_resourceBasic3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "html", "html update"),
					resource.TestCheckResourceAttr(resourceName, "subject", "subject"),
					resource.TestCheckResourceAttr(resourceName, "text", ""),
				),
			},
		},
	})
}

func TestAccSESTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ses_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template awstypes.Template

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SESServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_resourceBasic1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTemplateExists(ctx, resourceName, &template),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfses.ResourceTemplate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckTemplateExists(ctx context.Context, n string, v *awstypes.Template) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		output, err := tfses.FindTemplateByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ses_template" {
				continue
			}

			_, err := tfses.FindTemplateByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SES Template %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTemplateConfig_resourceBasic1(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = %[1]q
  subject = "subject"
  html    = "html"
}
`, name)
}

func testAccTemplateConfig_resourceBasic2(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = %[1]q
  subject = "subject"
  html    = "html"
  text    = "text"
}
`, name)
}

func testAccTemplateConfig_resourceBasic3(name string) string {
	return fmt.Sprintf(`
resource "aws_ses_template" "test" {
  name    = %[1]q
  subject = "subject"
  html    = "html update"
}
`, name)
}
