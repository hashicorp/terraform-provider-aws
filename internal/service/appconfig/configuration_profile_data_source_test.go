// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			acctest.PreCheckPartitionHasService(t, names.AppConfigEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppConfigServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConfigurationProfileDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationProfileDataSourceConfig_basic(appName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrApplicationID, appResourceName, names.AttrID),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrARN, "appconfig", regexache.MustCompile(`application/([a-z\d]{4,7})/configurationprofile/+.`)),
					resource.TestMatchResourceAttr(dataSourceName, "configuration_profile_id", regexache.MustCompile(`[a-z\d]{4,7}`)),
					resource.TestCheckResourceAttr(dataSourceName, "kms_key_identifier", "alias/"+rName),
					resource.TestCheckResourceAttr(dataSourceName, "location_uri", "hosted"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(dataSourceName, "retrieval_role_arn", ""),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrType, "AWS.Freeform"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "validator.*", map[string]string{
						names.AttrContent: "{\"$schema\":\"http://json-schema.org/draft-05/schema#\",\"description\":\"BasicFeatureToggle-1\",\"title\":\"$id$\"}",
						names.AttrType:    string(awstypes.ValidatorTypeJsonSchema),
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccConfigurationProfileDataSourceConfig_basic(appName, rName string) string {
	return acctest.ConfigCompose(
		testAccApplicationConfig_name(appName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/%[1]s"
  target_key_id = aws_kms_key.test.id
}

resource "aws_appconfig_configuration_profile" "test" {
  application_id     = aws_appconfig_application.test.id
  name               = %[1]q
  kms_key_identifier = aws_kms_alias.test.name
  location_uri       = "hosted"

  validator {
    content = jsonencode({
      "$schema"   = "http://json-schema.org/draft-05/schema#"
      title       = "$id$"
      description = "BasicFeatureToggle-1"
    })

    type = "JSON_SCHEMA"
  }
}

data "aws_appconfig_configuration_profile" "test" {
  application_id           = aws_appconfig_application.test.id
  configuration_profile_id = aws_appconfig_configuration_profile.test.configuration_profile_id
}
`, rName))
}
