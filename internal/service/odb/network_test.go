//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"strings"
	"testing"

	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
)

type odbNetworkResourceTest struct {
	displayNamePrefix string
}

var odbNetResourceTest = odbNetworkResourceTest{
	displayNamePrefix: "tf-odb-net",
}

// Basic test with bare minimum input
func TestOdbNetworkBasic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(odbNetResourceTest.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			odbNetResourceTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.basicOdbNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network),
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

func TestOdbNetworkWithAllParams(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(odbNetResourceTest.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			odbNetResourceTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.odbNetworkWithAllParams(rName, "julia.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network),
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

func TestAccODBNetworkUpdateManagedService(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network1, network2 odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(odbNetResourceTest.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			odbNetResourceTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.basicOdbNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: odbNetResourceTest.basicOdbNetworkWithActiveManagedService(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network2),
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

func TestAccODBNetworkDisableManagedService(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network1, network2 odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(odbNetResourceTest.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			odbNetResourceTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.basicOdbNetworkWithActiveManagedService(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: odbNetResourceTest.basicOdbNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network2),
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

func TestAccODBNetwork_Update_Tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network1, network2 odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(odbNetResourceTest.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			odbNetResourceTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.basicOdbNetwork(rName),

				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network1),
				),
			},
			{
				Config: odbNetResourceTest.updateOdbNetworkTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(network1.OdbNetworkId), *(network2.OdbNetworkId)) != 0 {
							return errors.New("should not  create a new cloud odb network")
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
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

func TestAccODBNetwork_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var network odbtypes.OdbNetwork
	rName := sdkacctest.RandomWithPrefix(odbNetResourceTest.displayNamePrefix)
	resourceName := "aws_odb_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			odbNetResourceTest.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.basicOdbNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					odbNetResourceTest.testAccCheckNetworkExists(ctx, resourceName, &network),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfodb.OdbNetwork, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (odbNetworkResourceTest) testAccCheckNetworkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_network" {
				continue
			}

			// TIP: ==== FINDERS ====
			// The find function should be exported. Since it won't be used outside of the package, it can be exported
			// in the `exports_test.go` file.
			_, err := tfodb.FindOdbNetworkByID(ctx, conn, rs.Primary.ID)
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

func (odbNetworkResourceTest) testAccCheckNetworkExists(ctx context.Context, name string, network *odbtypes.OdbNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetwork, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetwork, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := tfodb.FindOdbNetworkByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetwork, rs.Primary.ID, err)
		}

		*network = *resp

		return nil
	}
}

func (odbNetworkResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

	input := &odb.ListOdbNetworksInput{}

	_, err := conn.ListOdbNetworks(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

/*func testAccCheckNetworkNotRecreated(before, after *odb.DescribeNetworkResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.NetworkId), aws.ToString(after.NetworkId); before != after {
			return create.Error(names.ODB, create.ErrActionCheckingNotRecreated, tfodb.ResNameNetwork, aws.ToString(before.NetworkId), errors.New("recreated"))
		}

		return nil
	}
}*/

func (odbNetworkResourceTest) basicOdbNetwork(rName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access = "DISABLED"
  zero_etl_access = "DISABLED"
}

`, rName)
	return networkRes
}

func (odbNetworkResourceTest) basicOdbNetworkWithActiveManagedService(rName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access = "ENABLED"
  zero_etl_access = "ENABLED"
}

`, rName)
	return networkRes
}

func (odbNetworkResourceTest) odbNetworkWithAllParams(rName, customDomainName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access = "DISABLED"
  zero_etl_access = "DISABLED"
  custom_domain_name = %[2]q
}

`, rName, customDomainName)
	return networkRes
}

func (odbNetworkResourceTest) updateOdbNetworkDisplayName(rName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access = "DISABLED"
  zero_etl_access = "DISABLED"
}
`, rName)
	return networkRes
}

func (odbNetworkResourceTest) updateOdbNetworkTags(rName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access = "DISABLED"
  zero_etl_access = "DISABLED"
  tags = {
     "env"= "dev"
  }
}
`, rName)
	return networkRes
}
