package ssoadmin_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccSSOAdminPermissionSetDataSource_arn(t *testing.T) {
	dataSourceName := "data.aws_ssoadmin_permission_set.test"
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetDataSourceConfig_ssoByARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "relay_state", dataSourceName, "relay_state"),
					resource.TestCheckResourceAttrPair(resourceName, "session_duration", dataSourceName, "session_duration"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccSSOAdminPermissionSetDataSource_name(t *testing.T) {
	dataSourceName := "data.aws_ssoadmin_permission_set.test"
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionSetDataSourceConfig_ssoByName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "relay_state", dataSourceName, "relay_state"),
					resource.TestCheckResourceAttrPair(resourceName, "session_duration", dataSourceName, "session_duration"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccSSOAdminPermissionSetDataSource_nonExistent(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckInstances(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ssoadmin.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccPermissionSetDataSourceConfig_ssoByNameNonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func testAccSSOPermissionSetBaseDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  description  = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state  = "https://example.com"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName)
}

func testAccPermissionSetDataSourceConfig_ssoByARN(rName string) string {
	return acctest.ConfigCompose(
		testAccSSOPermissionSetBaseDataSourceConfig(rName),
		`
data "aws_ssoadmin_permission_set" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  arn          = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccPermissionSetDataSourceConfig_ssoByName(rName string) string {
	return acctest.ConfigCompose(
		testAccSSOPermissionSetBaseDataSourceConfig(rName),
		`
data "aws_ssoadmin_permission_set" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  name         = aws_ssoadmin_permission_set.test.name
}
`)
}

const testAccPermissionSetDataSourceConfig_ssoByNameNonExistent = `
data "aws_ssoadmin_instances" "test" {}

data "aws_ssoadmin_permission_set" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  name         = "does-not-exist"
}
`
