package mediastore_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccMediaStoreContainerPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_media_store_container_policy.test"

	rName = strings.ReplaceAll(rName, "-", "_")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, mediastore.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckContainerPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "container_name"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccContainerPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "container_name"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func testAccCheckContainerPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_container_policy" {
			continue
		}

		input := &mediastore.GetContainerPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetContainerPolicy(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, mediastore.ErrCodeContainerNotFoundException) {
				return nil
			}
			if tfawserr.ErrCodeEquals(err, mediastore.ErrCodePolicyNotFoundException) {
				return nil
			}
			if tfawserr.ErrMessageContains(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected MediaStore Container Policy to be destroyed, %s found", rs.Primary.ID)
	}
	return nil
}

func testAccCheckContainerPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreConn

		input := &mediastore.GetContainerPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetContainerPolicy(input)

		return err
	}
}

func testAccContainerPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_media_store_container" "test" {
  name = %[1]q
}

resource "aws_media_store_container_policy" "test" {
  container_name = aws_media_store_container.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid    = "lucky"
      Action = "mediastore:*"
      Principal = {
        AWS = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      }
      Effect   = "Allow"
      Resource = "arn:${data.aws_partition.current.partition}:mediastore:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:container/${aws_media_store_container.test.name}/*"
      Condition = {
        Bool = {
          "aws:SecureTransport" = "true"
        }
      }
    }]
  })
}
`, rName)
}
