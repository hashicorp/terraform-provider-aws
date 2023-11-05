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

func TestAccContactCaseDomain_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_connectcases_contact_case_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccContactCaseDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactCaseDomain_base(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccContactCaseDomainExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "domain_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_id"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_status"),
				),
			},
		},
	})
}

func TestAccContactCaseDomain_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_connectcases_contact_case_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccContactCaseDomainDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccContactCaseDomain_base(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccContactCaseDomainExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, connectcases.ResourceContactCaseDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccContactCaseDomainExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Connect Cases Contact Case Domain ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectCasesClient(ctx)

		_, err := connectcases.FindConnectCasesDomainById(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccContactCaseDomainDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectCasesClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connectcases_contact_case_domain" {
				continue
			}

			_, err := connectcases.FindConnectCasesDomainById(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Connect Cases Contace Case Domain %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccContactCaseDomain_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connectcases_contact_case_domain" "test" {
  name = %[1]q
}
`, rName)
}
