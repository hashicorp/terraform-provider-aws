// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glacier_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/glacier"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfglacier "github.com/hashicorp/terraform-provider-aws/internal/service/glacier"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlacierVault_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var vault glacier.DescribeVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "glacier", regexache.MustCompile(`vaults/.+`)),
					resource.TestCheckResourceAttr(resourceName, "notification.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "access_policy", ""),
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

func TestAccGlacierVault_notification(t *testing.T) {
	ctx := acctest.Context(t)
	var vault glacier.DescribeVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glacier_vault.test"
	snsResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_notification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "notification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "notification.0.events.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "notification.0.sns_topic", snsResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "notification.#", acctest.Ct0),
				),
			},
			{
				Config: testAccVaultConfig_notification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "notification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "notification.0.events.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "notification.0.sns_topic", snsResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccGlacierVault_policy(t *testing.T) {
	ctx := acctest.Context(t)
	var vault glacier.DescribeVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_policy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, "access_policy", regexache.MustCompile(`"Sid":"cross-account-upload".+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVaultConfig_policyUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestMatchResourceAttr(resourceName, "access_policy", regexache.MustCompile(`"Sid":"cross-account-upload1".+`)),
				),
			},
			{
				Config: testAccVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "access_policy", ""),
				),
			},
		},
	})
}

func TestAccGlacierVault_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var vault glacier.DescribeVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVaultConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVaultConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccGlacierVault_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var vault glacier.DescribeVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglacier.ResourceVault(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccGlacierVault_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	var vault glacier.DescribeVaultOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glacier_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlacierServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckVaultDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVaultConfig_policyOrder(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(ctx, resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "glacier", regexache.MustCompile(`vaults/.+`)),
					resource.TestCheckResourceAttr(resourceName, "notification.#", acctest.Ct0),
					resource.TestMatchResourceAttr(resourceName, "access_policy", regexache.MustCompile(fmt.Sprintf(`"Sid":"%s"`, rName))),
				),
			},
			{
				Config:   testAccVaultConfig_policyNewOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckVaultExists(ctx context.Context, n string, v *glacier.DescribeVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glacier Vault ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlacierClient(ctx)

		output, err := tfglacier.FindVaultByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckVaultDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlacierClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glacier_vault" {
				continue
			}

			_, err := tfglacier.FindVaultByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glacier Vault %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccVaultConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q
}
`, rName)
}

func testAccVaultConfig_notification(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  notification {
    sns_topic = aws_sns_topic.test.arn
    events    = ["ArchiveRetrievalCompleted", "InventoryRetrievalCompleted"]
  }
}
`, rName)
}

func testAccVaultConfig_policy(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  access_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[
       {
          "Sid":"cross-account-upload",
          "Principal": {
             "AWS": "*"
          },
          "Effect":"Allow",
          "Action": [
             "glacier:InitiateMultipartUpload",
             "glacier:AbortMultipartUpload",
             "glacier:CompleteMultipartUpload"
          ],
          "Resource": "arn:${data.aws_partition.current.partition}:glacier:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:vaults/%[1]s"
       }
    ]
}
EOF
}
`, rName)
}

func testAccVaultConfig_policyUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  access_policy = <<EOF
{
    "Version":"2012-10-17",
    "Statement":[
       {
          "Sid":"cross-account-upload1",
          "Principal": {
             "AWS": ["*"]
          },
          "Effect":"Allow",
          "Action": [
             "glacier:UploadArchive",
             "glacier:InitiateMultipartUpload",
             "glacier:AbortMultipartUpload",
             "glacier:CompleteMultipartUpload"
          ],
          "Resource": ["arn:${data.aws_partition.current.partition}:glacier:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:vaults/%[1]s"]
       }
    ]
}
EOF
}
`, rName)
}

func testAccVaultConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVaultConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_glacier_vault" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVaultConfig_policyOrder(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  access_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid = %[1]q
      Principal = {
        AWS = ["*"]
      }
      Effect = "Allow"
      Action = [
        "glacier:InitiateMultipartUpload",
        "glacier:AbortMultipartUpload",
        "glacier:CompleteMultipartUpload",
      ]
      Resource = [
        "arn:${data.aws_partition.current.partition}:glacier:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:vaults/%[1]s",
      ]
    }]
  })
}
`, rName)
}

func testAccVaultConfig_policyNewOrder(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_glacier_vault" "test" {
  name = %[1]q

  access_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid = %[1]q
      Principal = {
        AWS = ["*"]
      }
      Effect = "Allow"
      Action = [
        "glacier:CompleteMultipartUpload",
        "glacier:InitiateMultipartUpload",
        "glacier:AbortMultipartUpload",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:glacier:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:vaults/%[1]s"
    }]
  })
}
`, rName)
}
