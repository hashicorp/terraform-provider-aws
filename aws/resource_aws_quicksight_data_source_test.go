package aws

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_quicksight_data_source", &resource.Sweeper{
		Name: "aws_quicksight_data_source",
		F:    testSweepQuickSightDataSources,
	})
}

func testSweepQuickSightDataSources(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).quicksightconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	ctx := context.Background()
	awsAccountId := client.(*AWSClient).accountid

	input := &quicksight.ListDataSourcesInput{
		AwsAccountId: aws.String(awsAccountId),
	}

	err = conn.ListDataSourcesPagesWithContext(ctx, input, func(page *quicksight.ListDataSourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ds := range page.DataSources {
			if ds == nil {
				continue
			}

			r := resourceAwsQuickSightDataSource()

			d := r.Data(nil)

			d.SetId(fmt.Sprintf("%s/%s", awsAccountId, aws.StringValue(ds.DataSourceId)))

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing QuickSigth Data Sources: %w", err))
	}

	if err := testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping QuickSight Data Sources for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping QuickSight Data Source sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestQuickSightDataSourcePermissionsDiff(t *testing.T) {
	testCases := []struct {
		name            string
		oldPermissions  []interface{}
		newPermissions  []interface{}
		expectedGrants  []*quicksight.ResourcePermission
		expectedRevokes []*quicksight.ResourcePermission
	}{
		{
			name:            "no changes;empty",
			oldPermissions:  []interface{}{},
			newPermissions:  []interface{}{},
			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name: "no changes;same",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				}},

			expectedGrants:  nil,
			expectedRevokes: nil,
		},
		{
			name:           "grant only",
			oldPermissions: []interface{}{},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action2"}),
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke only",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			newPermissions: []interface{}{},
			expectedGrants: nil,
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action2"}),
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "grant new action",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action2"}),
					Principal: aws.String("principal1"),
				},
			},
			expectedRevokes: nil,
		},
		{
			name: "revoke old action",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"oldAction",
						"onlyOldAction",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"oldAction",
					}),
				},
			},
			expectedGrants: nil,
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"onlyOldAction"}),
					Principal: aws.String("principal1"),
				},
			},
		},
		{
			name: "multiple permissions",
			oldPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
				map[string]interface{}{
					"principal": "principal2",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action3",
						"action4",
					}),
				},
				map[string]interface{}{
					"principal": "principal3",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action5",
					}),
				},
			},
			newPermissions: []interface{}{
				map[string]interface{}{
					"principal": "principal1",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action1",
						"action2",
					}),
				},
				map[string]interface{}{
					"principal": "principal2",
					"actions": schema.NewSet(schema.HashString, []interface{}{
						"action3",
						"action5",
					}),
				},
			},
			expectedGrants: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action5"}),
					Principal: aws.String("principal2"),
				},
			},
			expectedRevokes: []*quicksight.ResourcePermission{
				{
					Actions:   aws.StringSlice([]string{"action1", "action4"}),
					Principal: aws.String("principal2"),
				},
				{
					Actions:   aws.StringSlice([]string{"action5"}),
					Principal: aws.String("principal3"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			toGrant, toRevoke := diffQuickSightDataSourcePermissions(testCase.oldPermissions, testCase.newPermissions)
			if !reflect.DeepEqual(toGrant, testCase.expectedGrants) {
				t.Fatalf("Expected: %v, got: %v", testCase.expectedGrants, toGrant)
			}

			if !reflect.DeepEqual(toRevoke, testCase.expectedRevokes) {
				t.Fatalf("Expected: %v, got: %v", testCase.expectedRevokes, toRevoke)
			}
		})
	}
}

func TestAccAWSQuickSightDataSource_basic(t *testing.T) {
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rId := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightDataSourceConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_id", rId),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("datasource/%s", rId)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.bucket", rName),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.key", rName),
					resource.TestCheckResourceAttr(resourceName, "type", quicksight.DataSourceTypeS3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSQuickSightDataSource_disappears(t *testing.T) {
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rId := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightDataSourceConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					acctest.CheckResourceDisappears(testAccProvider, resourceAwsQuickSightDataSource(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSQuickSightDataSource_Tags(t *testing.T) {
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rId := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckQuickSightDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightDataSourceConfigTags1(rId, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
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
				Config: testAccAWSQuickSightDataSourceConfigTags2(rId, rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSQuickSightDataSourceConfigTags1(rId, rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSQuickSightDataSource_Permissions(t *testing.T) {
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rId := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightDataSourceConfig_Permissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permission.*", map[string]*regexp.Regexp{
						"principal": regexp.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSourcePermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:PassDataSource"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSQuickSightDataSourceConfig_UpdatePermissions(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "permission.*", map[string]*regexp.Regexp{
						"principal": regexp.MustCompile(fmt.Sprintf(`user/default/%s`, rName)),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DescribeDataSourcePermissions"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:PassDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:UpdateDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:DeleteDataSource"),
					resource.TestCheckTypeSetElemAttr(resourceName, "permission.*.actions.*", "quicksight:UpdateDataSourcePermissions"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSQuickSightDataSourceConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "permission.#", "0"),
				),
			},
		},
	})
}

func testAccCheckQuickSightDataSourceExists(resourceName string, dataSource *quicksight.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, dataSourceId, err := resourceAwsQuickSightDataSourceParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).quicksightconn

		input := &quicksight.DescribeDataSourceInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSourceId: aws.String(dataSourceId),
		}

		output, err := conn.DescribeDataSource(input)

		if err != nil {
			return err
		}

		if output == nil || output.DataSource == nil {
			return fmt.Errorf("QuickSight Data Source (%s) not found", rs.Primary.ID)
		}

		*dataSource = *output.DataSource

		return nil
	}
}

func testAccCheckQuickSightDataSourceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).quicksightconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_data_source" {
			continue
		}

		awsAccountID, dataSourceId, err := resourceAwsQuickSightDataSourceParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := conn.DescribeDataSource(&quicksight.DescribeDataSourceInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSourceId: aws.String(dataSourceId),
		})

		if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.DataSource != nil {
			return fmt.Errorf("QuickSight Data Source (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSQuickSightDataSourceConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_s3_bucket" "test" {
  acl           = "public-read"
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket  = aws_s3_bucket.test.bucket
  key     = %[1]q
  content = <<EOF
{
  "fileLocations": [
      {
          "URIs": [
              "https://${aws_s3_bucket.test.bucket}.s3.${data.aws_partition.current.dns_suffix}/%[1]s"
          ]
      }
  ],
  "globalUploadSettings": {
      "format": "JSON"
  }
}
EOF
  acl     = "public-read"
}
`, rName)
}

func testAccAWSQuickSightDataSourceConfig(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccAWSQuickSightDataSourceConfigBase(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_bucket_object.test.key
      }
    }
  }

  type = "S3"
}
`, rId, rName))
}

func testAccAWSQuickSightDataSourceConfigTags1(rId, rName, key, value string) string {
	return acctest.ConfigCompose(
		testAccAWSQuickSightDataSourceConfigBase(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_bucket_object.test.key
      }
    }
  }

  tags = {
    %[3]q = %[4]q
  }

  type = "S3"
}
`, rId, rName, key, value))
}

func testAccAWSQuickSightDataSourceConfigTags2(rId, rName, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(
		testAccAWSQuickSightDataSourceConfigBase(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_bucket_object.test.key
      }
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
  type = "S3"
}
`, rId, rName, key1, value1, key2, value2))
}

func testAccAWSQuickSightDataSource_UserConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_user" "test" {
  aws_account_id = data.aws_caller_identity.current.account_id
  user_name      = %[1]q
  email          = %[2]q
  identity_type  = "QUICKSIGHT"
  user_role      = "AUTHOR"

  lifecycle {
    create_before_destroy = true
  }
}
`, rName, testAccDefaultEmailAddress)
}

func testAccAWSQuickSightDataSourceConfig_Permissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccAWSQuickSightDataSourceConfigBase(rName),
		testAccAWSQuickSightDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_bucket_object.test.key
      }
    }
  }

  permission {
    actions = [
      "quicksight:DescribeDataSource",
      "quicksight:DescribeDataSourcePermissions",
      "quicksight:PassDataSource"
    ]

    principal = aws_quicksight_user.test.arn
  }

  type = "S3"
}
`, rId, rName))
}

func testAccAWSQuickSightDataSourceConfig_UpdatePermissions(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccAWSQuickSightDataSourceConfigBase(rName),
		testAccAWSQuickSightDataSource_UserConfig(rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_source" "test" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.test.bucket
        key    = aws_s3_bucket_object.test.key
      }
    }
  }

  permission {
    actions = [
      "quicksight:DescribeDataSource",
      "quicksight:DescribeDataSourcePermissions",
      "quicksight:PassDataSource",
      "quicksight:UpdateDataSource",
      "quicksight:DeleteDataSource",
      "quicksight:UpdateDataSourcePermissions"
    ]

    principal = aws_quicksight_user.test.arn
  }

  type = "S3"
}
`, rId, rName))
}
