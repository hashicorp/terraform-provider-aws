package ec2_test

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Cannot run acceptance tests without a "real" image. Smallest image I could find was 385Mb
// Instead run tests and catch the failed state
func TestAccEC2ImageImport_badFileFormat(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_image_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageImportConfig_basic(t, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
				ExpectError: regexp.MustCompile(`create: unexpected state 'deleted', wanted target 'completed'. last error: ClientError: No valid partitions. Not a valid volume`),
			},
		},
	})
}

func TestAccEC2ImageImport_disk_container_conflict(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_image_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageImportConfig_disk_container_conflict(t, rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
				ExpectError: regexp.MustCompile(`url and user_bucket cannot be set on the same disk container`),
			},
		},
	})
}

func TestAccEC2ImageImport_boot_mode_bad_string(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_image_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageImportConfig_boot_mode_bad_string(t, rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
				ExpectError: regexp.MustCompile(`expected boot_mode to be one o`),
			},
		},
	})
}

func TestAccEC2ImageImport_no_os(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_image_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageImportConfig_no_os(t, rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
				ExpectError: regexp.MustCompile(`create: unexpected state 'deleted', wanted target 'completed'. last error: ClientError: Unknown OS / Missing OS files`),
			},
		},
	})
}

func TestAccEC2ImageImport_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ec2_image_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMIDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageImportConfig_tags1(t, rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
				ExpectError: regexp.MustCompile(`create: unexpected state 'deleted', wanted target 'completed'. last error: ClientError: No valid partitions. Not a valid volume`),
			},
			{
				Config: testAccImageImportConfig_tags2(t, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
				ExpectError: regexp.MustCompile(`create: unexpected state 'deleted', wanted target 'completed'. last error: ClientError: No valid partitions. Not a valid volume`),
			},
			{
				Config: testAccImageImportConfig_tags1(t, rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value3"),
				),
				ExpectError: regexp.MustCompile(`create: unexpected state 'deleted', wanted target 'completed'. last error: ClientError: No valid partitions. Not a valid volume`),
			},
		},
	})
}

func testAccImageImportNonBucketConfig(t *testing.T, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

# The following resources are for the *vmimport service user*
# See: https://docs.aws.amazon.com/vm-import/latest/userguide/vmie_prereqs.html#vmimport-role
resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.vmimport-trust.json
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.vmimport-access.json
}

data "aws_iam_policy_document" "vmimport-access" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "ec2:ModifySnapshotAttribute",
      "ec2:CopySnapshot",
      "ec2:RegisterImage",
      "ec2:Describe*"
    ]
    resources = [
      "*"
    ]
  }
}

data "aws_iam_policy_document" "vmimport-trust" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["vmie.amazonaws.com"]
    }

    actions = [
      "sts:AssumeRole"
    ]

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = ["vmimport"]
    }
  }
}
`, rName)
}

func testAccImageImportBaseConfig(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccImageImportNonBucketConfig(t, rName), fmt.Sprintf(`
resource "aws_s3_object" "test" {
  bucket         = aws_s3_bucket.test.id
  key            = "diskimage.vhd"
  content_base64 = %[1]q
}
`, testAccImageDisk(t)))
}

func testAccImageImportBaseConfigNoOs(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccImageImportNonBucketConfig(t, rName), fmt.Sprintf(`
resource "aws_s3_object" "test" {
  bucket         = aws_s3_bucket.test.id
  key            = "diskimage-no-os.vhd"
  content_base64 = %[1]q
}
`, testAccImageDiskNoOS(t)))
}

func testAccImageImportConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccImageImportBaseConfig(t, rName), `
resource "aws_ec2_image_import" "test" {
  disk_container {
    description = "test"
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.test.id
      s3_key    = aws_s3_object.test.key
    }
  }

  role_name = aws_iam_role.test.name
}
`)
}

func testAccImageImportConfig_no_os(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccImageImportBaseConfigNoOs(t, rName), `
resource "aws_ec2_image_import" "test" {
  disk_container {
    description = "test"
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.test.id
      s3_key    = aws_s3_object.test.key
    }
  }

  role_name = aws_iam_role.test.name
}
`)
}

func testAccImageImportConfig_disk_container_conflict(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccImageImportBaseConfig(t, rName), fmt.Sprintf(`
resource "aws_ec2_image_import" "test" {
  disk_container {
    description = "test"
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.test.id
      s3_key    = aws_s3_object.test.key
    }
    url = "url"
  }

  role_name = aws_iam_role.test.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccImageImportConfig_boot_mode_bad_string(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccImageImportBaseConfig(t, rName), fmt.Sprintf(`
resource "aws_ec2_image_import" "test" {
  disk_container {
    description = "test"
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.test.id
      s3_key    = aws_s3_object.test.key
    }
  }

  boot_mode = "not-uefi"

  role_name = aws_iam_role.test.name

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccImageImportConfig_tags1(t *testing.T, rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccImageImportBaseConfig(t, rName), fmt.Sprintf(`
resource "aws_ec2_image_import" "test" {
  disk_container {
    description = "test"
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.test.id
      s3_key    = aws_s3_object.test.key
    }
  }

  role_name = aws_iam_role.test.name

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccImageImportConfig_tags2(t *testing.T, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccImageImportBaseConfig(t, rName), fmt.Sprintf(`
resource "aws_ec2_image_import" "test" {
  disk_container {
    description = "test"
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.test.id
      s3_key    = aws_s3_object.test.key
    }
  }

  role_name = aws_iam_role.test.name

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccImageDisk(t *testing.T) string {
	// Take a compressed then base64'd disk image,
	// base64 decode, then decompress, then re-base64
	// the image, so it can be uploaded to s3.

	// little vmdk built by:
	// $ VBoxManage createmedium disk --filename ./image.vhd --sizebytes 512 --format vhd
	// $ cat image.vhd | gzip --best | base64
	b64_compressed := "H4sIAAAAAAACA0vOz0tNLsmsYGBgYGJgZIACJgZ1789hZUn5FQxsDIzhmUbZMHEEzSIIJJj///+QlV1rMXFVnLzHwteXYmWDDfYxjIIhA5IrigsSi4pT/0MBRJSNAZoWGBkUGBj+//9SNhpSo2AUDD+AyPOjYESW/6P1/4gGAAvDpVcACgAA"

	// TODO: They only support certain linux flavors. And it checks for a full operating system.
	// How can we import this? Pull from public bucket?
	// Expect fail?
	// little filesystem built by:
	// $ sudo dd if=/dev/zero of=vhd.img bs=4k count=55
	// $ sudo mkfs -t ext3 vhd.img
	// $ qemu-img convert -f raw -O vpc vhd.img image.vhd
	// $ cat image.vhd | gzip --best | base64
	// b64_compressed := "H4sIAAAAAAACA+3cz0tUQRwA8Nm3u265gnrw0smzhQch6pQKXQJR1INea9toMTVTQwLBDkbXzp1Cow5hQoeELp78Q4LQi9TBwIu89u2uS4qnyNL184GZNzO8/fGGYZgvzLzC9FSxMFdaCCFEIRVqotA1OLI6U5ycD9mQHiv1TFSa0xuhfs1l2pNSHO8erG9vLWaHh5ZX9tdGDwbGewPnRmFh9vHdJ7PFuKba2hRqYyEVOkOI472negoaUiYGLiTTH1xc5eV9uFGJ+EK4Vk4d9QiwmsJSNXXW2t+/+nIviQhGv6cqIUK1HmrBQlW+9rGbte9Il9PDZ19f73we+vCxeS2//PPKy//xrLe/rb7L99/59PbWjxfX299sJv+35chz/X0pQwwAgDPkcH2eqaz/O8r1jE4BAACABhPnknzJRggAAABoZLlkNy8AAADQyA73ASTnXw/Tv9x/sNtfztpO+v10uPTbfVlbNTgFS8/L2WbfCeMvdWT8/YnmY3Vnwc+ezWT+6Ttp/onq7zxINJVTslUqGROXdRsAAOdUcv6/JaSi7no5irq7q+/w2mltjh5Nz85dfTA9P3VfXwEAAMB5lT8W/++1VuN/AAAAoMG06QIAAAAQ/wMAAADifwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABOS2F6qliYKy2Ui1FIHbZGoWtwZHWmODkfsiE9VuqZqDSnN0L9msu0J6U43j1Y395azA4PLa/sr40eDIz36tXz4xf/D/tgAAwgAA=="
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64_compressed))
	zr, err := gzip.NewReader(decoder)

	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &out)

	_, err = io.Copy(encoder, zr)

	if err != nil {
		t.Fatal(err)
	}

	encoder.Close()

	return out.String()
}

func testAccImageDiskNoOS(t *testing.T) string {
	// Take a compressed then base64'd disk image,
	// base64 decode, then decompress, then re-base64
	// the image, so it can be uploaded to s3.

	// little vmdk built by:
	// $ VBoxManage createmedium disk --filename ./image.vhd --sizebytes 512 --format vhd
	// $ cat image.vhd | gzip --best | base64
	// b64_compressed := "H4sIAAAAAAACA0vOz0tNLsmsYGBgYGJgZIACJgZ1789hZUn5FQxsDIzhmUbZMHEEzSIIJJj///+QlV1rMXFVnLzHwteXYmWDDfYxjIIhA5IrigsSi4pT/0MBRJSNAZoWGBkUGBj+//9SNhpSo2AUDD+AyPOjYESW/6P1/4gGAAvDpVcACgAA"

	// TODO: They only support certain linux flavors. And it checks for a full operating system.
	// How can we import this? Pull from public bucket?
	// Expect fail?
	// little filesystem built by:
	// $ sudo dd if=/dev/zero of=vhd.img bs=4k count=55
	// $ sudo mkfs -t ext3 vhd.img
	// $ qemu-img convert -f raw -O vpc vhd.img image.vhd
	// $ cat image.vhd | gzip --best | base64
	b64_compressed := "H4sIAAAAAAACA+3cz0tUQRwA8Nm3u265gnrw0smzhQch6pQKXQJR1INea9toMTVTQwLBDkbXzp1Cow5hQoeELp78Q4LQi9TBwIu89u2uS4qnyNL184GZNzO8/fGGYZgvzLzC9FSxMFdaCCFEIRVqotA1OLI6U5ycD9mQHiv1TFSa0xuhfs1l2pNSHO8erG9vLWaHh5ZX9tdGDwbGewPnRmFh9vHdJ7PFuKba2hRqYyEVOkOI472negoaUiYGLiTTH1xc5eV9uFGJ+EK4Vk4d9QiwmsJSNXXW2t+/+nIviQhGv6cqIUK1HmrBQlW+9rGbte9Il9PDZ19f73we+vCxeS2//PPKy//xrLe/rb7L99/59PbWjxfX299sJv+35chz/X0pQwwAgDPkcH2eqaz/O8r1jE4BAACABhPnknzJRggAAABoZLlkNy8AAADQyA73ASTnXw/Tv9x/sNtfztpO+v10uPTbfVlbNTgFS8/L2WbfCeMvdWT8/YnmY3Vnwc+ezWT+6Ttp/onq7zxINJVTslUqGROXdRsAAOdUcv6/JaSi7no5irq7q+/w2mltjh5Nz85dfTA9P3VfXwEAAMB5lT8W/++1VuN/AAAAoMG06QIAAAAQ/wMAAADifwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABOS2F6qliYKy2Ui1FIHbZGoWtwZHWmODkfsiE9VuqZqDSnN0L9msu0J6U43j1Y395azA4PLa/sr40eDIz36tXz4xf/D/tgAAwgAA=="
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64_compressed))
	zr, err := gzip.NewReader(decoder)

	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &out)

	_, err = io.Copy(encoder, zr)

	if err != nil {
		t.Fatal(err)
	}

	encoder.Close()

	return out.String()
}
