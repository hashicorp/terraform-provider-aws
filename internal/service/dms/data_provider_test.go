// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdms "github.com/hashicorp/terraform-provider-aws/internal/service/dms"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDMSDataProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_data_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDataProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "data_provider_arn"),
					resource.TestCheckResourceAttr(resourceName, "data_provider_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "postgres"),
					resource.TestCheckResourceAttr(resourceName, "settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgres_settings.#", "1"),
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

func TestAccDMSDataProvider_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_data_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgres_settings.0.port", "5432"),
				),
			},
			{
				Config: testAccDataProviderConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "settings.0.postgres_settings.0.port", "5433"),
				),
			},
		},
	})
}

func TestAccDMSDataProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dms_data_provider.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DMSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDataProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDataProviderConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataProviderExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdms.ResourceDataProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataProviderExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSClient(ctx)

		_, err := tfdms.FindDataProviderByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckDataProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DMSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dms_data_provider" {
				continue
			}

			_, err := tfdms.FindDataProviderByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DMS Data Provider %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDataProviderConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  data_provider_name = %[1]q
  engine             = "postgres"

  settings {
    postgres_settings {
      server_name   = "%[1]s.example.com"
      port          = 5432
      database_name = "testdb"
      ssl_mode      = "none"
    }
  }
}
`, rName)
}

func testAccDataProviderConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_data_provider" "test" {
  data_provider_name = %[1]q
  engine             = "postgres"

  settings {
    postgres_settings {
      server_name   = "%[1]s.example.com"
      port          = 5433
      database_name = "testdb"
      ssl_mode      = "none"
    }
  }
}
`, rName)
}
