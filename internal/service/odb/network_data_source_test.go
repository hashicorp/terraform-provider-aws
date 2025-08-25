//Copyright (c) 2025, Oracle and/or its affiliates. All rights reserved.

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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
)

type odbNetworkDataSourceTest struct {
}

var odbNetworkDataSourceTestEntity = odbNetworkDataSourceTest{}

func TestAccODBNetworkDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	networkResource := "aws_odb_network.test_resource"
	networkDataSource := "data.aws_odb_network.test"
	rName := sdkacctest.RandomWithPrefix("tf-odb-net")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			odbNetworkDataSourceTestEntity.testAccOdbNetworkDataSourcePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             odbNetworkDataSourceTestEntity.testAccCheckOdbNetworkDataSourceDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: odbNetworkDataSourceTestEntity.basicOdbNetworkDataSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(networkResource, "id", networkDataSource, "id"),
				),
			},
		},
	})
}

func (odbNetworkDataSourceTest) testAccCheckOdbNetworkDataSourceDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_network" {
				continue
			}
			_, err := odbNetworkDataSourceTestEntity.findOdbNetwork(ctx, conn, rs.Primary.ID)
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

func (odbNetworkDataSourceTest) findOdbNetwork(ctx context.Context, conn *odb.Client, id string) (*odbtypes.OdbNetwork, error) {
	input := odb.GetOdbNetworkInput{
		OdbNetworkId: aws.String(id),
	}

	out, err := conn.GetOdbNetwork(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.OdbNetwork == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out.OdbNetwork, nil
}

func (odbNetworkDataSourceTest) basicOdbNetworkDataSource(rName string) string {
	networkRes := fmt.Sprintf(`




resource "aws_odb_network" "test_resource" {
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


data "aws_odb_network" "test" {
  id = aws_odb_network.test_resource.id
}


`, rName)
	return networkRes
}
func (odbNetworkDataSourceTest) testAccOdbNetworkDataSourcePreCheck(ctx context.Context, t *testing.T) {
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
