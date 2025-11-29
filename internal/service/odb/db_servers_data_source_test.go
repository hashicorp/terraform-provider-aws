// Copyright (c) HashiCorp, Inc.
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

type testDbServersListDataSource struct {
	displayNamePrefix string
}

var dbServersListDataSourceTestEntity = testDbServersListDataSource{
	displayNamePrefix: "Ofake-exa",
}

// Acceptance test access AWS and cost money to run.
func TestAccODBDBServersListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var dbServersList odb.ListDbServersOutput
	dataSourceName := "data.aws_odb_db_servers.test"
	exaInfraResourceName := "aws_odb_cloud_exadata_infrastructure.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             dbServersListDataSourceTestEntity.testAccCheckDBServersDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: dbServersListDataSourceTestEntity.testAccDBServersListDataSourceConfigBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					dbServersListDataSourceTestEntity.testAccCheckDBServersListExists(ctx, exaInfraResourceName, &dbServersList),
					resource.TestCheckResourceAttr(dataSourceName, "aws_odb_db_servers.db_servers.#", strconv.Itoa(len(dbServersList.DbServers))),
				),
			},
		},
	})
}

func (testDbServersListDataSource) testAccCheckDBServersListExists(ctx context.Context, name string, output *odb.ListDbServersOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDBServersList, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		var exaInfraId = &rs.Primary.ID

		resp, err := dbServersListDataSourceTestEntity.findDBServersList(ctx, conn, exaInfraId)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDBServersList, rs.Primary.ID, err)
		}
		*output = *resp
		return nil
	}
}

func (testDbServersListDataSource) testAccCheckDBServersDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_exadata_infrastructure" {
				continue
			}
			_, err := dbServersListDataSourceTestEntity.findExaInfra(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServersList, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServersList, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func (testDbServersListDataSource) findExaInfra(ctx context.Context, conn *odb.Client, id string) (*odbtypes.CloudExadataInfrastructure, error) {
	input := odb.GetCloudExadataInfrastructureInput{
		CloudExadataInfrastructureId: aws.String(id),
	}
	out, err := conn.GetCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return nil, err
	}
	if out == nil || out.CloudExadataInfrastructure == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}
	return out.CloudExadataInfrastructure, nil
}

func (testDbServersListDataSource) findDBServersList(ctx context.Context, conn *odb.Client, exaInfraId *string) (*odb.ListDbServersOutput, error) {
	inputWithExaId := odb.ListDbServersInput{
		CloudExadataInfrastructureId: exaInfraId,
	}
	output, err := conn.ListDbServers(ctx, &inputWithExaId)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (testDbServersListDataSource) testAccDBServersListDataSourceConfigBasic() string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(dbServersListDataSourceTestEntity.displayNamePrefix)
	exaInfra := dbServersListDataSourceTestEntity.exaInfra(exaInfraDisplayName)
	return fmt.Sprintf(`
%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}
`, exaInfra)
}

func (testDbServersListDataSource) exaInfra(rName string) string {
	exaRes := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name         = %[1]q
  shape                = "Exadata.X9M"
  storage_count        = 3
  compute_count        = 2
  availability_zone_id = "use1-az6"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    is_custom_action_timeout_enabled = true
    patching_mode                    = "ROLLING"
    preference                       = "NO_PREFERENCE"
  }
}
`, rName)
	return exaRes
}
