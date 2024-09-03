// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccImageBuilderImageDataSource_ARN_aws(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig_arn(),
				Check: resource.ComposeTestCheckFunc(
					acctest.MatchResourceAttrRegionalARNAccountID(dataSourceName, names.AttrARN, "imagebuilder", "aws", regexache.MustCompile(`image/amazon-linux-2-x86/x.x.x`)),
					acctest.MatchResourceAttrRegionalARNAccountID(dataSourceName, "build_version_arn", "imagebuilder", "aws", regexache.MustCompile(`image/amazon-linux-2-x86/\d+\.\d+\.\d+/\d+`)),
					acctest.CheckResourceAttrRFC3339(dataSourceName, "date_created"),
					resource.TestCheckNoResourceAttr(dataSourceName, "distribution_configuration_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "enhanced_image_metadata_enabled", acctest.CtFalse),
					resource.TestCheckNoResourceAttr(dataSourceName, "image_recipe_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "image_scanning_configuration.#", acctest.Ct0),
					resource.TestCheckResourceAttr(dataSourceName, "image_tests_configuration.#", acctest.Ct0),
					resource.TestCheckNoResourceAttr(dataSourceName, "infrastructure_configuration_arn"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, "Amazon Linux 2 x86"),
					resource.TestCheckResourceAttr(dataSourceName, "os_version", "Amazon Linux 2"),
					resource.TestCheckResourceAttr(dataSourceName, "output_resources.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "platform", imagebuilder.PlatformLinux),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestMatchResourceAttr(dataSourceName, names.AttrVersion, regexache.MustCompile(`\d+\.\d+\.\d+/\d+`)),
				),
			},
		},
	})
}

// Verify additional fields returned by Self owned Images
func TestAccImageBuilderImageDataSource_ARN_self(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_image.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig_arnSelf(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "build_version_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "date_created", resourceName, "date_created"),
					resource.TestCheckResourceAttrPair(dataSourceName, "distribution_configuration_arn", resourceName, "distribution_configuration_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enhanced_image_metadata_enabled", resourceName, "enhanced_image_metadata_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_recipe_arn", resourceName, "image_recipe_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_scanning_configuration.#", resourceName, "image_scanning_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_tests_configuration.#", resourceName, "image_tests_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "infrastructure_configuration_arn", resourceName, "infrastructure_configuration_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "os_version", resourceName, "os_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "output_resources.#", resourceName, "output_resources.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "platform", resourceName, "platform"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func TestAccImageBuilderImageDataSource_ARN_containerRecipe(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_imagebuilder_image.test"
	resourceName := "aws_imagebuilder_image.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ImageBuilderServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckImageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccImageDataSourceConfig_arnContainerRecipe(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "container_recipe_arn", resourceName, "container_recipe_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "image_scanning_configuration.#", resourceName, "image_scanning_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "output_resources.#", resourceName, "output_resources.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "output_resources.0.containers.#", resourceName, "output_resources.0.containers.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "output_resources.0.containers.0.image_uris.#", resourceName, "output_resources.0.containers.0.image_uris.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "output_resources.0.containers.0.region", resourceName, "output_resources.0.containers.0.region"),
				),
			},
		},
	})
}

func testAccImageDataSourceConfig_arn() string {
	return `
data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_imagebuilder_image" "test" {
  arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-2-x86/x.x.x"
}
`
}

func testAccImageDataSourceConfig_arnSelf(rName string) string {
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

resource "aws_imagebuilder_image" "test" {
  image_recipe_arn                 = aws_imagebuilder_image_recipe.test.arn
  infrastructure_configuration_arn = aws_imagebuilder_infrastructure_configuration.test.arn
}

data "aws_imagebuilder_image" "test" {
  arn = aws_imagebuilder_image.test.arn
}
`, rName)
}

func testAccImageDataSourceConfig_arnContainerRecipe(rName string) string {
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
  name         = %[1]q
  force_delete = true
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

  image_scanning_configuration {
    image_scanning_enabled = true

    ecr_configuration {
      repository_name = aws_ecr_repository.test.name
      container_tags  = ["foo", "bar"]
    }
  }
}

data "aws_imagebuilder_image" "test" {
  arn = aws_imagebuilder_image.test.arn
}
`, rName)
}
