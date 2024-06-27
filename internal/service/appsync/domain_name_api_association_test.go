// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappsync "github.com/hashicorp/terraform-provider-aws/internal/service/appsync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccDomainNameAPIAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ApiAssociation
	appsyncCertDomain := acctest.SkipIfEnvVarNotSet(t, "AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN")
	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name_api_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameAPIAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAPIAssociationConfig_basic(appsyncCertDomain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameAPIAssociationExists(ctx, resourceName, &association),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, "aws_appsync_domain_name.test", names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test", names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDomainNameAPIAssociationConfig_updated(appsyncCertDomain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameAPIAssociationExists(ctx, resourceName, &association),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDomainName, "aws_appsync_domain_name.test", names.AttrDomainName),
					resource.TestCheckResourceAttrPair(resourceName, "api_id", "aws_appsync_graphql_api.test2", names.AttrID),
				),
			},
		},
	})
}

func testAccDomainNameAPIAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var association awstypes.ApiAssociation
	appsyncCertDomain := acctest.SkipIfEnvVarNotSet(t, "AWS_APPSYNC_DOMAIN_NAME_CERTIFICATE_DOMAIN")
	rName := sdkacctest.RandString(8)
	resourceName := "aws_appsync_domain_name_api_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.AppSyncEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AppSyncServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDomainNameAPIAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDomainNameAPIAssociationConfig_basic(appsyncCertDomain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainNameAPIAssociationExists(ctx, resourceName, &association),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappsync.ResourceDomainNameAPIAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDomainNameAPIAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appsync_domain_name" {
				continue
			}

			_, err := tfappsync.FindDomainNameAPIAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Appsync Domain Name API Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckDomainNameAPIAssociationExists(ctx context.Context, n string, v *awstypes.ApiAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppSyncClient(ctx)

		output, err := tfappsync.FindDomainNameAPIAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDomainNameAPIAssociationConfig_base(domain, rName string) string {
	return fmt.Sprintf(`
data "aws_acm_certificate" "test" {
  domain      = "*.%[1]s"
  most_recent = true
}

resource "aws_appsync_domain_name" "test" {
  domain_name     = "%[2]s.%[1]s"
  certificate_arn = data.aws_acm_certificate.test.arn
}

resource "aws_appsync_graphql_api" "test" {
  authentication_type = "API_KEY"
  name                = %[2]q
}
`, domain, rName)
}

func testAccDomainNameAPIAssociationConfig_basic(domain, rName string) string {
	return acctest.ConfigCompose(testAccDomainNameAPIAssociationConfig_base(domain, rName), `
resource "aws_appsync_domain_name_api_association" "test" {
  api_id      = aws_appsync_graphql_api.test.id
  domain_name = aws_appsync_domain_name.test.domain_name
}
`)
}

func testAccDomainNameAPIAssociationConfig_updated(domain, rName string) string {
	return acctest.ConfigCompose(testAccDomainNameAPIAssociationConfig_base(domain, rName), `
resource "aws_appsync_graphql_api" "test2" {
  authentication_type = "API_KEY"
  name                = "%[1]s-2"
}

resource "aws_appsync_domain_name_api_association" "test" {
  api_id      = aws_appsync_graphql_api.test2.id
  domain_name = aws_appsync_domain_name.test.domain_name
}
`, rName)
}
