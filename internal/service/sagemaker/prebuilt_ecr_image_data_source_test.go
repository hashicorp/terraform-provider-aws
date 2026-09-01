// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"fmt"
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

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("kmeans"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "kmeans", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_region(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_XGBoost[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_explicitRegion("sagemaker-scikit-learn", "2.2-1.0.11.0"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-scikit-learn", "2.2-1.0.11.0")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseClarify(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_clarify[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-clarify-processing"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-clarify-processing", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseDataWrangler(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_dataWrangler[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-data-wrangler-container"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-data-wrangler-container", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseDebugger(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_debugger[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-debugger-rules"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-debugger-rules", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseInferentiaNeoInferentiaPyTorch(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_inferentiaNeo[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-neo-pytorch"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-neo-pytorch", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseInferentiaNeoTensorflowInferentia(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_inferentiaNeo[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-neo-tensorflow"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-neo-tensorflow", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseXGBoostSparkML(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_XGBoost[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-sparkml-serving"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-sparkml-serving", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseXGBoostHuggingFaceTEICPU(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_XGBoost[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("tei-cpu"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "tei-cpu", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseSageMakerCustomSageMakerChainer(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_SageMakerCustom[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-chainer"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-chainer", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseSageMakerCustomSageMakerMXNetServingEIA(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_SageMakerCustom[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-mxnet-serving-eia"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-mxnet-serving-eia", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseSageMakerCustomSageMakerTensorFlowEIA(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_SageMakerCustom[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-tensorflow-eia"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-tensorflow-eia", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseDeepLearningHuggingFacePyTorchTGIInference(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_deepLearning[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("huggingface-pytorch-tgi-inference"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "huggingface-pytorch-tgi-inference", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseDeepLearningNVIDIATritionInference(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_deepLearning[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-tritonserver"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-tritonserver", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseDeepLearningStabilityAI(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_deepLearning[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("stabilityai-pytorch-inference"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "stabilityai-pytorch-inference", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseSageMakerBasePython(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_SageMakerBasePython[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-base-python"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-base-python", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseSageMakerRL(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_SageMakerRL[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-rl-ray-container"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-rl-ray-container", "1")),
				),
			},
		},
	})
}

func TestAccSageMakerPrebuiltECRImageDataSource_caseSpark(t *testing.T) {
	ctx := acctest.Context(t)
	expectedID := tfsagemaker.PrebuiltECRImageIDByRegion_spark[acctest.Region()]

	dataSourceName := "data.aws_sagemaker_prebuilt_ecr_image.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrebuiltECRImageDataSourceConfig_basic("sagemaker-spark-processing"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrID, expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_id", expectedID),
					resource.TestCheckResourceAttr(dataSourceName, "registry_path", tfsagemaker.PrebuiltECRImageCreatePath(expectedID, acctest.Region(), acctest.PartitionDNSSuffix(), "sagemaker-spark-processing", "1")),
				),
			},
		},
	})
}

func testAccPrebuiltECRImageDataSourceConfig_basic(repository_name string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = %[1]q
}
`, repository_name))
}

func testAccPrebuiltECRImageDataSourceConfig_explicitRegion(repository_name string, image_tag string) string {
	return acctest.ConfigCompose(fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_sagemaker_prebuilt_ecr_image" "test" {
  repository_name = %[1]q
  image_tag       = %[2]q
  region          = data.aws_region.current.region
}
`, repository_name, image_tag))
}
