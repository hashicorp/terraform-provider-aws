package imagebuilder_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/hashicorp/terraform-provider-aws/internal/service/imagebuilder"
)

func TestAccImageBuilderContainerRecipe_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", regexp.MustCompile(fmt.Sprintf("container-recipe/%s/1.0.0", rName))),
					resource.TestCheckResourceAttr(resourceName, "component.#", "1"),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "component.0.component_arn", "imagebuilder", "aws", "component/update-linux/x.x.x"),
					resource.TestCheckResourceAttr(resourceName, "container_type", "DOCKER"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARNAccountID(resourceName, "parent_image", "imagebuilder", "aws", "image/amazon-linux-x86-2/x.x.x"),
					resource.TestCheckResourceAttr(resourceName, "target_repository.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_repository.0.repository_name", "aws_ecr_repository.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_repository.0.service", "ECR"),
					resource.TestCheckResourceAttr(resourceName, "version", "1.0.0"),
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

func TestAccImageBuilderContainerRecipe_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_container_recipe.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckContainerRecipeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerRecipeNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerRecipeExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfimagebuilder.ResourceContainerRecipe(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckContainerRecipeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_container_recipe" {
			continue
		}

		input := &imagebuilder.GetContainerRecipeInput{
			ContainerRecipeArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetContainerRecipe(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Container Recipe (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Container Recipe (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckContainerRecipeExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

		input := &imagebuilder.GetContainerRecipeInput{
			ContainerRecipeArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetContainerRecipe(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Container Recipe (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccContainerRecipeBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_ecr_repository" "test" {
  name = %[1]q
}

`, rName)
}

func testAccContainerRecipeNameConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccContainerRecipeBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_imagebuilder_container_recipe" "test" {
  container_type = "DOCKER"
  name = %[1]q
  parent_image = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:image/amazon-linux-x86-2/x.x.x"
	version = "1.0.0"

	component {
		component_arn = "arn:${data.aws_partition.current.partition}:imagebuilder:${data.aws_region.current.name}:aws:component/update-linux/x.x.x"
	}

  dockerfile_template_data = <<EOF
FROM {{{ imagebuilder:parentImage }}}
{{{ imagebuilder:environments }}}
{{{ imagebuilder:components }}}
EOF

  target_repository {
    repository_name = aws_ecr_repository.test.name
    service         = "ECR"
  }
}
`, rName))
}
