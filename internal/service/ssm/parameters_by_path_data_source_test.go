package ssm_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssm"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSSMParametersByPathDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_parameters_by_path.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckParametersByPathDataSourceConfig(rName1, rName2, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "false"),
					resource.TestCheckResourceAttr(resourceName, "recursive", "false"),
				),
			},
		},
	})
}

func testAccCheckParametersByPathDataSourceConfig(rName1, rName2 string, withDecryption bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test1" {
  name  = "/%[1]s/param-a"
  type  = "String"
  value = "TestValueA"
}

resource "aws_ssm_parameter" "test2" {
  name  = "/%[1]s/param-b"
  type  = "String"
  value = "TestValueB"
}

resource "aws_ssm_parameter" "test3" {
  name  = "/%[2]s/param-c"
  type  = "String"
  value = "TestValueC"
}

data "aws_ssm_parameters_by_path" "test" {
  path            = "/%[1]s"
  with_decryption = %[3]t

  depends_on = [
    aws_ssm_parameter.test1,
    aws_ssm_parameter.test2,
    aws_ssm_parameter.test3,
  ]
}
`, rName1, rName2, withDecryption)
}

func TestAccSSMParametersByPathDataSource_withRecursion(t *testing.T) {
	resourceName := "data.aws_ssm_parameters_by_path.recursive"
	pathPrefix := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssm.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckParametersByPathDataSourceConfigWithRecursion(pathPrefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "recursive", "true"),
				),
			},
		},
	})
}

func testAccCheckParametersByPathDataSourceConfigWithRecursion(pathPrefix string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "top_level" {
  name  = "/%[1]s/top_param"
  type  = "String"
  value = "TestValueA"
}

resource "aws_ssm_parameter" "nested" {
  name  = "/%[1]s/nested/param"
  type  = "String"
  value = "TestValueB"
}

data "aws_ssm_parameters_by_path" "recursive" {
  path      = "/%[1]s/"
  recursive = true

  depends_on = [
    aws_ssm_parameter.top_level,
    aws_ssm_parameter.nested,
  ]
}
`, pathPrefix)
}
