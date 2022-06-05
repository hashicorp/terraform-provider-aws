package codecommit_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccCodeCommitRepository_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_codecommit_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("aws_codecommit_repository.test"),
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

func TestAccCodeCommitRepository_withChanges(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_codecommit_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("aws_codecommit_repository.test"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_repository.test", "description", "This is a test description"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_changes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("aws_codecommit_repository.test"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_repository.test", "description", "This is a test description - with changes"),
				),
			},
		},
	})
}

func TestAccCodeCommitRepository_CreateDefault_branch(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_codecommit_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_defaultBranch(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("aws_codecommit_repository.test"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_repository.test", "default_branch", "master"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"default_branch"},
			},
		},
	})
}

func TestAccCodeCommitRepository_CreateAndUpdateDefault_branch(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resourceName := "aws_codecommit_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("aws_codecommit_repository.test"),
					resource.TestCheckNoResourceAttr(
						"aws_codecommit_repository.test", "default_branch"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_defaultBranch(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists("aws_codecommit_repository.test"),
					resource.TestCheckResourceAttr(
						"aws_codecommit_repository.test", "default_branch", "master"),
				),
			},
		},
	})
}

func TestAccCodeCommitRepository_tags(t *testing.T) {
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codecommit_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, codecommit.EndpointsID),
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

func testAccCheckRepositoryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn
		out, err := conn.GetRepository(&codecommit.GetRepositoryInput{
			RepositoryName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if out.RepositoryMetadata.Arn == nil {
			return fmt.Errorf("No CodeCommit Repository Vault Found")
		}

		if *out.RepositoryMetadata.RepositoryName != rs.Primary.ID {
			return fmt.Errorf("CodeCommit Repository Mismatch - existing: %q, state: %q",
				*out.RepositoryMetadata.RepositoryName, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codecommit_repository" {
			continue
		}

		_, err := conn.GetRepository(&codecommit.GetRepositoryInput{
			RepositoryName: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeRepositoryDoesNotExistException) {
			continue
		}

		if err == nil {
			return fmt.Errorf("Repository still exists: %s", rs.Primary.ID)
		}
		return err
	}

	return nil
}

func testAccRepositoryConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = "test_repository_%d"
  description     = "This is a test description"
}
`, rInt)
}

func testAccRepositoryConfig_changes(rInt int) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = "test_repository_%d"
  description     = "This is a test description - with changes"
}
`, rInt)
}

func testAccRepositoryConfig_defaultBranch(rInt int) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = "test_repository_%d"
  description     = "This is a test description"
  default_branch  = "master"
}
`, rInt)
}

func testAccRepositoryConfig_tags1(r, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = "terraform-test-%s"

  tags = {
    %q = %q
  }
}
`, r, tag1Key, tag1Value)
}

func testAccRepositoryConfig_tags2(r, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = "terraform-test-%s"

  tags = {
    %q = %q
    %q = %q
  }
}
`, r, tag1Key, tag1Value, tag2Key, tag2Value)
}
