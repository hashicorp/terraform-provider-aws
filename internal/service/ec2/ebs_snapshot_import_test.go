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
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

func TestAccEC2EBSSnapshotImport_basic(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotImport_disappears(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEBSSnapshotImport(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSSnapshotImport_Disappears_s3Object(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parentResourceName := "aws_s3_object.test"
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportConfig_basic(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceObject(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSSnapshotImport_tags(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportConfig_tags1(t, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccEBSSnapshotImportConfig_tags2(t, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccEBSSnapshotImportConfig_tags1(t, rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotImport_storageTier(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckEBSSnapshotDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportConfig_storageTier(t, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "storage_tier", "archive"),
				),
			},
		},
	})
}

func testAccEBSSnapshotImportBaseConfig(t *testing.T, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket         = aws_s3_bucket.test.id
  key            = "diskimage.vhd"
  content_base64 = %[2]q
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
`, rName, testAccEBSSnapshotDisk(t))
}

func testAccEBSSnapshotImportConfig_basic(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotImportBaseConfig(t, rName), `
resource "aws_ebs_snapshot_import" "test" {
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

func testAccEBSSnapshotImportConfig_storageTier(t *testing.T, rName string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotImportBaseConfig(t, rName), fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
  disk_container {
    description = "test"
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.test.id
      s3_key    = aws_s3_object.test.key
    }
  }

  role_name    = aws_iam_role.test.name
  storage_tier = "archive"

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccEBSSnapshotImportConfig_tags1(t *testing.T, rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotImportBaseConfig(t, rName), fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
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

func testAccEBSSnapshotImportConfig_tags2(t *testing.T, rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotImportBaseConfig(t, rName), fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
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

func testAccEBSSnapshotDisk(t *testing.T) string {
	// Take a compressed then base64'd disk image,
	// base64 decode, then decompress, then re-base64
	// the image, so it can be uploaded to s3.

	// little vmdk built by:
	// $ VBoxManage createmedium disk --filename ./image.vhd --sizebytes 512 --format vhd
	// $ cat image.vhd | gzip --best | base64
	b64_compressed := "H4sIAAAAAAACA0vOz0tNLsmsYGBgYGJgZIACJgZ1789hZUn5FQxsDIzhmUbZMHEEzSIIJJj///+QlV1rMXFVnLzHwteXYmWDDfYxjIIhA5IrigsSi4pT/0MBRJSNAZoWGBkUGBj+//9SNhpSo2AUDD+AyPOjYESW/6P1/4gGAAvDpVcACgAA"
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
