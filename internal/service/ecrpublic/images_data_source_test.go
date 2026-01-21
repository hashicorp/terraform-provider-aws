// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecrpublic_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPublicImagesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecrpublic_images.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, endpoints.UsEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImagesDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRepositoryName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "images.#", "0"),
				),
			},
		},
	})
}

func TestAccECRPublicImagesDataSource_registryID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecrpublic_images.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, endpoints.UsEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccImagesDataSourceConfig_registryID(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrRepositoryName, rName),
					resource.TestCheckResourceAttrSet(dataSourceName, "registry_id"),
					resource.TestCheckResourceAttr(dataSourceName, "images.#", "0"),
				),
			},
		},
	})
}

func TestAccECRPublicImagesDataSource_registryIDValidation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, endpoints.UsEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccImagesDataSourceConfig_registryIDInvalid(rName),
				ExpectError: regexache.MustCompile(`must be a 12-digit AWS account ID`),
			},
		},
	})
}

func testAccImagesDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
}

data "aws_ecrpublic_images" "test" {
  repository_name = aws_ecrpublic_repository.test.repository_name
}
`, rName)
}

func testAccImagesDataSourceConfig_registryID(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
}

data "aws_ecrpublic_images" "test" {
  repository_name = aws_ecrpublic_repository.test.repository_name
  registry_id     = data.aws_caller_identity.current.account_id
}
`, rName)
}

func testAccImagesDataSourceConfig_registryIDInvalid(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
}

data "aws_ecrpublic_images" "test" {
  repository_name = aws_ecrpublic_repository.test.repository_name
  registry_id     = "invalid"
}
`, rName)
}
