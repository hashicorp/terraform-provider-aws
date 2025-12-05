// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Acceptance test access AWS and cost money to run.

type cloudExaDataInfraResourceTest struct {
	displayNamePrefix string
}

var exaInfraTestResource = cloudExaDataInfraResourceTest{
	displayNamePrefix: "Ofake-exa",
}

func TestAccODBCloudExadataInfrastructureResource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudExaDataInfrastructure odbtypes.CloudExadataInfrastructure
	resourceName := "aws_odb_cloud_exadata_infrastructure.test"
	rName := sdkacctest.RandomWithPrefix(exaInfraTestResource.displayNamePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			exaInfraTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             exaInfraTestResource.testAccCheckCloudExaDataInfraDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure),
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
func TestAccODBCloudExadataInfrastructureResource_withAllParameters(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudExaDataInfrastructure odbtypes.CloudExadataInfrastructure
	resourceName := "aws_odb_cloud_exadata_infrastructure.test"
	rName := sdkacctest.RandomWithPrefix(exaInfraTestResource.displayNamePrefix)
	domain := acctest.RandomDomainName()
	emailAddress1 := acctest.RandomEmailAddress(domain)
	emailAddress2 := acctest.RandomEmailAddress(domain)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			exaInfraTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             exaInfraTestResource.testAccCheckCloudExaDataInfraDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: exaInfraTestResource.exaDataInfraResourceWithAllConfig(rName, emailAddress1, emailAddress2),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure),
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

func TestAccODBCloudExadataInfrastructureResource_tagging(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudExaDataInfrastructure1 odbtypes.CloudExadataInfrastructure
	var cloudExaDataInfrastructure2 odbtypes.CloudExadataInfrastructure
	resourceName := "aws_odb_cloud_exadata_infrastructure.test"
	rName := sdkacctest.RandomWithPrefix(exaInfraTestResource.displayNamePrefix)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			exaInfraTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             exaInfraTestResource.testAccCheckCloudExaDataInfraDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfigWithTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure1),
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
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudExaDataInfrastructure1.CloudExadataInfrastructureId), *(cloudExaDataInfrastructure2.CloudExadataInfrastructureId)) != 0 {
							return errors.New("Should not  create a new cloud exa basicExaInfraDataSource after  update")
						}
						return nil
					}),
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

func TestAccODBCloudExadataInfrastructureResource_updateDisplayName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudExaDataInfrastructure1 odbtypes.CloudExadataInfrastructure
	var cloudExaDataInfrastructure2 odbtypes.CloudExadataInfrastructure
	resourceName := "aws_odb_cloud_exadata_infrastructure.test"
	rName := sdkacctest.RandomWithPrefix(exaInfraTestResource.displayNamePrefix)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			exaInfraTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             exaInfraTestResource.testAccCheckCloudExaDataInfraDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfig(rName + "-u"),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudExaDataInfrastructure1.CloudExadataInfrastructureId), *(cloudExaDataInfrastructure2.CloudExadataInfrastructureId)) == 0 {
							return errors.New("Should   create a new cloud exa basicExaInfraDataSource after update")
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

func TestAccODBCloudExadataInfrastructureResource_updateMaintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudExaDataInfrastructure1 odbtypes.CloudExadataInfrastructure
	var cloudExaDataInfrastructure2 odbtypes.CloudExadataInfrastructure
	resourceName := "aws_odb_cloud_exadata_infrastructure.test"
	rName := sdkacctest.RandomWithPrefix(exaInfraTestResource.displayNamePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			exaInfraTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             exaInfraTestResource.testAccCheckCloudExaDataInfraDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: exaInfraTestResource.basicWithCustomMaintenanceWindow(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudExaDataInfrastructure1.CloudExadataInfrastructureId), *(cloudExaDataInfrastructure2.CloudExadataInfrastructureId)) != 0 {
							return errors.New("Should not  create a new cloud exa basicExaInfraDataSource after  update")
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

func TestAccODBCloudExadataInfrastructureResource_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudExaDataInfrastructure odbtypes.CloudExadataInfrastructure

	rName := sdkacctest.RandomWithPrefix(exaInfraTestResource.displayNamePrefix)
	resourceName := "aws_odb_cloud_exadata_infrastructure.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			exaInfraTestResource.testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             exaInfraTestResource.testAccCheckCloudExaDataInfraDestroyed(ctx),
		Steps: []resource.TestStep{
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfodb.ResourceCloudExadataInfrastructure, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func (cloudExaDataInfraResourceTest) testAccCheckCloudExaDataInfraDestroyed(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_odb_cloud_exadata_infrastructure" {
				continue
			}
			_, err := tfodb.FindExadataInfraResourceByID(ctx, conn, rs.Primary.ID)
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

func (cloudExaDataInfraResourceTest) testAccCheckCloudExadataInfrastructureExists(ctx context.Context, name string, cloudExadataInfrastructure *odbtypes.CloudExadataInfrastructure) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudExadataInfrastructure, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudExadataInfrastructure, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

		resp, err := tfodb.FindExadataInfraResourceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudExadataInfrastructure, rs.Primary.ID, err)
		}

		*cloudExadataInfrastructure = *resp

		return nil
	}
}

func (cloudExaDataInfraResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

	input := odb.ListCloudExadataInfrastructuresInput{}

	_, err := conn.ListCloudExadataInfrastructures(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func (cloudExaDataInfraResourceTest) exaDataInfraResourceWithAllConfig(randomId, emailAddress1, emailAddress2 string) string {
	exaDataInfra := fmt.Sprintf(`


resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name                     = %[1]q
  shape                            = "Exadata.X11M"
  storage_count                    = 3
  compute_count                    = 2
  availability_zone_id             = "use1-az6"
  customer_contacts_to_send_to_oci = [{ email = "%[2]s" }, { email = "%[3]s" }]
  database_server_type             = "X11M"
  storage_server_type              = "X11M-HC"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    days_of_week                     = [{ name = "MONDAY" }, { name = "TUESDAY" }]
    hours_of_day                     = [11, 16]
    is_custom_action_timeout_enabled = true
    lead_time_in_weeks               = 3
    months                           = [{ name = "FEBRUARY" }, { name = "MAY" }, { name = "AUGUST" }, { name = "NOVEMBER" }]
    patching_mode                    = "ROLLING"
    preference                       = "CUSTOM_PREFERENCE"
    weeks_of_month                   = [2, 4]
  }
  tags = {
    "env" = "dev"
  }

}
`, randomId, emailAddress1, emailAddress2)
	return exaDataInfra
}
func (cloudExaDataInfraResourceTest) exaDataInfraResourceBasicConfig(displayName string) string {
	exaInfra := fmt.Sprintf(`
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
`, displayName)
	return exaInfra
}
func (cloudExaDataInfraResourceTest) exaDataInfraResourceBasicConfigWithTags(displayName string) string {
	exaInfra := fmt.Sprintf(`
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
  tags = {
    "env" = "dev"
  }
}
`, displayName)
	return exaInfra
}

func (cloudExaDataInfraResourceTest) basicWithCustomMaintenanceWindow(displayName string) string {
	exaInfra := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name         = %[1]q
  shape                = "Exadata.X9M"
  storage_count        = 3
  compute_count        = 2
  availability_zone_id = "use1-az6"
  maintenance_window {
    custom_action_timeout_in_mins    = 16
    days_of_week                     = [{ name = "MONDAY" }, { name = "TUESDAY" }]
    hours_of_day                     = [11, 16]
    is_custom_action_timeout_enabled = true
    lead_time_in_weeks               = 3
    months                           = [{ name = "FEBRUARY" }, { name = "MAY" }, { name = "AUGUST" }, { name = "NOVEMBER" }]
    patching_mode                    = "ROLLING"
    preference                       = "CUSTOM_PREFERENCE"
    weeks_of_month                   = [2, 4]
  }
}
`, displayName)
	return exaInfra
}
