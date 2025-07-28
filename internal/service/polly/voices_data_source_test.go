// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package polly_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccPollyVoicesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_polly_voices.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PollyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PollyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccVoicesDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					// verify a known voice is returned in the results
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "voices.*", map[string]string{
						"gender":               "Female",
						names.AttrLanguageCode: "en-US",
						names.AttrName:         "Kendra",
					}),
				),
			},
		},
	})
}

func TestAccPollyVoicesDataSource_languageCode(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_polly_voices.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.PollyEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.PollyServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccVoicesDataSourceConfig_languageCode(string(types.LanguageCodeEnUs)),
				Check: resource.ComposeTestCheckFunc(
					// verify a known voice is returned in the results
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "voices.*", map[string]string{
						"gender":               "Female",
						names.AttrLanguageCode: "en-US",
						names.AttrName:         "Kendra",
					}),
				),
			},
		},
	})
}

func testAccVoicesDataSourceConfig_basic() string {
	return `
data "aws_polly_voices" "test" {}
`
}

func testAccVoicesDataSourceConfig_languageCode(languageCode string) string {
	return fmt.Sprintf(`
data "aws_polly_voices" "test" {
  language_code = %[1]q
}
`, languageCode)
}
