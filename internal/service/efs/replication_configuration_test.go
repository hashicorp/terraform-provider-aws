// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEFSReplicationConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_efs_replication_configuration.test"
	fsResourceName := "aws_efs_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "destination.#", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "destination.0.file_system_id", regexache.MustCompile(`fs-.+`)),
					resource.TestCheckResourceAttr(resourceName, "destination.0.region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "destination.0.status", string(awstypes.ReplicationStatusEnabled)),
					resource.TestCheckResourceAttrPair(resourceName, "original_source_file_system_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_file_system_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_file_system_id", fsResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "source_file_system_region", acctest.Region()),
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

func TestAccEFSReplicationConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_efs_replication_configuration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationConfigurationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfefs.ResourceReplicationConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSReplicationConfiguration_allAttributes(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_efs_replication_configuration.test"
	fsResourceName := "aws_efs_file_system.test"
	kmsKeyResourceName := "aws_kms_key.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination.0.availability_zone_name", "data.aws_availability_zones.available", "names.0"),
					resource.TestMatchResourceAttr(resourceName, "destination.0.file_system_id", regexache.MustCompile(`fs-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "destination.0.kms_key_id", kmsKeyResourceName, names.AttrKeyID),
					resource.TestCheckResourceAttr(resourceName, "destination.0.region", acctest.AlternateRegion()),
					resource.TestCheckResourceAttr(resourceName, "destination.0.status", string(awstypes.ReplicationStatusEnabled)),
					resource.TestCheckResourceAttrPair(resourceName, "original_source_file_system_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_file_system_arn", fsResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "source_file_system_id", fsResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "source_file_system_region", acctest.Region()),
				),
			},
		},
	})
}

func TestAccEFSReplicationConfiguration_existingDestination(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	resourceName := "aws_efs_replication_configuration.test"
	destinationFsResourceName := "aws_efs_file_system.destination"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EFSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesPlusProvidersAlternate(ctx, t, &providers),
		CheckDestroy:             acctest.CheckWithProviders(testAccCheckReplicationConfigurationDestroyWithProvider(ctx), &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationConfig_existingDestination(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationConfigurationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(resourceName, "destination.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "destination.0.file_system_id", destinationFsResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "destination.0.status", string(awstypes.ReplicationStatusEnabled)),
				),
			},
		},
	})
}

func testAccCheckReplicationConfigurationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		_, err := tfefs.FindReplicationConfigurationByID(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckReplicationConfigurationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return testAccCheckReplicationConfigurationDestroyWithProvider(ctx)(s, acctest.Provider)
	}
}

func testAccCheckReplicationConfigurationDestroyWithProvider(ctx context.Context) acctest.TestCheckWithProviderFunc {
	return func(s *terraform.State, provider *schema.Provider) error {
		conn := provider.Meta().(*conns.AWSClient).EFSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_efs_replication_configuration" {
				continue
			}

			_, err := tfefs.FindReplicationConfigurationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EFS Replication Configuration %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccReplicationConfigurationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_replication_configuration" "test" {
  source_file_system_id = aws_efs_file_system.test.id

  destination {
    region = %[2]q
  }
}
`, rName, acctest.AlternateRegion())
}

func testAccReplicationConfigurationConfig_existingDestination(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_efs_file_system" "source" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_file_system" "destination" {
  provider = "awsalternate"

  protection {
    replication_overwrite = "DISABLED"
  }

  tags = {
    Name = %[1]q
  }

  lifecycle {
    ignore_changes = [protection]
  }
}

resource "aws_efs_replication_configuration" "test" {
  source_file_system_id = aws_efs_file_system.source.id

  destination {
    file_system_id = aws_efs_file_system.destination.id
    region         = %[2]q
  }
}
`, rName, acctest.AlternateRegion()))
}

func testAccReplicationConfigurationConfig_full(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider = "awsalternate"

  description             = %[1]q
  deletion_window_in_days = 7
}

data "aws_availability_zones" "available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_efs_replication_configuration" "test" {
  source_file_system_id = aws_efs_file_system.test.id

  destination {
    availability_zone_name = data.aws_availability_zones.available.names[0]
    kms_key_id             = aws_kms_key.test.key_id
    region                 = %[2]q
  }
}
`, rName, acctest.AlternateRegion()))
}
