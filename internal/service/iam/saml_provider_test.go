// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMSAMLProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	idpEntityIdModified := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_basic(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "iam", fmt.Sprintf("saml-provider/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
					resource.TestCheckResourceAttrSet(resourceName, "valid_until"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccSAMLProviderConfig_update(rName, idpEntityIdModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
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

func TestAccIAMSAMLProvider_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_tags1(rName, idpEntityId, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, resourceName),
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
				Config: testAccSAMLProviderConfig_tags2(rName, idpEntityId, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSAMLProviderConfig_tags1(rName, idpEntityId, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIAMSAMLProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	idpEntityId := fmt.Sprintf("https://%s", acctest.RandomDomainName())
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSAMLProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSAMLProviderConfig_basic(rName, idpEntityId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSAMLProviderExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceSAMLProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSAMLProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_saml_provider" {
				continue
			}

			_, err := tfiam.FindSAMLProviderByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM SAML Provider %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckSAMLProviderExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IAM SAML Provider ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		_, err := tfiam.FindSAMLProviderByARN(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccSAMLProviderConfig_basic(rName, idpEntityId string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })
}
`, rName, idpEntityId)
}

func testAccSAMLProviderConfig_update(rName, idpEntityIdModified string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata-modified.xml.tpl", { entity_id_modified = %[2]q })
}
`, rName, idpEntityIdModified)
}

func testAccSAMLProviderConfig_tags1(rName, idpEntityId, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, idpEntityId, tagKey1, tagValue1)
}

func testAccSAMLProviderConfig_tags2(rName, idpEntityId, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %[1]q
  saml_metadata_document = templatefile("./test-fixtures/saml-metadata.xml.tpl", { entity_id = %[2]q })

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, idpEntityId, tagKey1, tagValue1, tagKey2, tagValue2)
}
