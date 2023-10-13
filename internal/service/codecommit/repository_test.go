// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccCodeCommitRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_repository.test"
	var v codecommit.RepositoryMetadata

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "repository_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "default_branch"),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test description"),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "codecommit", rName),
					resource.TestCheckResourceAttrSet(resourceName, "repository_id"),
					resource.TestCheckResourceAttrSet(resourceName, "clone_url_http"),
					resource.TestCheckResourceAttrSet(resourceName, "clone_url_ssh"),
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

func TestAccCodeCommitRepository_withChanges(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_repository.test"
	var v1, v2 codecommit.RepositoryMetadata

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test description"),
				),
			},
			{
				Config: testAccRepositoryConfig_changes(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "description", "This is a test description - with changes"),
					resource.TestCheckResourceAttr(resourceName, "repository_name", rNameUpdated),
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

func TestAccCodeCommitRepository_CreateDefault_branch(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_repository.test"
	var v codecommit.RepositoryMetadata

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_defaultBranch(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_branch", "main"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codecommit_repository.test"
	var v1, v2 codecommit.RepositoryMetadata

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
					resource.TestCheckNoResourceAttr(resourceName, "default_branch"),
				),
			},
			{
				Config: testAccRepositoryConfig_defaultBranch(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "default_branch", "main"),
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

func TestAccCodeCommitRepository_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(10)
	resourceName := "aws_codecommit_repository.test"
	var v1, v2, v3 codecommit.RepositoryMetadata

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, codecommit.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v1),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryExists(ctx, resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func testAccCheckRepositoryExists(ctx context.Context, name string, v *codecommit.RepositoryMetadata) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn(ctx)
		out, err := conn.GetRepositoryWithContext(ctx, &codecommit.GetRepositoryInput{
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

		*v = *out.RepositoryMetadata

		return nil
	}
}

func testAccCheckRepositoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeCommitConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codecommit_repository" {
				continue
			}

			_, err := conn.GetRepositoryWithContext(ctx, &codecommit.GetRepositoryInput{
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
}

func testAccRepositoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
  description     = "This is a test description"
}
`, rName)
}

func testAccRepositoryConfig_changes(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
  description     = "This is a test description - with changes"
}
`, rName)
}

func testAccRepositoryConfig_defaultBranch(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
  description     = "This is a test description"
  default_branch  = "main"
}
`, rName)
}

func testAccRepositoryConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccRepositoryConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
