// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/redshift"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftEndpointAccess_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAccess
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAccessConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "endpoint_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "5439"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint.#", acctest.Ct1),
					acctest.CheckResourceAttrAccountID(resourceName, acctest.CtResourceOwner),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_group_name", "aws_redshift_subnet_group.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrClusterIdentifier),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrAddress),
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

func TestAccRedshiftEndpointAccess_sgs(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAccess
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAccessConfig_sgs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEndpointAccessConfig_sgsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccRedshiftEndpointAccess_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAccess
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAccessConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceEndpointAccess(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftEndpointAccess_disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAccess
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_access.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEndpointAccessDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAccessConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAccessExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), "aws_redshift_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointAccessDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_endpoint_access" {
				continue
			}

			_, err := tfredshift.FindEndpointAccessByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Endpoint Access %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEndpointAccessExists(ctx context.Context, n string, v *redshift.EndpointAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Endpoint Access ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		output, err := tfredshift.FindEndpointAccessByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEndpointAccessConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_redshift_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier                   = %[1]q
  availability_zone                    = data.aws_availability_zones.available.names[0]
  database_name                        = "mydb"
  master_username                      = "foo_test"
  master_password                      = "Mustbe8characters"
  node_type                            = "ra3.xlplus"
  automated_snapshot_retention_period  = 1
  allow_version_upgrade                = false
  skip_final_snapshot                  = true
  availability_zone_relocation_enabled = true
  publicly_accessible                  = false
}
`, rName))
}

func testAccEndpointAccessConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAccessConfigBase(rName), fmt.Sprintf(`
resource "aws_redshift_endpoint_access" "test" {
  endpoint_name      = %[1]q
  subnet_group_name  = aws_redshift_subnet_group.test.id
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}
`, rName))
}

func testAccEndpointAccessConfig_sgs(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAccessConfigBase(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_redshift_endpoint_access" "test" {
  endpoint_name          = %[1]q
  subnet_group_name      = aws_redshift_subnet_group.test.id
  cluster_identifier     = aws_redshift_cluster.test.cluster_identifier
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName))
}

func testAccEndpointAccessConfig_sgsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAccessConfigBase(rName), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id
}

resource "aws_redshift_endpoint_access" "test" {
  endpoint_name          = %[1]q
  subnet_group_name      = aws_redshift_subnet_group.test.id
  cluster_identifier     = aws_redshift_cluster.test.cluster_identifier
  vpc_security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
}
`, rName))
}
