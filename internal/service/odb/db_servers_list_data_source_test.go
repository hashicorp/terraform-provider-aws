// Copyright (c) 2025, Oracle and/or its affiliates. All rights reserved.

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

type testDbServersListDataSource struct {
	displayNamePrefix string
}

var dbServersListDataSourceTests = testDbServersListDataSource{
	displayNamePrefix: "Ofake-exa",
}

// Acceptance test access AWS and cost money to run.
func TestAccODBDbServersListDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var dbserverslist odb.ListDbServersOutput
	exaInfraDisplayName := sdkacctest.RandomWithPrefix(dbServersListDataSourceTests.displayNamePrefix)

	dataSourceName := "data.aws_odb_db_servers_list.test"
	exaInfraResourceName := "aws_odb_cloud_exadata_infrastructure.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)

		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             dbServersListDataSourceTests.testAccCheckDbServersDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: dbServersListDataSourceTests.testAccDbServersListDataSourceConfigBasic(dbServersListDataSourceTests.exaInfra(exaInfraDisplayName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					dbServersListDataSourceTests.testAccCheckDbServersListExists(ctx, exaInfraResourceName, &dbserverslist),
					resource.TestCheckResourceAttr(dataSourceName, "aws_odb_db_servers_list.db_servers.#", strconv.Itoa(len(dbserverslist.DbServers))),
				),
			},
		},
	})
}

func (testDbServersListDataSource) testAccCheckDbServersListExists(ctx context.Context, name string, output *odb.ListDbServersOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDbServersList, name, errors.New("not found"))
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		var exaInfraId = &rs.Primary.ID

		resp, err := dbServersListDataSourceTests.findDbServersList(ctx, conn, exaInfraId)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.DSNameDbServersList, rs.Primary.ID, err)
		}
		*output = *resp
		return nil
	}
}

func (testDbServersListDataSource) testAccCheckDbServersDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_exadata_infrastructure" {
				continue
			}
			_, err := dbServersListDataSourceTests.findExaInfra(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDbServersList, rs.Primary.ID, err)
			}
			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.DSNameDbServersList, rs.Primary.ID, errors.New("not destroyed"))
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
			return nil, &retry.NotFoundError{
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

func (testDbServersListDataSource) findDbServersList(ctx context.Context, conn *odb.Client, exaInfraId *string) (*odb.ListDbServersOutput, error) {
	inputWithExaId := &odb.ListDbServersInput{
		CloudExadataInfrastructureId: exaInfraId,
	}
	output, err := conn.ListDbServers(ctx, inputWithExaId)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (testDbServersListDataSource) testAccDbServersListDataSourceConfigBasic(exaInfra string) string {
	return fmt.Sprintf(`
%s

data "aws_odb_db_servers_list" "test" {
  cloud_exadata_infrastructure_id = aws_odb_cloud_exadata_infrastructure.test.id
}
`, exaInfra)
}

func (testDbServersListDataSource) exaInfra(rName string) string {
	exaRes := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = "%[1]s"
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
  customer_contacts_to_send_to_oci = ["abc@example.com"]
  maintenance_window = {
  		custom_action_timeout_in_mins = 16
		days_of_week =	[]
        hours_of_day =	[]
        is_custom_action_timeout_enabled = true
        lead_time_in_weeks = 0
        months = []
        patching_mode = "ROLLING"
        preference = "NO_PREFERENCE"
		weeks_of_month =[]
  }
}
`, rName)
	return exaRes
}
