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

func TestAccField_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_connectcases_field.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		// Connect Cases Fields cannot be deleted once applied, ensure parent resource Connect Cases Domain is destroyed instead.
		CheckDestroy: testAccDomainDestroy(ctx), //or acctest.CheckDestroyNoop T.B.D.
		Steps: []resource.TestStep{
			{
				Config: testAccField_base(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccFieldExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "namespace"),
					resource.TestCheckResourceAttr(resourceName, "type", "Text"),
					resource.TestCheckResourceAttrSet(resourceName, "field_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "field_id"),
				),
			},
		},
	})
}

func testAccFieldExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Connect Cases Field ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectCasesClient(ctx)

		_, err := connectcases.FindFieldByDomainAndID(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["domain_id"])

		return err
	}
}

func testAccField_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connectcases_domain" "test" {
  name = %[1]q
}

resource "aws_connectcases_field" "test" {
  name        = %[1]q
  description = "example description of field"
  domain_id   = aws_connectcases_domain.test.domain_id
  type        = "Text"
}
`, rName)
}
