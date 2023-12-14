// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect_test

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func testAccCheckVirtualInterfaceExists(ctx context.Context, name string, vif *directconnect.VirtualInterface) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn(ctx)

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualInterfacesWithContext(ctx, &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		for _, v := range resp.VirtualInterfaces {
			if aws.StringValue(v.VirtualInterfaceId) == rs.Primary.ID {
				*vif = *v

				return nil
			}
		}

		return fmt.Errorf("Direct Connect virtual interface (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckVirtualInterfaceDestroy(ctx context.Context, s *terraform.State, t string) error { // nosemgrep:ci.semgrep.acctest.naming.destroy-check-signature
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn(ctx)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != t {
			continue
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeVirtualInterfacesWithContext(ctx, &directconnect.DescribeVirtualInterfacesInput{
			VirtualInterfaceId: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrMessageContains(err, directconnect.ErrCodeClientException, "does not exist") {
			continue
		}
		if err != nil {
			return err
		}

		for _, v := range resp.VirtualInterfaces {
			if aws.StringValue(v.VirtualInterfaceId) == rs.Primary.ID && aws.StringValue(v.VirtualInterfaceState) != directconnect.VirtualInterfaceStateDeleted {
				return fmt.Errorf("[DESTROY ERROR] Direct Connect virtual interface (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}
