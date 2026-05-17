// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	awsTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCognitoIDPUserPoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var userpool awsTypes.UserPoolType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolExists(ctx, t, dataSourceName, &userpool),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "cognito-idp", regexache.MustCompile(`userpool/.*`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("user_pool_tags"), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolDataSource_schemaAttributes(t *testing.T) {
	ctx := acctest.Context(t)

	var userpool awsTypes.UserPoolType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDataSourceConfig_schemaAttributes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolExists(ctx, t, dataSourceName, &userpool),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					testSchemaAttributes(dataSourceName),
				),
			},
		},
	})
}

func TestAccCognitoIDPUserPoolDataSource_userPoolTags(t *testing.T) {
	ctx := acctest.Context(t)

	var userpool awsTypes.UserPoolType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_pool.test"
	resourceName := "aws_cognito_user_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDataSourceConfig_userPoolTags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolExists(ctx, t, dataSourceName, &userpool),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("user_pool_tags"), resourceName, tfjsonpath.New(names.AttrTagsAll), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New(names.AttrTags), resourceName, tfjsonpath.New(names.AttrTagsAll), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccCognitoIDPUserPoolDataSource_userPoolAddOns(t *testing.T) {
	ctx := acctest.Context(t)

	var userpool awsTypes.UserPoolType
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_user_pool.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CognitoIDPServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoolDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoolDataSourceConfig_userPoolAddOns(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPoolExists(ctx, t, dataSourceName, &userpool),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "cognito-idp", regexache.MustCompile(`userpool/.*`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "user_pool_add_ons.0.advanced_security_mode", "ENFORCED"),
					resource.TestCheckResourceAttr(dataSourceName, "user_pool_add_ons.0.advanced_security_additional_flows.0.custom_auth_mode", "ENFORCED"),
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
			return fmt.Errorf("error parsing schema_attributes.#: %w", err)
		}

		// Loop through the schema_attributes and check the mutable key in each attribute
		checksCompleted := map[string]bool{
			names.AttrEmail: false,
		}
		for i := range numAttributes {
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
data "aws_cognito_user_pool" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
}
`, rName)
}

func testAccUserPoolDataSourceConfig_schemaAttributes(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoolConfig_schemaAttributes(rName), `
data "aws_cognito_user_pool" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}
`)
}

func testAccUserPoolDataSourceConfig_userPoolTags(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_cognito_user_pool" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccUserPoolDataSourceConfig_userPoolAddOns(rName string) string {
	return fmt.Sprintf(`
data "aws_cognito_user_pool" "test" {
  user_pool_id = aws_cognito_user_pool.test.id
}

resource "aws_cognito_user_pool" "test" {
  name = %[1]q
  user_pool_add_ons {
    advanced_security_mode = "ENFORCED"
    advanced_security_additional_flows {
      custom_auth_mode = "ENFORCED"
    }
  }
}
`, rName)
}
