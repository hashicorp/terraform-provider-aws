// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type oracleDBNetworkDataSourceTest struct {
}

var oracleDBNetworkDataSourceTestEntity = oracleDBNetworkDataSourceTest{}

func TestAccODBNetworkDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	networkResource := "aws_odb_network.test_resource"
	networkDataSource := "data.aws_odb_network.test"
	rName := acctest.RandomWithPrefix(t, "tf-ora-net")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkDataSourceTestEntity.testAccNetworkDataSourcePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkDataSourceTestEntity.testAccCheckNetworkDataSourceDestroyed(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkDataSourceTestEntity.basicNetworkDataSource(rName, endpoints.UsWest2RegionID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(networkResource, names.AttrID, networkDataSource, names.AttrID),
				),
			},
		},
	})
}

func TestAccODBNetworkDataSource_ec2PlacementGroupIDs(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	networkResource := "aws_odb_network.test_resource"
	networkDataSource := "data.aws_odb_network.test"
	rName := acctest.RandomWithPrefix(t, "tf-ora-net")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkDataSourceTestEntity.testAccNetworkDataSourcePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkDataSourceTestEntity.testAccCheckNetworkDataSourceDestroyed(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkDataSourceTestEntity.basicNetworkDataSourceForEC2PlacementGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(networkResource, names.AttrID, networkDataSource, names.AttrID),
					resource.TestCheckResourceAttrWith(networkDataSource, "ec2_placement_group_ids.#", func(value string) error {
						count, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("parsing ec2_placement_group_ids count: %w", err)
						}
						if count <= 0 {
							return fmt.Errorf("expected ec2_placement_group_ids to be non-empty, got %d", count)
						}
						return nil
					}),
				),
			},
		},
	})
}

func (oracleDBNetworkDataSourceTest) testAccCheckNetworkDataSourceDestroyed(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_network" {
				continue
			}
			_, err := oracleDBNetworkDataSourceTestEntity.findNetwork(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
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

func (oracleDBNetworkDataSourceTest) findNetwork(ctx context.Context, conn *odb.Client, id string) (*odbtypes.OdbNetwork, error) {
	input := odb.GetOdbNetworkInput{
		OdbNetworkId: aws.String(id),
	}

	out, err := conn.GetOdbNetwork(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || out.OdbNetwork == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.OdbNetwork, nil
}

func (oracleDBNetworkDataSourceTest) basicNetworkDataSource(rName, rRegion string) string {
	networkRes := fmt.Sprintf(`




resource "aws_odb_network" "test_resource" {
  display_name                           = %[1]q
  availability_zone_id                   = "use1-az6"
  client_subnet_cidr                     = "10.2.0.0/24"
  backup_subnet_cidr                     = "10.2.1.0/24"
  s3_access                              = "DISABLED"
  zero_etl_access                        = "DISABLED"
  sts_access                             = "DISABLED"
  kms_access                             = "DISABLED"
  cross_region_s3_restore_sources_access = [%[2]q]
  tags = {
    "env" = "dev"
  }
}


data "aws_odb_network" "test" {
  id = aws_odb_network.test_resource.id
}


`, rName, rRegion)
	return networkRes
}

func (oracleDBNetworkDataSourceTest) testAccNetworkDataSourcePreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ODBClient(ctx)
	input := odb.ListOdbNetworksInput{}
	_, err := conn.ListOdbNetworks(ctx, &input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func (oracleDBNetworkDataSourceTest) basicNetworkDataSourceForEC2PlacementGroup(rName string) string {
	networkRes := fmt.Sprintf(`

resource "aws_odb_network" "test_resource" {
  display_name         = %[1]q
  availability_zone_id = "aps2-az3"
  client_subnet_cidr   = "10.2.0.0/24"
  backup_subnet_cidr   = "10.2.1.0/24"
  s3_access            = "DISABLED"
  zero_etl_access      = "DISABLED"
  sts_access           = "DISABLED"
  kms_access           = "DISABLED"
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
