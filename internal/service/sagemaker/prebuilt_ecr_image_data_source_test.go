// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerPrebuiltECRImageDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_factorMachines[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "kmeans", acctest.Ct1)),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_sparkML[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_explicitRegion,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-scikit-learn", "2.2-1.0.11.0")),
				),
			},
		},
	})
}

const testAccPrebuiltECRImageDataSourceConfig_basic = `
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "kmeans"
}
`

const testAccPrebuiltECRImageDataSourceConfig_explicitRegion = `
data "aws_region" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = "sagemaker-scikit-learn"
  image_tag       = "2.2-1.0.11.0"
  region          = data.aws_region.current.name
}
`
