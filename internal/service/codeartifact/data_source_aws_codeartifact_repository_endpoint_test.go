package codeartifact_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codeartifact"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSCodeArtifactRepositoryEndpointDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "npm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "pypi"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "maven"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "nuget"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactRepositoryEndpointDataSource_owner(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck: acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					acctest.CheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccCheckAWSCodeArtifactRepositoryEndpointBaseConfig(rName string) string {
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

func testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, format string) string {
	return acctest.ConfigCompose(
		testAccCheckAWSCodeArtifactRepositoryEndpointBaseConfig(rName),
		fmt.Sprintf(`
data "aws_codeartifact_repository_endpoint" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = %[1]q
}
`, format))
}

func testAccCheckAWSCodeArtifactRepositoryEndpointOwnerConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccCheckAWSCodeArtifactRepositoryEndpointBaseConfig(rName),
		`
data "aws_codeartifact_repository_endpoint" "test" {
  domain       = aws_codeartifact_domain.test.domain
  repository   = aws_codeartifact_repository.test.repository
  domain_owner = aws_codeartifact_domain.test.owner
  format       = "npm"
}
`)
}
