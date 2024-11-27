// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCBlockPublicAccessOptions_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		acctest.CtBasic: testAccVPCBlockPublicAccessOptions_basic,
		"update":        testAccVPCBlockPublicAccessOptions_update,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccVPCBlockPublicAccessOptions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_block_public_access_options.test"
	rMode := string(awstypes.InternetGatewayBlockModeBlockBidirectional)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("aws_region"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_block_mode"), knownvalue.StringExact(rMode)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVPCBlockPublicAccessOptions_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_block_public_access_options.test"
	rMode1 := string(awstypes.InternetGatewayBlockModeBlockBidirectional)
	rMode2 := string(awstypes.InternetGatewayBlockModeBlockIngress)
	rMode3 := string(awstypes.InternetGatewayBlockModeOff)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.EC2)
			testAccPreCheckVPCBlockPublicAccess(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode1),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_block_mode"), knownvalue.StringExact(rMode1)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_block_mode"), knownvalue.StringExact(rMode2)),
				},
			},
			{
				Config: testAccVPCBlockPublicAccessOptionsConfig_basic(rMode3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("internet_gateway_block_mode"), knownvalue.StringExact(rMode3)),
				},
			},
		},
	})
}

func testAccPreCheckVPCBlockPublicAccess(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeVpcBlockPublicAccessOptionsInput{}
	_, err := conn.DescribeVpcBlockPublicAccessOptions(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccVPCBlockPublicAccessOptionsConfig_basic(rMode string) string {
	return fmt.Sprintf(`
resource "aws_vpc_block_public_access_options" "test" {
  internet_gateway_block_mode = %[1]q
}
`, rMode)
}
