package ec2_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2SpotInstanceRequest_basic(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequestAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "terminate"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_disappears(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceSpotInstanceRequest(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_tags(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
			{
				Config: testAccSpotInstanceRequestConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccSpotInstanceRequestConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_keyName(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	keyPairResourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_KeyName(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttrPair(resourceName, "key_name", keyPairResourceName, "key_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_withLaunchGroup(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_withLaunchGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequestAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_group", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_withBlockDuration(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_withBlockDuration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequestAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "block_duration_minutes", "60"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_vpc(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestVPCConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequestAttributes(&sir),
					testAccCheckSpotInstanceRequestAttributesVPC(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_validUntil(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := testAccSpotInstanceRequestValidUntil(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestValidUntilConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequestAttributes(&sir),
					testAccCheckSpotInstanceRequestAttributesValidUntil(&sir, validUntil),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_withoutSpotPrice(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_withoutSpotPrice(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequestAttributesCheckSIRWithoutSpot(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_subnetAndSGAndPublicIPAddress(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_SubnetAndSGAndPublicIPAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequest_InstanceAttributes(&sir, rName),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_networkInterfaceAttributes(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_SubnetAndSGAndPublicIPAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					testAccCheckSpotInstanceRequest_InstanceAttributes(&sir, rName),
					testAccCheckSpotInstanceRequest_NetworkInterfaceAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_getPasswordData(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_getPasswordData(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttrSet(resourceName, "password_data"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"get_password_data", "password_data", "user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_interruptStop(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestInterruptConfig(rName, "stop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "stop"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_interruptHibernate(t *testing.T) {
	var sir ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestInterruptConfig(rName, "hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_interruptUpdate(t *testing.T) {
	var sir1, sir2 ec2.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckSpotInstanceRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestInterruptConfig(rName, "hibernate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir1),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
				),
			},
			{
				Config: testAccSpotInstanceRequestInterruptConfig(rName, "terminate"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(resourceName, &sir2),
					testAccCheckSpotInstanceRequestRecreated(&sir1, &sir2),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "terminate"),
				),
			},
		},
	})
}

func testAccSpotInstanceRequestValidUntil(t *testing.T) string {
	return testAccSpotInstanceRequestTime(t, "12h")
}

func testAccSpotInstanceRequestTime(t *testing.T, duration string) string {
	n := time.Now().UTC()
	d, err := time.ParseDuration(duration)
	if err != nil {
		t.Fatalf("err parsing time duration: %s", err)
	}
	return n.Add(d).Format(time.RFC3339)
}

func testAccCheckSpotInstanceRequestDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_spot_instance_request" {
			continue
		}

		_, err := tfec2.FindSpotInstanceRequestByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			// Now check if the associated Spot Instance was also destroyed.
			instanceID := rs.Primary.Attributes["spot_instance_id"]
			_, err := tfec2.FindInstanceByID(conn, instanceID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Instance %s still exists", instanceID)
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Spot Instance Request %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckSpotInstanceRequestExists(n string, v *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Spot Instance Request ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		output, err := tfec2.FindSpotInstanceRequestByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributes(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *sir.SpotPrice != "0.050000" {
			return fmt.Errorf("Unexpected spot price: %s", *sir.SpotPrice)
		}
		if *sir.State != ec2.SpotInstanceStateActive {
			return fmt.Errorf("Unexpected request state: %s", *sir.State)
		}
		if *sir.Status.Code != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", *sir.State)
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributesValidUntil(
	sir *ec2.SpotInstanceRequest, validUntil string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sir.ValidUntil.Format(time.RFC3339) != validUntil {
			return fmt.Errorf("Unexpected valid_until time: %s", sir.ValidUntil.String())
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributesCheckSIRWithoutSpot(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *sir.State != ec2.SpotInstanceStateActive {
			return fmt.Errorf("Unexpected request state: %s", *sir.State)
		}
		if *sir.Status.Code != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", *sir.State)
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequest_InstanceAttributes(v *ec2.SpotInstanceRequest, sgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		instance, err := tfec2.FindInstanceByID(conn, aws.StringValue(v.InstanceId))

		if err != nil {
			return err
		}

		for _, v := range instance.SecurityGroups {
			if aws.StringValue(v.GroupName) == sgName {
				return nil
			}
		}

		return fmt.Errorf("Error in matching Spot Instance Security Group, expected %s, got %v", sgName, instance.SecurityGroups)
	}
}

func testAccCheckSpotInstanceRequest_NetworkInterfaceAttributes(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nis := sir.LaunchSpecification.NetworkInterfaces
		if nis == nil || len(nis) != 1 {
			return fmt.Errorf("Expected exactly 1 network interface, found %d", len(nis))
		}

		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributesVPC(
	sir *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nis := sir.LaunchSpecification.NetworkInterfaces
		if nis == nil || len(nis) != 1 {
			return fmt.Errorf("Expected exactly 1 network interface, found %d", len(nis))
		}

		ni := nis[0]

		if ni.SubnetId == nil {
			return fmt.Errorf("Expected SubnetId not be non-empty for %s as the instance belongs to a VPC", *sir.InstanceId)
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequestRecreated(before, after *ec2.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(before.InstanceId), aws.StringValue(after.InstanceId); before == after {
			return fmt.Errorf("Spot Instance (%s) not recreated", before)
		}

		return nil
	}
}

func testAccSpotInstanceRequestConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName))
}

func testAccSpotInstanceRequestConfigTags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true

  tags = {
    %[2]q = %[3]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName, tagKey1, tagValue1))
}

func testAccSpotInstanceRequestConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccSpotInstanceRequestValidUntilConfig(rName string, validUntil string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  valid_until          = %[2]q
  wait_for_fulfillment = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName, validUntil))
}

func testAccSpotInstanceRequestConfig_withoutSpotPrice(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  wait_for_fulfillment = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName))
}

func testAccSpotInstanceRequestConfig_KeyName(rName, publicKey string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  key_name             = aws_key_pair.test.key_name
  wait_for_fulfillment = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName, publicKey))
}

func testAccSpotInstanceRequestConfig_withLaunchGroup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
  launch_group         = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName))
}

func testAccSpotInstanceRequestConfig_withBlockDuration(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                    = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type          = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price             = "0.05"
  wait_for_fulfillment   = true
  block_duration_minutes = 60

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName))
}

func testAccSpotInstanceRequestVPCConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
  subnet_id            = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName))
}

func testAccSpotInstanceRequestConfig_SubnetAndSGAndPublicIPAddress(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price                  = "0.05"
  wait_for_fulfillment        = true
  subnet_id                   = aws_subnet.test.id
  vpc_security_group_ids      = [aws_security_group.test.id]
  associate_public_ip_address = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName))
}

func testAccSpotInstanceRequestConfig_getPasswordData(rName, publicKey string) string {
	return acctest.ConfigCompose(
		testAccLatestWindowsServer2016CoreAMIConfig(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.win2016core-ami.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  key_name             = aws_key_pair.test.key_name
  wait_for_fulfillment = true
  get_password_data    = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = %[2]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName, publicKey))
}

func testAccSpotInstanceRequestInterruptConfig(rName, interruptionBehavior string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinuxHvmEbsAmi(),
		acctest.AvailableEC2InstanceTypeForRegion("c5.large", "c4.large"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                            = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type                  = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price                     = "0.07"
  wait_for_fulfillment           = true
  instance_interruption_behavior = %[2]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName, interruptionBehavior))
}
