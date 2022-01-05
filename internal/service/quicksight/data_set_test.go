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
	resourceName := "aws_quicksight_data_set.test"
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
					// resource.TestCheckResourceAttr(resourceName, "data_set_id", rId),
					// acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("dataset/%s", rId)),
					// resource.TestCheckResourceAttr(resourceName, "name", rName),
					// resource.TestCheckResourceAttr(resourceName, "parameters.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.#", "1"),
					// resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.bucket", rName),
					// resource.TestCheckResourceAttr(resourceName, "parameters.0.s3.0.manifest_file_location.0.key", rName),
					// resource.TestCheckResourceAttr(resourceName, "type", quicksight.DataSourceTypeS3),
					// resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccBaseDataSetConfig() string {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	return acctest.ConfigCompose(
		testAccBaseDataSourceConfig(rName),
		testAccDataSourceConfig(rName, rId),
	)
}

func testAccDataSetConfig(rId, rName string) string {
	return acctest.ConfigCompose(
		testAccBaseDataSetConfig(),
		fmt.Sprintf(`
resource "aws_quicksight_data_set" "test" {
  data_set_id = %[1]q
  name        = %[2]q

  import_mode = "SPICE"
  physical_table_map 
}
`, rId, rName))
}
