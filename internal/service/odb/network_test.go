// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type oracleDBNetworkResourceTest struct {
	displayNamePrefix string
}

var oracleDBNetworkResourceTestEntity = oracleDBNetworkResourceTest{
	displayNamePrefix: "tf-ora-net",
}

// Basic test with bare minimum input
func TestAccODBNetworkResource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(oracleDBNetworkResourceTestEntity.displayNamePrefix)
	resourceName := "aws_odb_network.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkResourceTestEntity.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkResourceTestEntity.basicNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network),
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

func TestAccODBNetworkResource_withAllParams(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network1 odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(oracleDBNetworkResourceTestEntity.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkResourceTestEntity.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkResourceTestEntity.networkWithAllParams(rName, "julia.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network1),
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

func TestAccODBNetworkResource_updateManagedService(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network1, network2 odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(oracleDBNetworkResourceTestEntity.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkResourceTestEntity.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkResourceTestEntity.basicNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: oracleDBNetworkResourceTestEntity.basicNetworkWithActiveManagedService(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(network1.OdbNetworkId), *(network2.OdbNetworkId)) != 0 {
							return errors.New("should not  create a new cloud odb network")
						}
						return nil
					}),
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

func TestAccODBNetworkResource_disableManagedService(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network1, network2 odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(oracleDBNetworkResourceTestEntity.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkResourceTestEntity.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkResourceTestEntity.basicNetworkWithActiveManagedService(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: oracleDBNetworkResourceTestEntity.basicNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(network1.OdbNetworkId), *(network2.OdbNetworkId)) != 0 {
							return errors.New("should not  create a new cloud odb network")
						}
						return nil
					}),
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

func TestAccODBNetworkResource_updateTags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network1, network2 odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(oracleDBNetworkResourceTestEntity.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkResourceTestEntity.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkResourceTestEntity.basicNetwork(rName),

				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network1),
				),
			},
			{
				Config: oracleDBNetworkResourceTestEntity.updateNetworkTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(network1.OdbNetworkId), *(network2.OdbNetworkId)) != 0 {
							return errors.New("should not  create a new cloud odb network")
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
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

func TestAccODBNetworkResource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(oracleDBNetworkResourceTestEntity.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkResourceTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkResourceTestEntity.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkResourceTestEntity.basicNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					oracleDBNetworkResourceTestEntity.testAccCheckNetworkExists(ctx, resourceName, &network),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfodb.OracleDBNetwork, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (oracleDBNetworkResourceTest) testAccCheckNetworkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_network" {
				continue
			}
			_, err := tfodb.FindOracleDBNetworkResourceByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameNetwork, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameNetwork, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func (oracleDBNetworkResourceTest) testAccCheckNetworkExists(ctx context.Context, name string, network *odbtypes.OdbNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetwork, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetwork, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := tfodb.FindOracleDBNetworkResourceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetwork, rs.Primary.ID, err)
		}

		*network = *resp

		return nil
	}
}

func (oracleDBNetworkResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
	input := odb.ListOdbNetworksInput{}
	_, err := conn.ListOdbNetworks(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func (oracleDBNetworkResourceTest) basicNetwork(rName string) string {
	networkRes := fmt.Sprintf(`




resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}


`, rName)
	return networkRes
}

func (oracleDBNetworkResourceTest) basicNetworkWithActiveManagedService(rName string) string {
	networkRes := fmt.Sprintf(`




resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "ENABLED"
  zero_etl_access      = "ENABLED"
}


`, rName)
	return networkRes
}

func (oracleDBNetworkResourceTest) networkWithAllParams(rName, customDomainName string) string {
	networkRes := fmt.Sprintf(`




resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
  custom_domain_name   = %[2]q
}


`, rName, customDomainName)
	return networkRes
}

func (oracleDBNetworkResourceTest) updateNetworkTags(rName string) string {
	networkRes := fmt.Sprintf(`




resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
  tags = {
    "env" = "dev"
  }
}
`, rName)
	return networkRes
}
