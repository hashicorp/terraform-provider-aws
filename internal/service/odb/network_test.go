// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"reflect"
	"testing"

	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
)

type odbNetworkResourceTest struct {
	displayNamePrefix string
}

var odbNetResourceTest = odbNetworkResourceTest{
	displayNamePrefix: "tf-odb-net",
}

func TestOdbNetworkAddRemovePerredCidrsUnitTest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName       string
		OldCidrs       []string
		NewCidrs       []string
		AddRemoveCidrs map[string]int
	}{
		{
			TestName:       "non empty new, empty old",
			NewCidrs:       []string{"10.0.0.0/24"},
			OldCidrs:       []string{},
			AddRemoveCidrs: map[string]int{"10.0.0.0/24": 1},
		},
		{
			TestName:       "non empty new, non empty old",
			NewCidrs:       []string{"10.0.0.0/24"},
			OldCidrs:       []string{"10.0.0.0/34"},
			AddRemoveCidrs: map[string]int{"10.0.0.0/24": 1, "10.0.0.0/34": -1},
		},
		{
			TestName:       "non empty new, non empty old all same",
			NewCidrs:       []string{"10.0.0.0/24"},
			OldCidrs:       []string{"10.0.0.0/24"},
			AddRemoveCidrs: map[string]int{},
		},
		{
			TestName:       "empty new, non empty old all ",
			NewCidrs:       []string{},
			OldCidrs:       []string{"10.0.0.0/24", "10.0.0.0/34"},
			AddRemoveCidrs: map[string]int{"10.0.0.0/24": -1, "10.0.0.0/34": -1},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()
			addRemoveCidrs := tfodb.FindAddRemovedCidrsFromOdbNetWork(testCase.NewCidrs, testCase.OldCidrs)

			if addRemoveCidrs != nil {
				if len(addRemoveCidrs) != len(testCase.AddRemoveCidrs) {
					t.Fatalf("expected %d addRemoveCidrs, got %d", len(testCase.AddRemoveCidrs), len(addRemoveCidrs))
				}
				if !reflect.DeepEqual(addRemoveCidrs, testCase.AddRemoveCidrs) {
					t.Fatalf("expected %v, got %v", testCase.AddRemoveCidrs, addRemoveCidrs)
				}
			} else {
				t.Error("addRemoveCidrs was nil")
			}

		})
	}
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
			//testAccPreCheck(ctx, t)
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

// with peered_cidr
func TestAccODBNetwork_only_with_peered_cidr(t *testing.T) {
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
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.basicOdbNetworkWithPeeredCidrs(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(state *terraform.State) error {
						fmt.Println(state)
						return nil
					},
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

// TestAccODBNetwork_basic_with_peered_cidr_vpc_arn
func TestAccODBNetwork_with_peered_cidr_vpc_arn(t *testing.T) {
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
			//testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetResourceTest.testAccCheckNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetResourceTest.basicOdbNetworkWithVPCArn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(state *terraform.State) error {
						fmt.Println(state)
						return nil
					},
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

// TestAccODBNetwork_Update
func TestAccODBNetwork_Delete_Create(t *testing.T) {
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
			//acctest.PreCheckPartitionHasService(t, names.ODBEndpointID)
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
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						fmt.Println(state)
						return nil
					}),
				),
			},
			{
				Config: odbNetResourceTest.updateOdbNetworkDisplayName(rName + "_foo"),
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

func TestAccODBNetwork_Update_Tags(t *testing.T) {
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
			//acctest.PreCheckPartitionHasService(t, names.ODBEndpointID)
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
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						fmt.Println(state)
						return nil
					}),
				),
			},
			{
				Config: odbNetResourceTest.updateOdbNetworkTags(rName),
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

func (odbNetworkResourceTest) basicOdbNetworkWithPeeredCidrs(rName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  peered_cidrs         = ["10.32.0.0/24", "172.16.2.0/24", "172.16.0.0/16"]
  tags = {
    "env"= "dev"
  }
}

`, rName)
	return networkRes
}

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

func (odbNetworkResourceTest) updateOdbNetworkDisplayName(rName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
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
 tags = {
    "env"= "dev"
    "foo"= "bar"
  }
}
`, rName)
	return networkRes
}

func (odbNetworkResourceTest) basicOdbNetworkWithVPCArn(rName string) string {
	networkRes := fmt.Sprintf(`


resource "aws_odb_network" "test" {
  display_name          = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  peer_vpc_arn         = "arn:aws:ec2:us-east-1:711387093194:vpc/vpc-0c6e9101b49f80ea3" 
  peered_cidrs         = ["10.13.0.0/24", "172.16.2.0/24", "172.16.0.0/16"]
  tags = {
    "env"= "dev"
  }
}

`, rName)
	return networkRes
}
