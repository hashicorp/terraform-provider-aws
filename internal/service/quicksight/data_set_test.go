package quicksight_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
)

func TestAccQuickSightDataSet_basic(t *testing.T) {
	var dataSet quicksight.DataSet
	resourceName := "aws_quicksight_data_set.dset"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		ErrorCheck:        acctest.ErrorCheck(t, quicksight.EndpointsID),
		CheckDestroy:      testAccCheckQuickSightDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSetExists(resourceName, &dataSet),
					resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("dataset/%s", rId)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "import_mode", "SPICE"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.0.name", "ColumnId-1"),
					resource.TestCheckResourceAttr(resourceName, "physical_table_map.0.s3_source.0.input_columns.0.type", "STRING"),
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

func testAccCheckQuickSightDataSetExists(resourceName string, dataSet *quicksight.DataSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, dataSetId, err := tfquicksight.ParseDataSetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn

		input := &quicksight.DescribeDataSetInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetId),
		}

		output, err := conn.DescribeDataSet(input)

		if err != nil {
			return err
		}

		if output == nil || output.DataSet == nil {
			return fmt.Errorf("QuickSight Data Set (%s) not found", rs.Primary.ID)
		}

		*dataSet = *output.DataSet

		return nil
	}
}

func testAccCheckQuickSightDataSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_data_set" {
			continue
		}

		awsAccountID, dataSetId, err := tfquicksight.ParseDataSetID(rs.Primary.ID)
		if err != nil {
			return err
		}

		output, err := conn.DescribeDataSet(&quicksight.DescribeDataSetInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSetId:    aws.String(dataSetId),
		})

		if tfawserr.ErrMessageContains(err, quicksight.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.DataSet != nil {
			return fmt.Errorf("QuickSight Data Set (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccBaseDataSetConfig(rId string, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		fmt.Sprintf(`
data "aws_partition" "partition" {}

resource "aws_s3_bucket" "s3bucket" {
	acl           = "public-read"
	bucket        = %[1]q
	force_destroy = true
}
		
resource "aws_s3_bucket_object" "object" {
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
		
resource "aws_quicksight_data_source" "dsource" {
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

func testAccDataSetConfig(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(rId, rName),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "dset" {
	
	data_set_id = %[1]q
	name        = %[2]q
	import_mode = "SPICE"
	physical_table_map {
		s3_source {
			data_source_arn = aws_quicksight_data_source.dsource.arn
			input_columns {
				name = "ColumnId-1"
				type = "STRING"
			}
		}
	}
}
`, rId, rName))
}
