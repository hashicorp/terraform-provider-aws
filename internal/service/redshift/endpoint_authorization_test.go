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

func TestAccRedshiftEndpointAuthorization_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrClusterIdentifier, "aws_redshift_cluster.test", names.AttrClusterIdentifier),
					resource.TestCheckResourceAttrPair(resourceName, "account", "data.aws_caller_identity.test", names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "grantee", "data.aws_caller_identity.test", names.AttrAccountID),
					acctest.CheckResourceAttrAccountID(resourceName, "grantor"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDelete},
			},
		},
	})
}

func TestAccRedshiftEndpointAuthorization_vpcs(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_vpcs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDelete},
			},
			{
				Config: testAccEndpointAuthorizationConfig_vpcsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", acctest.CtFalse),
				),
			},
			{
				Config: testAccEndpointAuthorizationConfig_vpcs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "vpc_ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "allowed_all_vpcs", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccRedshiftEndpointAuthorization_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceEndpointAuthorization(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftEndpointAuthorization_disappears_cluster(t *testing.T) {
	ctx := acctest.Context(t)
	var v redshift.EndpointAuthorization
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(18))
	resourceName := "aws_redshift_endpoint_authorization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckAlternateAccount(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesAlternate(ctx, t),
		CheckDestroy:             testAccCheckEndpointAuthorizationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointAuthorizationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointAuthorizationExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceCluster(), "aws_redshift_cluster.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEndpointAuthorizationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_endpoint_authorization" {
				continue
			}

			_, err := tfredshift.FindEndpointAuthorizationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift Endpoint Authorization %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckEndpointAuthorizationExists(ctx context.Context, n string, v *redshift.EndpointAuthorization) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Endpoint Authorization ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		output, err := tfredshift.FindEndpointAuthorizationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccEndpointAuthorizationConfigBase(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2),
		acctest.ConfigAlternateAccountProvider(),
		fmt.Sprintf(`
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

data "aws_caller_identity" "test" {
  provider = "awsalternate"
}
`, rName))
}

func testAccEndpointAuthorizationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAuthorizationConfigBase(rName), `
resource "aws_redshift_endpoint_authorization" "test" {
  account            = data.aws_caller_identity.test.account_id
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
}
`)
}

func testAccEndpointAuthorizationConfig_vpcs(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAuthorizationConfigBase(rName), fmt.Sprintf(`
resource "aws_vpc" "test2" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshift_endpoint_authorization" "test" {
  account            = data.aws_caller_identity.test.account_id
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  vpc_ids            = [aws_vpc.test2.id]
}
`, rName))
}

func testAccEndpointAuthorizationConfig_vpcsUpdated(rName string) string {
	return acctest.ConfigCompose(testAccEndpointAuthorizationConfigBase(rName), fmt.Sprintf(`
resource "aws_vpc" "test2" {
  provider = "awsalternate"

  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test3" {
  provider = "awsalternate"

  cidr_block = "11.0.0.0/16"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_redshift_endpoint_authorization" "test" {
  account            = data.aws_caller_identity.test.account_id
  cluster_identifier = aws_redshift_cluster.test.cluster_identifier
  vpc_ids            = [aws_vpc.test2.id, aws_vpc.test3.id]
}
`, rName))
}
