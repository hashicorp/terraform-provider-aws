// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2SpotInstanceRequest_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
					testAccCheckSpotInstanceRequestAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "terminate"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceSpotInstanceRequest(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var sir1, sir2, sir3 awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"user_data_replace_on_change", "wait_for_fulfillment"},
			},
			{
				Config: testAccSpotInstanceRequestConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir2),
					testAccCheckSpotInstanceRequestIDsEqual(&sir2, &sir1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccSpotInstanceRequestConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir3),
					testAccCheckSpotInstanceRequestIDsEqual(&sir3, &sir2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_keyName(t *testing.T) {
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	keyPairResourceName := "aws_key_pair.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_keyName(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_launchGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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

func TestAccEC2SpotInstanceRequest_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_vpc(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	validUntil := testAccSpotInstanceRequestValidUntil(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_validUntil(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_noPrice(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_subnetAndSGAndPublicIPAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
					testAccCheckSpotInstanceRequest_InstanceAttributes(ctx, &sir, rName),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_subnetAndSGAndPublicIPAddress(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
					testAccCheckSpotInstanceRequest_InstanceAttributes(ctx, &sir, rName),
					testAccCheckSpotInstanceRequest_NetworkInterfaceAttributes(&sir),
					resource.TestCheckResourceAttr(resourceName, "associate_public_ip_address", acctest.CtTrue),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	publicKey, _, err := sdkacctest.RandSSHKeyPair(acctest.DefaultEmailAddress)
	if err != nil {
		t.Fatalf("error generating random SSH key: %s", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_getPasswordData(rName, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_interrupt(rName, "stop", acctest.CtFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_interrupt(rName, "hibernate", acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
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
	ctx := acctest.Context(t)
	var sir1, sir2 awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_interrupt(rName, "hibernate", acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir1),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "hibernate"),
				),
			},
			{
				Config: testAccSpotInstanceRequestConfig_interrupt(rName, "terminate", acctest.CtFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir2),
					testAccCheckSpotInstanceRequestIDsNotEqual(&sir1, &sir2),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behavior", "terminate"),
				),
			},
		},
	})
}

func TestAccEC2SpotInstanceRequest_withInstanceProfile(t *testing.T) {
	ctx := acctest.Context(t)
	var sir awstypes.SpotInstanceRequest
	resourceName := "aws_spot_instance_request.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSpotInstanceRequestDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSpotInstanceRequestConfig_withInstanceProfile(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSpotInstanceRequestExists(ctx, resourceName, &sir),
					resource.TestCheckResourceAttrSet(resourceName, "iam_instance_profile"),
					resource.TestCheckResourceAttr(resourceName, "spot_bid_status", "fulfilled"),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
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

func testAccCheckSpotInstanceRequestDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_spot_instance_request" {
				continue
			}

			_, err := tfec2.FindSpotInstanceRequestByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				// Now check if the associated Spot Instance was also destroyed.
				instanceID := rs.Primary.Attributes["spot_instance_id"]
				_, err := tfec2.FindInstanceByID(ctx, conn, instanceID)

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
}

func testAccCheckSpotInstanceRequestExists(ctx context.Context, n string, v *awstypes.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Spot Instance Request ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		output, err := tfec2.FindSpotInstanceRequestByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributes(
	sir *awstypes.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v := aws.ToString(sir.SpotPrice); v != "0.050000" {
			return fmt.Errorf("Unexpected spot price: %s", v)
		}
		if sir.State != awstypes.SpotInstanceStateActive {
			return fmt.Errorf("Unexpected request state: %s", sir.State)
		}
		if v := aws.ToString(sir.Status.Code); v != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", v)
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributesValidUntil(
	sir *awstypes.SpotInstanceRequest, validUntil string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sir.ValidUntil.Format(time.RFC3339) != validUntil {
			return fmt.Errorf("Unexpected valid_until time: %s", sir.ValidUntil.String())
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributesCheckSIRWithoutSpot(
	sir *awstypes.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sir.State != awstypes.SpotInstanceStateActive {
			return fmt.Errorf("Unexpected request state: %s", sir.State)
		}
		if v := aws.ToString(sir.Status.Code); v != "fulfilled" {
			return fmt.Errorf("Unexpected bid status: %s", sir.State)
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequest_InstanceAttributes(ctx context.Context, v *awstypes.SpotInstanceRequest, sgName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		instance, err := tfec2.FindInstanceByID(ctx, conn, aws.ToString(v.InstanceId))

		if err != nil {
			return err
		}

		for _, v := range instance.SecurityGroups {
			if aws.ToString(v.GroupName) == sgName {
				return nil
			}
		}

		return fmt.Errorf("Error in matching Spot Instance Security Group, expected %s, got %v", sgName, instance.SecurityGroups)
	}
}

func testAccCheckSpotInstanceRequest_NetworkInterfaceAttributes(
	sir *awstypes.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		nis := sir.LaunchSpecification.NetworkInterfaces
		if nis == nil || len(nis) != 1 {
			return fmt.Errorf("Expected exactly 1 network interface, found %d", len(nis))
		}

		return nil
	}
}

func testAccCheckSpotInstanceRequestAttributesVPC(
	sir *awstypes.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if sir.LaunchSpecification.SubnetId == nil {
			return fmt.Errorf("Expected SubnetId not be non-empty for %s as the instance belongs to a VPC", *sir.InstanceId)
		}
		return nil
	}
}

func testAccCheckSpotInstanceRequestIDsEqual(sir1, sir2 *awstypes.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(sir1.SpotInstanceRequestId) != aws.ToString(sir2.SpotInstanceRequestId) {
			return fmt.Errorf("Spot Instance Request IDs are not equal")
		}

		return nil
	}
}

func testAccCheckSpotInstanceRequestIDsNotEqual(sir1, sir2 *awstypes.SpotInstanceRequest) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(sir1.SpotInstanceRequestId) == aws.ToString(sir2.SpotInstanceRequestId) {
			return fmt.Errorf("Spot Instance Request IDs are equal")
		}

		return nil
	}
}

func testAccSpotInstanceRequestConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_validUntil(rName string, validUntil string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_noPrice(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_keyName(rName, publicKey string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_launchGroup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_vpc(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
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
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_subnetAndSGAndPublicIPAddress(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                         = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
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

func testAccSpotInstanceRequestConfig_interrupt(rName, interruptionBehavior, encrypted string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("c5.large", "c4.large"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  ami                            = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type                  = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price                     = "0.07"
  wait_for_fulfillment           = true
  instance_interruption_behavior = %[2]q

  root_block_device {
    encrypted   = %[3]q
    volume_type = "gp2"
    volume_size = 11
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_tag" "test" {
  resource_id = aws_spot_instance_request.test.spot_instance_id
  key         = "Name"
  value       = %[1]q
}
`, rName, interruptionBehavior, encrypted))
}

func testAccSpotInstanceRequestConfig_withInstanceProfile(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.AvailableEC2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {
                "Service": "ec2.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
        }
    ]
}
EOF
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

resource "aws_spot_instance_request" "test" {
  ami                  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  iam_instance_profile = aws_iam_instance_profile.test.name
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
