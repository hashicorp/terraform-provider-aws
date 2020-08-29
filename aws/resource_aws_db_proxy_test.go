package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_db_proxy", &resource.Sweeper{
		Name: "aws_db_proxy",
		F:    testSweepRdsDbProxies,
	})
}

func testSweepRdsDbProxies(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).rdsconn

	err = conn.DescribeDBProxiesPages(&rds.DescribeDBProxiesInput{}, func(out *rds.DescribeDBProxiesOutput, lastPage bool) bool {
		for _, dbpg := range out.DBProxies {
			if dbpg == nil {
				continue
			}

			input := &rds.DeleteDBProxyInput{
				DBProxyName: dbpg.DBProxyName,
			}
			name := aws.StringValue(dbpg.DBProxyName)

			log.Printf("[INFO] Deleting DB Proxy: %s", name)

			_, err := conn.DeleteDBProxy(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DB Proxy %s: %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Proxy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving DB Proxies: %s", err)
	}

	return nil
}

func TestAccAWSDBProxy_basic(t *testing.T) {
	var v rds.DBProxy
	resourceName := "aws_db_proxy.test"
	name := fmt.Sprintf("tf-acc-db-proxy-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "engine_family", "MYSQL"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db-proxy:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auth.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						"auth_scheme": "SECRETS",
						"description": "test",
						"iam_auth":    "DISABLED",
					}),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "1800"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "require_tls", "false"),
					resource.TestCheckResourceAttr(resourceName, "vpc_subnet_ids.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.0", "id"),
					tfawsresource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`))),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDBProxy_disappears(t *testing.T) {
	var v rds.DBProxy
	resourceName := "aws_db_proxy.test"
	name := fmt.Sprintf("tf-acc-db-proxy-%d", acctest.RandInt())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbProxy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDBProxyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBProxies(
			&rds.DescribeDBProxiesInput{
				DBProxyName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBProxies) != 0 &&
				*resp.DBProxies[0].DBProxyName == rs.Primary.ID {
				return fmt.Errorf("DB Proxy still exists")
			}
		}

		if !isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDBProxyExists(n string, v *rds.DBProxy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Proxy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBProxiesInput{
			DBProxyName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBProxies(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBProxies) != 1 ||
			*resp.DBProxies[0].DBProxyName != rs.Primary.ID {
			return fmt.Errorf("DB Proxy not found")
		}

		*v = *resp.DBProxies[0]

		return nil
	}
}

func testAccAWSDBProxyConfig(n string) string {
	return fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%s"
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = []
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }

  tags = {
    Name = "%[1]s"
  }
}

# Secrets Manager setup

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}

# IAM setup

resource "aws_iam_role" "test" {
  name               = "%[1]s"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

data "aws_iam_policy_document" "assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["rds.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "secretsmanager:GetRandomPassword",
      "secretsmanager:CreateSecret",
      "secretsmanager:ListSecrets",
    ]
    resources = ["*"]
  }

  statement {
    actions   = ["secretsmanager:*"]
    resources = [aws_secretsmanager_secret.test.arn]
  }
}

# VPC setup

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "%[1]s"
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-${count.index}"
  }
}
`, n)
}
