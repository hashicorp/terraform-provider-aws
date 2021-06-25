package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

/*
arn:aws:iam::123456789012:root
arn:aws:iam::123456789012:user/JohnDoe
arn:aws:iam::123456789012:user/division_abc/subdivision_xyz/JaneDoe
arn:aws:iam::123456789012:group/Developers
arn:aws:iam::123456789012:group/division_abc/subdivision_xyz/product_A/Developers
arn:aws:iam::123456789012:role/S3Access
arn:aws:iam::123456789012:role/application_abc/component_xyz/RDSAccess
arn:aws:iam::123456789012:role/aws-service-role/access-analyzer.amazonaws.com/AWSServiceRoleForAccessAnalyzer
arn:aws:iam::123456789012:role/service-role/QuickSightAction
arn:aws:iam::123456789012:policy/UsersManageOwnCredentials
arn:aws:iam::123456789012:policy/division_abc/subdivision_xyz/UsersManageOwnCredentials
arn:aws:iam::123456789012:instance-profile/Webserver
arn:aws:sts::123456789012:federated-user/JohnDoe
arn:aws:sts::123456789012:assumed-role/Accounting-Role/JaneDoe
arn:aws:iam::123456789012:mfa/JaneDoeMFA
arn:aws:iam::123456789012:u2f/user/JohnDoe/default (U2F security key)
arn:aws:iam::123456789012:server-certificate/ProdServerCert
arn:aws:iam::123456789012:server-certificate/division_abc/subdivision_xyz/ProdServerCert
arn:aws:iam::123456789012:saml-provider/ADFSProvider
arn:aws:iam::123456789012:oidc-provider/GoogleProvider
*/

func TestAssumedRoleRoleSessionName(t *testing.T) {
	testCases := []struct {
		Name                string
		ARN                 string
		ExpectedRoleName    string
		ExpectedSessionName string
		ExpectedError       bool
	}{
		{
			Name:                "not an ARN",
			ARN:                 "abcd",
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
			ExpectedError:       true,
		},
		{
			Name:                "regular role ARN",
			ARN:                 "arn:aws:iam::111122223333:role/role_name", //lintignore:AWSAT005
			ExpectedRoleName:    "role_name",
			ExpectedSessionName: "",
			ExpectedError:       false,
		},
		{
			Name:                "assumed role ARN",
			ARN:                 "arn:aws:sts::444433332222:assumed-role/something_something-admin/sessionIDNotPartOfRoleARN", //lintignore:AWSAT005
			ExpectedRoleName:    "something_something-admin",
			ExpectedSessionName: "sessionIDNotPartOfRoleARN",
			ExpectedError:       false,
		},
		{
			Name:                "'assumed-role' part of ARN resource",
			ARN:                 "arn:aws:iam::444433332222:user/assumed-role-but-not-really", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
			ExpectedError:       true,
		},
		{
			Name:                "user ARN",
			ARN:                 "arn:aws:iam::123456789012:user/Bob", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
			ExpectedError:       true,
		},
		{
			Name:                "assumed role from AWS example",
			ARN:                 "arn:aws:sts::123456789012:assumed-role/example-role/AWSCLI-Session", //lintignore:AWSAT005
			ExpectedRoleName:    "example-role",
			ExpectedSessionName: "AWSCLI-Session",
			ExpectedError:       false,
		},
		{
			Name:                "multiple slashes in resource",                                         // not sure this is even valid
			ARN:                 "arn:aws:sts::123456789012:assumed-role/path/role-name/AWSCLI-Session", //lintignore:AWSAT005
			ExpectedRoleName:    "role-name",
			ExpectedSessionName: "AWSCLI-Session",
			ExpectedError:       false,
		},
		{
			Name:                "not an sts ARN",
			ARN:                 "arn:aws:iam::123456789012:assumed-role/example-role/AWSCLI-Session", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
			ExpectedError:       true,
		},
		{
			Name:                "role with path",
			ARN:                 "arn:aws:iam::123456789012:role/this/is/the/path/role-name", //lintignore:AWSAT005
			ExpectedRoleName:    "role-name",
			ExpectedSessionName: "",
			ExpectedError:       false,
		},
		{
			Name:                "wrong service",
			ARN:                 "arn:aws:ec2::123456789012:role/role-name", //lintignore:AWSAT005
			ExpectedRoleName:    "",
			ExpectedSessionName: "",
			ExpectedError:       true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			role, session, err := roleSessionNameFromARN(testCase.ARN)

			if err != nil && !testCase.ExpectedError {
				t.Errorf("for %s: got error (%s), expected none", testCase.ARN, err)
			}

			if err == nil && testCase.ExpectedError {
				t.Errorf("for %s: got no error, expected an error", testCase.ARN)
			}

			if testCase.ExpectedRoleName != role || testCase.ExpectedSessionName != session {
				t.Errorf("for %s: got role %s, session %s; expected role %s, session %s", testCase.ARN, role, session, testCase.ExpectedRoleName, testCase.ExpectedSessionName)
			}
		})
	}
}

func TestAccAWSDataSourceIAMAssumedRoleSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_assumed_role_source.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMAssumedRoleSourceConfig(rName, "/", "session-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "source_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_path", resourceName, "path"),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", "session-id"),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMAssumedRoleSource_withPath(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_assumed_role_source.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMAssumedRoleSourceConfig(rName, "/this/is/a/long/path/", "session-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "source_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_path", resourceName, "path"),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", "session-id"),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMAssumedRoleSource_notAssumedRole(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_assumed_role_source.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMAssumedRoleSourceNotAssumedConfig(rName, "/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "source_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_path", resourceName, "path"),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", ""),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMAssumedRoleSource_notAssumedRoleWithPath(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_assumed_role_source.test"
	resourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMAssumedRoleSourceNotAssumedConfig(rName, "/this/is/a/long/path/"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "source_arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_path", resourceName, "path"),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", ""),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMAssumedRoleSource_notAssumedRoleUser(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_iam_assumed_role_source.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, iam.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMAssumedRoleSourceUserConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceAttrGlobalARN(dataSourceName, "arn", "iam", fmt.Sprintf("user/division/extra-division/not-assumed-role/%[1]s", rName)),
					resource.TestCheckResourceAttr(dataSourceName, "role_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "role_path", ""),
					resource.TestCheckResourceAttr(dataSourceName, "session_name", ""),
				),
			},
		},
	})
}

func testAccAwsIAMAssumedRoleSourceConfig(rName, path, sessionID string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q
  path = %[2]q

  assume_role_policy = jsonencode({
    "Version" = "2012-10-17"

    "Statement" = [{
      "Action" = "sts:AssumeRole"
      "Principal" = {
        "Service" = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      "Effect" = "Allow"
    }]
  })
}

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

data "aws_iam_assumed_role_source" "test" {
  arn = "arn:${data.aws_partition.current.partition}:sts::${data.aws_caller_identity.current.account_id}:assumed-role/${aws_iam_role.test.name}/%[3]s"
}
`, rName, path, sessionID)
}

func testAccAwsIAMAssumedRoleSourceNotAssumedConfig(rName, path string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = %[2]q

  assume_role_policy = jsonencode({
    "Version" = "2012-10-17"

    "Statement" = [{
      "Action" = "sts:AssumeRole"
      "Principal" = {
        "Service" = "ec2.${data.aws_partition.current.dns_suffix}"
      }
      "Effect" = "Allow"
    }]
  })
}

data "aws_iam_assumed_role_source" "test" {
  arn = aws_iam_role.test.arn
}
`, rName, path)
}

func testAccAwsIAMAssumedRoleSourceUserConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

data "aws_iam_assumed_role_source" "test" {
  arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:user/division/extra-division/not-assumed-role/%[1]s"
}
`, rName)
}
