// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfecrpublic "github.com/hashicorp/terraform-provider-aws/internal/service/ecrpublic"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECRPublicRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrRepositoryName, rName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "ecr-public", "repository/"+rName),
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

func TestAccECRPublicRepository_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
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
				Config: testAccRepositoryConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccECRPublicRepository_CatalogData_aboutText(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_catalogDataAboutText(rName, "about_text_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.about_text", "about_text_1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_catalogDataAboutText(rName, "about_text_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.about_text", "about_text_2"),
				),
			},
		},
	})
}

func TestAccECRPublicRepository_CatalogData_architectures(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_catalogDataArchitectures(rName, "Linux"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.0", "Linux"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_catalogDataArchitectures(rName, "Windows"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.0", "Windows"),
				),
			},
		},
	})
}

func TestAccECRPublicRepository_CatalogData_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_catalogDataDescription(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.description", "description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_catalogDataDescription(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.description", "description 2"),
				),
			},
		},
	})
}

func TestAccECRPublicRepository_CatalogData_operatingSystems(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_catalogDataOperatingSystems(rName, "ARM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.0", "ARM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_catalogDataOperatingSystems(rName, "x86"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.0", "x86"),
				),
			},
		},
	})
}

func TestAccECRPublicRepository_CatalogData_usageText(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_catalogDataUsageText(rName, "usage text 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.usage_text", "usage text 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_catalogDataUsageText(rName, "usage text 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.usage_text", "usage text 2"),
				),
			},
		},
	})
}

func TestAccECRPublicRepository_CatalogData_logoImageBlob(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_catalogDataLogoImageBlob(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "catalog_data.0.logo_image_blob"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"catalog_data.0.logo_image_blob"},
			},
		},
	})
}

func TestAccECRPublicRepository_Basic_forceDestroy(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_forceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrRepositoryName, rName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, names.AttrARN, "ecr-public", "repository/"+rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
		},
	})
}

func TestAccECRPublicRepository_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Repository
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckRegion(t, names.USEast1RegionID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECRPublicServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecrpublic.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryExists(ctx context.Context, name string, res *awstypes.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Public repository ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRPublicClient(ctx)

		output, err := conn.DescribeRepositories(ctx, &ecrpublic.DescribeRepositoriesInput{
			RepositoryNames: []string{aws.ToString(&rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(output.Repositories) == 0 {
			return fmt.Errorf("ECR Public repository %s not found", rs.Primary.ID)
		}

		res = &output.Repositories[0]

		return nil
	}
}

func testAccCheckRepositoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRPublicClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecrpublic_repository" {
				continue
			}

			repositoryName := rs.Primary.Attributes[names.AttrRepositoryName]
			input := ecrpublic.DescribeRepositoriesInput{

				RepositoryNames: []string{repositoryName},
			}

			out, err := conn.DescribeRepositories(ctx, &input)

			if errs.IsA[*awstypes.RepositoryNotFoundException](err) {
				return nil
			}

			if err != nil {
				return err
			}

			for _, repository := range out.Repositories {
				if aws.ToString(repository.RepositoryName) == rs.Primary.Attributes[names.AttrRepositoryName] {
					return fmt.Errorf("ECR Public repository still exists: %s", rs.Primary.Attributes[names.AttrRepositoryName])
				}
			}
		}

		return nil
	}
}

func testAccRepositoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
}
`, rName)
}

func testAccRepositoryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRepositoryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRepositoryConfig_forceDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  force_destroy   = true
}
`, rName)
}

func testAccRepositoryConfig_catalogDataAboutText(rName string, aboutText string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    about_text = %[2]q
  }
}
`, rName, aboutText)
}

func testAccRepositoryConfig_catalogDataArchitectures(rName string, architecture string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    architectures = [%[2]q]
  }
}
`, rName, architecture)
}

func testAccRepositoryConfig_catalogDataDescription(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    description = %[2]q
  }
}
`, rName, description)
}

func testAccRepositoryConfig_catalogDataOperatingSystems(rName string, operatingSystem string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    operating_systems = [%[2]q]
  }
}
`, rName, operatingSystem)
}

func testAccRepositoryConfig_catalogDataUsageText(rName string, usageText string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    usage_text = %[2]q
  }
}
`, rName, usageText)
}

func testAccRepositoryConfig_catalogDataLogoImageBlob(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    logo_image_blob = filebase64("test-fixtures/terraform_logo.png")
  }
}
`, rName)
}
