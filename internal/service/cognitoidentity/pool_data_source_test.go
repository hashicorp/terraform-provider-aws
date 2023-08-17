// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccCognitoIdentityPoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// var ip cognitoidentity.IdentityPool
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_cognito_identity_pool.test"
	resourceName := "aws_cognito_identity_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, cognitoidentity.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, cognitoidentity.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPoolDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPoolDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),

					// testAccCheckPoolExists(ctx, dataSourceName, &ip),
					// resource.TestCheckResourceAttrPair(dataSourceName, "address_family", resourceName, "address_family"),
					// resource.TestCheckResourceAttr(dataSourceName, "auto_minor_version_upgrade", "false"),
					// resource.TestCheckResourceAttrSet(dataSourceName, "maintenance_window_start_time.0.day_of_week"),
					// resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "user.*", map[string]string{
					// 	"console_access": "false",
					// 	"groups.#":       "0",
					// 	"username":       "Test",
					// 	"password":       "TestTest1234",
					// }),
					// acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "cognitoidentity", regexp.MustCompile(`pool:+.`)),
				),
			},
		},
	})
}

func testAccPoolDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_identity_pool" "test" {
	identity_pool_name               = "%s"
	allow_unauthenticated_identities = false
}

data "aws_cognito_identity_pool" "test" {
	identity_pool_name = aws_cognito_identity_pool.test.identity_pool_name
}
`, rName)
}
