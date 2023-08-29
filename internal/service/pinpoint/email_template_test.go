// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccPinpointEmailTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpoint_email_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template pinpoint.EmailTemplateResponse

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_resourceBasic1(rName),
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
		},
	})
}

func TestAccPinpointEmailTemplate_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_pinpoint_email_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var template pinpoint.EmailTemplateResponse

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, pinpoint.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEmailTemplateDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig_resourceBasic1(rName),
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

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

	input := &pinpoint.ListTemplatesInput{}

	_, err := conn.ListTemplatesWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %v", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckEmailTemplateExists(ctx context.Context, pr string, template *pinpoint.EmailTemplateResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Email Template ID is set")
		}

		input := &pinpoint.GetEmailTemplateInput{
			TemplateName: aws.String(rs.Primary.ID),
		}

		templateOutput, err := conn.GetEmailTemplateWithContext(ctx, input)
		if err != nil {
			return err
		}

		if templateOutput == nil || templateOutput.EmailTemplateResponse == nil {
			return fmt.Errorf("Email Template (%s) not found", rs.Primary.ID)
		}

		*template = *templateOutput.EmailTemplateResponse

		return nil
	}
}

func testAccCheckEmailTemplateDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).PinpointConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_pinpoint_email_template" {
				continue
			}

			err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
				input := pinpoint.GetEmailTemplateInput{
					TemplateName: aws.String(rs.Primary.ID),
				}

				gto, err := conn.GetEmailTemplateWithContext(ctx, &input)
				if err != nil {
					if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
						return nil
					}
					return retry.NonRetryableError(err)
				}
				if gto != nil {
					return retry.RetryableError(fmt.Errorf("Template exists: %s", rs.Primary.ID))
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccEmailTemplateConfig_resourceBasic1(name string) string {
	return fmt.Sprintf(`
resource "aws_pinpoint_email_template" "test" {
  name        = "%s"
  subject     = "subject"
  html        = "html"
  description = "description"
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
