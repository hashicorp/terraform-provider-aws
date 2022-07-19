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

func TestAccSageMakerModelPackageGroup_basic(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccModelPackageGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "model_package_group_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("model-package-group/%s", rName)),
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

func TestAccSageMakerModelPackageGroup_description(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccModelPackageGroupConfig_description(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "model_package_group_description", rName),
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

func TestAccSageMakerModelPackageGroup_tags(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccModelPackageGroupConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupExists(resourceName, &mpg),
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
				Config: testAccModelPackageGroupConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccModelPackageGroupConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccSageMakerModelPackageGroup_disappears(t *testing.T) {
	var mpg sagemaker.DescribeModelPackageGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, sagemaker.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckModelPackageGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccModelPackageGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckModelPackageGroupExists(resourceName, &mpg),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceModelPackageGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckModelPackageGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_model_package_group" {
			continue
		}

		ModelPackageGroup, err := tfsagemaker.FindModelPackageGroupByName(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "does not exist") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SageMaker Model Package Group (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(ModelPackageGroup.ModelPackageGroupName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Model Package Group %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckModelPackageGroupExists(n string, mpg *sagemaker.DescribeModelPackageGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Model Package Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindModelPackageGroupByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccModelPackageGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q
}
`, rName)
}

func testAccModelPackageGroupConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name        = %[1]q
  model_package_group_description = %[1]q
}
`, rName)
}

func testAccModelPackageGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccModelPackageGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
