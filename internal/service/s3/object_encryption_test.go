// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3Object_encryptionUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccObjectConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				// Switch to KMS (aws:kms) which IS supported for in-place update
				Config: testAccObjectConfig_updateable_kms(rName, "initial content", "aws:kms", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectSSE(ctx, resourceName, "aws:kms"),
				),
			},
		},
	})
}

func TestAccS3Object_encryptionUpdate_Versioned(t *testing.T) {
	ctx := acctest.Context(t)
	var obj1, obj2 s3.GetObjectOutput
	resourceName := "aws_s3_object.object"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Create object with default encryption (none/AES256)
			{
				Config: testAccObjectConfig_updateable_kms(rName, names.AttrContent, "AES256", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj1),
					resource.TestCheckResourceAttrSet(resourceName, "version_id"),
				),
			},
			// Step 2: Update to SSE-KMS Encryption
			// This IS supported by UpdateObjectEncryption and should be in-place.
			{
				Config: testAccObjectConfig_updateable_kms(rName, names.AttrContent, "aws:kms", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckObjectExists(ctx, resourceName, &obj2),
					testAccCheckObjectSSE(ctx, resourceName, "aws:kms"),
					// THE KEY CHECK: Version ID should be identical
					testAccCheckObjectVersionIDEquals(&obj2, &obj1),
				),
			},
		},
	})
}

// Helper that constructs a config with a KMS key but allows optional usage
func testAccObjectConfig_updateable_kms(rName string, content string, sse string, useKMS bool) string {
	kmsRef := ""
	if useKMS {
		kmsRef = `kms_key_id = aws_kms_key.test.arn`
	}

	// If sse is AES256, we don't pass kms_key_id usually, but if we do, it's ignored or error?
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test"
  deletion_window_in_days = 7
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_object" "object" {
  bucket                 = aws_s3_bucket.test.id
  key                    = "updateable-key"
  content                = %[2]q
  server_side_encryption = %[3]q
  %[4]s
}
`, rName, content, sse, kmsRef)
}
