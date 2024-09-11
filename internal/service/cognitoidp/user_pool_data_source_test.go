// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var userpool awsTypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			// acctest.PreCheckPartitionHasService(t, names.CognitoIDPServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolExists(ctx, dataSourceName, &userpool),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "cognito-idp", regexache.MustCompile(`userpool/.*`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolDataSource_schemaAttributes(t *testing.T) {
	ctx := acctest.Context(t)

	var userpool awsTypes.UserPoolType
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDataSourceConfig_schemaAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolExists(ctx, dataSourceName, &userpool),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					testSchemaAttributes(dataSourceName),
				),
			},
		},
	})
}

func testSchemaAttributes(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		// Get the number of schema_attributes
		numAttributesStr, ok := rs.Primary.Attributes["schema_attributes.#"]
		if !ok {
			return fmt.Errorf("schema_attributes not found in resource %s", n)
		}
		numAttributes, err := strconv.Atoi(numAttributesStr)
		if err != nil {
			return fmt.Errorf("error parsing schema_attributes.#: %s", err)
		}

		// Loop through the schema_attributes and check the mutable key in each attribute
		checksCompleted := map[string]bool{
			names.AttrEmail: false,
		}
		for i := 0; i < numAttributes; i++ {
			// Get the attribute
			attribute := fmt.Sprintf("schema_attributes.%d.name", i)
			name, ok := rs.Primary.Attributes[attribute]
			if name == "" || !ok {
				return fmt.Errorf("attribute not found at %s", name)
			}
			if name == names.AttrEmail {
				if rs.Primary.Attributes[fmt.Sprintf("schema_attributes.%d.mutable", i)] != acctest.CtFalse {
					return fmt.Errorf("mutable is not false for attribute %v", name)
				}
				checksCompleted[names.AttrEmail] = true
			}
		}
		for k, v := range checksCompleted {
			if !v {
				return fmt.Errorf("attribute %v not found in schema_attributes", k)
			}
		}

		return nil
	}
}

func testAccUserPoolDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}

data "aws_cognito_user_pool" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}
`, rName)
}

func testAccUserPoolDataSourceConfig_schemaAttributes(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolConfig_schemaAttributes(rName),
		`
data "aws_cognito_user_pool" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}
`)
}
