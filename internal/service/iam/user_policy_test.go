// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMUserPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policy1 := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Action":"iam:*","Effect":"Allow","Resource":"*"}}`
	resourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserPolicyConfig_basic(rName, strconv.Quote("NonJSONString")),
				ExpectError: regexache.MustCompile("invalid JSON"),
			},
			{
				Config: testAccUserPolicyConfig_basic(rName, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName, &userPolicy),
					testAccCheckUserPolicyExpectedPolicies(ctx, t, userResourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, policy1),
					resource.TestCheckResourceAttr(resourceName, "user", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPolicyConfig_basic(rName, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName, &userPolicy),
					testAccCheckUserPolicyExpectedPolicies(ctx, t, userResourceName, 1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPolicy, policy2),
				),
			},
		},
	})
}

func TestAccIAMUserPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policy := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	resourceName := "aws_iam_user_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_basic(rName, strconv.Quote(policy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName, &userPolicy),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceUserPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUserPolicy_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policy := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	resourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_nameGenerated(rName, strconv.Quote(policy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName, &userPolicy),
					testAccCheckUserPolicyExpectedPolicies(ctx, t, userResourceName, 1),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, sdkid.UniqueIdPrefix),
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

func TestAccIAMUserPolicy_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policy := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	resourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_namePrefix(rName, "tf-acc-test-prefix-", strconv.Quote(policy)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName, &userPolicy),
					testAccCheckUserPolicyExpectedPolicies(ctx, t, userResourceName, 1),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
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

func TestAccIAMUserPolicy_multiplePolicies(t *testing.T) {
	ctx := acctest.Context(t)
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policy1 := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Action":"iam:*","Effect":"Allow","Resource":"*"}}`
	resourceName1 := "aws_iam_user_policy.test1"
	resourceName2 := "aws_iam_user_policy.test2"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_multiplePolicies(rName, strconv.Quote(policy1), strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName1, &userPolicy),
					testAccCheckUserPolicyExists(ctx, t, resourceName2, &userPolicy),
					testAccCheckUserPolicyExpectedPolicies(ctx, t, userResourceName, 2),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPolicy, policy1),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPolicy, policy2),
				),
			},
			{
				Config: testAccUserPolicyConfig_multiplePolicies(rName, strconv.Quote(policy2), strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName1, &userPolicy),
					testAccCheckUserPolicyExists(ctx, t, resourceName2, &userPolicy),
					testAccCheckUserPolicyExpectedPolicies(ctx, t, userResourceName, 2),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPolicy, policy2),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPolicy, policy2),
				),
			},
		},
	})
}

func TestAccIAMUserPolicy_policyOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, t, resourceName, &userPolicy),
					testAccCheckUserPolicyExpectedPolicies(ctx, t, userResourceName, 1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccUserPolicyConfig_newOrder(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccCheckUserPolicyExists(ctx context.Context, t *testing.T, n string, v *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		output, err := tfiam.FindUserPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["user"], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}

		*v = output

		return nil
	}
}

func testAccCheckUserPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_user_policy" {
				continue
			}

			_, err := tfiam.FindUserPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["user"], rs.Primary.Attributes[names.AttrName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IAM User Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckUserPolicyExpectedPolicies(ctx context.Context, t *testing.T, n string, want int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		var got int

		input := &iam.ListUserPoliciesInput{
			UserName: aws.String(rs.Primary.ID),
		}

		pages := iam.NewListUserPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				return err
			}

			got += len(page.PolicyNames)
		}

		if got != want {
			return fmt.Errorf("Got %d IAM User Policies for %s, want %v", got, rs.Primary.ID, want)
		}

		return nil
	}
}

func testAccUserPolicyUserConfig_base(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
  path = %[2]q
}
`, rName, path)
}

func testAccUserPolicyConfig_basic(rName, policy string) string {
	return acctest.ConfigCompose(testAccUserPolicyUserConfig_base(rName, "/"), fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  name   = %[1]q
  user   = aws_iam_user.test.name
  policy = %[2]s
}
`, rName, policy))
}

func testAccUserPolicyConfig_nameGenerated(rName, policy string) string {
	return acctest.ConfigCompose(testAccUserPolicyUserConfig_base(rName, "/"), fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  user   = aws_iam_user.test.name
  policy = %[1]s
}
`, policy))
}

func testAccUserPolicyConfig_namePrefix(rName, namePrefix, policy string) string {
	return acctest.ConfigCompose(testAccUserPolicyUserConfig_base(rName, "/"), fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  name_prefix = %[1]q
  user        = aws_iam_user.test.name
  policy      = %[2]s
}
`, namePrefix, policy))
}

func testAccUserPolicyConfig_multiplePolicies(rName, policy1, policy2 string) string {
	return acctest.ConfigCompose(testAccUserPolicyUserConfig_base(rName, "/"), fmt.Sprintf(`
resource "aws_iam_user_policy" "test1" {
  name   = %[1]q
  user   = aws_iam_user.test.name
  policy = %[2]s
}

resource "aws_iam_user_policy" "test2" {
  name   = "%[1]s-2"
  user   = aws_iam_user.test.name
  policy = %[3]s
}
`, rName, policy1, policy2))
}

func testAccUserPolicyConfig_order(rName string) string {
	return acctest.ConfigCompose(testAccUserPolicyUserConfig_base(rName, "/"), fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  name = %[1]q
  user = aws_iam_user.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": [
      "ec2:DescribeScheduledInstances",
      "ec2:DescribeScheduledInstanceAvailability",
      "ec2:DescribeFastSnapshotRestores",
      "ec2:DescribeElasticGpus"
    ],
    "Resource": "*"
  }
}
EOF
}
`, rName))
}

func testAccUserPolicyConfig_newOrder(rName string) string {
	return acctest.ConfigCompose(testAccUserPolicyUserConfig_base(rName, "/"), fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  name = %[1]q
  user = aws_iam_user.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": [
      "ec2:DescribeElasticGpus",
      "ec2:DescribeFastSnapshotRestores",
      "ec2:DescribeScheduledInstances",
      "ec2:DescribeScheduledInstanceAvailability"
    ],
    "Resource": "*"
  }
}
EOF
}
`, rName))
}
