// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	rName := sdkacctest.RandomWithPrefix("tf-ora-net")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			oracleDBNetworkDataSourceTestEntity.testAccNetworkDataSourcePreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             oracleDBNetworkDataSourceTestEntity.testAccCheckNetworkDataSourceDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: oracleDBNetworkDataSourceTestEntity.basicNetworkDataSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(networkResource, names.AttrID, networkDataSource, names.AttrID),
				),
			},
		},
	})
}

func (oracleDBNetworkDataSourceTest) testAccCheckNetworkDataSourceDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

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
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.OdbNetwork == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out.OdbNetwork, nil
}

func (oracleDBNetworkDataSourceTest) basicNetworkDataSource(rName string) string {
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
  cross_region_s3_restore_sources_access = ["us-west-2"]
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
func (oracleDBNetworkDataSourceTest) testAccNetworkDataSourcePreCheck(ctx context.Context, t *testing.T) {
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
