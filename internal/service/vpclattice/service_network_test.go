// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfvpclattice "github.com/hashicorp/terraform-provider-aws/internal/service/vpclattice"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestIDFromIDOrARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		idOrARN string
		want    string
	}{
		{
			idOrARN: "",
			want:    "",
		},
		{
			idOrARN: "sn-1234567890abcdefg",
			want:    "sn-1234567890abcdefg",
		},
		{
			idOrARN: "arn:aws:vpc-lattice:us-east-1:123456789012:servicenetwork/sn-1234567890abcdefg", //lintignore:AWSAT003,AWSAT005
			want:    "sn-1234567890abcdefg",
		},
	}
	for _, testCase := range testCases {
		if got, want := tfvpclattice.IDFromIDOrARN(testCase.idOrARN), testCase.want; got != want {
			t.Errorf("IDFromIDOrARN(%q) = %v, want %v", testCase.idOrARN, got, want)
		}
	}
}

func TestSuppressEquivalentIDOrARN(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		old  string
		new  string
		want bool
	}{
		{
			old:  "sn-1234567890abcdefg",
			new:  "sn-1234567890abcdefg",
			want: true,
		},
		{
			old:  "sn-1234567890abcdefg",
			new:  "sn-1234567890abcdefh",
			want: false,
		},
		{
			old:  "arn:aws:vpc-lattice:us-east-1:123456789012:servicenetwork/sn-1234567890abcdefg", //lintignore:AWSAT003,AWSAT005
			new:  "sn-1234567890abcdefg",
			want: true,
		},
		{
			old:  "sn-1234567890abcdefg",
			new:  "arn:aws:vpc-lattice:us-east-1:123456789012:servicenetwork/sn-1234567890abcdefg", //lintignore:AWSAT003,AWSAT005
			want: true,
		},
		{
			old:  "arn:aws:vpc-lattice:us-east-1:123456789012:servicenetwork/sn-1234567890abcdefg", //lintignore:AWSAT003,AWSAT005
			new:  "sn-1234567890abcdefh",
			want: false,
		},
	}
	for _, testCase := range testCases {
		if got, want := tfvpclattice.SuppressEquivalentIDOrARN("test_property", testCase.old, testCase.new, nil), testCase.want; got != want {
			t.Errorf("SuppressEquivalentIDOrARN(%q, %q) = %v, want %v", testCase.old, testCase.new, got, want)
		}
	}
}

func TestAccVPCLatticeServiceNetwork_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetwork vpclattice.GetServiceNetworkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkExists(ctx, resourceName, &servicenetwork),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile("servicenetwork/.+$")),
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

func TestAccVPCLatticeServiceNetwork_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetwork vpclattice.GetServiceNetworkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkExists(ctx, resourceName, &servicenetwork),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfvpclattice.ResourceServiceNetwork(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCLatticeServiceNetwork_full(t *testing.T) {
	ctx := acctest.Context(t)
	var servicenetwork vpclattice.GetServiceNetworkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceNetworkDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkConfig_full(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkExists(ctx, resourceName, &servicenetwork),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "auth_type", "AWS_IAM"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "vpc-lattice", regexache.MustCompile("servicenetwork/.+$")),
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

func TestAccVPCLatticeServiceNetwork_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var serviceNetwork1, serviceNetwork2, serviceNetwork3 vpclattice.GetServiceNetworkOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpclattice_service_network.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkExists(ctx, resourceName, &serviceNetwork1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceNetworkConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkExists(ctx, resourceName, &serviceNetwork2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServiceNetworkConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceNetworkExists(ctx, resourceName, &serviceNetwork3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckServiceNetworkDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_vpclattice_service_network" {
				continue
			}

			_, err := tfvpclattice.FindServiceNetworkByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("VPC Lattice Service Network %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceNetworkExists(ctx context.Context, name string, servicenetwork *vpclattice.GetServiceNetworkOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetwork, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.VPCLattice, create.ErrActionCheckingExistence, tfvpclattice.ResNameServiceNetwork, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).VPCLatticeClient(ctx)
		resp, err := tfvpclattice.FindServiceNetworkByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*servicenetwork = *resp

		return nil
	}
}

// func testAccCheckServiceNetworkNotRecreated(before, after *vpclattice.DescribeServiceNetworkResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.StringValue(before.ServiceNetworkId), aws.StringValue(after.ServiceNetworkId); before != after {
// 			return create.Error(names.VPCLattice, create.ErrActionCheckingNotRecreated, tfvpclattice.ResNameServiceNetwork, aws.StringValue(before.ServiceNetworkId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccServiceNetworkConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}
`, rName)
}

func testAccServiceNetworkConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name      = %[1]q
  auth_type = "AWS_IAM"
}
`, rName)
}

func testAccServiceNetworkConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceNetworkConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpclattice_service_network" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
