package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2EBSSnapshotCopy_basic(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_tags(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccEBSSnapshotCopyTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEBSSnapshotCopyTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_withDescription(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyWithDescriptionConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttr(resourceName, "description", "Copy Snapshot Acceptance Test"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_withRegions(t *testing.T) {
	var providers []*schema.Provider
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyWithRegionsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
				),
			},
		},
	})

}

func TestAccEC2EBSSnapshotCopy_withKMS(t *testing.T) {
	var snapshot ec2.Snapshot
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyWithKMSConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_storageTier(t *testing.T) {
	var v ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyStorageTierConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "archive"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotCopy_disappears(t *testing.T) {
	var snapshot ec2.Snapshot
	resourceName := "aws_ebs_snapshot_copy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotCopyConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &snapshot),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEBSSnapshotCopy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccEBSSnapshotCopyBaseConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), `
data "aws_region" "current" {}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
  }
}
`)
}

func testAccEBSSnapshotCopyConfig() string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(), `
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
}
`)
}

func testAccEBSSnapshotCopyStorageTierConfig() string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(), `
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
  storage_tier       = "archive"
}
`)
}

func testAccEBSSnapshotCopyTags1Config(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(), fmt.Sprintf(`
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
    "%s" = "%s"
  }
}
`, tagKey1, tagValue1))
}

func testAccEBSSnapshotCopyTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(), fmt.Sprintf(`
resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    Name = "testAccEBSSnapshotCopyConfig"
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEBSSnapshotCopyWithDescriptionConfig() string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(), `
resource "aws_ebs_snapshot_copy" "test" {
  description        = "Copy Snapshot Acceptance Test"
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name

  tags = {
    Name = "testAccEBSSnapshotCopyWithDescriptionConfig"
  }
}
`)
}

func testAccEBSSnapshotCopyWithRegionsConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAlternateRegionProvider(), `
data "aws_availability_zones" "alternate_available" {
  provider = "awsalternate"
  state    = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_region" "alternate" {
  provider = "awsalternate"
}

resource "aws_ebs_volume" "test" {
  provider          = "awsalternate"
  availability_zone = data.aws_availability_zones.alternate_available.names[0]
  size              = 1

  tags = {
    Name = "testAccEBSSnapshotCopyWithRegionsConfig"
  }
}

resource "aws_ebs_snapshot" "test" {
  provider  = "awsalternate"
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "testAccEBSSnapshotCopyWithRegionsConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.alternate.name

  tags = {
    Name = "testAccEBSSnapshotCopyWithRegionsConfig"
  }
}
`)
}

func testAccEBSSnapshotCopyWithKMSConfig() string {
	return acctest.ConfigCompose(testAccEBSSnapshotCopyBaseConfig(), `
resource "aws_kms_key" "test" {
  description             = "testAccEBSSnapshotCopyWithKMSConfig"
  deletion_window_in_days = 7
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = aws_ebs_snapshot.test.id
  source_region      = data.aws_region.current.name
  encrypted          = true
  kms_key_id         = aws_kms_key.test.arn

  tags = {
    Name = "testAccEBSSnapshotCopyWithKMSConfig"
  }
}
`)
}
