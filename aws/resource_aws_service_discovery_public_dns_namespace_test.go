package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
)

func init() {
	resource.AddTestSweepers("aws_service_discovery_public_dns_namespace", &resource.Sweeper{
		Name: "aws_service_discovery_public_dns_namespace",
		F:    testSweepServiceDiscoveryPublicDnsNamespaces,
		Dependencies: []string{
			"aws_service_discovery_service",
		},
	})
}

func testSweepServiceDiscoveryPublicDnsNamespaces(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sdconn
	var sweeperErrs *multierror.Error

	input := &servicediscovery.ListNamespacesInput{
		Filters: []*servicediscovery.NamespaceFilter{
			{
				Condition: aws.String(servicediscovery.FilterConditionEq),
				Name:      aws.String(servicediscovery.NamespaceFilterNameType),
				Values:    aws.StringSlice([]string{servicediscovery.NamespaceTypeDnsPublic}),
			},
		},
	}

	err = conn.ListNamespacesPages(input, func(page *servicediscovery.ListNamespacesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, namespace := range page.Namespaces {
			if namespace == nil {
				continue
			}

			id := aws.StringValue(namespace.Id)
			input := &servicediscovery.DeleteNamespaceInput{
				Id: namespace.Id,
			}

			log.Printf("[INFO] Deleting Service Discovery Public DNS Namespace: %s", id)
			output, err := conn.DeleteNamespace(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Service Discovery Public DNS Namespace (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if output != nil && output.OperationId != nil {
				if _, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Service Discovery Public DNS Namespace (%s) deletion: %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Service Discovery Public DNS Namespaces sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Service Discovery Public DNS Namespaces: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSServiceDiscoveryPublicDnsNamespace_basic(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := acctest.RandStringFromCharSet(5, acctest.CharSetAlpha) + ".terraformtesting.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPublicDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPublicDnsNamespaceExists("aws_service_discovery_public_dns_namespace.test"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_public_dns_namespace.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_public_dns_namespace.test", "hosted_zone"),
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

func TestAccAWSServiceDiscoveryPublicDnsNamespace_longname(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := acctest.RandStringFromCharSet(64-len("terraformtesting.com"), acctest.CharSetAlpha) + ".terraformtesting.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPublicDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPublicDnsNamespaceExists("aws_service_discovery_public_dns_namespace.test"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_public_dns_namespace.test", "arn"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_public_dns_namespace.test", "hosted_zone"),
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

func testAccCheckAwsServiceDiscoveryPublicDnsNamespaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sdconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_public_dns_namespace" {
			continue
		}

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		if err != nil {
			if isAWSErr(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsServiceDiscoveryPublicDnsNamespaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).sdconn

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		return err
	}
}

func testAccServiceDiscoveryPublicDnsNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name        = %q
  description = "test"
}
`, rName)
}
