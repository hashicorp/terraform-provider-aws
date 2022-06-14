package ds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfds "github.com/hashicorp/terraform-provider-aws/internal/service/ds"
)

func TestAccDSConditionalForwarder_Condition_basic(t *testing.T) {
	resourceName := "aws_directory_service_conditional_forwarder.fwd"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domainName := acctest.RandomDomainName()
	ip1, ip2, ip3 := "8.8.8.8", "1.1.1.1", "8.8.4.4"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckDirectoryService(t) },
		ErrorCheck:        acctest.ErrorCheck(t, directoryservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckConditionalForwarderDestroy,
		Steps: []resource.TestStep{
			// test create
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(
						resourceName,
						[]string{ip1, ip2},
					),
				),
			},
			// test update
			{
				Config: testAccConditionalForwarderConfig_basic(rName, domainName, ip1, ip3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConditionalForwarderExists(
						resourceName,
						[]string{ip1, ip3},
					),
				),
			},
			// test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckConditionalForwarderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_conditional_forwarder" {
			continue
		}

		directoryId, domainName, err := tfds.ParseConditionalForwarderID(rs.Primary.ID)
		if err != nil {
			return err
		}

		res, err := conn.DescribeConditionalForwarders(&directoryservice.DescribeConditionalForwardersInput{
			DirectoryId:       aws.String(directoryId),
			RemoteDomainNames: []*string{aws.String(domainName)},
		})

		if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
			continue
		}

		if err != nil {
			return err
		}

		if len(res.ConditionalForwarders) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Conditional Forwarder to be gone, but was still found")
		}
	}

	return nil
}

func testAccCheckConditionalForwarderExists(name string, dnsIps []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		directoryId, domainName, err := tfds.ParseConditionalForwarderID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DSConn

		res, err := conn.DescribeConditionalForwarders(&directoryservice.DescribeConditionalForwardersInput{
			DirectoryId:       aws.String(directoryId),
			RemoteDomainNames: []*string{aws.String(domainName)},
		})

		if err != nil {
			return err
		}

		if len(res.ConditionalForwarders) == 0 {
			return fmt.Errorf("No Conditional Fowrwarder found")
		}

		cfd := res.ConditionalForwarders[0]

		if dnsIps != nil {
			if len(dnsIps) != len(cfd.DnsIpAddrs) {
				return fmt.Errorf("DnsIpAddrs length mismatch")
			}

			for k, v := range cfd.DnsIpAddrs {
				if *v != dnsIps[k] {
					return fmt.Errorf("DnsIp mismatch, '%s' != '%s' at index '%d'", *v, dnsIps[k], k)
				}
			}
		}

		return nil
	}
}

func testAccConditionalForwarderConfig_basic(rName, domain, ip1, ip2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_directory_service_conditional_forwarder" "fwd" {
  directory_id = aws_directory_service_directory.test.id

  remote_domain_name = "test.example.com"

  dns_ips = [
    %[2]q,
    %[3]q,
  ]
}

resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = aws_subnet.test[*].id
  }

  tags = {
    Name = "terraform-testacc-directory-service-conditional-forwarder"
  }
}
`, domain, ip1, ip2),
	)
}
