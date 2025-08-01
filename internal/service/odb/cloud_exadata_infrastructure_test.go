//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb_test

import (
	"context"
	"errors"
	"fmt"
	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfodb "github.com/hashicorp/terraform-provider-aws/internal/service/odb"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Acceptance test access AWS and cost money to run.

type cloudExaDataInfraResourceTest struct {
	displayNamePrefix string
}

var exaInfraTestResource = cloudExaDataInfraResourceTest{
	displayNamePrefix: "Ofake-exa",
}

func TestAccODBCloudExadataInfrastructureCreate_basic(t *testing.T) {
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
func TestAccODBCloudExadataInfrastructureCreateWithAllParameters(t *testing.T) {
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
				Config: exaInfraTestResource.exaDataInfraResourceWithAllConfig(rName),
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

func TestAccODBCloudExadataInfrastructureTagging(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cloudExaDataInfrastructure1 odbtypes.CloudExadataInfrastructure
	var cloudExaDataInfrastructure2 odbtypes.CloudExadataInfrastructure
	var cloudExaDataInfrastructure3 odbtypes.CloudExadataInfrastructure
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
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfigAddTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure2),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudExaDataInfrastructure1.CloudExadataInfrastructureId), *(cloudExaDataInfrastructure2.CloudExadataInfrastructureId)) != 0 {
							return errors.New("Should not  create a new cloud exa basicExaInfraDataSource after  update")
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.env", "dev"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: exaInfraTestResource.exaDataInfraResourceBasicConfigRemoveTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					exaInfraTestResource.testAccCheckCloudExadataInfrastructureExists(ctx, resourceName, &cloudExaDataInfrastructure3),
					resource.ComposeTestCheckFunc(func(state *terraform.State) error {
						if strings.Compare(*(cloudExaDataInfrastructure1.CloudExadataInfrastructureId), *(cloudExaDataInfrastructure3.CloudExadataInfrastructureId)) != 0 {
							return errors.New("Should not  create a new cloud exa basicExaInfraDataSource after  update")
						}
						return nil
					}),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccODBCloudExadataUpdateMaintenanceWindow(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "maintenance_window.preference", "NO_PREFERENCE"),
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
					resource.TestCheckResourceAttr(resourceName, "maintenance_window.preference", "CUSTOM_PREFERENCE"),
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

func TestAccODBCloudExadataInfrastructure_disappears(t *testing.T) {
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
			_, err := tfodb.FindOdbExadataInfraResourceByID(ctx, conn, rs.Primary.ID)
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

		resp, err := tfodb.FindOdbExadataInfraResourceByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ODB, create.ErrActionCheckingExistence, tfodb.ResNameCloudExadataInfrastructure, rs.Primary.ID, err)
		}

		*cloudExadataInfrastructure = *resp

		return nil
	}
}

func (cloudExaDataInfraResourceTest) testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ODBClient(ctx)

	input := &odb.ListCloudExadataInfrastructuresInput{}

	_, err := conn.ListCloudExadataInfrastructures(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

/*
	func testAccCheckCloudExadataInfrastructureNotRecreated(before, after *odb.DescribeCloudExadataInfrastructureResponse) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			if before, after := aws.ToString(before.CloudExadataInfrastructureId), aws.ToString(after.CloudExadataInfrastructureId); before != after {
				return create.Error(names.ODB, create.ErrActionCheckingNotRecreated, tfodb.ResNameCloudExadataInfrastructure, aws.ToString(before.CloudExadataInfrastructureId), errors.New("recreated"))
			}

			return nil
		}
	}
*/
func (cloudExaDataInfraResourceTest) exaDataInfraResourceWithAllConfig(randomId string) string {
	exaDataInfra := fmt.Sprintf(`

resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = %[1]q
  shape             	= "Exadata.X11M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
  customer_contacts_to_send_to_oci = ["abc@example.com"]
  database_server_type = "X11M"
  storage_server_type = "X11M-HC"
  maintenance_window = {
  		custom_action_timeout_in_mins = 16
		days_of_week =	["MONDAY", "TUESDAY"]
        hours_of_day =	[11,16]
        is_custom_action_timeout_enabled = true
        lead_time_in_weeks = 3
        months = ["FEBRUARY","MAY","AUGUST","NOVEMBER"]
        patching_mode = "ROLLING"
        preference = "CUSTOM_PREFERENCE"
		weeks_of_month =[2,4]
  }
  tags = {
    "env"= "dev"
  }

}
`, randomId)
	return exaDataInfra
}
func (cloudExaDataInfraResourceTest) exaDataInfraResourceBasicConfig(displayName string) string {
	exaInfra := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = %[1]q
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
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
`, displayName)
	return exaInfra
}
func (cloudExaDataInfraResourceTest) exaDataInfraResourceBasicConfigAddTags(displayName string) string {
	exaInfra := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = %[1]q
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
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
   tags = {
    "env"= "dev"
  }
}
`, displayName)
	return exaInfra
}

func (cloudExaDataInfraResourceTest) exaDataInfraResourceBasicConfigRemoveTags(displayName string) string {
	exaInfra := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = %[1]q
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
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
`, displayName)
	return exaInfra
}
func (cloudExaDataInfraResourceTest) basicWithCustomMaintenanceWindow(displayName string) string {
	exaInfra := fmt.Sprintf(`
resource "aws_odb_cloud_exadata_infrastructure" "test" {
  display_name          = %[1]q
  shape             	= "Exadata.X9M"
  storage_count      	= 3
  compute_count         = 2
  availability_zone_id 	= "use1-az6"
  maintenance_window = {
  		custom_action_timeout_in_mins = 16
		days_of_week =	["MONDAY", "TUESDAY"]
        hours_of_day =	[11,16]
        is_custom_action_timeout_enabled = true
        lead_time_in_weeks = 3
        months = ["FEBRUARY","MAY","AUGUST","NOVEMBER"]
        patching_mode = "ROLLING"
        preference = "CUSTOM_PREFERENCE"
		weeks_of_month =[2,4]
  }
}
`, displayName)
	return exaInfra
}
