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

func TestAccDataExchangeRevision_basic(t *testing.T) {
	var proj dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(dataexchange.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, dataexchange.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataExchangeRevisionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataExchangeRevisionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataExchangeRevisionExists(resourceName, &proj),
					resource.TestCheckResourceAttrPair(resourceName, "data_set_id", "aws_dataexchange_data_set.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "dataexchange", regexp.MustCompile(`data-sets/.+/revisions/.+`)),
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

func TestAccDataExchangeRevision_tags(t *testing.T) {
	var proj dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(dataexchange.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, dataexchange.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataExchangeRevisionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataExchangeRevisionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataExchangeRevisionExists(resourceName, &proj),
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
				Config: testAccDataExchangeRevisionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataExchangeRevisionExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDataExchangeRevisionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataExchangeRevisionExists(resourceName, &proj),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccDataExchangeRevision_disappears(t *testing.T) {
	var proj dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(dataexchange.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, dataexchange.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataExchangeRevisionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataExchangeRevisionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataExchangeRevisionExists(resourceName, &proj),
					acctest.CheckResourceDisappears(acctest.Provider, tfdataexchange.ResourceRevision(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfdataexchange.ResourceRevision(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataExchangeRevision_disappears_dataSet(t *testing.T) {
	var proj dataexchange.GetRevisionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_dataexchange_revision.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(dataexchange.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, dataexchange.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDataExchangeRevisionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataExchangeRevisionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataExchangeRevisionExists(resourceName, &proj),
					acctest.CheckResourceDisappears(acctest.Provider, tfdataexchange.ResourceDataSet(), "aws_dataexchange_data_set.test"),
					acctest.CheckResourceDisappears(acctest.Provider, tfdataexchange.ResourceRevision(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDataExchangeRevisionExists(n string, v *dataexchange.GetRevisionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeConn

		dataSetId, revisionId, err := tfdataexchange.RevisionParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := tfdataexchange.FindRevisionById(conn, dataSetId, revisionId)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("DataExchange Revision not found")
		}

		*v = *resp

		return nil
	}
}

func testAccCheckDataExchangeRevisionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataExchangeConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dataexchange_revision" {
			continue
		}

		dataSetId, revisionId, err := tfdataexchange.RevisionParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Try to find the resource
		_, err = tfdataexchange.FindRevisionById(conn, dataSetId, revisionId)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DataExchange Revision %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccDataExchangeRevisionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id
}
`, rName)
}

func testAccDataExchangeRevisionConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDataExchangeRevisionConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_dataexchange_data_set" "test" {
  asset_type  = "S3_SNAPSHOT"
  description = %[1]q
  name        = %[1]q
}

resource "aws_dataexchange_revision" "test" {
  data_set_id = aws_dataexchange_data_set.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
