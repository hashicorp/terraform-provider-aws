// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Acceptance test access AWS and cost money to run.
type cloudExaDataInfraDataSourceTest struct {
	displayNamePrefix string
}

var exaInfraDataSourceTestEntity = cloudExaDataInfraDataSourceTest{
	displayNamePrefix: "Ofake-exa",
}

func TestAccODBCloudExadataInfrastructureDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	exaInfraResource := "aws_odb_cloud_exadata_infrastructure.test"
	exaInfraDataSource := "data.aws_odb_cloud_exadata_infrastructure.test"
	displayNameSuffix := sdkacctest.RandomWithPrefix(exaInfraDataSourceTestEntity.displayNamePrefix)
	domain := acctest.RandomDomainName()
	emailAddress1 := acctest.RandomEmailAddress(domain)
	emailAddress2 := acctest.RandomEmailAddress(domain)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             exaInfraDataSourceTestEntity.testAccCheckCloudExadataInfrastructureDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: exaInfraDataSourceTestEntity.basicExaInfraDataSource(displayNameSuffix, emailAddress1, emailAddress2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(exaInfraResource, names.AttrID, exaInfraDataSource, names.AttrID),
					resource.TestCheckResourceAttr(exaInfraDataSource, "shape", "Exadata.X9M"),
					resource.TestCheckResourceAttr(exaInfraDataSource, names.AttrStatus, "AVAILABLE"),
					resource.TestCheckResourceAttr(exaInfraDataSource, "storage_count", "3"),
					resource.TestCheckResourceAttr(exaInfraDataSource, "compute_count", "2"),
				),
			},
		},
	})
}

func (cloudExaDataInfraDataSourceTest) testAccCheckCloudExadataInfrastructureDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_exadata_infrastructure" {
				continue
			}
			_, err := tfodb.FindExaDataInfraForDataSourceByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
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

func (cloudExaDataInfraDataSourceTest) basicExaInfraDataSource(displayNameSuffix, emailAddress1, emailAddress2 string) string {
	testData := fmt.Sprintf(`




resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name                     = %[1]q
  shape                            = "Exadata.X9M"
  storage_count                    = 3
  compute_count                    = 2
  availability_zone_id             = "use1-az6"
  customer_contacts_to_send_to_oci = [{ email = "%[2]s" }, { email = "%[3]s" }]
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    is_custom_action_timeout_enabled = true
    patching_mode                    = "ROLLING"
    preference                       = "NO_PREFERENCE"
  }
}

data "aws_odb_cloud_exadata_infrastructure" "test" {
  id = aws_odb_cloud_exadata_infrastructure.test.id
}
`, displayNameSuffix, emailAddress1, emailAddress2)
	return testData
}
