package dataexchange_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/dataexchange"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdataexchange "github.com/hashicorp/terraform-provider-aws/internal/service/dataexchange"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDataExchangeDataSet_basic(t *testing.T) {
	var proj dataexchange.GetDataSetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_dataexchange_data_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(dataexchange.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, dataexchange.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dataexchange", regexp.MustCompile(`data-sets/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataSetConfig_basic(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "description", rNameUpdated),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dataexchange", regexp.MustCompile(`data-sets/.+`)),
				),
			},
		},
	})
}

func TestAccDataExchangeDataSet_tags(t *testing.T) {
	var proj dataexchange.GetDataSetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_data_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(dataexchange.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, dataexchange.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(resourceName, &proj),
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
				Config: testAccDataSetConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDataSetConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDataExchangeDataSet_disappears(t *testing.T) {
	var proj dataexchange.GetDataSetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_data_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(dataexchange.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, dataexchange.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSetExists(resourceName, &proj),
					acctest.CheckResourceDisappears(acctest.Provider, tfdataexchange.ResourceDataSet(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdataexchange.ResourceDataSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataSetExists(n string, v *dataexchange.GetDataSetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeConn
		resp, err := tfdataexchange.FindDataSetById(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DataExchange DataSet not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckDataSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dataexchange_data_set" {
			continue
		}

		// Try to find the resource
		_, err := tfdataexchange.FindDataSetById(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DataExchange DataSet %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccDataSetConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}
`, rName)
}

func testAccDataSetConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDataSetConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
