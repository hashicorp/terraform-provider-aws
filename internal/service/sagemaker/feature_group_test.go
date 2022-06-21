package sagemaker_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccSageMakerFeatureGroup_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":                         testAccFeatureGroup_basic,
		"description":                   testAccFeatureGroup_description,
		"disappears":                    TestAccSageMakerFeatureGroup_disappears,
		"multipleFeatures":              testAccFeatureGroup_multipleFeatures,
		"offlineConfig_basic":           testAccFeatureGroup_offlineConfig_basic,
		"offlineConfig_createCatalog":   testAccFeatureGroup_offlineConfig_createCatalog,
		"offlineConfig_providedCatalog": TestAccSageMakerFeatureGroup_Offline_providedCatalog,
		"onlineConfigSecurityConfig":    testAccFeatureGroup_onlineConfigSecurityConfig,
		"tags":                          testAccFeatureGroup_tags,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccFeatureGroup_basic(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "event_time_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "record_identifier_feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.enable_online_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_name", rName),
					resource.TestCheckResourceAttr(resourceName, "feature_definition.0.feature_type", "String"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("feature-group/%s", rName)),
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

func testAccFeatureGroup_description(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
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
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFeatureGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFeatureGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccFeatureGroup_multipleFeatures(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_multi(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
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
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_onlineSecurity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.enable_online_store", "true"),
					resource.TestCheckResourceAttr(resourceName, "online_store_config.0.security_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "online_store_config.0.security_config.0.kms_key_id", "aws_kms_key.test", "arn"),
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

func testAccFeatureGroup_offlineConfig_basic(t *testing.T) {
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_offlineBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", "true"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "0"),
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
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_offlineCreateGlueCatalog(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", "false"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.catalog", "AwsDataCatalog"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.database", "sagemaker_featurestore"),
					resource.TestMatchResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.0.table_name", regexp.MustCompile(fmt.Sprintf("^%s-", rName))),
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
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"
	glueTableResourceName := "aws_glue_catalog_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_offlineCreateGlueCatalogProvidedCatalog(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					resource.TestCheckResourceAttr(resourceName, "feature_group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.disable_glue_table_creation", "true"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.s3_storage_config.0.s3_uri", fmt.Sprintf("s3://%s/prefix/", rName)),
					resource.TestCheckResourceAttr(resourceName, "offline_store_config.0.data_catalog_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.catalog", glueTableResourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.database", glueTableResourceName, "database_name"),
					resource.TestCheckResourceAttrPair(resourceName, "offline_store_config.0.data_catalog_config.0.table_name", glueTableResourceName, "name"),
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
	var featureGroup sagemaker.DescribeFeatureGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_feature_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFeatureGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFeatureGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFeatureGroupExists(resourceName, &featureGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceFeatureGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFeatureGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_feature_group" {
			continue
		}

		_, err := tfsagemaker.FindFeatureGroupByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SageMaker Feature Group %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFeatureGroupExists(n string, v *sagemaker.DescribeFeatureGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SageMaker Feature Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

		output, err := tfsagemaker.FindFeatureGroupByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFeatureGroupBaseConfig(rName string) string {
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

func testAccFeatureGroupOfflineBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_acl" "test" {
  bucket = aws_s3_bucket.test.id
  acl    = "private"
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
        "${aws_s3_bucket.test.arn}",
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
	return acctest.ConfigCompose(testAccFeatureGroupBaseConfig(rName), fmt.Sprintf(`
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

func testAccFeatureGroupConfig_description(rName string) string {
	return acctest.ConfigCompose(testAccFeatureGroupBaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccFeatureGroupBaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccFeatureGroupBaseConfig(rName), fmt.Sprintf(`
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
		testAccFeatureGroupBaseConfig(rName),
		testAccFeatureGroupOfflineBaseConfig(rName),
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

func testAccFeatureGroupConfig_offlineCreateGlueCatalog(rName string) string {
	return acctest.ConfigCompose(
		testAccFeatureGroupBaseConfig(rName),
		testAccFeatureGroupOfflineBaseConfig(rName),
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
		testAccFeatureGroupBaseConfig(rName),
		testAccFeatureGroupOfflineBaseConfig(rName),
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
	return acctest.ConfigCompose(testAccFeatureGroupBaseConfig(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccFeatureGroupBaseConfig(rName), fmt.Sprintf(`
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
