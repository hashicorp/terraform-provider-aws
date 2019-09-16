package aws

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccAWSCognitoUserPoolAddCustomAttribute_basic(t *testing.T) {
	userPoolName := fmt.Sprintf("tf-acc-cognito-user-pool-%s", acctest.RandString(7))
	attributeName := "brand_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSCognitoIdentityProvider(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCognitoUserPoolAddCustomAttributeConfig_basic(userPoolName, attributeName),
				Check: resource.ComposeAggregateTestCheckFunc(
					//resource.TestCheckResourceAttr("aws_cognito_user_pool.pool", "schema.#", "0"),
					resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.#", "1"),
					//resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.3072888811.attribute_data_type", "String"),
					//resource.TestCheckResourceAttr("aws_cognito_user_pool_schema_custom_attributes.custom_attribute_1", "schema.3072888811.name", attributeName),
				),
			},
		},
	})
}

func testAccAWSCognitoUserPoolAddCustomAttributeConfig_basic(userPoolName, attributeName string) string {
	return fmt.Sprintf(`
	resource "aws_cognito_user_pool" "pool" {
		name = "%s"
	}
	
	resource "aws_cognito_user_pool_schema_custom_attributes" "custom_attribute_1" {
		user_pool_id        = "${aws_cognito_user_pool.pool.id}"
		schema {
			attribute_data_type      = "String"
			developer_only_attribute = false
			mutable                  = true
			name                     = "%s"
			required                 = false
		  }
		schema {
			attribute_data_type      = "String"
			developer_only_attribute = false
			mutable                  = true
			name                     = "updated_at"
			required                 = false
		  }
	}
	`, userPoolName, attributeName)
}
