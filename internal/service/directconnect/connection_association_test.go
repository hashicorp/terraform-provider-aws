// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDirectConnectConnectionAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dx_connection_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAssociationExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccDirectConnectConnectionAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dx_connection_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfdirectconnect.ResourceConnectionAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDirectConnectConnectionAssociation_lagOnConnection(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_dx_connection_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAssociationConfig_lagOn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAssociationExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func TestAccDirectConnectConnectionAssociation_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_dx_connection_association.test1"
	resourceName2 := "aws_dx_connection_association.test2"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.DirectConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConnectionAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccConnectionAssociationConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConnectionAssociationExists(ctx, t, resourceName1),
					testAccCheckConnectionAssociationExists(ctx, t, resourceName2),
				),
			},
		},
	})
}

func testAccCheckConnectionAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_dx_connection_association" {
				continue
			}

			err := tfdirectconnect.FindConnectionLAGAssociation(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["lag_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Direct Connect Connection (%s) LAG (%s) Association still exists", rs.Primary.ID, rs.Primary.Attributes["lag_id"])
		}

		return nil
	}
}

func testAccCheckConnectionAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).DirectConnectClient(ctx)

		return tfdirectconnect.FindConnectionLAGAssociation(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["lag_id"])
	}
}

func testAccConnectionAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[1]
}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = local.location_code
  force_destroy         = true
}

resource "aws_dx_connection_association" "test" {
  connection_id = aws_dx_connection.test.id
  lag_id        = aws_dx_lag.test.id
}
`, rName)
}

func testAccConnectionAssociationConfig_lagOn(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[1]
}

resource "aws_dx_connection" "test1" {
  name      = "%[1]s-1"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_connection" "test2" {
  name      = "%[1]s-2"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connection_id         = aws_dx_connection.test1.id
  connections_bandwidth = "1Gbps"
  location              = local.location_code
}

resource "aws_dx_connection_association" "test" {
  connection_id = aws_dx_connection.test2.id
  lag_id        = aws_dx_lag.test.id
}
`, rName)
}

func testAccConnectionAssociationConfig_multiple(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

locals {
  location_code = tolist(data.aws_dx_locations.test.location_codes)[1]
}

resource "aws_dx_connection" "test1" {
  name      = "%[1]s-1"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_connection" "test2" {
  name      = "%[1]s-2"
  bandwidth = "1Gbps"
  location  = local.location_code
}

resource "aws_dx_lag" "test" {
  name                  = %[1]q
  connections_bandwidth = "1Gbps"
  location              = local.location_code
  force_destroy         = true
}

resource "aws_dx_connection_association" "test1" {
  connection_id = aws_dx_connection.test1.id
  lag_id        = aws_dx_lag.test.id
}

resource "aws_dx_connection_association" "test2" {
  connection_id = aws_dx_connection.test2.id
  lag_id        = aws_dx_lag.test.id
}
`, rName)
}
