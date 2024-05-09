// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusVPCStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusVPCIPv6CIDRBlockAssociationStateV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, _, err := findVPCIPv6CIDRBlockAssociationByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.Ipv6CidrBlockState, string(output.Ipv6CidrBlockState.State), nil
	}
}

func statusVPCAttributeValueV2(ctx context.Context, conn *ec2.Client, id string, attribute awstypes.VpcAttributeName) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		attributeValue, err := findVPCAttributeV2(ctx, conn, id, attribute)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return attributeValue, strconv.FormatBool(attributeValue), nil
	}
}

func statusNetworkInterfaceV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNetworkInterfaceByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNetworkInterfaceAttachmentV2(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNetworkInterfaceAttachmentByIDV2(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func StatusVPNConnectionState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNConnectionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusVPNConnectionRouteState(ctx context.Context, conn *ec2.Client, vpnConnectionID, cidrBlock string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNConnectionRouteByVPNConnectionIDAndCIDR(ctx, conn, vpnConnectionID, cidrBlock)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusVPNGatewayState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusVPNGatewayVPCAttachmentState(ctx context.Context, conn *ec2.Client, vpnGatewayID, vpcID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPNGatewayVPCAttachment(ctx, conn, vpnGatewayID, vpcID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func StatusCustomerGatewayState(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCustomerGatewayByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}
