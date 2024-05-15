// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codegurureviewer_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcodegurureviewer "github.com/hashicorp/terraform-provider-aws/internal/service/codegurureviewer"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Repository types of "BitBucket and GitHubEnterpriseServer cannot be tested, as they require their CodeStar Connection to be in "AVAILABLE" status vs "PENDING", requiring console interaction
// However, this has been manually tested successfully
func TestAccCodeGuruReviewerRepositoryAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation types.RepositoryAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codegurureviewer_repository_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeGuruReviewerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruReviewerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codeguru-reviewer", regexache.MustCompile(`association:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrID, "codeguru-reviewer", regexache.MustCompile(`association:+.`)),
					resource.TestCheckResourceAttr(resourceName, "repository.0.bitbucket.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "repository.0.github_enterprise_server.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.0.encryption_option", "AWS_OWNED_CMK"),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_KMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation types.RepositoryAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codegurureviewer_repository_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeGuruReviewerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruReviewerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_kms_key(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codeguru-reviewer", regexache.MustCompile(`association:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrID, "codeguru-reviewer", regexache.MustCompile(`association:+.`)),
					resource.TestCheckResourceAttr(resourceName, "repository.0.bitbucket.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "repository.0.github_enterprise_server.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.0.encryption_option", "CUSTOMER_MANAGED_CMK"),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_S3Repository(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation types.RepositoryAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codegurureviewer_repository_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeGuruReviewerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruReviewerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_s3_repository(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "codeguru-reviewer", regexache.MustCompile(`association:+.`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrID, "codeguru-reviewer", regexache.MustCompile(`association:+.`)),
					resource.TestCheckResourceAttr(resourceName, "repository.0.bitbucket.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.codecommit.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.github_enterprise_server.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.0.bucket_name", "codeguru-reviewer-"+rName),
					resource.TestCheckResourceAttr(resourceName, "repository.0.s3_bucket.0.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "kms_key_details.0.encryption_option", "AWS_OWNED_CMK"),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation types.RepositoryAssociation
	resourceName := "aws_codegurureviewer_repository_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeGuruReviewerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruReviewerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRepositoryAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryAssociationConfig_tags_1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccRepositoryAssociationConfig_tags_2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRepositoryAssociationConfig_tags_1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRepositoryAssociationExists(ctx, resourceName, &repositoryassociation),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccCodeGuruReviewerRepositoryAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var repositoryassociation types.RepositoryAssociation
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codegurureviewer_repository_association.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CodeGuruReviewerEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CodeGuruReviewerServiceID),
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
		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruReviewerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_codegurureviewer_repository_association" {
				continue
			}

			_, err := tfcodegurureviewer.FindRepositoryAssociationByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CodeGuru Reviewer Repository Association (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRepositoryAssociationExists(ctx context.Context, n string, v *types.RepositoryAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruReviewerClient(ctx)

		output, err := tfcodegurureviewer.FindRepositoryAssociationByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CodeGuruReviewerClient(ctx)

	input := &codegurureviewer.ListRepositoryAssociationsInput{}
	_, err := conn.ListRepositoryAssociations(ctx, input)

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
  bucket = "codeguru-reviewer-%[1]s"
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
