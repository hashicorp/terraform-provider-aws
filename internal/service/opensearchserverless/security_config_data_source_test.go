// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOpenSearchServerlessSecurityConfigDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var securityconfig types.SecurityConfigDetail
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opensearchserverless_security_config.test"
	dataSourceName := "data.aws_opensearchserverless_security_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.OpenSearchServerlessEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OpenSearchServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSecurityConfigDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityConfigDataSourceConfig_basic(rName, names.AttrDescription, "test-fixtures/idp-metadata.xml"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecurityConfigExists(ctx, dataSourceName, &securityconfig),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_version", resourceName, "config_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_modified_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_options.metadata", resourceName, "saml_options.metadata"),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_options.session_timeout", resourceName, "saml_options.session_timeout"),
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
