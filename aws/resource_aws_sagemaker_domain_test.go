package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/sagemaker/finder"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_domain", &resource.Sweeper{
		Name: "aws_sagemaker_domain",
		F:    testSweepSagemakerDomains,
		Dependencies: []string{
			"aws_efs_mount_target",
			"aws_efs_file_system",
		},
	})
}

func testSweepSagemakerDomains(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sagemakerconn

	err = conn.ListDomainsPages(&sagemaker.ListDomainsInput{}, func(page *sagemaker.ListDomainsOutput, lastPage bool) bool {
		for _, instance := range page.Domains {
			domainArn := aws.StringValue(instance.DomainArn)
			domainID, err := decodeSagemakerDomainID(domainArn)
			if err != nil {
				log.Printf("[ERROR] Error parsing sagemaker domain arn (%s): %s", domainArn, err)
			}
			input := &sagemaker.DeleteDomainInput{
				DomainId: aws.String(domainID),
			}

			log.Printf("[INFO] Deleting SageMaker domain: %s", domainArn)
			if _, err := conn.DeleteDomain(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker domain (%s): %s", domainArn, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker domains: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerDomain_basic(t *testing.T) {
	var notebook sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestCheckResourceAttr(resourceName, "auth_mode", "IAM"),
					resource.TestCheckResourceAttr(resourceName, "app_network_access_type", "PublicInternetOnly"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "default_user_settings.0.execution_role", "aws_iam_role.test", "arn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "sagemaker", regexp.MustCompile(`domain/.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttrSet(resourceName, "url"),
					resource.TestCheckResourceAttrSet(resourceName, "home_efs_file_system_id"),
					testAccCheckAWSSagemakerDomainDeleteImplicitResources(resourceName),
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

func TestAccAWSSagemakerDomain_efsKms(t *testing.T) {
	var notebook sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigEFSKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					resource.TestCheckResourceAttrPair(resourceName, "home_efs_file_system_kms_key_id", "aws_kms_key.test", "arn"),
					testAccCheckAWSSagemakerDomainDeleteImplicitResources(resourceName),
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

func TestAccAWSSagemakerDomain_tags(t *testing.T) {
	var notebook sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
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
				Config: testAccAWSSagemakerDomainConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSagemakerDomainConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					testAccCheckAWSSagemakerDomainDeleteImplicitResources(resourceName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerDomain_securityGroup(t *testing.T) {
	var notebook sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainConfigSecurityGroup1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.security_groups.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSagemakerDomainConfigSecurityGroup2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_user_settings.0.security_groups.#", "2"),
					testAccCheckAWSSagemakerDomainDeleteImplicitResources(resourceName),
				),
			},
		},
	})
}

func TestAccAWSSagemakerDomain_disappears(t *testing.T) {
	var notebook sagemaker.DescribeDomainOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSagemakerDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerDomainBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerDomainExists(resourceName, &notebook),
					testAccCheckAWSSagemakerDomainDeleteImplicitResources(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSagemakerDomain(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerDomainDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sagemakerconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_domain" {
			continue
		}

		domain, err := finder.DomainByName(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		domainArn := aws.StringValue(domain.DomainArn)
		domainID, err := decodeSagemakerDomainID(domainArn)
		if err != nil {
			return err
		}

		if domainID == rs.Primary.ID {
			return fmt.Errorf("sagemaker domain %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerDomainExists(n string, codeRepo *sagemaker.DescribeDomainOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sagemakerconn
		resp, err := finder.DomainByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccCheckAWSSagemakerDomainDeleteImplicitResources(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Sagemaker domain not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Sagemaker domain name not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).efsconn
		efsFsID := rs.Primary.Attributes["home_efs_file_system_id"]
		vpcID := rs.Primary.Attributes["vpc_id"]

		resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			FileSystemId: aws.String(efsFsID),
		})

		if err != nil {
			return fmt.Errorf("Sagemaker domain EFS mount targets not found: %w", err)
		}

		//reusing EFS mount target delete for wait logic
		mountTargets := resp.MountTargets
		for _, mt := range mountTargets {
			r := resourceAwsEfsMountTarget()
			d := r.Data(nil)
			mtID := aws.StringValue(mt.MountTargetId)
			d.SetId(mtID)
			err := r.Delete(d, testAccProvider.Meta())
			if err != nil {
				return fmt.Errorf("Sagemaker domain EFS mount target (%s) failed to delete: %w", mtID, err)
			}
		}

		r := resourceAwsEfsFileSystem()
		d := r.Data(nil)
		d.SetId(efsFsID)
		err = r.Delete(d, testAccProvider.Meta())
		if err != nil {
			return fmt.Errorf("Sagemaker domain EFS file system (%s) failed to delete: %w", efsFsID, err)
		}

		var filters []*ec2.Filter
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("vpc-id"),
			Values: aws.StringSlice([]string{vpcID}),
		})

		req := &ec2.DescribeSecurityGroupsInput{
			Filters: filters,
		}

		ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn

		sgResp, err := ec2conn.DescribeSecurityGroups(req)
		if err != nil {
			return fmt.Errorf("error reading security groups: %w", err)
		}

		//revoke permissions
		for _, sg := range sgResp.SecurityGroups {
			sgID := aws.StringValue(sg.GroupId)

			if len(sg.IpPermissions) > 0 {
				req := &ec2.RevokeSecurityGroupIngressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissions,
				}
				_, err = ec2conn.RevokeSecurityGroupIngress(req)

				if err != nil {
					return fmt.Errorf("Error revoking security group %s rules: %w", sgID, err)
				}
			}

			if len(sg.IpPermissionsEgress) > 0 {
				req := &ec2.RevokeSecurityGroupEgressInput{
					GroupId:       sg.GroupId,
					IpPermissions: sg.IpPermissionsEgress,
				}
				_, err = ec2conn.RevokeSecurityGroupEgress(req)

				if err != nil {
					return fmt.Errorf("Error revoking security group %s rules: %w", sgID, err)
				}
			}
		}

		for _, sg := range sgResp.SecurityGroups {
			sgID := aws.StringValue(sg.GroupId)
			if aws.StringValue(sg.GroupName) != "default" {
				r := resourceAwsSecurityGroup()
				d := r.Data(nil)
				d.SetId(sgID)
				err = r.Delete(d, testAccProvider.Meta())
				if err != nil {
					return fmt.Errorf("Sagemaker domain EFS file system sg (%s) failed to delete: %w", sgID, err)
				}
			}
		}

		return nil
	}
}

func testAccAWSSagemakerDomainConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/"
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["sagemaker.amazonaws.com"]
    }
  }
}
`, rName)
}

func testAccAWSSagemakerDomainBasicConfig(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigSecurityGroup1(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = "%[1]s"
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role  = aws_iam_role.test.arn
    security_groups = [aws_security_sg.test.id]
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigSecurityGroup2(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_security_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role  = aws_iam_role.test.arn
    security_groups = [aws_security_sg.test.id, aws_security_sg.test2.id]
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigEFSKMS(rName string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test %s"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_sagemaker_domain" "test" {
  domain_name                     = %[1]q
  auth_mode                       = "IAM"
  vpc_id                          = aws_vpc.test.id
  subnet_ids                      = [aws_subnet.test.id]
  home_efs_file_system_kms_key_id = aws_kms_key.test.arn

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }
}
`, rName)
}

func testAccAWSSagemakerDomainConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSagemakerDomainConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSagemakerDomainConfigBase(rName) + fmt.Sprintf(`
resource "aws_sagemaker_domain" "test" {
  domain_name = %[1]q
  auth_mode   = "IAM"
  vpc_id      = aws_vpc.test.id
  subnet_ids  = [aws_subnet.test.id]

  default_user_settings {
    execution_role = aws_iam_role.test.arn
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
