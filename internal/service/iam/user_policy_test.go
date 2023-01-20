package iam_test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMUserPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policy1 := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Action":"iam:*","Effect":"Allow","Resource":"*"}}`
	policyResourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccUserPolicyConfig_name(rName, strconv.Quote("NonJSONString")),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
			{
				Config: testAccUserPolicyConfig_name(rName, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
					resource.TestMatchResourceAttr(policyResourceName, "id", regexp.MustCompile(fmt.Sprintf("^%[1]s:%[1]s$", rName))),
					resource.TestCheckResourceAttr(policyResourceName, "name", rName),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy1),
					resource.TestCheckResourceAttr(policyResourceName, "user", rName),
				),
			},
			{
				ResourceName:      policyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPolicyConfig_name(rName, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy2),
				),
			},
		},
	})
}

func TestAccIAMUserPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var out iam.GetUserPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyExists(ctx, resourceName, &out),
					testAccCheckUserPolicyDisappears(ctx, &out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUserPolicy_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policy1 := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Action":"iam:*","Effect":"Allow","Resource":"*"}}`
	policyResourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_namePrefix(rName, acctest.ResourcePrefix, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
					resource.TestMatchResourceAttr(policyResourceName, "id", regexp.MustCompile(fmt.Sprintf("^%s:%s.+$", rName, acctest.ResourcePrefix))),
					resource.TestCheckResourceAttr(policyResourceName, "name_prefix", acctest.ResourcePrefix),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy1),
				),
			},
			{
				ResourceName:            policyResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccUserPolicyConfig_namePrefix(rName, acctest.ResourcePrefix, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy2),
				),
			},
		},
	})
}

func TestAccIAMUserPolicy_generatedName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policy1 := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Action":"iam:*","Effect":"Allow","Resource":"*"}}`
	policyResourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_generatedName(rName, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
					resource.TestMatchResourceAttr(policyResourceName, "id", regexp.MustCompile(fmt.Sprintf("^%s:.+$", rName))),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy1),
				),
			},
			{
				ResourceName:      policyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPolicyConfig_generatedName(rName, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy2),
				),
			},
		},
	})
}

func TestAccIAMUserPolicy_multiplePolicies(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policy1 := `{"Version":"2012-10-17","Statement":{"Action":"*","Effect":"Allow","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Action":"iam:*","Effect":"Allow","Resource":"*"}}`
	policyResourceName1 := "aws_iam_user_policy.test"
	policyResourceName2 := "aws_iam_user_policy.test2"
	userResourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_name(rName, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName1),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName1, "name", rName),
					resource.TestCheckResourceAttr(policyResourceName1, "policy", policy1),
				),
			},
			{
				ResourceName:      policyResourceName1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccUserPolicyConfig_multiplePolicies(rName, strconv.Quote(policy1), strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName1),
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName2),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 2),
					resource.TestCheckResourceAttr(policyResourceName1, "policy", policy1),
					resource.TestCheckResourceAttr(policyResourceName2, "name", fmt.Sprintf("%s-2", rName)),
					resource.TestCheckResourceAttr(policyResourceName2, "policy", policy2),
				),
			},
			{
				Config: testAccUserPolicyConfig_multiplePolicies(rName, strconv.Quote(policy2), strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName1),
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName2),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 2),
					resource.TestCheckResourceAttr(policyResourceName1, "policy", policy2),
				),
			},
			{
				Config: testAccUserPolicyConfig_name(rName, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName1),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
				),
			},
		},
	})
}

func TestAccIAMUserPolicy_policyOrder(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyResourceName := "aws_iam_user_policy.test"
	userResourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, iam.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyConfig_order(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicy(ctx, userResourceName, policyResourceName),
					testAccCheckUserPolicyExpectedPolicies(ctx, userResourceName, 1),
				),
			},
			{
				Config:   testAccUserPolicyConfig_newOrder(rName),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckUserPolicyExists(ctx context.Context, resource string, res *iam.GetUserPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Policy name is set")
		}

		user, name, err := tfiam.UserPolicyParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()

		resp, err := conn.GetUserPolicyWithContext(ctx, &iam.GetUserPolicyInput{
			PolicyName: aws.String(name),
			UserName:   aws.String(user),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckUserPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_user_policy" {
				continue
			}

			user, name, err := tfiam.UserPolicyParseID(rs.Primary.ID)
			if err != nil {
				return err
			}

			request := &iam.GetUserPolicyInput{
				PolicyName: aws.String(name),
				UserName:   aws.String(user),
			}

			getResp, err := conn.GetUserPolicyWithContext(ctx, request)

			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				continue
			}

			if err != nil {
				return fmt.Errorf("Error reading IAM policy %s from user %s: %s", name, user, err)
			}

			if getResp != nil {
				return fmt.Errorf("Found IAM user policy, expected none: %s", getResp)
			}
		}

		return nil
	}
}

func testAccCheckUserPolicyDisappears(ctx context.Context, out *iam.GetUserPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()

		params := &iam.DeleteUserPolicyInput{
			PolicyName: out.PolicyName,
			UserName:   out.UserName,
		}

		_, err := conn.DeleteUserPolicyWithContext(ctx, params)
		return err
	}
}

func testAccCheckUserPolicy(ctx context.Context,
	iamUserResource string,
	iamUserPolicyResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamUserResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamUserResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, ok := s.RootModule().Resources[iamUserPolicyResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamUserPolicyResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()
		username, name, err := tfiam.UserPolicyParseID(policy.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetUserPolicyWithContext(ctx, &iam.GetUserPolicyInput{
			UserName:   aws.String(username),
			PolicyName: aws.String(name),
		})

		return err
	}
}

func testAccCheckUserPolicyExpectedPolicies(ctx context.Context, iamUserResource string, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamUserResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamUserResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn()
		userPolicies, err := conn.ListUserPoliciesWithContext(ctx, &iam.ListUserPoliciesInput{
			UserName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(userPolicies.PolicyNames) != expected {
			return fmt.Errorf("Expected (%d) IAM user policies for user (%s), found: %d", expected, rs.Primary.ID, len(userPolicies.PolicyNames))
		}

		return nil
	}
}

func testAccUserPolicyUserBaseConfig(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
  path = %[2]q
}
`, rName, path)
}

func testAccUserPolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyUserBaseConfig(rName, "/"),
		fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  name = %[1]q
  user = aws_iam_user.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, rName))
}

func testAccUserPolicyConfig_name(rName, policy string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyUserBaseConfig(rName, "/"),
		fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  name   = %[1]q
  user   = aws_iam_user.test.name
  policy = %[2]s
}
`, rName, policy))
}

func testAccUserPolicyConfig_namePrefix(rName, prefix, policy string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyUserBaseConfig(rName, "/"),
		fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  name_prefix = %[1]q
  user        = aws_iam_user.test.name
  policy      = %[2]s
}
`, prefix, policy))
}

func testAccUserPolicyConfig_generatedName(rName, policy string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyUserBaseConfig(rName, "/"),
		fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
  user   = aws_iam_user.test.name
  policy = %[1]s
}
`, policy))
}

func testAccUserPolicyConfig_multiplePolicies(rName, policy1, policy2 string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyUserBaseConfig(rName, "/"),
		fmt.Sprintf(`
resource "aws_iam_user_policy" "test" {
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
	return acctest.ConfigCompose(
		testAccUserPolicyUserBaseConfig(rName, "/"),
		fmt.Sprintf(`
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
	return acctest.ConfigCompose(
		testAccUserPolicyUserBaseConfig(rName, "/"),
		fmt.Sprintf(`
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
