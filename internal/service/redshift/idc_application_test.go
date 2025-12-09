// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfredshift "github.com/hashicorp/terraform-provider-aws/internal/service/redshift"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftIdcApplication_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdcApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdcApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_arn", "aws_iam_role.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "idc_display_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "idc_instance_arn", "data.aws_ssoadmin_instances.test", "arns.0"),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "identity_namespace", rName),
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

func TestAccRedshiftIdcApplication_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdcApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdcApplicationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfredshift.ResourceIdcApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRedshiftIdcApplication_authorizedTokenIssuerList(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdcApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_authorizedTokenIssuerList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdcApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "authorized_token_issuer_list.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "authorized_token_issuer_list.0.trusted_token_issuer_arn", "aws_ssoadmin_trusted_token_issuer.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "authorized_token_issuer_list.0.authorized_audiences_list.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authorized_token_issuer_list.0.authorized_audiences_list.0", "client_id"),
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

func TestAccRedshiftIdcApplication_serviceIntegrations(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_redshift_idc_application.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIdcApplicationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIdcApplicationConfig_serviceIntegrations(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIdcApplicationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "redshift_idc_application_name", rName),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.lake_formation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "service_integrations.0.lake_formation.0.lake_formation_query.authorization", "Enabled"),
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

func testAccCheckIdcApplicationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_redshift_idc_application" {
				continue
			}
			_, err := tfredshift.FindIDCApplicationByARN(ctx, conn, rs.Primary.ID)

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

func testAccCheckIdcApplicationExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Redshift IDC Application is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RedshiftConn(ctx)

		_, err := tfredshift.FindIDCApplicationByARN(ctx, conn, rs.Primary.ID)

		return err
	}
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

func testAccIdcApplicationConfig_serviceIntegrations(rName string) string {
	return acctest.ConfigCompose(testAccIdcApplicationConfig_baseIAMRole(rName), fmt.Sprintf(`


data "aws_ssoadmin_instances" "test" {}

resource "aws_redshift_idc_application" "test" {
  iam_role_arn                  = aws_iam_role.test.arn
  idc_display_name              = %[1]q
  idc_instance_arn              = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  redshift_idc_application_name = %[1]q
  service_integrations {
    lake_formation {
      lake_formation_query = {
        authorization = "Enabled"
      }
    }
  }
}


`, rName))
}
