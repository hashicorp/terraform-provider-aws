// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/datasync"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccDataSyncLocationObjectStorage_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var locationObjectStorage datasync.DescribeLocationObjectStorageOutput
	resourceName := "aws_datasync_location_object_storage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationObjectStorageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationObjectStorageConfig_basic(rName, domain),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage),
					resource.TestCheckResourceAttr(resourceName, "access_key", ""),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", rName),
					resource.TestCheckNoResourceAttr(resourceName, "secret_key"),
					resource.TestCheckResourceAttr(resourceName, "server_certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "server_hostname", domain),
					resource.TestCheckResourceAttr(resourceName, "server_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "server_protocol", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "uri", fmt.Sprintf("object-storage://%s/%s/", domain, rName)),
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

func TestAccDataSyncLocationObjectStorage_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var locationObjectStorage datasync.DescribeLocationObjectStorageOutput
	resourceName := "aws_datasync_location_object_storage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationObjectStorageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationObjectStorageConfig_basic(rName, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfdatasync.ResourceLocationObjectStorage(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccDataSyncLocationObjectStorage_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var locationObjectStorage datasync.DescribeLocationObjectStorageOutput
	resourceName := "aws_datasync_location_object_storage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationObjectStorageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationObjectStorageConfig_tags1(rName, domain, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage),
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
				Config: testAccLocationObjectStorageConfig_tags2(rName, domain, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationObjectStorageConfig_tags1(rName, domain, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccDataSyncLocationObjectStorage_serverCertificate(t *testing.T) {
	ctx := acctest.Context(t)
	var locationObjectStorage datasync.DescribeLocationObjectStorageOutput
	resourceName := "aws_datasync_location_object_storage.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := acctest.RandomDomainName()
	caKey := acctest.TLSRSAPrivateKeyPEM(t, 2048)
	caCertificate := acctest.TLSRSAX509SelfSignedCACertificatePEM(t, caKey)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, datasync.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLocationObjectStorageDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccLocationObjectStorageConfig_serverCertificate(rName, domain, caCertificate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLocationObjectStorageExists(ctx, resourceName, &locationObjectStorage),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "server_certificate", caCertificate),
					resource.TestCheckResourceAttr(resourceName, "server_hostname", domain),
					resource.TestCheckResourceAttr(resourceName, "server_port", "443"),
					resource.TestCheckResourceAttr(resourceName, "server_protocol", "HTTPS"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/test/"),
					resource.TestCheckResourceAttr(resourceName, "uri", fmt.Sprintf("object-storage://%s/%s/test/", domain, rName)),
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

func testAccCheckLocationObjectStorageDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_datasync_location_object_storage" {
				continue
			}

			_, err := tfdatasync.FindLocationObjectStorageByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("DataSync Location Object Storage %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckLocationObjectStorageExists(ctx context.Context, n string, v *datasync.DescribeLocationObjectStorageOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn(ctx)

		output, err := tfdatasync.FindLocationObjectStorageByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccLocationObjectStorageConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test[0].availability_zone", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_default_route_table.test]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationObjectStorageConfig_basic(rName, domain string) string {
	return acctest.ConfigCompose(testAccLocationObjectStorageConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_object_storage" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  server_hostname = %[2]q
  bucket_name     = %[1]q
  server_protocol = "HTTP"
  server_port     = 8080
}
`, rName, domain))
}

func testAccLocationObjectStorageConfig_tags1(rName, domain, key1, value1 string) string {
	return acctest.ConfigCompose(testAccLocationObjectStorageConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_object_storage" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  server_hostname = %[2]q
  bucket_name     = %[1]q

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, domain, key1, value1))
}

func testAccLocationObjectStorageConfig_tags2(rName, domain, key1, value1, key2, value2 string) string {
	return acctest.ConfigCompose(testAccLocationObjectStorageConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_object_storage" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  server_hostname = %[2]q
  bucket_name     = %[1]q

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, domain, key1, value1, key2, value2))
}

func testAccLocationObjectStorageConfig_serverCertificate(rName, domain, certificate string) string {
	return acctest.ConfigCompose(testAccLocationObjectStorageConfig_base(rName), fmt.Sprintf(`
resource "aws_datasync_location_object_storage" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  server_hostname = %[2]q
  bucket_name     = %[1]q
  subdirectory    = "/test/"

  server_certificate = "%[3]s"
}
`, rName, domain, acctest.TLSPEMEscapeNewlines(certificate)))
}
