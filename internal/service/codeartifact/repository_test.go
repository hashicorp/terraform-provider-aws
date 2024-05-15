// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codeartifact_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodeartifact "github.com/hashicorp/terraform-provider-aws/internal/service/codeartifact"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codeartifact", fmt.Sprintf("repository/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "repository", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
					resource.TestCheckResourceAttrPair(resourceName, "administrator_account", "aws_codeartifact_domain.test", names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func testAccRepository_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, "codeartifact") },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
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
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRepositoryConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccRepository_owner(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_owner(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "codeartifact", fmt.Sprintf("repository/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "repository", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
					resource.TestCheckResourceAttrPair(resourceName, "administrator_account", "aws_codeartifact_domain.test", names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", acctest.Ct0),
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

func testAccRepository_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_desc(rName, "desc"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_desc(rName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc2"),
				),
			},
		},
	})
}

func testAccRepository_upstreams(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_upstreams1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_upstreams2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
					resource.TestCheckResourceAttr(resourceName, "upstream.1.repository_name", fmt.Sprintf("%s-upstream2", rName)),
				),
			},
			{
				Config: testAccRepositoryConfig_upstreams1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
				),
			},
		},
	})
}

func testAccRepository_externalConnection(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_externalConnection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.external_connection_name", "public:npmjs"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.package_format", "npm"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.status", "AVAILABLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", acctest.Ct0),
				),
			},
			{
				Config: testAccRepositoryConfig_externalConnection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.external_connection_name", "public:npmjs"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.package_format", "npm"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.status", "AVAILABLE"),
				),
			},
		},
	})
}

func testAccRepository_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodeartifact.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactClient(ctx)

		_, err := tfcodeartifact.FindRepositoryByThreePartKey(ctx, conn, rs.Primary.Attributes["domain_owner"], rs.Primary.Attributes[names.AttrDomain], rs.Primary.Attributes["repository"])

		return err
	}
}

func testAccCheckRepositoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codeartifact_repository" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactClient(ctx)

			_, err := tfcodeartifact.FindRepositoryByThreePartKey(ctx, conn, rs.Primary.Attributes["domain_owner"], rs.Primary.Attributes[names.AttrDomain], rs.Primary.Attributes["repository"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeArtifact Repository %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccRepositoryConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}
`, rName)
}

func testAccRepositoryConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccRepositoryConfig_base(rName), fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}
`, rName))
}

func testAccRepositoryConfig_owner(rName string) string {
	return acctest.ConfigCompose(testAccRepositoryConfig_base(rName), fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository   = %[1]q
  domain       = aws_codeartifact_domain.test.domain
  domain_owner = aws_codeartifact_domain.test.owner
}
`, rName))
}

func testAccRepositoryConfig_desc(rName, desc string) string {
	return acctest.ConfigCompose(testAccRepositoryConfig_base(rName), fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository  = %[1]q
  domain      = aws_codeartifact_domain.test.domain
  description = %[2]q
}
`, rName, desc))
}

func testAccRepositoryConfig_upstreams1(rName string) string {
	return acctest.ConfigCompose(testAccRepositoryConfig_base(rName), fmt.Sprintf(`
resource "aws_codeartifact_repository" "upstream1" {
  repository = "%[1]s-upstream1"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  upstream {
    repository_name = aws_codeartifact_repository.upstream1.repository
  }
}
`, rName))
}

func testAccRepositoryConfig_upstreams2(rName string) string {
	return testAccRepositoryConfig_base(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "upstream1" {
  repository = "%[1]s-upstream1"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "upstream2" {
  repository = "%[1]s-upstream2"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  upstream {
    repository_name = aws_codeartifact_repository.upstream1.repository
  }

  upstream {
    repository_name = aws_codeartifact_repository.upstream2.repository
  }
}
`, rName)
}

func testAccRepositoryConfig_externalConnection(rName string) string {
	return testAccRepositoryConfig_base(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  external_connections {
    external_connection_name = "public:npmjs"
  }
}
`, rName)
}

func testAccRepositoryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccRepositoryConfig_base(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRepositoryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccRepositoryConfig_base(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
