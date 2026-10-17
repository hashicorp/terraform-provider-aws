// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package codeartifact_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codeartifact"
	awstypes "github.com/aws/aws-sdk-go-v2/service/codeartifact/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcodeartifact "github.com/hashicorp/terraform-provider-aws/internal/service/codeartifact"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPackageOriginConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.PackageOriginConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_package_origin_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageOriginConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageOriginConfigurationConfig_basic(rName, "BLOCK", "ALLOW"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackageOriginConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomain, rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", names.AttrOwner),
					resource.TestCheckResourceAttr(resourceName, "repository", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrFormat, "npm"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, rName),
					resource.TestCheckResourceAttr(resourceName, "package", rName),
					resource.TestCheckResourceAttr(resourceName, "restrictions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.publish", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.upstream", "ALLOW"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccPackageOriginConfigurationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "package",
			},
		},
	})
}

func testAccPackageOriginConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.PackageOriginConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_package_origin_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageOriginConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageOriginConfigurationConfig_basic(rName, "BLOCK", "ALLOW"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackageOriginConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.publish", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.upstream", "ALLOW"),
				),
			},
			{
				Config: testAccPackageOriginConfigurationConfig_basic(rName, "ALLOW", "BLOCK"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackageOriginConfigurationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.publish", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, "restrictions.0.upstream", "BLOCK"),
				),
			},
		},
	})
}

func testAccPackageOriginConfiguration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.PackageOriginConfiguration
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_package_origin_configuration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.CodeArtifactEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeArtifactServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPackageOriginConfigurationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPackageOriginConfigurationConfig_basic(rName, "BLOCK", "ALLOW"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPackageOriginConfigurationExists(ctx, t, resourceName, &v),
					testAccCheckPackageDisappears(ctx, t, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPackageOriginConfigurationExists(ctx context.Context, t *testing.T, n string, v *awstypes.PackageOriginConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CodeArtifactClient(ctx)

		output, err := tfcodeartifact.FindPackageOriginConfiguration(ctx, conn,
			rs.Primary.Attributes[names.AttrDomain],
			rs.Primary.Attributes["domain_owner"],
			rs.Primary.Attributes["repository"],
			awstypes.PackageFormat(rs.Primary.Attributes[names.AttrFormat]),
			rs.Primary.Attributes[names.AttrNamespace],
			rs.Primary.Attributes["package"],
		)
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// Package origin configurations cannot be deleted; destroying the resource
// resets the restrictions to the CodeArtifact defaults (publish and upstream
// both ALLOW). The package (and its containing domain and repository) may also
// have been deleted entirely.
func testAccCheckPackageOriginConfigurationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codeartifact_package_origin_configuration" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).CodeArtifactClient(ctx)

			output, err := tfcodeartifact.FindPackageOriginConfiguration(ctx, conn,
				rs.Primary.Attributes[names.AttrDomain],
				rs.Primary.Attributes["domain_owner"],
				rs.Primary.Attributes["repository"],
				awstypes.PackageFormat(rs.Primary.Attributes[names.AttrFormat]),
				rs.Primary.Attributes[names.AttrNamespace],
				rs.Primary.Attributes["package"],
			)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if output.Restrictions.Publish == awstypes.AllowPublishAllow && output.Restrictions.Upstream == awstypes.AllowUpstreamAllow {
				continue
			}

			return fmt.Errorf("CodeArtifact Package Origin Configuration %s still exists", rs.Primary.Attributes["package"])
		}

		return nil
	}
}

func testAccCheckPackageDisappears(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CodeArtifactClient(ctx)

		input := codeartifact.DeletePackageInput{
			Domain:      aws.String(rs.Primary.Attributes[names.AttrDomain]),
			DomainOwner: aws.String(rs.Primary.Attributes["domain_owner"]),
			Format:      awstypes.PackageFormat(rs.Primary.Attributes[names.AttrFormat]),
			Package:     aws.String(rs.Primary.Attributes["package"]),
			Repository:  aws.String(rs.Primary.Attributes["repository"]),
		}
		if namespace := rs.Primary.Attributes[names.AttrNamespace]; namespace != "" {
			input.Namespace = aws.String(namespace)
		}

		_, err := conn.DeletePackage(ctx, &input)

		return err
	}
}

func testAccPackageOriginConfigurationImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		return fmt.Sprintf("%s,%s,%s,%s,%s",
			rs.Primary.Attributes[names.AttrDomain],
			rs.Primary.Attributes["repository"],
			rs.Primary.Attributes[names.AttrFormat],
			rs.Primary.Attributes[names.AttrNamespace],
			rs.Primary.Attributes["package"],
		), nil
	}
}

func testAccPackageOriginConfigurationConfig_basic(rName, publish, upstream string) string {
	return acctest.ConfigCompose(testAccRepositoryConfig_base(rName), fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_package_origin_configuration" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = "npm"
  namespace  = %[1]q
  package    = %[1]q

  restrictions {
    publish  = %[2]q
    upstream = %[3]q
  }
}
`, rName, publish, upstream))
}
