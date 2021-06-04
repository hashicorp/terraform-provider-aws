package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

func TestAccAWSSageMakerPrebuiltECRImage_basic(t *testing.T) {
	expectedID := sageMakerPrebuiltECRImageIDByRegion_FactorMachines[atest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { atest.PreCheck(t) },
		ErrorCheck: atest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:  atest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSageMakerPrebuiltECRImageConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", dataSourceAwsSageMakerPrebuiltECRImageCreatePath(expectedID, atest.Region(), atest.PartitionDNSSuffix(), "kmeans", "1")),
				),
			},
		},
	})
}

func TestAccAWSSageMakerPrebuiltECRImage_region(t *testing.T) {
	expectedID := sageMakerPrebuiltECRImageIDByRegion_SparkML[atest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { atest.PreCheck(t) },
		ErrorCheck: atest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:  atest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSageMakerPrebuiltECRImageExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", dataSourceAwsSageMakerPrebuiltECRImageCreatePath(expectedID, atest.Region(), atest.PartitionDNSSuffix(), "sagemaker-scikit-learn", "2.2-1.0.11.0")),
				),
			},
		},
	})
}

const testAccCheckAwsSageMakerPrebuiltECRImageConfig = `
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
`

const testAccCheckAwsSageMakerPrebuiltECRImageExplicitRegionConfig = `
data "aws_region" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-scikit-learn"
  image_tag       = "2.2-1.0.11.0"
  region          = data.aws_region.current.name
}
`
