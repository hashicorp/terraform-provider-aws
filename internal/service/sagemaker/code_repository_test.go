// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerCodeRepository_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var repo sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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

func TestAccSageMakerCodeRepository_Git_branch(t *testing.T) {
	ctx := acctest.Context(t)
	var repo sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryConfig_gitBranch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.branch", "master"),
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

func TestAccSageMakerCodeRepository_Git_secret(t *testing.T) {
	ctx := acctest.Context(t)
	var repo sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryConfig_gitSecret(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttrPair(resourceName, "git_config.0.secret_arn", "aws_secretsmanager_secret.test", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodeRepositoryConfig_gitSecretUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttrPair(resourceName, "git_config.0.secret_arn", "aws_secretsmanager_secret.test2", names.AttrARN),
				),
			},
		},
	})
}

func TestAccSageMakerCodeRepository_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var repo sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDeviceFleetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryConfig_basicTags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodeRepositoryConfig_basicTags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccCodeRepositoryConfig_basicTags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccSageMakerCodeRepository_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var repo sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCodeRepositoryDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(ctx, resourceName, &repo),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceCodeRepository(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfsagemaker.ResourceCodeRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCodeRepositoryDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_sagemaker_code_repository" {
				continue
			}

			codeRepository, err := tfsagemaker.FindCodeRepositoryByName(ctx, conn, rs.Primary.ID)

			if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "Cannot find CodeRepository") {
				continue
			}

			if err != nil {
				return fmt.Errorf("reading SageMaker Code Repository (%s): %w", rs.Primary.ID, err)
			}

			if aws.StringValue(codeRepository.CodeRepositoryName) == rs.Primary.ID {
				return fmt.Errorf("sagemaker Code Repository %q still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

func testAccCheckCodeRepositoryExists(ctx context.Context, n string, codeRepo *sagemaker.DescribeCodeRepositoryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Code Repository ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn(ctx)
		resp, err := tfsagemaker.FindCodeRepositoryByName(ctx, conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccCodeRepositoryConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }
}
`, rName)
}

func testAccCodeRepositoryConfig_gitBranch(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    branch         = "master"
  }
}
`, rName)
}

func testAccCodeRepositoryConfig_gitSecret(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    secret_arn     = aws_secretsmanager_secret.test.arn
  }

  depends_on = [aws_secretsmanager_secret_version.test]
}
`, rName)
}

func testAccCodeRepositoryConfig_gitSecretUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test2" {
  name = "%[1]s-2"
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    secret_arn     = aws_secretsmanager_secret.test2.arn
  }

  depends_on = [aws_secretsmanager_secret_version.test2]
}
`, rName)
}

func testAccCodeRepositoryConfig_basicTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccCodeRepositoryConfig_basicTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
