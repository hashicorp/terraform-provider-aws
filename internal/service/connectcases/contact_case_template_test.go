// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connectcases_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/service/connectcases"
)

func TestAccTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_connectcases_template.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameField := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameField2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// Connect Cases Templates cannot be deleted once applied, ensure parent resource Connect Cases Domain is destroyed instead.
		CheckDestroy: testAccDomainDestroy(ctx), //or acctest.CheckDestroyNoop T.B.D.
		Steps: []resource.TestStep{
			{
				Config: testAccTemplate_base(rName, rNameField, rNameField2),
				Check: resource.ComposeTestCheckFunc(
					testAccTemplateExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "template_arn"),
				),
			},
		},
	})
}

func testAccTemplateExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Connect Case Template ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectCasesClient(ctx)

		_, err := connectcases.FindTemplateByDomainAndId(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["domain_id"])

		return err
	}
}

func testAccTemplate_base(rName, rNameField, rNameField2 string) string {
	return fmt.Sprintf(`
resource "aws_connectcases_domain" "test" {
  name = %[1]q
}

resource "aws_connectcases_field" "test" {
  name        = %[2]q
  description = "example description of field"
  domain_id   = aws_connectcases_domain.test.domain_id
  type        = "Text"
}

resource "aws_connectcases_field" "test2" {
  name        = %[3]q
  description = "example description of field"
  domain_id   = aws_connectcases_domain.test.domain_id
  type        = "Text"
}

resource "aws_connectcases_template" "test" {
  name        = %[3]q
  description = "example description of template"
  domain_id   = aws_connectcases_domain.test.domain_id
  status      = "Inactive"

  layout_configuration {
    default_layout = aws_connectcases_layout.test.id
  }

  required_fields {
    field_id = aws_connectcases_field.test.field_id
  }
}

resource "aws_connectcases_layout" "test" {
  name      = %[1]q
  domain_id = aws_connectcases_domain.test.domain_id

  content {
    more_info {
      sections {
        name = "more_info_example"
        field_group {
          fields {
            id = aws_connectcases_field.test.field_id
          }
        }
      }
    }
    top_panel {
      sections {
        name = "top_panel_example"
        field_group {
          fields {
            id = aws_connectcases_field.test2.field_id
          }
        }
      }
    }
  }
}
`, rName, rNameField, rNameField2)
}
