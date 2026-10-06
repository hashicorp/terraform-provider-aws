// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("additional_os_capabilities"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact(string(awstypes.CapabilityAll)),
					})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("base_image_arn"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("base_image_arn"), resourceName, tfjsonpath.New("base_image_arn"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("base_image_version"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("base_image_version"), resourceName, tfjsonpath.New("base_image_version"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("build_role_arn"), tfknownvalue.GlobalARNExact("iam", "role/"+rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("code_artifact"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrURI: knownvalue.StringExact(fmt.Sprintf("s3://%s/code.zip", rName)),
						}),
					})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("cpu_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"architecture": knownvalue.StringExact(string(awstypes.ArchitectureArm64)),
						}),
					})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCreatedAt), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("environment_variables"), knownvalue.MapExact(map[string]knownvalue.Check{
						"KEY1": knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("image_arn"), checkImageARN(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("image_identifier"), checkImageARN(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("image_version"), knownvalue.NotNull()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("image_version"), resourceName, tfjsonpath.New("image_version"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrState), tfknownvalue.StringExact(awstypes.MicrovmImageVersionStateSuccessful)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.MicrovmImageVersionStatusActive)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func testAccImageVersionDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccImageConfig_base(rName), fmt.Sprintf(`
resource "aws_lambdamicrovms_image" "test" {
  name                       = %[1]q
  description                = %[1]q
  base_image_arn             = "arn:${data.aws_partition.current.partition}:lambda:${data.aws_region.current.region}:aws:microvm-image:al2023-1"
  build_role_arn             = aws_iam_role.test.arn
  additional_os_capabilities = ["ALL"]

  environment_variables = {
    KEY1 = "value1"
  }

  code_artifact {
    uri = "s3://${aws_s3_bucket.test.bucket}/${aws_s3_object.test.key}"
  }

  cpu_configuration {
    architecture = "ARM_64"
  }
}

data "aws_lambdamicrovms_image_version" "test" {
  image_identifier = aws_lambdamicrovms_image.test.arn
  image_version    = aws_lambdamicrovms_image.test.image_version
}
`, rName))
}
