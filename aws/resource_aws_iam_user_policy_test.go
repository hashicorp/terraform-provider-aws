package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func TestAccAWSIAMUserPolicy_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	policy1 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"*","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"iam:*","Resource":"*"}}`
	policyName := fmt.Sprintf("foo_policy_%d", rInt)
	policyResourceName := "aws_iam_user_policy.foo"
	userResourceName := "aws_iam_user.user"
	userName := fmt.Sprintf("test_user_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccIAMUserPolicyConfig_name(rInt, strconv.Quote("NonJSONString")),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
			{
				Config: testAccIAMUserPolicyConfig_name(rInt, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
					resource.TestMatchResourceAttr(policyResourceName, "id", regexp.MustCompile(fmt.Sprintf("^%s:%s$", userName, policyName))),
					resource.TestCheckResourceAttr(policyResourceName, "name", policyName),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy1),
					resource.TestCheckResourceAttr(policyResourceName, "user", userName),
				),
			},
			{
				ResourceName:      policyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMUserPolicyConfig_name(rInt, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy2),
				),
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_disappears(t *testing.T) {
	var out iam.GetUserPolicyOutput
	suffix := sdkacctest.RandStringFromCharSet(10, sdkacctest.CharSetAlpha)
	resourceName := fmt.Sprintf("aws_iam_user_policy.foo_%s", suffix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIamUserPolicyConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicyExists(resourceName, &out),
					testAccCheckIAMUserPolicyDisappears(&out),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_namePrefix(t *testing.T) {
	rInt := sdkacctest.RandInt()
	policy1 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"*","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"iam:*","Resource":"*"}}`
	policyNamePrefix := "foo_policy_"
	policyResourceName := "aws_iam_user_policy.foo"
	userResourceName := "aws_iam_user.user"
	userName := fmt.Sprintf("test_user_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyConfig_namePrefix(rInt, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
					resource.TestMatchResourceAttr(policyResourceName, "id", regexp.MustCompile(fmt.Sprintf("^%s:%s.+$", userName, policyNamePrefix))),
					resource.TestCheckResourceAttr(policyResourceName, "name_prefix", policyNamePrefix),
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
				Config: testAccIAMUserPolicyConfig_namePrefix(rInt, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy2),
				),
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_generatedName(t *testing.T) {
	rInt := sdkacctest.RandInt()
	policy1 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"*","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"iam:*","Resource":"*"}}`
	policyResourceName := "aws_iam_user_policy.foo"
	userResourceName := "aws_iam_user.user"
	userName := fmt.Sprintf("test_user_%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyConfig_generatedName(rInt, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
					resource.TestMatchResourceAttr(policyResourceName, "id", regexp.MustCompile(fmt.Sprintf("^%s:.+$", userName))),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy1),
				),
			},
			{
				ResourceName:      policyResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMUserPolicyConfig_generatedName(rInt, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName, "policy", policy2),
				),
			},
		},
	})
}

func TestAccAWSIAMUserPolicy_multiplePolicies(t *testing.T) {
	rInt := sdkacctest.RandInt()
	policy1 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"*","Resource":"*"}}`
	policy2 := `{"Version":"2012-10-17","Statement":{"Effect":"Allow","Action":"iam:*","Resource":"*"}}`
	policyName1 := fmt.Sprintf("foo_policy_%d", rInt)
	policyName2 := fmt.Sprintf("bar_policy_%d", rInt)
	policyResourceName1 := "aws_iam_user_policy.foo"
	policyResourceName2 := "aws_iam_user_policy.bar"
	userResourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMUserPolicyConfig_name(rInt, strconv.Quote(policy1)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName1),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
					resource.TestCheckResourceAttr(policyResourceName1, "name", policyName1),
					resource.TestCheckResourceAttr(policyResourceName1, "policy", policy1),
				),
			},
			{
				ResourceName:      policyResourceName1,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMUserPolicyConfig_multiplePolicies(rInt, strconv.Quote(policy1), strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName1),
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName2),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 2),
					resource.TestCheckResourceAttr(policyResourceName1, "policy", policy1),
					resource.TestCheckResourceAttr(policyResourceName2, "name", policyName2),
					resource.TestCheckResourceAttr(policyResourceName2, "policy", policy2),
				),
			},
			{
				Config: testAccIAMUserPolicyConfig_multiplePolicies(rInt, strconv.Quote(policy2), strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName1),
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName2),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 2),
					resource.TestCheckResourceAttr(policyResourceName1, "policy", policy2),
				),
			},
			{
				Config: testAccIAMUserPolicyConfig_name(rInt, strconv.Quote(policy2)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMUserPolicy(userResourceName, policyResourceName1),
					testAccCheckIAMUserPolicyExpectedPolicies(userResourceName, 1),
				),
			},
		},
	})
}

func testAccCheckIAMUserPolicyExists(resource string, res *iam.GetUserPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Policy name is set")
		}

		user, name, err := resourceAwsIamUserPolicyParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetUserPolicy(&iam.GetUserPolicyInput{
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

func testAccCheckIAMUserPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user_policy" {
			continue
		}

		user, name, err := resourceAwsIamUserPolicyParseId(rs.Primary.ID)
		if err != nil {
			return err
		}

		request := &iam.GetUserPolicyInput{
			PolicyName: aws.String(name),
			UserName:   aws.String(user),
		}

		getResp, err := conn.GetUserPolicy(request)
		if err != nil {
			if iamerr, ok := err.(awserr.Error); ok && iamerr.Code() == "NoSuchEntity" {
				// none found, that's good
				return nil
			}
			return fmt.Errorf("Error reading IAM policy %s from user %s: %s", name, user, err)
		}

		if getResp != nil {
			return fmt.Errorf("Found IAM user policy, expected none: %s", getResp)
		}
	}

	return nil
}

func testAccCheckIAMUserPolicyDisappears(out *iam.GetUserPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		params := &iam.DeleteUserPolicyInput{
			PolicyName: out.PolicyName,
			UserName:   out.UserName,
		}

		_, err := conn.DeleteUserPolicy(params)
		return err
	}
}

func testAccCheckIAMUserPolicy(
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		username, name, err := resourceAwsIamUserPolicyParseId(policy.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetUserPolicy(&iam.GetUserPolicyInput{
			UserName:   aws.String(username),
			PolicyName: aws.String(name),
		})

		return err
	}
}

func testAccCheckIAMUserPolicyExpectedPolicies(iamUserResource string, expected int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[iamUserResource]
		if !ok {
			return fmt.Errorf("Not Found: %s", iamUserResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		userPolicies, err := conn.ListUserPolicies(&iam.ListUserPoliciesInput{
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

func testAccAwsIamUserPolicyConfig(suffix string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user_%[1]s" {
  name = "tf_test_user_test_%[1]s"
  path = "/"
}

resource "aws_iam_user_policy" "foo_%[1]s" {
  name = "tf_test_policy_test_%[1]s"
  user = "${aws_iam_user.user_%[1]s.name}"

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
`, suffix)
}

func testAccIAMUserPolicyConfig_name(rInt int, policy string) string {
	return fmt.Sprintf(`
%s

resource "aws_iam_user_policy" "foo" {
  name   = "foo_policy_%d"
  user   = aws_iam_user.user.name
  policy = %s
}
`, testAccAWSUserConfig(fmt.Sprintf("test_user_%d", rInt), "/"), rInt, policy)
}

func testAccIAMUserPolicyConfig_namePrefix(rInt int, policy string) string {
	return fmt.Sprintf(`
%s

resource "aws_iam_user_policy" "foo" {
  name_prefix = "foo_policy_"
  user        = aws_iam_user.user.name
  policy      = %s
}
`, testAccAWSUserConfig(fmt.Sprintf("test_user_%d", rInt), "/"), policy)
}

func testAccIAMUserPolicyConfig_generatedName(rInt int, policy string) string {
	return fmt.Sprintf(`
%s

resource "aws_iam_user_policy" "foo" {
  user   = aws_iam_user.user.name
  policy = %s
}
`, testAccAWSUserConfig(fmt.Sprintf("test_user_%d", rInt), "/"), policy)
}

func testAccIAMUserPolicyConfig_multiplePolicies(rInt int, policy1, policy2 string) string {
	return fmt.Sprintf(`
%[1]s

resource "aws_iam_user_policy" "foo" {
  name   = "foo_policy_%[2]d"
  user   = aws_iam_user.user.name
  policy = %[3]s
}

resource "aws_iam_user_policy" "bar" {
  name   = "bar_policy_%[2]d"
  user   = aws_iam_user.user.name
  policy = %[4]s
}
`, testAccAWSUserConfig(fmt.Sprintf("test_user_%d", rInt), "/"), rInt, policy1, policy2)
}
