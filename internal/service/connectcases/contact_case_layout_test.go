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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccContactCaseLayout_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_connectcases_contact_case_layout.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccContactCaseLayoutDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactCaseLayout_base(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccContactCaseLayoutExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "layout_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
				),
			},
		},
	})
}

func TestAccContactCaseLayout_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_connectcases_contact_case_layout.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccContactCaseLayoutDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactCaseLayout_base(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccContactCaseLayoutExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, connectcases.ResourceContactCaseLayout(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccContactCaseLayoutExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Connect Cases Contact Case Layout ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectCasesClient(ctx)

		//Have to fix the Domain ID here
		_, err := connectcases.FindConnectCasesLayoutById(ctx, conn, rs.Primary.ID, "")

		return err
	}
}

func testAccContactCaseLayoutDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectCasesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connectcases_contact_case_layout" {
				continue
			}

			//Have to fix the Domain ID here
			_, err := connectcases.FindConnectCasesLayoutById(ctx, conn, rs.Primary.ID, "")

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Cases Contace Case Layout %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccContactCaseLayout_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connectcases_contact_case_layout" "test" {
  name = %[1]q

  content {
	more_info {
		sections {
			name = "more_info_example"
			field_group {
				fields {
					id = "more_info_example"
				}
			}
		}
	}
	top_panel {
		sections {
			name = "more_info_example"
			field_group {
				fields {
					id = "more_info_example"
				}
			}
		}
	}
  }
}
`, rName)
}
