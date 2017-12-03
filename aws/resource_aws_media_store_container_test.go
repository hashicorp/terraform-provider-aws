package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsMediaStoreContainer_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerExists("aws_media_store_container.test"),
				),
			},
			{
				Config: testAccMediaStoreContainerConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerExists("aws_media_store_container.test"),
					resource.TestCheckResourceAttrSet("aws_media_store_container.test", "policy"),
				),
			},
		},
	})
}

func testAccCheckAwsMediaStoreContainerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_container" {
			continue
		}

		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeContainer(input)
		if err != nil {
			if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
				return nil
			}
			return err
		}

		if *resp.Container.Status != mediastore.ContainerStatusDeleting {
			return fmt.Errorf("Media Store Container (%s) not deleted", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAwsMediaStoreContainerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccMediaStoreContainerConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}`, rName)
}

func testAccMediaStoreContainerConfig_Update(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "test" {}

resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
	policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "MediaStoreFullAccess",
    "Action": [ "mediastore:*" ],
    "Principal": {"AWS" : "*"},
    "Effect": "Allow",
    "Resource": "arn:aws:mediastore:us-west-2:${data.aws_caller_identity.test.account_id}:container/tf_mediastore_%s/*",
    "Condition": {
      "Bool": { "aws:SecureTransport": "true" }
    }
  }]
}
POLICY
}`, rName, rName)
}
