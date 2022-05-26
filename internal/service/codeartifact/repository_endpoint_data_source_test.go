package codeartifact_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codeartifact"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccRepositoryEndpointDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRepositoryEndpointBasicConfig(rName, "npm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckRepositoryEndpointBasicConfig(rName, "pypi"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckRepositoryEndpointBasicConfig(rName, "maven"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckRepositoryEndpointBasicConfig(rName, "nuget"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccRepositoryEndpointDataSource_owner(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckRepositoryEndpointOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccCheckRepositoryEndpointBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}
`, rName)
}

func testAccCheckRepositoryEndpointBasicConfig(rName, format string) string {
	return acctest.ConfigCompose(
		testAccCheckRepositoryEndpointBaseConfig(rName),
		fmt.Sprintf(`
data "aws_codeartifact_repository_endpoint" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = %[1]q
}
`, format))
}

func testAccCheckRepositoryEndpointOwnerConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccCheckRepositoryEndpointBaseConfig(rName),
		`
data "aws_codeartifact_repository_endpoint" "test" {
  domain       = aws_codeartifact_domain.test.domain
  repository   = aws_codeartifact_repository.test.repository
  domain_owner = aws_codeartifact_domain.test.owner
  format       = "npm"
}
`)
}
