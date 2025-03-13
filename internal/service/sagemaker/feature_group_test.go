// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerFeatureGroup_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic:                      testAccFeatureGroup_basic,
		"storageType":                        testAccFeatureGroup_storageType,
		"featureDefinition_collectionType":   testAccFeatureGroup_featureDefinition_collectionType,
		"featureDefinition_collectionConfig": testAccFeatureGroup_featureDefinition_collectionConfig,
		"description":                        testAccFeatureGroup_description,
		acctest.CtDisappears:                 TestAccSageMakerFeatureGroup_disappears,
		"multipleFeatures":                   testAccFeatureGroup_multipleFeatures,
		"offlineConfig_basic":                testAccFeatureGroup_offlineConfig_basic,
		"offlineConfig_format":               testAccFeatureGroup_offlineConfig_format,
		"offlineConfig_createCatalog":        testAccFeatureGroup_offlineConfig_createCatalog,
		"offlineConfig_providedCatalog":      TestAccSageMakerFeatureGroup_Offline_providedCatalog,
		"onlineConfigSecurityConfig":         testAccFeatureGroup_onlineConfigSecurityConfig,
		"onlineConfig_TTLDuration":           testAccFeatureGroup_onlineConfigTTLDuration,
		"throughputConfig":                   testAccFeatureGroup_throughputConfig,
		"tags":                               testAccFeatureGroup_tags,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccFeatureGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_time_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "record_identifier_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.enable_online_store", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					acctest.CheckResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("feature-group/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_storageType(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_storageType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.storage_type", "InMemory"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_description(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFeatureGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccFeatureGroup_multipleFeatures(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_multi(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_type", "Integral"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_onlineConfigSecurityConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_onlineSecurity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.enable_online_store", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.security_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "online_store_config.0.security_config.0.kms_key_id", "aws_kms_key.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_onlineConfigTTLDuration(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup1, featureGroup2 sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_TTLDuration(rName, "Seconds"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup1),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.ttl_duration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.ttl_duration.0.unit", "Seconds"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.ttl_duration.0.value", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureGroupConfig_TTLDuration(rName, "Minutes"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup2),
					func(*terraform.State) error {
						if !aws.ToTime(featureGroup1.CreationTime).Equal(aws.ToTime(featureGroup1.CreationTime)) {
							return errors.New("SageMaker AI Feature Group was recreated")
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.ttl_duration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.ttl_duration.0.unit", "Minutes"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.ttl_duration.0.value", "1"),
				),
			},
		},
	})
}

func testAccFeatureGroup_offlineConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_offlineBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.table_format", "Glue"),
					resource.TestCheckResourceAttr(resourceName, "throughput_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_config.0.throughput_mode", "OnDemand"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_offlineConfig_format(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_offlineTableFormat(rName, "Iceberg"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.table_format", "Iceberg"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_offlineConfig_createCatalog(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_offlineCreateGlueCatalog(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.catalog", "AwsDataCatalog"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.database", "sagemaker_featurestore"),
					resource.TestCheckResourceAttrSet(resourceName, "offline_store_config.0.data_catalog_config.0.table_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSageMakerFeatureGroup_Offline_providedCatalog(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"
	glueTableResourceName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_offlineCreateGlueCatalogProvidedCatalog(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.catalog", glueTableResourceName, names.AttrCatalogID),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.database", glueTableResourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.table_name", glueTableResourceName, names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_throughputConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_throughputConfigOnDemand(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_config.0.throughput_mode", "OnDemand"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureGroupConfig_throughputConfigProvisioned(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "throughput_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "throughput_config.0.throughput_mode", "Provisioned"),
					resource.TestCheckResourceAttr(resourceName, "throughput_config.0.provisioned_write_capacity_units", "1"),
				),
			},
		},
	})
}

func testAccFeatureGroup_featureDefinition_collectionType(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_featureDefinition_collectionType(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_name", fmt.Sprintf("%s-1", rName)),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.collection_type", "List"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.collection_config.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeatureGroup_featureDefinition_collectionConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_featureDefinition_collectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_name", fmt.Sprintf("%s-1", rName)),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.feature_type", "Fractional"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.collection_type", "Vector"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.collection_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.collection_config.0.vector_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.1.collection_config.0.vector_config.0.dimension", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSageMakerFeatureGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFeatureGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(ctx, resourceName, &featureGroup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceFeatureGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFeatureGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_feature_group" {
				continue
			}

			_, err := tfsagemaker.FindFeatureGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("SageMaker AI Feature Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFeatureGroupExists(ctx context.Context, n string, v *sagemaker.DescribeFeatureGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker AI Feature Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerClient(ctx)

		output, err := tfsagemaker.FindFeatureGroupByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFeatureGroupConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccFeatureGroupConfig_baseOffline(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_iam_policy" "test" {
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [{
      "Effect" : "Allow",
      "Resource" : [
        aws_s3_bucket.test.arn,
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Action" : [
        "s3:*"
      ]
    }]
  })
}
`, rName)
}

func testAccFeatureGroupConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }
}
`, rName))
}

func testAccFeatureGroupConfig_storageType(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
    storage_type        = "InMemory"
  }
}
`, rName))
}

func testAccFeatureGroupConfig_TTLDuration(rName, unit string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true

    ttl_duration {
      unit  = %[2]q
      value = 1
    }
  }
}
`, rName, unit))
}

func testAccFeatureGroupConfig_description(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn
  description                    = %[1]q

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }
}
`, rName))
}

func testAccFeatureGroupConfig_multi(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  feature_definition {
    feature_name = "%[1]s-2"
    feature_type = "Integral"
  }

  online_store_config {
    enable_online_store = true
  }
}
`, rName))
}

func testAccFeatureGroupConfig_onlineSecurity(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true

    security_config {
      kms_key_id = aws_kms_key.test.arn
    }
  }
}
`, rName))
}

func testAccFeatureGroupConfig_offlineBasic(rName string) string {
	return acctest.ConfigCompose(
		testAccFeatureGroupConfig_base(rName),
		testAccFeatureGroupConfig_baseOffline(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = true

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccFeatureGroupConfig_offlineTableFormat(rName, format string) string {
	return acctest.ConfigCompose(
		testAccFeatureGroupConfig_base(rName),
		testAccFeatureGroupConfig_baseOffline(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = false
    table_format                = %[2]q

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test, aws_iam_role_policy_attachment.test2]
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSageMakerFeatureStoreAccess"
}
`, rName, format))
}

func testAccFeatureGroupConfig_offlineCreateGlueCatalog(rName string) string {
	return acctest.ConfigCompose(
		testAccFeatureGroupConfig_base(rName),
		testAccFeatureGroupConfig_baseOffline(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = false

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccFeatureGroupConfig_offlineCreateGlueCatalogProvidedCatalog(rName string) string {
	return acctest.ConfigCompose(
		testAccFeatureGroupConfig_base(rName),
		testAccFeatureGroupConfig_baseOffline(rName),
		fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  name          = %[1]q
  database_name = aws_glue_catalog_database.test.name
}

resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = true

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }

    data_catalog_config {
      catalog    = aws_glue_catalog_table.test.catalog_id
      database   = aws_glue_catalog_table.test.database_name
      table_name = aws_glue_catalog_table.test.name
    }
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccFeatureGroupConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value))
}

func testAccFeatureGroupConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  online_store_config {
    enable_online_store = true
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccFeatureGroupConfig_featureDefinition_collectionType(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  feature_definition {
    feature_name    = "%[1]s-1"
    feature_type    = "String"
    collection_type = "List"
  }

  online_store_config {
    enable_online_store = true
    storage_type        = "InMemory"
  }
}
`, rName))
}

func testAccFeatureGroupConfig_featureDefinition_collectionConfig(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupConfig_base(rName), fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  feature_definition {
    feature_name    = "%[1]s-1"
    feature_type    = "Fractional"
    collection_type = "Vector"
    collection_config {
      vector_config {
        dimension = 2
      }
    }
  }

  online_store_config {
    enable_online_store = true
    storage_type        = "InMemory"
  }
}
`, rName))
}

func testAccFeatureGroupConfig_throughputConfigOnDemand(rName string) string {
	return acctest.ConfigCompose(
		testAccFeatureGroupConfig_base(rName),
		testAccFeatureGroupConfig_baseOffline(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = true

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }
  }

  throughput_config {
    throughput_mode = "OnDemand"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}

func testAccFeatureGroupConfig_throughputConfigProvisioned(rName string) string {
	return acctest.ConfigCompose(
		testAccFeatureGroupConfig_base(rName),
		testAccFeatureGroupConfig_baseOffline(rName),
		fmt.Sprintf(`
resource "aws_sagemaker_feature_group" "test" {
  feature_group_name             = %[1]q
  record_identifier_feature_name = %[1]q
  event_time_feature_name        = %[1]q
  role_arn                       = aws_iam_role.test.arn

  feature_definition {
    feature_name = %[1]q
    feature_type = "String"
  }

  offline_store_config {
    disable_glue_table_creation = true

    s3_storage_config {
      s3_uri = "s3://${aws_s3_bucket.test.bucket}/prefix/"
    }
  }

  throughput_config {
    throughput_mode                  = "Provisioned"
    provisioned_write_capacity_units = 1
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName))
}
