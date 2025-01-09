// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_endpoint_security_group_association", name="VPC Endpoint Security Group Association")
func resourceVPCEndpointSecurityGroupAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointSecurityGroupAssociationCreate,
		ReadWithoutTimeout:   resourceVPCEndpointSecurityGroupAssociationRead,
		DeleteWithoutTimeout: resourceVPCEndpointSecurityGroupAssociationDelete,

		Schema: map[string]*schema.Schema{
			"replace_default_association": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVPCEndpointID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPCEndpointSecurityGroupAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcEndpointID := d.Get(names.AttrVPCEndpointID).(string)
	securityGroupID := d.Get("security_group_id").(string)
	replaceDefaultAssociation := d.Get("replace_default_association").(bool)

	defaultSecurityGroupID := ""
	if replaceDefaultAssociation {
		vpcEndpoint, err := findVPCEndpointByID(ctx, conn, vpcEndpointID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint (%s): %s", vpcEndpointID, err)
		}

		vpcID := aws.ToString(vpcEndpoint.VpcId)

		defaultSecurityGroup, err := findVPCDefaultSecurityGroup(ctx, conn, vpcID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) default Security Group: %s", vpcID, err)
		}

		defaultSecurityGroupID = aws.ToString(defaultSecurityGroup.GroupId)

		if defaultSecurityGroupID == securityGroupID {
			return sdkdiag.AppendErrorf(diags, "%s is the default Security Group for EC2 VPC (%s)", securityGroupID, vpcID)
		}

		foundDefaultAssociation := false

		for _, group := range vpcEndpoint.Groups {
			if aws.ToString(group.GroupId) == defaultSecurityGroupID {
				foundDefaultAssociation = true
				break
			}
		}

		if !foundDefaultAssociation {
			return sdkdiag.AppendErrorf(diags, "no association of default Security Group (%s) with VPC Endpoint (%s)", defaultSecurityGroupID, vpcEndpointID)
		}
	}

	err := createVPCEndpointSecurityGroupAssociation(ctx, conn, vpcEndpointID, securityGroupID)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(vpcEndpointSecurityGroupAssociationCreateID(vpcEndpointID, securityGroupID))

	if replaceDefaultAssociation {
		// Delete the existing VPC endpoint/default security group association.
		if err := deleteVPCEndpointSecurityGroupAssociation(ctx, conn, vpcEndpointID, defaultSecurityGroupID); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceVPCEndpointSecurityGroupAssociationRead(ctx, d, meta)...)
}

func resourceVPCEndpointSecurityGroupAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcEndpointID := d.Get(names.AttrVPCEndpointID).(string)
	securityGroupID := d.Get("security_group_id").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive
	id := fmt.Sprintf("%s/%s", vpcEndpointID, securityGroupID)

	err := findVPCEndpointSecurityGroupAssociationExists(ctx, conn, vpcEndpointID, securityGroupID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] VPC Endpoint Security Group Association (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading VPC Security Group Association (%s): %s", id, err)
	}

	return diags
}

func resourceVPCEndpointSecurityGroupAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcEndpointID := d.Get(names.AttrVPCEndpointID).(string)
	securityGroupID := d.Get("security_group_id").(string)
	replaceDefaultAssociation := d.Get("replace_default_association").(bool)

	if replaceDefaultAssociation {
		vpcEndpoint, err := findVPCEndpointByID(ctx, conn, vpcEndpointID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading VPC Endpoint (%s): %s", vpcEndpointID, err)
		}

		vpcID := aws.ToString(vpcEndpoint.VpcId)

		defaultSecurityGroup, err := findVPCDefaultSecurityGroup(ctx, conn, vpcID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) default Security Group: %s", vpcID, err)
		}

		// Add back the VPC endpoint/default security group association.
		err = createVPCEndpointSecurityGroupAssociation(ctx, conn, vpcEndpointID, aws.ToString(defaultSecurityGroup.GroupId))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if err := deleteVPCEndpointSecurityGroupAssociation(ctx, conn, vpcEndpointID, securityGroupID); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	return diags
}

// createVPCEndpointSecurityGroupAssociation creates the specified VPC endpoint/security group association.
func createVPCEndpointSecurityGroupAssociation(ctx context.Context, conn *ec2.Client, vpcEndpointID, securityGroupID string) error {
	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:       aws.String(vpcEndpointID),
		AddSecurityGroupIds: []string{securityGroupID},
	}

	log.Printf("[DEBUG] Creating VPC Endpoint Security Group Association: %v", input)
	_, err := conn.ModifyVpcEndpoint(ctx, input)

	if err != nil {
		return fmt.Errorf("creating VPC Endpoint (%s) Security Group (%s) Association: %w", vpcEndpointID, securityGroupID, err)
	}

	return nil
}

// deleteVPCEndpointSecurityGroupAssociation deletes the specified VPC endpoint/security group association.
func deleteVPCEndpointSecurityGroupAssociation(ctx context.Context, conn *ec2.Client, vpcEndpointID, securityGroupID string) error {
	input := &ec2.ModifyVpcEndpointInput{
		VpcEndpointId:          aws.String(vpcEndpointID),
		RemoveSecurityGroupIds: []string{securityGroupID},
	}

	log.Printf("[DEBUG] Deleting VPC Endpoint Security Group Association: %v", input)
	_, err := conn.ModifyVpcEndpoint(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPCEndpointIdNotFound, errCodeInvalidGroupNotFound, errCodeInvalidParameter) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting VPC Endpoint (%s) Security Group (%s) Association: %w", vpcEndpointID, securityGroupID, err)
	}

	return nil
}
