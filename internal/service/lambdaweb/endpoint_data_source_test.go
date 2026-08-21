// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdaweb/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaWebEndpointDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambdaweb_endpoint.test"
	resourceName := "aws_lambdaweb_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaWebServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccEndpointDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDomainName, resourceName, names.AttrDomainName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEndpointType, string(awstypes.EndpointTypeHomeRegion)),
					resource.TestCheckResourceAttr(dataSourceName, "auth_type", string(awstypes.AuthTypeApplicationManaged)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, string(awstypes.EndpointStateActive)),
				),
			},
		},
	})
}

func testAccEndpointDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccEndpointConfig_basic(rName), `
data "aws_lambdaweb_endpoint" "test" {
  function_name = aws_lambdaweb_endpoint.test.function_name
  endpoint_name = aws_lambdaweb_endpoint.test.endpoint_name
}
`)
}
