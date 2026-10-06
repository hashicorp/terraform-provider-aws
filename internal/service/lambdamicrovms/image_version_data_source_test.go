// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaMicroVMsImageVersionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_lambdamicrovms_image_version.test"
	resourceName := "aws_lambdamicrovms_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaMicroVMsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccImageVersionDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "base_image_arn", resourceName, "base_image_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "build_role_arn", resourceName, "build_role_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "code_artifact.0.uri", resourceName, "code_artifact.0.uri"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_version", resourceName, "image_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrState), knownvalue.StringExact(string(awstypes.MicrovmImageVersionStateSuccessful))),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact(string(awstypes.MicrovmImageVersionStatusActive))),
				},
			},
		},
	})
}

func testAccImageVersionDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccImageConfig_basic(rName), `
data "aws_lambdamicrovms_image_version" "test" {
  image_identifier = aws_lambdamicrovms_image.test.arn
  image_version    = aws_lambdamicrovms_image.test.image_version
}
`)
}
