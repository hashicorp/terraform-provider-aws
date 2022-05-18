package sagemaker_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func TestAccSageMakerPrebuiltECRImageDataSource_basic(t *testing.T) {
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_factorMachines[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "kmeans", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_region(t *testing.T) {
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_sparkML[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageConfig_explicitRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-scikit-learn", "2.2-1.0.11.0")),
				),
			},
		},
	})
}

const testAccPrebuiltECRImageConfig_basic = `
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
`

const testAccPrebuiltECRImageConfig_explicitRegion = `
data "aws_region" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-scikit-learn"
  image_tag       = "2.2-1.0.11.0"
  region          = data.aws_region.current.name
}
`
