// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.AppStreamServiceID, testAccErrorCheckSkip)
}

// testAccErrorCheckSkip skips AppStream tests that have error messages indicating unsupported features
func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"ResourceNotFoundException: The image",
		"InvalidParameterValueException: The AppStream 2.0 user pool feature",
	)
}

func TestAccAppStreamFleet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var fleetOutput awstypes.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, "idle_disconnect_timeout_in_seconds", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "stream_view", string(awstypes.StreamViewApp)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
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

func TestAccAppStreamFleet_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var fleetOutput awstypes.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_basic(rName, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceFleet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAppStreamFleet_completeWithStop(t *testing.T) {
	ctx := acctest.Context(t)
	var fleetOutput awstypes.Fleet
	resourceName := "aws_appstream_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	descriptionUpdated := "Updated Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	instanceTypeUpdate := "stream.standard.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_complete(rName, description, fleetType, instanceType),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "stream_view", string(awstypes.StreamViewDesktop)),
				),
			},
			{
				Config: testAccFleetConfig_complete(rName, descriptionUpdated, fleetType, instanceTypeUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceTypeUpdate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionUpdated),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, "stream_view", string(awstypes.StreamViewDesktop)),
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

func TestAccAppStreamFleet_completeWithoutStop(t *testing.T) {
	ctx := acctest.Context(t)
	var fleetOutput awstypes.Fleet
	resourceName := "aws_appstream_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	displayName := "display name of a test"
	displayNameUpdated := "display name of a test updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_completeNoStopping(rName, description, fleetType, instanceType, displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayName),
				),
			},
			{
				Config: testAccFleetConfig_completeNoStopping(rName, description, fleetType, instanceType, displayNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, names.AttrDisplayName, displayNameUpdated),
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

func TestAccAppStreamFleet_withTags(t *testing.T) {
	ctx := acctest.Context(t)
	var fleetOutput awstypes.Fleet
	resourceName := "aws_appstream_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description := "Description of a test"
	fleetType := "ON_DEMAND"
	instanceType := "stream.standard.small"
	displayName := "display name of a test"
	displayNameUpdated := "display name of a test updated"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_tags(rName, description, fleetType, instanceType, displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", names.AttrValue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", names.AttrValue),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
			{
				Config: testAccFleetConfig_tags(rName, description, fleetType, instanceType, displayNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, description),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Key", names.AttrValue),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Key", names.AttrValue),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
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

func TestAccAppStreamFleet_emptyDomainJoin(t *testing.T) {
	ctx := acctest.Context(t)
	var fleetOutput awstypes.Fleet
	resourceName := "aws_appstream_fleet.test"
	instanceType := "stream.standard.small"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_emptyDomainJoin(rName, instanceType, `""`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, "stream_view", string(awstypes.StreamViewApp)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
			{
				Config: testAccFleetConfig_emptyDomainJoin(rName, instanceType, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					resource.TestCheckResourceAttr(resourceName, "stream_view", string(awstypes.StreamViewApp)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
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

func TestAccAppStreamFleet_multiSession(t *testing.T) {
	ctx := acctest.Context(t)
	var fleetOutput awstypes.Fleet
	resourceName := "aws_appstream_fleet.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	instanceType := "stream.standard.small"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig_multiSession(rName, instanceType, 1, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, "max_sessions_per_instance", "5"),
					resource.TestCheckResourceAttr(resourceName, "compute_capacity.0.desired_sessions", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFleetConfig_multiSession(rName, instanceType, 2, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(ctx, resourceName, &fleetOutput),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceType, instanceType),
					resource.TestCheckResourceAttr(resourceName, "max_sessions_per_instance", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "compute_capacity.0.desired_sessions", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.FleetStateRunning)),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedTime),
				),
			},
		},
	})
}

func testAccCheckFleetExists(ctx context.Context, resourceName string, appStreamFleet *awstypes.Fleet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)
		resp, err := conn.DescribeFleets(ctx, &appstream.DescribeFleetsInput{Names: []string{rs.Primary.ID}})

		if err != nil {
			return err
		}

		if resp == nil || len(resp.Fleets) == 0 {
			return fmt.Errorf("appstream fleet %q does not exist", rs.Primary.ID)
		}

		*appStreamFleet = resp.Fleets[0]

		return nil
	}
}

func testAccCheckFleetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_fleet" {
				continue
			}

			resp, err := conn.DescribeFleets(ctx, &appstream.DescribeFleetsInput{Names: []string{rs.Primary.ID}})

			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				continue
			}

			if err != nil {
				return err
			}

			if resp != nil && len(resp.Fleets) > 0 {
				return fmt.Errorf("appstream fleet %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccFleetConfig_basic(name, instanceType string) string {
	// "Amazon-AppStream2-Sample-Image-03-11-2023" is not available in GovCloud
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test" {
  name          = %[1]q
  image_name    = "Amazon-AppStream2-Sample-Image-03-11-2023"
  instance_type = %[2]q

  compute_capacity {
    desired_instances = 1
  }
}
`, name, instanceType)
}

func testAccFleetConfig_complete(name, description, fleetType, instanceType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_fleet" "test" {
  name      = %[1]q
  image_arn = "arn:${data.aws_partition.current.partition}:appstream:${data.aws_region.current.name}::image/Amazon-AppStream2-Sample-Image-03-11-2023"

  compute_capacity {
    desired_instances = 1
  }

  description                        = %[2]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[3]q
  instance_type                      = %[4]q
  max_user_duration_in_seconds       = 1000
  stream_view                        = "DESKTOP"

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }
}
`, name, description, fleetType, instanceType))
}

func testAccFleetConfig_completeNoStopping(name, description, fleetType, instanceType, displayName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_fleet" "test" {
  name       = %[1]q
  image_name = "Amazon-AppStream2-Sample-Image-03-11-2023"

  compute_capacity {
    desired_instances = 1
  }

  description                        = %[2]q
  display_name                       = %[5]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[3]q
  instance_type                      = %[4]q
  max_user_duration_in_seconds       = 1000

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }
}
`, name, description, fleetType, instanceType, displayName))
}

func testAccFleetConfig_tags(name, description, fleetType, instanceType, displayName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_fleet" "test" {
  name       = %[1]q
  image_name = "Amazon-AppStream2-Sample-Image-03-11-2023"

  compute_capacity {
    desired_instances = 1
  }

  description                        = %[2]q
  display_name                       = %[5]q
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = %[3]q
  instance_type                      = %[4]q
  max_user_duration_in_seconds       = 1000

  tags = {
    Key = "value"
  }

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }
}
`, name, description, fleetType, instanceType, displayName))
}

func testAccFleetConfig_emptyDomainJoin(name, instanceType, empty string) string {
	// "Amazon-AppStream2-Sample-Image-03-11-2023" is not available in GovCloud
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test" {
  name          = %[1]q
  image_name    = "Amazon-AppStream2-Sample-Image-03-11-2023"
  instance_type = %[2]q

  compute_capacity {
    desired_instances = 1
  }

  domain_join_info {
    directory_name                         = %[3]s
    organizational_unit_distinguished_name = %[3]s
  }
}
`, name, instanceType, empty)
}

func testAccFleetConfig_multiSession(name, instanceType string, desiredSessions, maxSessionsPerInstance int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id
}

resource "aws_appstream_fleet" "test" {
  name      = %[1]q
  image_arn = "arn:${data.aws_partition.current.partition}:appstream:${data.aws_region.current.name}::image/AppStream-WinServer2019-01-26-2024"

  compute_capacity {
    desired_sessions = %[3]d
  }

  description                        = "Description for a multi-session fleet"
  idle_disconnect_timeout_in_seconds = 70
  enable_default_internet_access     = false
  fleet_type                         = "ON_DEMAND"
  instance_type                      = %[2]q
  max_sessions_per_instance          = %[4]d
  max_user_duration_in_seconds       = 1000
  stream_view                        = "DESKTOP"

  vpc_config {
    subnet_ids = aws_subnet.test[*].id
  }
}
`, name, instanceType, desiredSessions, maxSessionsPerInstance))
}
