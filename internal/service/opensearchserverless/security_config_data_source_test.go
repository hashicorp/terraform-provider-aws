// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessSecurityConfigDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"
	dataSourceName := "data.aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigDataSourceConfig_basic(rName, names.AttrDescription, "test-fixtures/idp-metadata.xml"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, dataSourceName, &securityconfig),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_version", resourceName, "config_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_options.0.metadata", resourceName, "saml_options.0.metadata"),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_options.0.session_timeout", resourceName, "saml_options.0.session_timeout"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityConfigDataSource_iamFederationOptions(t *testing.T) {
	ctx := acctest.Context(t)

	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"
	dataSourceName := "data.aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigDataSourceConfig_iamFederationOptions(rName, names.AttrDescription, string(types.IamIdentityCenterGroupAttributeGroupId), string(types.IamIdentityCenterUserAttributeUserId)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, dataSourceName, &securityconfig),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_version", resourceName, "config_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_federation_options.0.group_attribute", resourceName, "iam_federation_options.0.group_attribute"),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_federation_options.0.user_attribute", resourceName, "iam_federation_options.0.user_attribute"),
				),
			},
		},
	})
}

func TestAccOpenSearchServerlessSecurityConfigDataSource_iamIdentityCenterOptions(t *testing.T) {
	ctx := acctest.Context(t)

	var securityconfig types.SecurityConfigDetail
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"
	dataSourceName := "data.aws_opensearchserverless_security_config.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			acctest.PreCheckSSOAdminInstances(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigDataSourceConfig_iamIdentityCenterOptions(rName, names.AttrDescription, string(types.IamIdentityCenterGroupAttributeGroupId), string(types.IamIdentityCenterUserAttributeUserId)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, t, dataSourceName, &securityconfig),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_version", resourceName, "config_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_identity_center_options.0.instance_arn", resourceName, "iam_identity_center_options.0.instance_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_identity_center_options.0.group_attribute", resourceName, "iam_identity_center_options.0.group_attribute"),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_identity_center_options.0.user_attribute", resourceName, "iam_identity_center_options.0.user_attribute"),
				),
			},
		},
	})
}

func testAccSecurityConfigDataSourceConfig_basic(rName, description, samlOptions string) string {
	return fmt.Sprintf(`
resource "aws_opensearchserverless_security_config" "test" {
  name        = %[1]q
  description = %[2]q
  type        = "saml"

  saml_options {
    metadata = file("%[3]s")
  }
}

data "aws_opensearchserverless_security_config" "test" {
  id = aws_opensearchserverless_security_config.test.id
}
`, rName, description, samlOptions)
}

func testAccSecurityConfigDataSourceConfig_iamFederationOptions(rName, description, groupAttribute, userAttribute string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_opensearchserverless_security_config" "test" {
  name        = %[1]q
  description = %[2]q
  type        = "iamfederation"

  iam_federation_options {
    group_attribute = %[3]q
    user_attribute  = %[4]q
  }
}

data "aws_opensearchserverless_security_config" "test" {
  id = aws_opensearchserverless_security_config.test.id
}
`, rName, description, groupAttribute, userAttribute)
}

func testAccSecurityConfigDataSourceConfig_iamIdentityCenterOptions(rName, description, groupAttribute, userAttribute string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_opensearchserverless_security_config" "test" {
  name        = %[1]q
  description = %[2]q
  type        = "iamidentitycenter"

  iam_identity_center_options {
    instance_arn    = tolist(data.aws_ssoadmin_instances.test.arns)[0]
    group_attribute = %[3]q
    user_attribute  = %[4]q
  }
}

data "aws_opensearchserverless_security_config" "test" {
  id = aws_opensearchserverless_security_config.test.id
}
`, rName, description, groupAttribute, userAttribute)
}
