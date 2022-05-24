package sagemaker_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
)

func TestAccSageMakerImage_basic(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "image_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("image/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
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

func TestAccSageMakerImage_description(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccImageDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
				),
			},
		},
	})
}

func TestAccSageMakerImage_displayName(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageDisplayName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "display_name", ""),
				),
			},
			{
				Config: testAccImageDisplayName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "display_name", rName),
				),
			},
		},
	})
}

func TestAccSageMakerImage_tags(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
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
				Config: testAccImageTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccImageTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerImage_disappears(t *testing.T) {
	var image sagemaker.DescribeImageOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName, &image),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceImage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckImageDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_image" {
			continue
		}

		Image, err := tfsagemaker.FindImageByName(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SageMaker Image (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(Image.ImageName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Image %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckImageExists(n string, image *sagemaker.DescribeImageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Image ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindImageByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*image = *resp

		return nil
	}
}

func testAccImageBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.${data.aws_partition.current.dns_suffix}"]
    }
  }
}
`, rName)
}

func testAccImageBasicConfig(rName string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn
}
`, rName)
}

func testAccImageDescription(rName string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name  = %[1]q
  role_arn    = aws_iam_role.test.arn
  description = %[1]q
}
`, rName)
}

func testAccImageDisplayName(rName string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name   = %[1]q
  role_arn     = aws_iam_role.test.arn
  display_name = %[1]q
}
`, rName)
}

func testAccImageTags1Config(rName, tagKey1, tagValue1 string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccImageTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccImageBaseConfig(rName) + fmt.Sprintf(`
resource "aws_sagemaker_image" "test" {
  image_name = %[1]q
  role_arn   = aws_iam_role.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
