// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftIDCApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIDCApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "idc_display_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "idc_instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "identity_namespace", rName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "redshift_idc_application_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "redshift_idc_application_arn",
			},
		},
	})
}

func TestAccRedshiftIDCApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIDCApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, acctest.Provider, tfredshift.ResourceIdcApplication, resourceName, IDCApplicationDisappearsStateFunc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftIDCApplication_authorizedTokenIssuerList(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIDCApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_authorizedTokenIssuerList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "authorized_token_issuer_list.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "authorized_token_issuer_list.0.trusted_token_issuer_arn", "aws_ssoadmin_trusted_token_issuer.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "authorized_token_issuer_list.0.authorized_audiences_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorized_token_issuer_list.0.authorized_audiences_list.0", "client_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "redshift_idc_application_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "redshift_idc_application_arn",
			},
		},
	})
}

func TestAccRedshiftIDCApplication_serviceIntegrationsLakehouse(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIDCApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_serviceIntegrationsLakehouse(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.lake_formation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.lake_formation.0.lake_formation_query.0.authorization", "Enabled"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "redshift_idc_application_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "redshift_idc_application_arn",
			},
		},
	})
}

func TestAccRedshiftIDCApplication_serviceIntegrationsRedshift(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIDCApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_serviceIntegrationsRedshift(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.redshift.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.redshift.0.connect.0.authorization", "Enabled"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "redshift_idc_application_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "redshift_idc_application_arn",
			},
		},
	})
}

func TestAccRedshiftIDCApplication_serviceIntegrationsS3AccessGrants(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIDCApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_serviceIntegrationsS3AccessGrants(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.s3_access_grants.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.s3_access_grants.0.read_write_access.0.authorization", "Enabled"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "redshift_idc_application_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "redshift_idc_application_arn",
			},
		},
	})
}

func TestAccRedshiftIDCApplication_tags(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIDCApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "redshift_idc_application_arn"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "redshift_idc_application_arn",
			},
			{
				Config: testAccIdcApplicationConfig_tags(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIDCApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckIDCApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_idc_application" {
				continue
			}

			arn := rs.Primary.Attributes["redshift_idc_application_arn"]

			_, err := tfredshift.FindIDCApplicationByARN(ctx, conn, arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Redshift IDC Application %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIDCApplicationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		arn := rs.Primary.Attributes["redshift_idc_application_arn"]

		if arn == "" {
			return fmt.Errorf("Redshift IDC Application is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftClient(ctx)

		_, err := tfredshift.FindIDCApplicationByARN(ctx, conn, arn)

		return err
	}
}

func IDCApplicationDisappearsStateFunc(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	v, ok := is.Attributes["redshift_idc_application_arn"]
	if !ok {
		return errors.New(`Identifying attribute "redshift_idc_application_arn" not defined`)
	}

	if err := fwdiag.DiagnosticsError(state.SetAttribute(ctx, path.Root("redshift_idc_application_arn"), v)); err != nil {
		return err
	}

	return nil
}

func testAccIdcApplicationConfig_baseIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/service-role/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": [
                    "redshift-serverless.amazonaws.com",
                    "redshift.amazonaws.com"
                ]
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test1" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSSSOMemberAccountAdministrator"
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonRedshiftFullAccess"
}

`, rName)
}

func testAccIdcApplicationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccIdcApplicationConfig_baseIAMRole(rName), fmt.Sprintf(`


data "aws_ssoadmin_instances" "test" {}

resource "aws_redshift_idc_application" "test" {
  iam_role_arn                  = aws_iam_role.test.arn
  idc_display_name              = %[1]q
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  identity_namespace            = %[1]q
  redshift_idc_application_name = %[1]q
}
`, rName))
}

func testAccIdcApplicationConfig_authorizedTokenIssuerList(rName string) string {
	return acctest.ConfigCompose(testAccIdcApplicationConfig_baseIAMRole(rName), fmt.Sprintf(`
resource "aws_ssoadmin_trusted_token_issuer" "test" {
  name                      = %[1]q
  instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  trusted_token_issuer_type = "OIDC_JWT"

  trusted_token_issuer_configuration {
    oidc_jwt_configuration {
      claim_attribute_path          = "email"
      identity_store_attribute_path = "emails.value"
      issuer_url                    = "https://example.com"
      jwks_retrieval_option         = "OPEN_ID_DISCOVERY"
    }
  }
}

data "aws_ssoadmin_instances" "test" {}

resource "aws_redshift_idc_application" "test" {
  authorized_token_issuer_list {
    authorized_audiences_list = ["client_id"]
    trusted_token_issuer_arn  = aws_ssoadmin_trusted_token_issuer.test.arn
  }
  iam_role_arn                  = aws_iam_role.test.arn
  idc_display_name              = %[1]q
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  redshift_idc_application_name = %[1]q
}
`, rName))
}

func testAccIdcApplicationConfig_serviceIntegrationsLakehouse(rName string) string {
	return acctest.ConfigCompose(testAccIdcApplicationConfig_baseIAMRole(rName), fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_redshift_idc_application" "test" {
  iam_role_arn                  = aws_iam_role.test.arn
  idc_display_name              = %[1]q
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  redshift_idc_application_name = %[1]q
  service_integrations {
    lake_formation {
      lake_formation_query {
        authorization = "Enabled"
      }
    }
  }
}
`, rName))
}

func testAccIdcApplicationConfig_serviceIntegrationsRedshift(rName string) string {
	return acctest.ConfigCompose(testAccIdcApplicationConfig_baseIAMRole(rName), fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_redshift_idc_application" "test" {
  iam_role_arn                  = aws_iam_role.test.arn
  idc_display_name              = %[1]q
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  redshift_idc_application_name = %[1]q
  service_integrations {
    redshift {
      connect {
        authorization = "Enabled"
      }
    }
  }
}
`, rName))
}

func testAccIdcApplicationConfig_serviceIntegrationsS3AccessGrants(rName string) string {
	return acctest.ConfigCompose(testAccIdcApplicationConfig_baseIAMRole(rName), fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_redshift_idc_application" "test" {
  iam_role_arn                  = aws_iam_role.test.arn
  idc_display_name              = %[1]q
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  redshift_idc_application_name = %[1]q
  service_integrations {
    s3_access_grants {
      read_write_access {
        authorization = "Enabled"
      }
    }
  }
}
`, rName))
}

func testAccIdcApplicationConfig_tags(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccIdcApplicationConfig_baseIAMRole(rName), fmt.Sprintf(`


data "aws_ssoadmin_instances" "test" {}

resource "aws_redshift_idc_application" "test" {
  iam_role_arn                  = aws_iam_role.test.arn
  idc_display_name              = %[1]q
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  identity_namespace            = %[1]q
  redshift_idc_application_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}
