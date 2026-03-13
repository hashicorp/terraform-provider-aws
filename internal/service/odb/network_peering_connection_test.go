// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type oracleDBNwkPeeringResourceTest struct {
	vpcNamePrefix               string
	odbPeeringDisplayNamePrefix string
	odbNwkDisplayNamePrefix     string
}

var oracleDBNwkPeeringTestResource = oracleDBNwkPeeringResourceTest{
	vpcNamePrefix:               "odb-vpc",
	odbPeeringDisplayNamePrefix: "odb-peering",
	odbNwkDisplayNamePrefix:     "odb-net",
}

func TestOdbNetworkAddRemovePeerCIDRUnitTest(t *testing.T) {
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
			addRemoveCidrs := tfodb.ResourceNetworkPeeringConnection.FindAddRemovePeeredNetworkCIDR(testCase.NewCidrs, testCase.OldCidrs)
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
func TestAccODBNetworkPeeringConnection_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var odbPeeringResource odb.GetOdbPeeringConnectionOutput
	odbPeeringDisplayName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbPeeringDisplayNamePrefix)
	vpcName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.vpcNamePrefix)
	odbNetName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbNwkDisplayNamePrefix)
	resourceName := "aws_odb_network_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNwkPeeringTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNwkPeeringTestResource.testAccCheckNetworkPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNwkPeeringTestResource.basicConfig(vpcName, odbNetName, odbPeeringDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &odbPeeringResource),
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

func TestAccODBNetworkPeeringConnection_withARN(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var odbPeeringResource odb.GetOdbPeeringConnectionOutput
	odbPeeringDisplayName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbPeeringDisplayNamePrefix)
	vpcName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.vpcNamePrefix)
	odbNetName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbNwkDisplayNamePrefix)
	resourceName := "aws_odb_network_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNwkPeeringTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNwkPeeringTestResource.testAccCheckNetworkPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNwkPeeringTestResource.basicConfigWithARN(vpcName, odbNetName, odbPeeringDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &odbPeeringResource),
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

func TestAccODBNetworkPeeringConnection_variables(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	odbPeeringDisplayName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbPeeringDisplayNamePrefix)
	vpcName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.vpcNamePrefix)
	odbNetName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbNwkDisplayNamePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			vmClusterTestEntity.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNwkPeeringTestResource.testAccCheckNetworkPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			// nosemgrep:ci.semgrep.acctest.checks.replace-planonly-checks
			{
				Config:             oracleDBNwkPeeringTestResource.basicConfig_useVariables(vpcName, odbNetName, odbPeeringDisplayName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccODBNetworkPeeringConnection_tagging(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var odbPeeringResource odb.GetOdbPeeringConnectionOutput
	odbPeeringDisplayName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbPeeringDisplayNamePrefix)
	vpcName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.vpcNamePrefix)
	odbNetName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbNwkDisplayNamePrefix)
	resourceName := "aws_odb_network_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNwkPeeringTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNwkPeeringTestResource.testAccCheckNetworkPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNwkPeeringTestResource.basicConfig(vpcName, odbNetName, odbPeeringDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &odbPeeringResource),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: oracleDBNwkPeeringTestResource.basicConfigNoTag(vpcName, odbNetName, odbPeeringDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &odbPeeringResource),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
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

func TestAccODBNetworkPeeringConnection_addRemovePeeredCIDR(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var peering1, peering2, peering3 odb.GetOdbPeeringConnectionOutput
	odbPeeringDisplayName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbPeeringDisplayNamePrefix)
	vpcName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.vpcNamePrefix)
	odbNetName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbNwkDisplayNamePrefix)
	resourceName := "aws_odb_network_peering_connection.test"
	basicConfig, removedCidr, addedCidr := oracleDBNwkPeeringTestResource.addRemovePeeredNetworkCIDRConfig(vpcName, odbNetName, odbPeeringDisplayName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNwkPeeringTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNwkPeeringTestResource.testAccCheckNetworkPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: basicConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &peering1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: removedCidr,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &peering2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(peering1.OdbPeeringConnection.OdbPeeringConnectionId), *(peering2.OdbPeeringConnection.OdbPeeringConnectionId)) != 0 {
							return errors.New("should not  create a new odb network peering connection")
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
			{
				Config: addedCidr,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &peering3),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(peering1.OdbPeeringConnection.OdbPeeringConnectionId), *(peering3.OdbPeeringConnection.OdbPeeringConnectionId)) != 0 {
							return errors.New("should not  create a new odb network peering connection")
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

func TestAccODBNetworkPeeringConnection_updateTagAndRemovePeeredCIDR(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	vpcName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.vpcNamePrefix)
	odbNetName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbNwkDisplayNamePrefix)
	resourceName := "aws_odb_network_peering_connection.test"
	var peering1, peering2 odb.GetOdbPeeringConnectionOutput
	basicConfig, removedCidr := oracleDBNwkPeeringTestResource.basicConfigWithMultiplePeerCIDR(vpcName, odbNetName, resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNwkPeeringTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNwkPeeringTestResource.testAccCheckNetworkPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: basicConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &peering1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: removedCidr,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &peering2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(peering1.OdbPeeringConnection.OdbPeeringConnectionId), *(peering2.OdbPeeringConnection.OdbPeeringConnectionId)) != 0 {
							return errors.New("should not  create a new odb network peering connection")
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

func TestAccODBNetworkPeeringConnection_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var odbPeering odb.GetOdbPeeringConnectionOutput
	odbPeeringDisplayName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbPeeringDisplayNamePrefix)
	vpcName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.vpcNamePrefix)
	odbNetDisplayName := sdkacctest.RandomWithPrefix(oracleDBNwkPeeringTestResource.odbPeeringDisplayNamePrefix)
	resourceName := "aws_odb_network_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNwkPeeringTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNwkPeeringTestResource.testAccCheckNetworkPeeringConnectionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNwkPeeringTestResource.basicConfig(vpcName, odbNetDisplayName, odbPeeringDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkPeeringConnectionExists(ctx, resourceName, &odbPeering),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfodb.OracleDBNetworkPeeringConnection, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (oracleDBNwkPeeringResourceTest) testAccCheckNetworkPeeringConnectionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_network_peering_connection" {
				continue
			}
			_, err := oracleDBNwkPeeringTestResource.findOracleDBNetworkPeering(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameNetworkPeeringConnection, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameNetworkPeeringConnection, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func testAccCheckNetworkPeeringConnectionExists(ctx context.Context, name string, odbPeeringConnection *odb.GetOdbPeeringConnectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetworkPeeringConnection, name, errors.New("not found"))
		}
		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetworkPeeringConnection, name, errors.New("not set"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := oracleDBNwkPeeringTestResource.findOracleDBNetworkPeering(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameNetworkPeeringConnection, rs.Primary.ID, err)
		}
		*odbPeeringConnection = *resp
		return nil
	}
}

func (oracleDBNwkPeeringResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
	input := odb.ListOdbPeeringConnectionsInput{}
	_, err := conn.ListOdbPeeringConnections(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func (oracleDBNwkPeeringResourceTest) findOracleDBNetworkPeering(ctx context.Context, conn *odb.Client, id string) (output *odb.GetOdbPeeringConnectionOutput, err error) {
	input := odb.GetOdbPeeringConnectionInput{
		OdbPeeringConnectionId: &id,
	}
	out, err := conn.GetOdbPeeringConnection(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}
	if out == nil {
		return nil, errors.New("odb Network Peering Connection resource can not be nil")
	}
	return out, nil
}

func (oracleDBNwkPeeringResourceTest) basicConfig(vpcName, odbNetName, odbPeeringName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_odb_network" "test" {
  display_name         = %[2]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}

resource "aws_odb_network_peering_connection" "test" {
  display_name    = %[3]q
  odb_network_id  = aws_odb_network.test.id
  peer_network_id = aws_vpc.test.id
  tags = {
    "env" = "dev"
  }
}
`, vpcName, odbNetName, odbPeeringName)
}

func (oracleDBNwkPeeringResourceTest) basicConfig_useVariables(vpcName, odbNetName, odbPeeringName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

variable odb_network_id {
  default     = "odbnet_3l9st3litg"
  type        = string
  description = "ODB Network"
}

resource "aws_odb_network_peering_connection" "test" {
  display_name    = %[3]q
  odb_network_id  = var.odb_network_id
  peer_network_id = aws_vpc.test.id
  tags = {
    "env" = "dev"
  }
}
`, vpcName, odbNetName, odbPeeringName)
}

func (oracleDBNwkPeeringResourceTest) basicConfigWithARN(vpcName, odbNetName, odbPeeringName string) string {
	return fmt.Sprintf(`

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_odb_network" "test" {
  display_name         = %[2]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}

resource "aws_odb_network_peering_connection" "test" {
  display_name    = %[3]q
  odb_network_id  = aws_odb_network.test.arn
  peer_network_id = aws_vpc.test.id
  tags = {
    "env" = "dev"
  }
}
`, vpcName, odbNetName, odbPeeringName)
}

func (oracleDBNwkPeeringResourceTest) basicConfigNoTag(vpcName, odbNetName, odbPeeringName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_odb_network" "test" {
  display_name         = %[2]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}

resource "aws_odb_network_peering_connection" "test" {
  display_name    = %[3]q
  odb_network_id  = aws_odb_network.test.id
  peer_network_id = aws_vpc.test.id

}
`, vpcName, odbNetName, odbPeeringName)
}

func (oracleDBNwkPeeringResourceTest) addRemovePeeredNetworkCIDRConfig(vpcName, odbNetName, odbPeeringName string) (string, string, string) {
	odbPeeringBasic := fmt.Sprintf(`




%[1]s

%[2]s

resource "aws_odb_network_peering_connection" "test" {
  display_name    = %[3]q
  depends_on      = [aws_ec2_transit_gateway_vpc_attachment.test]
  odb_network_id  = aws_odb_network.test.id
  peer_network_id = aws_vpc.test.id
}
`, oracleDBNwkPeeringTestResource.configVPCForAddRemoveCIDR(vpcName), oracleDBNwkPeeringTestResource.oracleDataBaseNetworkConfig(odbNetName), odbPeeringName)

	odbPeeringRemoved := fmt.Sprintf(`






%[1]s

%[2]s

resource "aws_odb_network_peering_connection" "test" {
  display_name       = %[3]q
  odb_network_id     = aws_odb_network.test.id
  depends_on         = [aws_ec2_transit_gateway_vpc_attachment.test]
  peer_network_id    = aws_vpc.test.id
  peer_network_cidrs = ["13.0.0.0/16"]

}
`, oracleDBNwkPeeringTestResource.configVPCForAddRemoveCIDR(vpcName), oracleDBNwkPeeringTestResource.oracleDataBaseNetworkConfig(odbNetName), odbPeeringName)

	odbPeeringAdded := fmt.Sprintf(`






%[1]s

%[2]s

resource "aws_odb_network_peering_connection" "test" {
  display_name       = %[3]q
  odb_network_id     = aws_odb_network.test.id
  depends_on         = [aws_ec2_transit_gateway_vpc_attachment.test]
  peer_network_id    = aws_vpc.test.id
  peer_network_cidrs = ["13.0.0.0/16", "16.1.0.0/16"]

}
`, oracleDBNwkPeeringTestResource.configVPCForAddRemoveCIDR(vpcName), oracleDBNwkPeeringTestResource.oracleDataBaseNetworkConfig(odbNetName), odbPeeringName)
	return odbPeeringBasic, odbPeeringRemoved, odbPeeringAdded
}

func (oracleDBNwkPeeringResourceTest) configVPCForAddRemoveCIDR(vpcName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "13.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
resource "aws_vpc_ipv4_cidr_block_association" "secondary" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "16.1.0.0/16"
}


# Subnets in primary CIDR
resource "aws_subnet" "secondary_a" {
  depends_on = [
    aws_vpc_ipv4_cidr_block_association.secondary
  ]
  vpc_id               = aws_vpc.test.id
  cidr_block           = "16.1.1.0/24"
  availability_zone_id = "use1-az6"

  tags = {
    Name = "secondary_a"
  }
}


resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
  subnet_ids         = [aws_subnet.secondary_a.id]
}








`, vpcName)
}

func (oracleDBNwkPeeringResourceTest) oracleDataBaseNetworkConfig(displayName string) string {
	return fmt.Sprintf(`
resource "aws_odb_network" "test" {
  display_name         = %[1]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
}
`, displayName)
}

func (oracleDBNwkPeeringResourceTest) basicConfigWithMultiplePeerCIDR(vpcName, networkName, peerNetworkConnectionName string) (string, string) {
	peeringWithTags := fmt.Sprintf(`
 %[1]s

resource "aws_vpc" "test" {
  cidr_block = "13.0.0.0/16"
  tags = {
    Name = %[2]q
  }
}
resource "aws_vpc_ipv4_cidr_block_association" "secondary" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "16.1.0.0/16"
}
resource "aws_vpc_ipv4_cidr_block_association" "tertiary" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "19.1.0.0/16"
}

resource "aws_odb_network_peering_connection" "test" {
  odb_network_id  = aws_odb_network.test.id
  display_name    = %[3]q
  depends_on      = [aws_vpc_ipv4_cidr_block_association.secondary, aws_vpc_ipv4_cidr_block_association.tertiary]
  peer_network_id = aws_vpc.test.id
  tags = {
    "env" = "dev"
  }

}




`, oracleDBNwkPeeringTestResource.oracleDataBaseNetworkConfig(networkName), vpcName, peerNetworkConnectionName)

	peeringWithoutTags := fmt.Sprintf(`
 %[1]s

resource "aws_vpc" "test" {
  cidr_block = "13.0.0.0/16"
  tags = {
    Name = %[2]q
  }
}
resource "aws_vpc_ipv4_cidr_block_association" "secondary" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "16.1.0.0/16"
}
resource "aws_vpc_ipv4_cidr_block_association" "tertiary" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "19.1.0.0/16"
}

resource "aws_odb_network_peering_connection" "test" {
  odb_network_id     = aws_odb_network.test.id
  display_name       = %[3]q
  depends_on         = [aws_vpc_ipv4_cidr_block_association.secondary, aws_vpc_ipv4_cidr_block_association.tertiary]
  peer_network_id    = aws_vpc.test.id
  peer_network_cidrs = ["13.0.0.0/16", "16.1.0.0/16"]
  tags = {
    "env" = "dev"
  }

}




`, oracleDBNwkPeeringTestResource.oracleDataBaseNetworkConfig(networkName), vpcName, peerNetworkConnectionName)

	return peeringWithTags, peeringWithoutTags
}
