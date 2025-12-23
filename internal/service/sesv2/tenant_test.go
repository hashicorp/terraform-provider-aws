// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfsesv2 "github.com/hashicorp/terraform-provider-aws/internal/service/sesv2"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSESV2Tenant_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName1 := acctest.RandomWithPrefix(t, "tf-acc-test")
	rName2 := acctest.RandomWithPrefix(t, "tf-acc-test-new")
	resourceName := "aws_sesv2_tenant.test"

	var tenantID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, tfsesv2.ResNameTenant),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTenantDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTenantConfig_basic(rName1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tenant_name", rName1),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources[resourceName]
						tenantID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccTenantConfig_basic(rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTenantExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "tenant_name", rName2),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					testAccCheckTenantRecreated(resourceName, &tenantID),
					testAccCheckTenantDoesNotExist(ctx, rName1),
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

func testAccCheckTenantDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sesv2_tenant" {
				continue
			}

			tenantName := rs.Primary.Attributes["tenant_name"]
			_, err := tfsesv2.FindTenantByName(ctx, conn, tenantName)

			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameTenant, tenantName, err)
			}

			return create.Error(names.SESV2, create.ErrActionCheckingDestroyed, tfsesv2.ResNameTenant, tenantName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckTenantExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameTenant, name, errors.New("not found"))
		}

		tenantName, ok := rs.Primary.Attributes["tenant_name"]
		if !ok || tenantName == "" {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameTenant, name, errors.New("tenant_name attribute not found or empty"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindTenantByName(ctx, conn, tenantName)
		if err != nil {
			return create.Error(names.SESV2, create.ErrActionCheckingExistence, tfsesv2.ResNameTenant, tenantName, err)
		}

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

	input := &sesv2.ListTenantsInput{}

	_, err := conn.ListTenants(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckTenantRecreated(name string, oldID *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("resource not found: %s", name)
		}

		if rs.Primary.ID == *oldID {
			return fmt.Errorf("tenant was not recreated (ID did not change)")
		}

		*oldID = rs.Primary.ID
		return nil
	}
}

func testAccCheckTenantDoesNotExist(ctx context.Context, tenantName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Client(ctx)

		_, err := tfsesv2.FindTenantByName(ctx, conn, tenantName)
		if retry.NotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("tenant %q still exists", tenantName)
	}
}

func testAccTenantConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sesv2_tenant" "test" {
  tenant_name             = %[1]q
	tags = {
		"testkey" = "testvalue"
	}
}
`, rName)
}
