package rekognition_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccRekognitionCollection_Resource_basic(t *testing.T) {
	var collection rekognition.DescribeCollectionOutput
	resourceName := "aws_rekognition_collection.test"
	rName := fmt.Sprintf("test-collection-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rekognition.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, "collection_id", rName),
					resource.TestCheckResourceAttrSet(resourceName, "collection_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "face_count"),
					resource.TestCheckResourceAttrSet(resourceName, "face_model_version"),
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

func TestAccRekognitionCollection_Resource_tags(t *testing.T) {
	var collection rekognition.DescribeCollectionOutput
	resourceName := "aws_rekognition_collection.test"
	rName := fmt.Sprintf("test-collection-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, rekognition.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCollectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCollectionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(resourceName, &collection),
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
				Config: testAccCollectionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCollectionConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCollectionExists(resourceName, &collection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckCollectionExists(n string, res *rekognition.DescribeCollectionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No rekognition id is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionConn

		collection, err := conn.DescribeCollection(&rekognition.DescribeCollectionInput{
			CollectionId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *collection

		return nil
	}
}

func testAccCheckCollectionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_rekognition_collection" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).RekognitionConn
		_, err := conn.DescribeCollection(&rekognition.DescribeCollectionInput{
			CollectionId: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Collection %s still exists", rs.Primary.ID)
		}

		if tfawserr.ErrCodeEquals(err, rekognition.ErrCodeResourceNotFoundException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccCollectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
  collection_id = "%s"
}
`, rName)
}

func testAccCollectionConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
  collection_id = "%s"

  tags = {
    %q = %q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccCollectionConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_rekognition_collection" "test" {
  collection_id = "%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
