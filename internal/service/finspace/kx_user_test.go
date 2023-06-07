package finspace_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tffinspace "github.com/hashicorp/terraform-provider-aws/internal/service/finspace"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFinSpaceKxUser_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_basic(rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, "name", userName),
				),
			},
		},
	})
}

func TestAccFinSpaceKxUser_disappears(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_basic(rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(resourceName, &kxuser),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tffinspace.ResourceKxUser(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFinSpaceKxUser_updateRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_basic(rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(resourceName, &kxuser),
				),
			},
			{
				Config: testAccKxUserConfig_updateRole(rName, "updated"+rName, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(resourceName, &kxuser),
				),
			},
		},
	})
}

func TestAccFinSpaceKxUser_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var kxuser finspace.GetKxUserOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userName := sdkacctest.RandString(sdkacctest.RandIntRange(1, 50))
	resourceName := "aws_finspace_kx_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, finspace.ServiceID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, finspace.ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKxUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKxUserConfig_tags1(rName, userName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccKxUserConfig_tags2(rName, userName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccKxUserConfig_tags1(rName, userName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKxUserExists(resourceName, &kxuser),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckKxUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient()
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_finspace_kx_user" {
			continue
		}

		input := &finspace.GetKxUserInput{
			UserName:      aws.String(rs.Primary.Attributes["name"]),
			EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
		}
		_, err := conn.GetKxUser(ctx, input)
		if err != nil {
			var nfe *types.ResourceNotFoundException
			if errors.As(err, &nfe) {
				return nil
			}
			return err
		}

		return create.Error(names.FinSpace, create.ErrActionCheckingDestroyed, tffinspace.ResNameKxUser, rs.Primary.ID, errors.New("not destroyed"))
	}

	return nil
}

func testAccCheckKxUserExists(name string, kxuser *finspace.GetKxUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxUser, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxUser, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).FinSpaceClient()
		ctx := context.Background()
		resp, err := conn.GetKxUser(ctx, &finspace.GetKxUserInput{
			UserName:      aws.String(rs.Primary.Attributes["name"]),
			EnvironmentId: aws.String(rs.Primary.Attributes["environment_id"]),
		})

		if err != nil {
			return create.Error(names.FinSpace, create.ErrActionCheckingExistence, tffinspace.ResNameKxUser, rs.Primary.ID, err)
		}

		*kxuser = *resp

		return nil
	}
}

func testAccKxUserConfig_basic(rName, userName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_user" "test" {
  name           = %[2]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.test.arn
}
`, rName, userName)
}

func testAccKxUserConfig_updateRole(rName, rName2, userName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role" "updated" {
  name = %[2]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_user" "test" {
  name           = %[3]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.updated.arn
}
`, rName, rName2, userName)
}

func testAccKxUserConfig_tags1(rName, userName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_user" "test" {
  name           = %[2]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.test.arn
  tags = {
    %[3]q = %[4]q
  }
}

`, rName, userName, tagKey1, tagValue1)
}

func testAccKxUserConfig_tags2(rName, userName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_finspace_kx_environment" "test" {
  name       = %[1]q
  kms_key_id = aws_kms_key.test.arn
}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_finspace_kx_user" "test" {
  name           = %[2]q
  environment_id = aws_finspace_kx_environment.test.id
  iam_role       = aws_iam_role.test.arn
  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, userName, tagKey1, tagValue1, tagKey2, tagValue2)
}
