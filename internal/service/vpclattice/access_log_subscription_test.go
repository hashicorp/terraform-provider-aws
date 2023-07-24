// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeAccessLogSubscription_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"
	serviceNetworkResourceName := "aws_vpclattice_service_network.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_basicS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					resource.TestCheckResourceAttrPair(resourceName, "resource_identifier", serviceNetworkResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", s3BucketResourceName, "arn"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", names.VPCLatticeEndpointID, regexp.MustCompile(`accesslogsubscription/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccVPCLatticeAccessLogSubscription_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_basicS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceAccessLogSubscription(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCLatticeAccessLogSubscription_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var accesslogsubscription1, accesslogsubscription2, accesslogsubscription3 vpclattice.GetAccessLogSubscriptionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_access_log_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccessLogSubscriptionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccessLogSubscriptionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription1),
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
				Config: testAccAccessLogSubscriptionConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAccessLogSubscriptionConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccessLogSubscriptionExists(ctx, resourceName, &accesslogsubscription3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckAccessLogSubscriptionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_access_log_subscription" {
				continue
			}

			_, err := conn.GetAccessLogSubscription(ctx, &vpclattice.GetAccessLogSubscriptionInput{
				AccessLogSubscriptionIdentifier: aws.String(rs.Primary.ID),
			})
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.VPCLattice, create.ErrActionCheckingDestroyed, tfvpclattice.ResNameAccessLogSubscription, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckAccessLogSubscriptionExists(ctx context.Context, name string, accesslogsubscription *vpclattice.GetAccessLogSubscriptionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameAccessLogSubscription, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameAccessLogSubscription, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)
		resp, err := conn.GetAccessLogSubscription(ctx, &vpclattice.GetAccessLogSubscriptionInput{
			AccessLogSubscriptionIdentifier: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameAccessLogSubscription, rs.Primary.ID, err)
		}

		*accesslogsubscription = *resp

		return nil
	}
}

func testAccAccessLogSubscriptionConfig_basicS3(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service_network.test.id
  destination_arn     = aws_s3_bucket.test.arn
}
`, rName)
}

func testAccAccessLogSubscriptionConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service_network.test.id
  destination_arn     = aws_s3_bucket.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAccessLogSubscriptionConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_vpclattice_access_log_subscription" "test" {
  resource_identifier = aws_vpclattice_service_network.test.id
  destination_arn     = aws_s3_bucket.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
