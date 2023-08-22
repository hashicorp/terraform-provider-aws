// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/appconfig"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAppConfigConfigurationProfileDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	appResourceName := "aws_appconfig_application.test"
	dataSourceName := "data.aws_appconfig_configuration_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, appconfig.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, appconfig.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileDataSourceConfig_basic(appName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "application_id", appResourceName, "id"),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "appconfig", regexp.MustCompile(`application/([a-z\d]{4,7})/configurationprofile/+.`)),
					resource.TestMatchResourceAttr(dataSourceName, "configuration_profile_id", regexp.MustCompile(`[a-z\d]{4,7}`)),
					resource.TestCheckResourceAttr(dataSourceName, "location_uri", "hosted"),
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "retrieval_role_arn", ""),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(dataSourceName, "type", "AWS.Freeform"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "validator.*", map[string]string{
						"content": "{\"$schema\":\"http://json-schema.org/draft-05/schema#\",\"description\":\"BasicFeatureToggle-1\",\"title\":\"$id$\"}",
						"type":    appconfig.ValidatorTypeJsonSchema,
					}),
				),
			},
		},
	})
}

func testAccConfigurationProfileDataSourceConfig_basic(appName, rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(appName),
		fmt.Sprintf(`
resource "aws_appconfig_configuration_profile" "test" {
  application_id = aws_appconfig_application.test.id
  name           = %[1]q
  location_uri   = "hosted"

  validator {
    content = jsonencode({
      "$schema"   = "http://json-schema.org/draft-05/schema#"
      title       = "$id$"
      description = "BasicFeatureToggle-1"
    })

    type = "JSON_SCHEMA"
  }

  tags = {
    key1 = "value1"
  }
}

data "aws_appconfig_configuration_profile" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
}
`, rName))
}
