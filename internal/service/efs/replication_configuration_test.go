package efs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEFSReplicationConfiguration_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	replicationName := "aws_efs_replication_configuration.test"
	efsName := "aws_efs_file_system.test"
	region := acctest.Region()

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckEfsReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationBasic(region),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(replicationName, "creation_time"),
					resource.TestCheckResourceAttrPair(replicationName, "original_source_file_system_arn", efsName, "arn"),
					resource.TestCheckResourceAttrPair(replicationName, "source_file_system_arn", efsName, "arn"),
					resource.TestCheckResourceAttrPair(replicationName, "source_file_system_id", efsName, "id"),
					resource.TestCheckResourceAttr(replicationName, "source_file_system_region", region),
					resource.TestCheckResourceAttr(replicationName, "destination.#", "1"),
					resource.TestMatchResourceAttr(replicationName, "destination.0.file_system_id", regexp.MustCompile(`fs-.+`)),
					resource.TestCheckResourceAttr(replicationName, "destination.0.region", region),
					resource.TestCheckResourceAttr(replicationName, "destination.0.status", efs.ReplicationStatusEnabled),
					//cleanupOtherRegion(alternateRegion, v.Destinations[0].FileSystemId) //TODO ??
				),
			},
			{
				ResourceName:      replicationName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEFSReplicationConfiguration_allAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	replicationName := "aws_efs_replication_configuration.test"
	efsName := "aws_efs_file_system.test"
	kmsName := "aws_kms_key.test"
	alternateRegion := acctest.AlternateRegion()
	azName := alternateRegion + "a"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckEfsReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationFull(azName, alternateRegion),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(replicationName, "creation_time"),
					resource.TestCheckResourceAttrPair(replicationName, "original_source_file_system_arn", efsName, "arn"),
					resource.TestCheckResourceAttrPair(replicationName, "source_file_system_arn", efsName, "arn"),
					resource.TestCheckResourceAttrPair(replicationName, "source_file_system_id", efsName, "id"),
					resource.TestCheckResourceAttr(replicationName, "source_file_system_region", acctest.Region()),
					resource.TestCheckResourceAttr(replicationName, "destination.#", "1"),
					resource.TestCheckResourceAttr(replicationName, "destination.0.availability_zone_name", azName),
					resource.TestMatchResourceAttr(replicationName, "destination.0.file_system_id", regexp.MustCompile(`fs-.+`)),
					resource.TestCheckResourceAttrPair(replicationName, "destination.0.kms_key_id", kmsName, "key_id"),
					resource.TestCheckResourceAttr(replicationName, "destination.0.region", alternateRegion),
					resource.TestCheckResourceAttr(replicationName, "destination.0.status", efs.ReplicationStatusEnabled),
					//cleanupOtherRegion(alternateRegion, v.Destinations[0].FileSystemId) //TODO ??
				),
			},
		},
	})

	//TODO internal/service/cloudfront/distribution_test.go:1329 might be an example for deleting resources
}

func TestAccEFSReplicationConfiguration_disappears(t *testing.T) {
	replicationName := "aws_efs_replication_configuration.test"
	alternateRegion := acctest.AlternateRegion()
	azName := alternateRegion + "a"

	var providers []*schema.Provider

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, efs.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      acctest.CheckWithProviders(testAccCheckEfsReplicationConfigurationDestroy, &providers),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationConfigurationFull(azName, alternateRegion),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceDisappears(acctest.Provider, tfefs.ResourceReplicationConfiguration(), replicationName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEfsReplicationConfigurationDestroy(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*conns.AWSClient).EFSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_replication_configuration" {
			continue
		}

		_, err := tfefs.FindReplicationConfigurationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Replication Configuration for %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccReplicationConfigurationBasic(region string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {}

resource "aws_efs_replication_configuration" "test" {
  source_file_system_id = aws_efs_file_system.test.id

  destination {
    region = %[1]q
  }
}
`, region)
}

func testAccReplicationConfigurationFull(azName, region string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateRegionProvider(),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  provider = "awsalternate"

  description = "terraform test"
}

resource "aws_efs_file_system" "test" {}

resource "aws_efs_replication_configuration" "test" {
  source_file_system_id = aws_efs_file_system.test.id

  destination {
    availability_zone_name = %[1]q
    kms_key_id             = aws_kms_key.test.key_id
    region                 = %[2]q
  }
}
`, azName, region))
}
