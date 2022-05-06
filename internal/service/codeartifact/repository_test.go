package codeartifact_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodeartifact "github.com/hashicorp/terraform-provider-aws/internal/service/codeartifact"
)

func testAccRepository_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codeartifact", fmt.Sprintf("repository/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "repository", rName),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "administrator_account", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "0"),
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

func testAccRepository_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService("codeartifact", t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
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
				Config: testAccRepositoryConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRepositoryConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccRepository_owner(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codeartifact", fmt.Sprintf("repository/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "repository", rName),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "administrator_account", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "0"),
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

func testAccRepository_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryDescConfig(rName, "desc"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "desc"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryDescConfig(rName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "desc2"),
				),
			},
		},
	})
}

func testAccRepository_upstreams(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryUpstreams1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryUpstreams2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
					resource.TestCheckResourceAttr(resourceName, "upstream.1.repository_name", fmt.Sprintf("%s-upstream2", rName)),
				),
			},
			{
				Config: testAccRepositoryUpstreams1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
				),
			},
		},
	})
}

func testAccRepository_externalConnection(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryExternalConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.external_connection_name", "public:npmjs"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.package_format", "npm"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.status", "AVAILABLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "0"),
				),
			},
			{
				Config: testAccRepositoryExternalConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.external_connection_name", "public:npmjs"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.package_format", "npm"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.status", "AVAILABLE"),
				),
			},
		},
	})
}

func testAccRepository_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, codeartifact.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodeartifact.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no CodeArtifact repository set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn
		owner, domain, repo, err := tfcodeartifact.DecodeRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = conn.DescribeRepository(&codeartifact.DescribeRepositoryInput{
			Repository:  aws.String(repo),
			Domain:      aws.String(domain),
			DomainOwner: aws.String(owner),
		})

		return err
	}
}

func testAccCheckRepositoryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codeartifact_repository" {
			continue
		}

		owner, domain, repo, err := tfcodeartifact.DecodeRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn
		resp, err := conn.DescribeRepository(&codeartifact.DescribeRepositoryInput{
			Repository:  aws.String(repo),
			Domain:      aws.String(domain),
			DomainOwner: aws.String(owner),
		})

		if err == nil {
			if aws.StringValue(resp.Repository.Name) == repo &&
				aws.StringValue(resp.Repository.DomainName) == domain &&
				aws.StringValue(resp.Repository.DomainOwner) == owner {
				return fmt.Errorf("CodeArtifact Repository %s in Domain %s still exists", repo, domain)
			}
		}

		if tfawserr.ErrCodeEquals(err, codeartifact.ErrCodeResourceNotFoundException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccRepositoryBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}
`, rName)
}

func testAccRepositoryBasicConfig(rName string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}
`, rName)
}

func testAccRepositoryOwnerConfig(rName string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository   = %[1]q
  domain       = aws_codeartifact_domain.test.domain
  domain_owner = aws_codeartifact_domain.test.owner
}
`, rName)
}

func testAccRepositoryDescConfig(rName, desc string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository  = %[1]q
  domain      = aws_codeartifact_domain.test.domain
  description = %[2]q
}
`, rName, desc)
}

func testAccRepositoryUpstreams1Config(rName string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "upstream1" {
  repository = "%[1]s-upstream1"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  upstream {
    repository_name = aws_codeartifact_repository.upstream1.repository
  }
}
`, rName)
}

func testAccRepositoryUpstreams2Config(rName string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "upstream1" {
  repository = "%[1]s-upstream1"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "upstream2" {
  repository = "%[1]s-upstream2"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  upstream {
    repository_name = aws_codeartifact_repository.upstream1.repository
  }

  upstream {
    repository_name = aws_codeartifact_repository.upstream2.repository
  }
}
`, rName)
}

func testAccRepositoryExternalConnectionConfig(rName string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  external_connections {
    external_connection_name = "public:npmjs"
  }
}
`, rName)
}

func testAccRepositoryConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRepositoryConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
