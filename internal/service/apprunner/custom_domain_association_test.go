// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apprunner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfapprunner "github.com/hashicorp/terraform-provider-aws/internal/service/apprunner"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppRunnerCustomDomainAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"
	serviceResourceName := "aws_apprunner_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "certificate_validation_records.#", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_target"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDomainName, domain),
					resource.TestCheckResourceAttr(resourceName, "enable_www_subdomain", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "pending_certificate_dns_validation"),
					resource.TestCheckResourceAttrPair(resourceName, "service_arn", serviceResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"dns_target"},
			},
		},
	})
}

func TestAccAppRunnerCustomDomainAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	domain := acctest.SkipIfEnvVarNotSet(t, "APPRUNNER_CUSTOM_DOMAIN")
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_apprunner_custom_domain_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppRunnerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomDomainAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomDomainAssociationConfig_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomDomainAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfapprunner.ResourceCustomDomainAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCustomDomainAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_apprunner_connection" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).AppRunnerClient(ctx)

			_, err := tfapprunner.FindCustomDomainByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["service_arn"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("App Runner Custom Domain Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckCustomDomainAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppRunnerClient(ctx)

		_, err := tfapprunner.FindCustomDomainByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDomainName], rs.Primary.Attributes["service_arn"])

		return err
	}
}

func testAccCustomDomainAssociationConfig_basic(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_apprunner_service" "test" {
  service_name = %[1]q

  source_configuration {
    auto_deployments_enabled = false
    image_repository {
      image_configuration {
        port = "80"
      }
      image_identifier      = "public.ecr.aws/nginx/nginx:latest"
      image_repository_type = "ECR_PUBLIC"
    }
  }
}

resource "aws_apprunner_custom_domain_association" "test" {
  domain_name = %[2]q
  service_arn = aws_apprunner_service.test.arn
}
`, rName, domain)
}
