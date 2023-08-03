// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codegurureviewer_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codegurureviewer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcodegurureviewer "github.com/hashicorp/terraform-provider-aws/internal/service/codegurureviewer"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Repository types of "BitBucket and GitHubEnterpriseServer cannot be tested, as they require their CodeStar Connection to be in "AVAILABLE" status vs "PENDING", requiring console interaction
// However, this has been manually tested successfully
func TestAccCodeGuruReviewerRepositoryAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation codegurureviewer.DescribeRepositoryAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codegurureviewer_repository_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, codegurureviewer.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codegurureviewer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codeguru-reviewer", regexp.MustCompile(`association:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "id", "codeguru-reviewer", regexp.MustCompile(`association:+.`)),
					resource.TestCheckResourceAttr(resourceName, "repository.0.bitbucket.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.github_enterprise_server.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.0.encryption_option", "AWS_OWNED_CMK"),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_KMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation codegurureviewer.DescribeRepositoryAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codegurureviewer_repository_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, codegurureviewer.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codegurureviewer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_kms_key(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codeguru-reviewer", regexp.MustCompile(`association:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "id", "codeguru-reviewer", regexp.MustCompile(`association:+.`)),
					resource.TestCheckResourceAttr(resourceName, "repository.0.bitbucket.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.github_enterprise_server.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.0.encryption_option", "CUSTOMER_MANAGED_CMK"),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_S3Repository(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation codegurureviewer.DescribeRepositoryAssociationOutput
	rName := "codeguru-reviewer-" + sdkacctest.RandString(10)
	resourceName := "aws_codegurureviewer_repository_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, codegurureviewer.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codegurureviewer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_s3_repository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "codeguru-reviewer", regexp.MustCompile(`association:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "id", "codeguru-reviewer", regexp.MustCompile(`association:+.`)),
					resource.TestCheckResourceAttr(resourceName, "repository.0.bitbucket.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.github_enterprise_server.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.0.bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.0.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.0.encryption_option", "AWS_OWNED_CMK"),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation codegurureviewer.DescribeRepositoryAssociationOutput
	resourceName := "aws_codegurureviewer_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, codegurureviewer.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codegurureviewer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_tags_1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccRepositoryAssociationConfig_tags_2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRepositoryAssociationConfig_tags_1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation codegurureviewer.DescribeRepositoryAssociationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codegurureviewer_repository_association.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, codegurureviewer.EndpointsID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, codegurureviewer.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfcodegurureviewer.ResourceRepositoryAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruReviewerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codegurureviewer_repository_association" {
				continue
			}

			input := &codegurureviewer.DescribeRepositoryAssociationInput{
				AssociationArn: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeRepositoryAssociationWithContext(ctx, input)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, codegurureviewer.ErrCodeNotFoundException) {
					return nil
				}
				return err
			}

			return fmt.Errorf("CodeGuru Reviewer Association Connection %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryAssociationExists(ctx context.Context, name string, repositoryassociation *codegurureviewer.DescribeRepositoryAssociationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CodeGuruReviewer, create.ErrActionCheckingExistence, tfcodegurureviewer.ResNameRepositoryAssociation, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CodeGuruReviewer, create.ErrActionCheckingExistence, tfcodegurureviewer.ResNameRepositoryAssociation, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruReviewerConn(ctx)
		resp, err := conn.DescribeRepositoryAssociationWithContext(ctx, &codegurureviewer.DescribeRepositoryAssociationInput{
			AssociationArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.CodeGuruReviewer, create.ErrActionCheckingExistence, tfcodegurureviewer.ResNameRepositoryAssociation, rs.Primary.ID, err)
		}

		*repositoryassociation = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruReviewerConn(ctx)

	input := &codegurureviewer.ListRepositoryAssociationsInput{}
	_, err := conn.ListRepositoryAssociationsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRepositoryAssociationConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccRepositoryAssociation_codecommit_repository(rName), `
resource "aws_codegurureviewer_repository_association" "test" {
  repository {
    codecommit {
      name = aws_codecommit_repository.test.repository_name
    }
  }
}
`)
}

func testAccRepositoryAssociationConfig_kms_key(rName string) string {
	return acctest.ConfigCompose(testAccRepositoryAssociation_codecommit_repository(rName), testAccRepositoryAssociation_kms_key(), `
resource "aws_codegurureviewer_repository_association" "test" {
  repository {
    codecommit {
      name = aws_codecommit_repository.test.repository_name
    }
  }

  kms_key_details {
    encryption_option = "CUSTOMER_MANAGED_CMK"
    kms_key_id        = aws_kms_key.test.key_id
  }
}
`)
}

func testAccRepositoryAssociationConfig_tags_1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccRepositoryAssociation_codecommit_repository(rName), fmt.Sprintf(`
resource "aws_codegurureviewer_repository_association" "test" {
  repository {
    codecommit {
      name = aws_codecommit_repository.test.repository_name
    }
  }
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccRepositoryAssociationConfig_tags_2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccRepositoryAssociation_codecommit_repository(rName), fmt.Sprintf(`
resource "aws_codegurureviewer_repository_association" "test" {
  repository {
    codecommit {
      name = aws_codecommit_repository.test.repository_name
    }
  }
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccRepositoryAssociationConfig_s3_repository(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_codegurureviewer_repository_association" "test" {
  repository {
    s3_bucket {
      bucket_name = aws_s3_bucket.test.id
      name        = "test"
    }
  }
}
`, rName)
}

func testAccRepositoryAssociation_codecommit_repository(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
  description     = "This is a test description"
  lifecycle {
    ignore_changes = [
      tags["codeguru-reviewer"]
    ]
  }
}
`, rName)
}

func testAccRepositoryAssociation_kms_key() string {
	return `
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}
`
}
