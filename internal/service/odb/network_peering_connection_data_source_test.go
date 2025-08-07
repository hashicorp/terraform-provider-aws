//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"testing"
)

type odbPeeringDataSourceTest struct {
	odbNetDisplayNamePrefix            string
	odbNetworkPeeringDisplayNamePrefix string
	vpcNamePrefix                      string
}

var odbPeeringDSTest = odbPeeringDataSourceTest{
	odbNetDisplayNamePrefix:            "tf",
	odbNetworkPeeringDisplayNamePrefix: "tf",
	vpcNamePrefix:                      "tf",
}

// Acceptance test access AWS and cost money to run.
func TestAccODBNetworkPeeringConnectionDataSource_basic(t *testing.T) {

	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	networkPeeringResource := "aws_odb_network_peering_connection.test"
	networkPerringDataSource := "data.aws_odb_network_peering_connection.test"
	odbNetPeeringDisplayName := sdkacctest.RandomWithPrefix(odbPeeringDSTest.odbNetworkPeeringDisplayNamePrefix)
	odbNetDispName := sdkacctest.RandomWithPrefix(odbPeeringDSTest.odbNetDisplayNamePrefix)
	vpcName := sdkacctest.RandomWithPrefix(odbPeeringDSTest.vpcNamePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbPeeringDSTest.testAccCheckCloudOdbNetworkPeeringDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbPeeringDSTest.basicPeeringConfig(vpcName, odbNetDispName, odbNetPeeringDisplayName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(networkPeeringResource, "id", networkPerringDataSource, "id"),
				),
			},
		},
	})
}

func (odbPeeringDataSourceTest) testAccCheckCloudOdbNetworkPeeringDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_network_peering_connection" {
				continue
			}
			_, err := odbPeeringDSTest.findOdbPeering(ctx, conn, rs.Primary.ID)

			if err != nil {
				if tfresource.NotFound(err) {
					return nil
				}
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameNetworkPeeringConnection, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameNetworkPeeringConnection, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func (odbPeeringDataSourceTest) findOdbPeering(ctx context.Context, conn *odb.Client, id string) (output *odb.GetOdbPeeringConnectionOutput, err error) {
	input := odb.GetOdbPeeringConnectionInput{
		OdbPeeringConnectionId: &id,
	}
	out, err := conn.GetOdbPeeringConnection(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
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
func (odbPeeringDataSourceTest) basicPeeringConfig(vpcName, odbNetDisplayName, odbPeeringDisplayName string) string {

	testData := fmt.Sprintf(`

resource "aws_vpc" "test" {
  cidr_block       = "10.0.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_odb_network" "test" {
  display_name          = %[2]q
  availability_zone_id = "use1-az6"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access = "DISABLED"
  zero_etl_access = "DISABLED"
}

resource "aws_odb_network_peering_connection" "test" {
  display_name = %[3]q
  odb_network_id = aws_odb_network.test.id
  peer_network_id = aws_vpc.test.id
  
}

data "aws_odb_network_peering_connection" "test" {
  id=aws_odb_network_peering_connection.test.id
}

`, vpcName, odbNetDisplayName, odbPeeringDisplayName)
	return testData
}
