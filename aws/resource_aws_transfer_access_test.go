package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tftransfer "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/transfer/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

//TODO: Test with EFS
//TODO: Test with Posix Profile
//TODO: Test with Policy

func testAccAWSTransferAccess_basic(t *testing.T) {
	var conf transfer.DescribedAccess
	resourceName := "aws_transfer_access.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		ErrorCheck:   testAccErrorCheck(t, transfer.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferAccessDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferAccessBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferAccessExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "external_id", "S-1-1-12-1234567890-123456789-1234567890-1234"),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/"+rName+"/"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "PATH"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferAccessUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferAccessExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "external_id", "S-1-1-12-1234567890-123456789-1234567890-1234"),
					resource.TestCheckResourceAttr(resourceName, "home_directory", "/"+rName+"/test"),
					resource.TestCheckResourceAttr(resourceName, "home_directory_type", "PATH"),
				),
			},
		},
	})
}

func testAccCheckAWSTransferAccessExists(n string, v *transfer.DescribedAccess) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Access ID is set")
		}

		serverID, externalID, err := tftransfer.AccessParseResourceID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing Transfer Access ID: %s", err)
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn

		output, err := finder.AccessByID(conn, serverID, externalID)

		if err != nil {
			return err
		}

		*v = *output.Access

		return nil
	}
}

func testAccCheckAWSTransferAccessDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_access" {
			continue
		}

		externalID := rs.Primary.Attributes["external_id"]
		serverID := rs.Primary.Attributes["server_id"]
		_, err := finder.AccessByID(conn, serverID, externalID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Transfer Access %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAWSTransferAccessConfigBase(rName string) string {
	return fmt.Sprintf(`
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
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_subnet" "test2" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  map_public_ip_on_launch = true
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [
      aws_subnet.test.id,
      aws_subnet.test2.id
    ]
  }
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": "transfer.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowFullAccesstoCloudWatchLogs",
    "Effect": "Allow",
    "Action": [
      "logs:*"
    ],
    "Resource": "*"
  },
{
                "Sid": "AllowFullAccesstoS3",
                "Effect": "Allow",
                "Action": [
                  "s3:*"
                ],
                "Resource": "*"
              }]
}
POLICY
}

  resource "aws_s3_bucket" "test" {
    bucket = %[1]q
	acl = "private"
}

  resource "aws_transfer_server" "test" {
  identity_provider_type = "AWS_DIRECTORY_SERVICE"
  directory_id			 = "${aws_directory_service_directory.test.id}"
  logging_role           = aws_iam_role.test.arn  
}
`, rName)
}

func testAccAWSTransferAccessBasicConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferAccessConfigBase(rName),
		`
		resource "aws_transfer_access" "test" {
		  external_id = "S-1-1-12-1234567890-123456789-1234567890-1234"
		  server_id = aws_transfer_server.test.id
		  role = aws_iam_role.test.arn
		  home_directory = "/${aws_s3_bucket.test.id}/"
		  home_directory_type = "PATH"		  		 
		}
		`)
}

func testAccAWSTransferAccessUpdatedConfig(rName string) string {
	return composeConfig(
		testAccAWSTransferAccessConfigBase(rName),
		`
		resource "aws_transfer_access" "test" {
		  external_id = "S-1-0-09-0987654321-098765432-0987654321-0987"
		  server_id = aws_transfer_server.test.id
          role = aws_iam_role.test.arn
		  home_directory = "/${aws_s3_bucket.test.id}/test"
		  home_directory_type = "PATH"		
		}
		`)
}
