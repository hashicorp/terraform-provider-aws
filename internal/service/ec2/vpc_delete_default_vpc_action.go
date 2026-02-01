// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// deleteDefaultVPCPollInterval defines polling cadence for VPC deletion.
const deleteDefaultVPCPollInterval = 10 * time.Second

// @Action(aws_vpc_delete_default_vpc, name="Delete Default VPC")
func newDeleteDefaultVPCAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &deleteDefaultVPCAction{}, nil
}

var (
	_ action.Action = (*deleteDefaultVPCAction)(nil)
)

type deleteDefaultVPCAction struct {
	framework.ActionWithModel[deleteDefaultVPCActionModel]
}

type deleteDefaultVPCActionModel struct {
	framework.WithRegionModel
	Timeout types.Int64 `tfsdk:"timeout"`
}

func (a *deleteDefaultVPCAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deletes the default VPC in the current region. This action removes the default VPC and all its dependencies including internet gateways, subnets, route tables, network ACLs, and security groups.",
		Attributes: map[string]schema.Attribute{
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for VPC deletion to complete (default: 600, min: 60, max: 3600)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(60),
					int64validator.AtMost(3600),
				},
			},
		},
	}
}

func (a *deleteDefaultVPCAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config deleteDefaultVPCActionModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default timeout
	timeout := 600 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting delete default VPC action", map[string]any{
		names.AttrTimeout: timeout.String(),
	})

	// Get EC2 client for current region
	conn := a.Meta().EC2Client(ctx)

	resp.SendProgress(action.InvokeProgressEvent{
		Message: "Finding default VPC...",
	})

	// Find default VPC
	vpcID, err := a.findDefaultVPC(ctx, conn)
	if err != nil {
		if !retry.NotFound(err) {
			resp.Diagnostics.AddError(
				"Failed to Find Default VPC",
				fmt.Sprintf("Could not find default VPC: %s", err),
			)
			return
		}

		resp.SendProgress(action.InvokeProgressEvent{
			Message: "No default VPC found in this region",
		})
		tflog.Info(ctx, "No default VPC found, nothing to delete")
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Found default VPC %s", vpcID),
	})

	// Delete VPC and dependencies
	if err := a.deleteDefaultVPC(ctx, conn, vpcID, timeout, resp); err != nil {
		resp.Diagnostics.AddError(
			"Failed to Delete Default VPC",
			fmt.Sprintf("Could not delete default VPC %s: %s", vpcID, err),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Successfully deleted default VPC %s", vpcID),
	})

	tflog.Info(ctx, "Delete default VPC action completed successfully", map[string]any{
		names.AttrVPCID: vpcID,
	})
}

// findDefaultVPC finds the default VPC in the current region
func (a *deleteDefaultVPCAction) findDefaultVPC(ctx context.Context, conn *ec2.Client) (string, error) {
	input := &ec2.DescribeVpcsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"isDefault": "true",
			},
		),
	}

	vpc, err := findVPC(ctx, conn, input)
	if err != nil {
		return "", err
	}

	return aws.ToString(vpc.VpcId), nil
}

// deleteDefaultVPCInRegion deletes the default VPC and all its dependencies
func (a *deleteDefaultVPCAction) deleteDefaultVPC(ctx context.Context, conn *ec2.Client, vpcID string, timeout time.Duration, resp *action.InvokeResponse) error {
	tflog.Info(ctx, "Deleting default VPC", map[string]any{
		names.AttrVPCID: vpcID,
	})

	// Delete dependencies
	progressFn := func(msg string) {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: msg,
		})
	}

	progressFn(fmt.Sprintf("Deleting dependencies for VPC %s...", vpcID))

	if err := deleteDefaultVPCDependencies(ctx, conn, vpcID, progressFn); err != nil {
		return fmt.Errorf("deleting VPC dependencies: %w", err)
	}

	// Delete the VPC
	progressFn(fmt.Sprintf("Deleting VPC %s...", vpcID))

	input := &ec2.DeleteVpcInput{
		VpcId: aws.String(vpcID),
	}

	_, err := conn.DeleteVpc(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCIDNotFound) {
			// VPC already deleted
			return nil
		}
		return fmt.Errorf("deleting VPC %s: %w", vpcID, err)
	}

	// Wait for VPC to be deleted
	progressFn(fmt.Sprintf("Waiting for VPC %s to be deleted...", vpcID))

	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[struct{}], error) {
		_, derr := findVPCByID(ctx, conn, vpcID)
		if derr != nil {
			if tfawserr.ErrCodeEquals(derr, errCodeInvalidVPCIDNotFound) || retry.NotFound(derr) {
				// VPC has been deleted
				return actionwait.FetchResult[struct{}]{
					Status: actionwait.Status("deleted"),
				}, nil
			}
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("describing VPC: %w", derr)
		}
		// VPC still exists, in deleting state
		return actionwait.FetchResult[struct{}]{
			Status: actionwait.Status("deleting"),
		}, nil
	}, actionwait.Options[struct{}]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(deleteDefaultVPCPollInterval),
		ProgressInterval: 30 * time.Second,
		SuccessStates:    []actionwait.Status{actionwait.Status("deleted")},
		TransitionalStates: []actionwait.Status{
			actionwait.Status("deleting"),
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			progressFn(fmt.Sprintf("VPC %s is being deleted, continuing to wait...", vpcID))
		},
	})
	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		if errors.As(err, &timeoutErr) {
			return fmt.Errorf("timeout waiting for VPC %s to be deleted: %w", vpcID, err)
		}
		return fmt.Errorf("waiting for VPC %s to be deleted: %w", vpcID, err)
	}

	progressFn(fmt.Sprintf("VPC %s has been successfully deleted", vpcID))

	return nil
}

// deleteDefaultVPCDependencies deletes all dependencies of a VPC in the correct order
func deleteDefaultVPCDependencies(ctx context.Context, conn *ec2.Client, vpcID string, progressFn func(string)) error {
	// 1. Delete Internet Gateways
	progressFn("Deleting internet gateways...")
	if err := deleteVPCInternetGateways(ctx, conn, vpcID, progressFn); err != nil {
		return fmt.Errorf("deleting internet gateways: %w", err)
	}

	// 2. Delete Subnets
	progressFn("Deleting subnets...")
	if err := deleteVPCSubnets(ctx, conn, vpcID, progressFn); err != nil {
		return fmt.Errorf("deleting subnets: %w", err)
	}

	// 3. Delete Route Tables (non-main)
	progressFn("Deleting route tables...")
	if err := deleteVPCRouteTables(ctx, conn, vpcID, progressFn); err != nil {
		return fmt.Errorf("deleting route tables: %w", err)
	}

	// 4. Delete Security Groups (non-default)
	progressFn("Deleting security groups...")
	if err := deleteVPCSecurityGroups(ctx, conn, vpcID, progressFn); err != nil {
		return fmt.Errorf("deleting security groups: %w", err)
	}

	// 5. Delete Network ACLs (non-default)
	progressFn("Deleting network ACLs...")
	if err := deleteVPCNetworkACLs(ctx, conn, vpcID, progressFn); err != nil {
		return fmt.Errorf("deleting network ACLs: %w", err)
	}

	return nil
}

// deleteVPCInternetGateways detaches and deletes all internet gateways attached to the VPC
func deleteVPCInternetGateways(ctx context.Context, conn *ec2.Client, vpcID string, progressFn func(string)) error {
	input := &ec2.DescribeInternetGatewaysInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"attachment.vpc-id": vpcID,
			},
		),
	}

	output, err := conn.DescribeInternetGateways(ctx, input)
	if err != nil {
		return fmt.Errorf("describing internet gateways: %w", err)
	}

	for _, igw := range output.InternetGateways {
		igwID := aws.ToString(igw.InternetGatewayId)

		// Detach from VPC
		detachInput := ec2.DetachInternetGatewayInput{
			InternetGatewayId: aws.String(igwID),
			VpcId:             aws.String(vpcID),
		}
		_, err := conn.DetachInternetGateway(ctx, &detachInput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
				continue
			}
			return fmt.Errorf("detaching internet gateway %s: %w", igwID, err)
		}

		// Delete IGW
		deleteInput := ec2.DeleteInternetGatewayInput{
			InternetGatewayId: aws.String(igwID),
		}
		_, err = conn.DeleteInternetGateway(ctx, &deleteInput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
				continue
			}
			return fmt.Errorf("deleting internet gateway %s: %w", igwID, err)
		} else {
			progressFn(fmt.Sprintf("deleted internet gateway %s", igwID))
		}
	}

	return nil
}

// deleteVPCSubnets deletes all subnets in the VPC
func deleteVPCSubnets(ctx context.Context, conn *ec2.Client, vpcID string, progressFn func(string)) error {
	input := &ec2.DescribeSubnetsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"vpc-id": vpcID,
			},
		),
	}

	output, err := conn.DescribeSubnets(ctx, input)
	if err != nil {
		return fmt.Errorf("describing subnets: %w", err)
	}

	for _, subnet := range output.Subnets {
		subnetID := aws.ToString(subnet.SubnetId)

		input := ec2.DeleteSubnetInput{
			SubnetId: aws.String(subnetID),
		}
		_, err := conn.DeleteSubnet(ctx, &input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidSubnetIDNotFound) {
				continue
			}
			return fmt.Errorf("deleting subnet %s: %w", subnetID, err)
		} else {
			progressFn(fmt.Sprintf("deleted subnet %s", subnetID))
		}
	}

	return nil
}

// deleteVPCRouteTables deletes all non-main route tables in the VPC
func deleteVPCRouteTables(ctx context.Context, conn *ec2.Client, vpcID string, progressFn func(string)) error {
	input := &ec2.DescribeRouteTablesInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"vpc-id": vpcID,
			},
		),
	}

	output, err := conn.DescribeRouteTables(ctx, input)
	if err != nil {
		return fmt.Errorf("describing route tables: %w", err)
	}

	for _, rt := range output.RouteTables {
		// Skip main route table (deleted with VPC)
		isMain := false
		for _, assoc := range rt.Associations {
			if aws.ToBool(assoc.Main) {
				isMain = true
				break
			}
		}
		if isMain {
			continue
		}

		rtID := aws.ToString(rt.RouteTableId)

		// Disassociate from subnets first
		for _, assoc := range rt.Associations {
			if assoc.RouteTableAssociationId != nil && !aws.ToBool(assoc.Main) {
				disassociateInput := ec2.DisassociateRouteTableInput{
					AssociationId: assoc.RouteTableAssociationId,
				}
				_, err := conn.DisassociateRouteTable(ctx, &disassociateInput)
				if err != nil {
					if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
						continue
					}
					return fmt.Errorf("disassociating route table %s: %w", rtID, err)
				}
			}
		}

		// Delete route table
		deleteInput := ec2.DeleteRouteTableInput{
			RouteTableId: aws.String(rtID),
		}
		_, err := conn.DeleteRouteTable(ctx, &deleteInput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
				continue
			}
			return fmt.Errorf("deleting route table %s: %w", rtID, err)
		} else {
			progressFn(fmt.Sprintf("deleted route table %s", rtID))
		}
	}

	return nil
}

// deleteVPCSecurityGroups deletes all non-default security groups in the VPC
func deleteVPCSecurityGroups(ctx context.Context, conn *ec2.Client, vpcID string, progressFn func(string)) error {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"vpc-id": vpcID,
			},
		),
	}

	output, err := conn.DescribeSecurityGroups(ctx, input)
	if err != nil {
		return fmt.Errorf("describing security groups: %w", err)
	}

	for _, sg := range output.SecurityGroups {
		// Skip default security group (deleted with VPC)
		if aws.ToString(sg.GroupName) == "default" {
			continue
		}

		sgID := aws.ToString(sg.GroupId)

		// Revoke all ingress rules
		if len(sg.IpPermissions) > 0 {
			ingressInput := ec2.RevokeSecurityGroupIngressInput{
				GroupId:       aws.String(sgID),
				IpPermissions: sg.IpPermissions,
			}
			_, err := conn.RevokeSecurityGroupIngress(ctx, &ingressInput)
			if err != nil {
				if !tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound) {
					return fmt.Errorf("revoking ingress rules for security group %s: %w", sgID, err)
				}
			}
		}

		// Revoke all egress rules
		if len(sg.IpPermissionsEgress) > 0 {
			egressInput := ec2.RevokeSecurityGroupEgressInput{
				GroupId:       aws.String(sgID),
				IpPermissions: sg.IpPermissionsEgress,
			}
			_, err := conn.RevokeSecurityGroupEgress(ctx, &egressInput)
			if err != nil {
				if !tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound) {
					return fmt.Errorf("revoking egress rules for security group %s: %w", sgID, err)
				}
			}
		}

		// Delete security group
		deleteInput := ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(sgID),
		}
		_, err := conn.DeleteSecurityGroup(ctx, &deleteInput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidGroupNotFound) {
				continue
			}
			return fmt.Errorf("deleting security group %s: %w", sgID, err)
		} else {
			progressFn(fmt.Sprintf("deleted security group %s", sgID))
		}
	}

	return nil
}

// deleteVPCNetworkACLs deletes all non-default network ACLs in the VPC
func deleteVPCNetworkACLs(ctx context.Context, conn *ec2.Client, vpcID string, progressFn func(string)) error {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(
			map[string]string{
				"vpc-id": vpcID,
			},
		),
	}

	output, err := conn.DescribeNetworkAcls(ctx, input)
	if err != nil {
		return fmt.Errorf("describing network ACLs: %w", err)
	}

	for _, nacl := range output.NetworkAcls {
		// Skip default network ACL (deleted with VPC)
		if aws.ToBool(nacl.IsDefault) {
			continue
		}

		naclID := aws.ToString(nacl.NetworkAclId)

		// Delete network ACL
		input := ec2.DeleteNetworkAclInput{
			NetworkAclId: aws.String(naclID),
		}
		_, err := conn.DeleteNetworkAcl(ctx, &input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkACLIDNotFound) {
				continue
			}
			return fmt.Errorf("deleting network ACL %s: %w", naclID, err)
		} else {
			progressFn(fmt.Sprintf("deleted network ACL %s", naclID))
		}
	}

	return nil
}
