package mediaconvert_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/mediaconvert"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMediaConvertQueueDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_convert_queue.test"
	dataSourceName := "data.aws_media_convert_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
				),
			},
		},
	})
}

func TestAccMediaConvertQueueDataSource_withStatus(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_convert_queue.test"
	dataSourceName := "data.aws_media_convert_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_status(rName, mediaconvert.QueueStatusActive),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceName, "status"),
				),
			},
		},
	})
}

func TestAccMediaConvertQueueDataSource_withTags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_convert_queue.test"
	dataSourceName := "data.aws_media_convert_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, mediaconvert.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueueDataSourceConfig_tags(rName, "foo", "bar", "fizz", "buzz"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.foo", dataSourceName, "tags.foo"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.fizz", dataSourceName, "tags.fizz"),
				),
			},
		},
	})
}

func testAccQueueDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccQueueConfig_basic(rName), `
data "aws_media_convert_queue" "test" {
	id = aws_media_convert_queue.test.id
}
`)
}

func testAccQueueDataSourceConfig_status(rName, status string) string {
	return acctest.ConfigCompose(testAccQueueConfig_status(rName, status), `
data "aws_media_convert_queue" "test" {
	id = aws_media_convert_queue.test.id
}
`)
}

func testAccQueueDataSourceConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccQueueConfig_tags(rName, tagKey1, tagValue1, tagKey2, tagValue2), `
data "aws_media_convert_queue" "test" {
	id = aws_media_convert_queue.test.id
}
`)
}
