// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_network_acl_association", name="Network ACL Association")
func resourceNetworkACLAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkACLAssociationCreate,
		ReadWithoutTimeout:   resourceNetworkACLAssociationRead,
		DeleteWithoutTimeout: resourceNetworkACLAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"network_acl_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkACLAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	associationID, err := networkACLAssociationCreate(ctx, conn, d.Get("network_acl_id").(string), d.Get(names.AttrSubnetID).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(associationID)

	return append(diags, resourceNetworkACLAssociationRead(ctx, d, meta)...)
}

func resourceNetworkACLAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findNetworkACLAssociationByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network ACL Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL Association (%s): %s", d.Id(), err)
	}

	association := outputRaw.(*awstypes.NetworkAclAssociation)

	d.Set("network_acl_id", association.NetworkAclId)
	d.Set(names.AttrSubnetID, association.SubnetId)

	return diags
}

func resourceNetworkACLAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeNetworkAclsInput{
		Filters: newAttributeFilterList(map[string]string{
			"association.association-id": d.Id(),
		}),
	}

	nacl, err := findNetworkACL(ctx, conn, input)

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL for Association (%s): %s", d.Id(), err)
	}

	vpcID := aws.ToString(nacl.VpcId)
	defaultNACL, err := findVPCDefaultNetworkACL(ctx, conn, vpcID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC (%s) default NACL: %s", vpcID, err)
	}

	if err := networkACLAssociationDelete(ctx, conn, d.Id(), aws.ToString(defaultNACL.NetworkAclId)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

// networkACLAssociationCreate creates an association between the specified NACL and subnet.
// The subnet's current association is replaced and the new association's ID is returned.
func networkACLAssociationCreate(ctx context.Context, conn *ec2.Client, naclID, subnetID string) (string, error) {
	association, err := findNetworkACLAssociationBySubnetID(ctx, conn, subnetID)

	if err != nil {
		return "", fmt.Errorf("reading EC2 Network ACL Association for EC2 Subnet (%s): %w", subnetID, err)
	}

	input := &ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: association.NetworkAclAssociationId,
		NetworkAclId:  aws.String(naclID),
	}

	log.Printf("[DEBUG] Creating EC2 Network ACL Association: %#v", input)
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return conn.ReplaceNetworkAclAssociation(ctx, input)
	}, errCodeInvalidAssociationIDNotFound)

	if err != nil {
		return "", fmt.Errorf("creating EC2 Network ACL (%s) Association: %w", naclID, err)
	}

	return aws.ToString(outputRaw.(*ec2.ReplaceNetworkAclAssociationOutput).NewAssociationId), nil
}

// networkACLAssociationsCreate creates associations between the specified NACL and subnets.
func networkACLAssociationsCreate(ctx context.Context, conn *ec2.Client, naclID string, subnetIDs []interface{}) error {
	for _, v := range subnetIDs {
		subnetID := v.(string)
		_, err := networkACLAssociationCreate(ctx, conn, naclID, subnetID)

		if tfresource.NotFound(err) {
			// Subnet has been deleted.
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// networkACLAssociationDelete deletes the specified association between a NACL and subnet.
// The subnet's current association is replaced by an association with the VPC's default NACL.
func networkACLAssociationDelete(ctx context.Context, conn *ec2.Client, associationID, naclID string) error {
	input := &ec2.ReplaceNetworkAclAssociationInput{
		AssociationId: aws.String(associationID),
		NetworkAclId:  aws.String(naclID),
	}

	log.Printf("[DEBUG] Deleting EC2 Network ACL Association: %s", associationID)
	_, err := conn.ReplaceNetworkAclAssociation(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidAssociationIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Network ACL Association (%s): %w", associationID, err)
	}

	return nil
}

// networkACLAssociationsDelete deletes the specified NACL associations for the specified subnets.
// Each subnet's current association is replaced by an association with the specified VPC's default NACL.
func networkACLAssociationsDelete(ctx context.Context, conn *ec2.Client, vpcID string, subnetIDs []interface{}) error {
	defaultNACL, err := findVPCDefaultNetworkACL(ctx, conn, vpcID)

	if err != nil {
		return fmt.Errorf("reading EC2 VPC (%s) default NACL: %w", vpcID, err)
	}

	for _, v := range subnetIDs {
		subnetID := v.(string)
		association, err := findNetworkACLAssociationBySubnetID(ctx, conn, subnetID)

		if tfresource.NotFound(err) {
			// Subnet has been deleted.
			continue
		}

		if err != nil {
			return fmt.Errorf("reading EC2 Network ACL Association for EC2 Subnet (%s): %w", subnetID, err)
		}

		if err := networkACLAssociationDelete(ctx, conn, aws.ToString(association.NetworkAclAssociationId), aws.ToString(defaultNACL.NetworkAclId)); err != nil {
			return err
		}
	}

	return nil
}
