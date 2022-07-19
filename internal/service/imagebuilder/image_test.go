package imagebuilder_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
)

func TestAccImageBuilderImage_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	imageRecipeResourceName := "aws_imagebuilder_image_recipe.test"
	infrastructureConfigurationResourceName := "aws_imagebuilder_infrastructure_configuration.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", regexp.MustCompile(fmt.Sprintf("image/%s/1.0.0/[1-9][0-9]*", rName))),
					resource.TestCheckNoResourceAttr(resourceName, "container_recipe_arn"),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckNoResourceAttr(resourceName, "distribution_configuration_arn"),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "image_recipe_arn", imageRecipeResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "720"),
					resource.TestCheckResourceAttrPair(resourceName, "infrastructure_configuration_arn", infrastructureConfigurationResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(resourceName, "os_version", "Amazon Linux 2"),
					resource.TestCheckResourceAttr(resourceName, "output_resources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "version", regexp.MustCompile(`1.0.0/[1-9][0-9]*`)),
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

func TestAccImageBuilderImage_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_required(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfimagebuilder.ResourceImage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderImage_distributionARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	distributionConfigurationResourceName := "aws_imagebuilder_distribution_configuration.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_distributionConfigurationARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "distribution_configuration_arn", distributionConfigurationResourceName, "arn"),
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

func TestAccImageBuilderImage_enhancedImageMetadataEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_enhancedMetadataEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enhanced_image_metadata_enabled", "false"),
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

func TestAccImageBuilderImage_ImageTests_imageTestsEnabled(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_testsConfigurationTestsEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.image_tests_enabled", "false"),
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

func TestAccImageBuilderImage_ImageTests_timeoutMinutes(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_testsConfigurationTimeoutMinutes(rName, 721),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_tests_configuration.0.timeout_minutes", "721"),
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

func TestAccImageBuilderImage_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
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
				Config: testAccImageConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccImageConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccImageBuilderImage_containerRecipeARN(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_image.test"
	containerRecipeResourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccImageConfig_containerRecipeARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckImageExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "container_recipe_arn", containerRecipeResourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckImageDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_image_pipeline" {
			continue
		}

		input := &imagebuilder.GetImageInput{
			ImageBuildVersionArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetImage(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Image (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckImageExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

		input := &imagebuilder.GetImageInput{
			ImageBuildVersionArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetImage(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Image (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccImageBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_imagebuilder_component" "update-linux" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/1.0.0"
}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.AmazonSSMManagedInstanceCore,
    aws_iam_role_policy_attachment.EC2InstanceProfileForImageBuilder,
  ]
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}

resource "aws_iam_role_policy_attachment" "AmazonSSMManagedInstanceCore" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilder" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilder"
  role       = aws_iam_role.test.name
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  ingress {
    from_port = 0
    protocol  = -1
    self      = true
    to_port   = 0
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id
}

resource "aws_imagebuilder_image_recipe" "test" {
  component {
    component_arn = data.aws_imagebuilder_component.update-linux.arn
  }

  name         = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
  version      = "1.0.0"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_default_security_group.test.id]
  subnet_id             = aws_subnet.test.id

  depends_on = [aws_default_route_table.test]
}
`, rName)
}

func testAccImageConfig_distributionConfigurationARN(rName string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}

resource "aws_imagebuilder_image" "test" {
  distribution_configuration_arn   = aws_imagebuilder_distribution_configuration.test.arn
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName))
}

func testAccImageConfig_enhancedMetadataEnabled(rName string, enhancedImageMetadataEnabled bool) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  enhanced_image_metadata_enabled  = %[2]t
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName, enhancedImageMetadataEnabled))
}

func testAccImageConfig_testsConfigurationTestsEnabled(rName string, imageTestsEnabled bool) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  image_tests_configuration {
    image_tests_enabled = %[2]t
  }
}
`, rName, imageTestsEnabled))
}

func testAccImageConfig_testsConfigurationTimeoutMinutes(rName string, timeoutMinutes int) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  image_tests_configuration {
    timeout_minutes = %[2]d
  }
}
`, rName, timeoutMinutes))
}

func testAccImageConfig_required(rName string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`)
}

func testAccImageConfig_tags1(rName string, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccImageConfig_tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccImageBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccImageConfig_containerRecipeARN(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }
}

resource "aws_default_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    cidr_blocks = ["0.0.0.0/0"]
    from_port   = 0
    protocol    = "-1"
    to_port     = 0
  }

  ingress {
    from_port = 0
    protocol  = -1
    self      = true
    to_port   = 0
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  map_public_ip_on_launch = true
  vpc_id                  = aws_vpc.test.id
}

resource "aws_iam_role" "test" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      Sid = ""
    }]
  })
  name = %[1]q
}

resource "aws_iam_role_policy_attachment" "AmazonSSMManagedInstanceCore" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonSSMManagedInstanceCore"
  role       = aws_iam_role.test.name
}

resource "aws_iam_role_policy_attachment" "EC2InstanceProfileForImageBuilderECRContainerBuilds" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/EC2InstanceProfileForImageBuilderECRContainerBuilds"
  role       = aws_iam_role.test.name
}

resource "aws_iam_instance_profile" "test" {
  name = aws_iam_role.test.name
  role = aws_iam_role.test.name

  depends_on = [
    aws_iam_role_policy_attachment.AmazonSSMManagedInstanceCore,
    aws_iam_role_policy_attachment.EC2InstanceProfileForImageBuilderECRContainerBuilds
  ]
}

resource "aws_ecr_repository" "test" {
  name = %[1]q
}

data "aws_imagebuilder_component" "update-linux" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/1.0.0"
}

resource "aws_imagebuilder_container_recipe" "test" {
  component {
    component_arn = data.aws_imagebuilder_component.update-linux.arn
  }

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  name           = %[1]q
  container_type = "DOCKER"
  parent_image   = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-latest/x.x.x"
  version        = "1.0.0"
  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test.name
  name                  = %[1]q
  security_group_ids    = [aws_default_security_group.test.id]
  subnet_id             = aws_subnet.test.id

  depends_on = [aws_default_route_table.test]
}

resource "aws_imagebuilder_image" "test" {
  container_recipe_arn             = aws_imagebuilder_container_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}
`, rName)
}
