package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfconnect "github.com/hashicorp/terraform-provider-aws/aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_connect_instance", &resource.Sweeper{
		Name: "aws_connect_instance",
		F:    testSweepConnectInstance,
	})
}

func testSweepConnectInstance(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).connectconn
	ctx := context.Background()
	var errs *multierror.Error
	sweepResources := make([]*testSweepResource, 0)

	// MaxResults:  Maximum value of 10. https://docs.aws.amazon.com/connect/latest/APIReference/API_ListInstances.html
	input := &connect.ListInstancesInput{MaxResults: aws.Int64(tfconnect.ListInstancesMaxResults)}

	err = conn.ListInstancesPagesWithContext(ctx, input, func(page *connect.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, instanceSummary := range page.InstanceSummaryList {
			if instanceSummary == nil {
				continue
			}

			id := aws.StringValue(instanceSummary.Id)

			log.Printf("[INFO] Deleting Connect Instance (%s)", id)
			r := resourceAwsConnectInstance()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Connect Instances: %w", err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Connect Instances for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Connect Instances sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccAwsConnectInstance_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":     testAccAwsConnectInstance_basic,
		"directory": testAccAwsConnectInstance_directory,
		"saml":      testAccAwsConnectInstance_saml,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAwsConnectInstance_basic(t *testing.T) {
	var v connect.DescribeInstanceOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "connect", regexp.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices_enabled", "true"), //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", "false"),       //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", "true"),             //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", "true"), //verified default result from ListInstanceAttributes()
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeConnectManaged),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", "true"),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexp.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", "true"),
					acctest.MatchResourceAttrGlobalARN(resourceName, "service_role", "iam", regexp.MustCompile(`role/aws-service-role/connect.amazonaws.com/.+`)),
					resource.TestCheckResourceAttr(resourceName, "status", connect.InstanceStatusActive),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsConnectInstanceConfigBasicFlipped(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "connect", regexp.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttr(resourceName, "auto_resolve_best_voices_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "contact_flow_logs_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "contact_lens_enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttr(resourceName, "early_media_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "inbound_calls_enabled", "false"),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexp.MustCompile(rName)),
					resource.TestCheckResourceAttr(resourceName, "outbound_calls_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", connect.InstanceStatusActive),
				),
			},
		},
	})
}

func testAccAwsConnectInstance_directory(t *testing.T) {
	var v connect.DescribeInstanceOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceConfigDirectory(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeExistingDirectory),
					resource.TestCheckResourceAttr(resourceName, "status", connect.InstanceStatusActive),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"directory_id"},
			},
		},
	})
}

func testAccAwsConnectInstance_saml(t *testing.T) {
	var v connect.DescribeInstanceOutput
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_instance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceConfigSAML(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "identity_management_type", connect.DirectoryTypeSaml),
					testAccCheckAwsConnectInstanceExists(resourceName, &v),
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

func testAccCheckAwsConnectInstanceExists(resourceName string, instance *connect.DescribeInstanceOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect instance not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect instance ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		input := &connect.DescribeInstanceInput{
			InstanceId: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeInstance(input)
		if err != nil {
			return err
		}
		if output == nil {
			return fmt.Errorf("Connect instance %q does not exist", rs.Primary.ID)
		}

		*instance = *output

		return nil
	}
}

func testAccCheckAwsConnectInstanceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_instance" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		instanceID := rs.Primary.ID

		input := &connect.DescribeInstanceInput{
			InstanceId: aws.String(instanceID),
		}

		_, connectErr := conn.DescribeInstance(input)
		// Verify the error is what we want
		if connectErr != nil {
			if awsErr, ok := connectErr.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return connectErr
		}
	}
	return nil
}

func testAccAwsConnectInstanceConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccAwsConnectInstanceConfigBasicFlipped(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  auto_resolve_best_voices_enabled = false
  contact_flow_logs_enabled        = true
  contact_lens_enabled             = false
  early_media_enabled              = false
  identity_management_type         = "CONNECT_MANAGED"
  inbound_calls_enabled            = false
  instance_alias                   = %[1]q
  outbound_calls_enabled           = false
}
`, rName)
}

func testAccAwsConnectInstanceConfigDirectory(rName string) string {
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
    Name = "terraform-testacc-directory-service-directory-tags"
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-test"
  }
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.test.id
    subnet_ids = [aws_subnet.test1.id, aws_subnet.test2.id]
  }
}

resource "aws_connect_instance" "test" {
  directory_id             = aws_directory_service_directory.test.id
  identity_management_type = "EXISTING_DIRECTORY"
  instance_alias           = %[1]q
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccAwsConnectInstanceConfigSAML(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "SAML"
  instance_alias           = %[1]q
  inbound_calls_enabled    = true
  outbound_calls_enabled   = true
}
`, rName)
}
