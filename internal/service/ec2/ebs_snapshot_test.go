package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2EBSSnapshot_basic(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-basic-%s", sdkacctest.RandString(7))
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "standard"),
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

func TestAccEC2EBSSnapshot_storageTier(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-basic-%s", sdkacctest.RandString(7))
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotStorageTierConfig(rName, "archive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "archive"),
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

func TestAccEC2EBSSnapshot_outpost(t *testing.T) {
	var v ec2.Snapshot
	outpostDataSourceName := "data.aws_outposts_outpost.test"
	resourceName := "aws_ebs_snapshot.test"
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-basic-%s", sdkacctest.RandString(7))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotOutpostConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "outpost_arn", outpostDataSourceName, "arn"),
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

func TestAccEC2EBSSnapshot_tags(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-desc-%s", sdkacctest.RandString(7))
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotBasicTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccEBSSnapshotBasicTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEBSSnapshotBasicTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshot_withDescription(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-desc-%s", sdkacctest.RandString(7))
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotWithDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
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

func TestAccEC2EBSSnapshot_withKMS(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-kms-%s", sdkacctest.RandString(7))
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotWithKMSConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_id", kmsKeyResourceName, "arn"),
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

func TestAccEC2EBSSnapshot_disappears(t *testing.T) {
	var v ec2.Snapshot
	rName := fmt.Sprintf("tf-acc-ebs-snapshot-basic-%s", sdkacctest.RandString(7))
	resourceName := "aws_ebs_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEBSSnapshot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		output, err := tfec2.FindSnapshotById(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckEBSSnapshotDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_snapshot" {
			continue
		}

		_, err := tfec2.FindSnapshotById(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EBS Snapshot %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccEBSSnapshotBaseConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotBasicConfig(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), `
resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`)
}

func testAccEBSSnapshotStorageTierConfig(rName, tier string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id    = aws_ebs_volume.test.id
  storage_tier = %[1]q

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, tier))
}

func testAccEBSSnapshotOutpostConfig(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_ebs_snapshot" "test" {
  volume_id   = aws_ebs_volume.test.id
  outpost_arn = data.aws_outposts_outpost.test.arn

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`)
}

func testAccEBSSnapshotBasicTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "%s"
    "%s" = "%s"
  }

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEBSSnapshotBasicTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id

  tags = {
    Name = "%s"
    "%s" = "%s"
    "%s" = "%s"
  }

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEBSSnapshotWithDescriptionConfig(rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotBaseConfig(rName), fmt.Sprintf(`
resource "aws_ebs_snapshot" "test" {
  volume_id   = aws_ebs_volume.test.id
  description = %[1]q
}
`, rName))
}

func testAccEBSSnapshotWithKMSConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  size              = 1
  encrypted         = true
  kms_key_id        = aws_kms_key.test.arn

  tags = {
    Name = %[1]q
  }
}

resource "aws_ebs_snapshot" "test" {
  volume_id = aws_ebs_volume.test.id
}
`, rName))
}
