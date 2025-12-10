// Copyright IBM Corp. 2014, 2025
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

type testDbServerDataSourceTest struct {
	exaDisplayNamePrefix string
}

var dbServerDataSourceTestEntity = testDbServerDataSourceTest{
	exaDisplayNamePrefix: "Ofake-exa",
}

// Acceptance test access AWS and cost money to run.
func TestAccODBDBServerDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var dbServer odb.GetDbServerOutput

	dataSourceName := "data.aws_odb_db_server.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             dbServerDataSourceTestEntity.testAccCheckDBServersDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: dbServerDataSourceTestEntity.basicDBServerDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					dbServerDataSourceTestEntity.testAccCheckDBServerExists(ctx, dataSourceName, &dbServer),
				),
			},
		},
	})
}

func (testDbServerDataSourceTest) testAccCheckDBServerExists(ctx context.Context, name string, output *odb.GetDbServerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDBServer, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		var dbServerId = rs.Primary.ID
		var attributes = rs.Primary.Attributes
		exaId := attributes["exadata_infrastructure_id"]
		resp, err := dbServerDataSourceTestEntity.findDBServer(ctx, conn, &dbServerId, &exaId)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDBServer, rs.Primary.ID, err)
		}
		*output = *resp
		return nil
	}
}

func (testDbServerDataSourceTest) testAccCheckDBServersDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_exadata_infrastructure" {
				continue
			}
			err := dbServerDataSourceTestEntity.findExaInfra(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServer, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDBServer, rs.Primary.ID, errors.New("not destroyed"))
		}
		return nil
	}
}

func (testDbServerDataSourceTest) findExaInfra(ctx context.Context, conn *odb.Client, id string) error {
	input := odb.GetCloudExadataInfrastructureInput{
		CloudExadataInfrastructureId: aws.String(id),
	}
	out, err := conn.GetCloudExadataInfrastructure(ctx, &input)
	if err != nil {
		if errs.IsA[*odbtypes.ResourceNotFoundException](err) {
			return &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}
		return err
	}
	if out == nil || out.CloudExadataInfrastructure == nil {
		return tfresource.NewEmptyResultError(&input)
	}
	return nil
}

func (testDbServerDataSourceTest) findDBServer(ctx context.Context, conn *odb.Client, dbServerId *string, exaInfraId *string) (*odb.GetDbServerOutput, error) {
	inputWithExaId := odb.GetDbServerInput{
		DbServerId:                   dbServerId,
		CloudExadataInfrastructureId: exaInfraId,
	}
	output, err := conn.GetDbServer(ctx, &inputWithExaId)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (testDbServerDataSourceTest) basicDBServerDataSourceConfig() string {
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(dbServersListDataSourceTestEntity.displayNamePrefix)
	exaInfra := dbServerDataSourceTestEntity.exaInfra(exaInfraDisplayName)

	return fmt.Sprintf(`
%s

data "aws_odb_db_servers" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}

data "aws_odb_db_server" "test" {
  id                              = data.aws_odb_db_servers.test.db_servers[0].id
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}
`, exaInfra)
}

func (testDbServerDataSourceTest) exaInfra(rName string) string {
	exaRes := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name         = "%[1]s"
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
