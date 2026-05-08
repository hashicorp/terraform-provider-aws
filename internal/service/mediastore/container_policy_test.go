// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediastore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfmediastore "github.com/hashicorp/terraform-provider-aws/internal/service/mediastore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccMediaStoreContainerPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_media_store_container_policy.test"

	rName = strings.ReplaceAll(rName, "-", "_")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "container_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerPolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "container_name"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
		},
	})
}

func TestAccMediaStoreContainerPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_media_store_container_policy.test"

	rName = strings.ReplaceAll(rName, "-", "_")

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.MediaStoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckContainerPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccContainerPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerPolicyExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfmediastore.ResourceContainerPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContainerPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).MediaStoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_media_store_container_policy" {
				continue
			}

			_, err := tfmediastore.FindContainerPolicyByContainerName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Expected MediaStore Container Policy to be destroyed, %s found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckContainerPolicyExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.ProviderMeta(ctx, t).MediaStoreClient(ctx)

		_, err := tfmediastore.FindContainerPolicyByContainerName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("retrieving MediaStore Container Policy (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccContainerPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_media_store_container" "test" {
  name = %[1]q
}

resource "aws_media_store_container_policy" "test" {
  container_name = aws_media_store_container.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "lucky"
      Action = "mediastore:*"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Effect   = "Allow"
      Resource = "arn:${data.aws_partition.current.partition}:mediastore:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:container/${aws_media_store_container.test.name}/*"
      Condition = {
        Bool = {
          "aws:SecureTransport" = "true"
        }
      }
    }]
  })
}
`, rName)
}
