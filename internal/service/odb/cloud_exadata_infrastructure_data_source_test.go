// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	"testing"
)

// Acceptance test access AWS and cost money to run.
func TestAccODBCloudExadataInfrastructureDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	exaInfraResource := "aws_odb_cloud_exadata_infrastructure.test"
	exaInfraDataSource := "data.aws_odb_cloud_exadata_infrastructure.test"
	displayNameSuffix := sdkacctest.RandomWithPrefix("tf_")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCloudExadataInfrastructureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: basicExaInfraDataSource(displayNameSuffix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(exaInfraResource, "id", exaInfraDataSource, "id"),
					resource.TestCheckResourceAttr(exaInfraDataSource, "shape", "Exadata.X9M"),
					resource.TestCheckResourceAttr(exaInfraDataSource, "status", "AVAILABLE"),
					resource.TestCheckResourceAttr(exaInfraDataSource, "storage_count", "3"),
					resource.TestCheckResourceAttr(exaInfraDataSource, "compute_count", "2"),
				),
			},
		},
	})
}

func testAccCheckCloudExadataInfrastructureDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_exadata_infrastructure" {
				continue
			}
			_, err := tfodb.FindOdbExaDataInfraForDataSourceByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudExadataInfrastructure, rs.Primary.ID, err)
			}

			return create.Error(names.ODB, create.ErrActionCheckingDestroyed, tfodb.ResNameCloudExadataInfrastructure, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func basicExaInfraDataSource(displayNameSuffix string) string {

	testData := fmt.Sprintf(`


resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = "Ofake_exa_%[1]s"
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

data "aws_odb_cloud_exadata_infrastructure" "test" {
    id = aws_odb_cloud_exadata_infrastructure.test.id
}
`, displayNameSuffix)
	//fmt.Println(testData)
	return testData
}
