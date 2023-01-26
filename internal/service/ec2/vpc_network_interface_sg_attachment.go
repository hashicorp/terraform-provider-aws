package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceNetworkInterfaceSGAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInterfaceSGAttachmentCreate,
		ReadWithoutTimeout:   resourceNetworkInterfaceSGAttachmentRead,
		DeleteWithoutTimeout: resourceNetworkInterfaceSGAttachmentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceNetworkInterfaceSGAttachmentImport,
		},

		Schema: map[string]*schema.Schema{
			"network_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNetworkInterfaceSGAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	networkInterfaceID := d.Get("network_interface_id").(string)
	sgID := d.Get("security_group_id").(string)
	mutexKey := "network_interface_sg_attachment_" + networkInterfaceID
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	eni, err := FindNetworkInterfaceByID(ctx, conn, networkInterfaceID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface (%s): %s", networkInterfaceID, err)
	}

	groupIDs := []string{sgID}

	for _, group := range eni.Groups {
		if group == nil {
			continue
		}

		groupID := aws.StringValue(group.GroupId)

		if groupID == sgID {
			return sdkdiag.AppendErrorf(diags, "EC2 Security Group (%s) already attached to EC2 Network Interface (%s)", sgID, networkInterfaceID)
		}

		groupIDs = append(groupIDs, groupID)
	}

	input := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: aws.String(networkInterfaceID),
		Groups:             aws.StringSlice(groupIDs),
	}

	log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
	_, err = conn.ModifyNetworkInterfaceAttributeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s): %s", networkInterfaceID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", sgID, networkInterfaceID))

	return append(diags, resourceNetworkInterfaceSGAttachmentRead(ctx, d, meta)...)
}

func resourceNetworkInterfaceSGAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	networkInterfaceID := d.Get("network_interface_id").(string)
	sgID := d.Get("security_group_id").(string)
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindNetworkInterfaceSecurityGroup(ctx, conn, networkInterfaceID, sgID)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface (%s) Security Group (%s) Attachment not found, removing from state", networkInterfaceID, sgID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface (%s) Security Group (%s) Attachment: %s", networkInterfaceID, sgID, err)
	}

	groupIdentifier := outputRaw.(*ec2.GroupIdentifier)

	d.Set("network_interface_id", networkInterfaceID)
	d.Set("security_group_id", groupIdentifier.GroupId)

	return diags
}

func resourceNetworkInterfaceSGAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	networkInterfaceID := d.Get("network_interface_id").(string)
	sgID := d.Get("security_group_id").(string)
	mutexKey := "network_interface_sg_attachment_" + networkInterfaceID
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	eni, err := FindNetworkInterfaceByID(ctx, conn, networkInterfaceID)

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface (%s): %s", networkInterfaceID, err)
	}

	groupIDs := []string{}

	for _, group := range eni.Groups {
		if group == nil {
			continue
		}

		groupID := aws.StringValue(group.GroupId)

		if groupID == sgID {
			continue
		}

		groupIDs = append(groupIDs, groupID)
	}

	input := &ec2.ModifyNetworkInterfaceAttributeInput{
		NetworkInterfaceId: aws.String(networkInterfaceID),
		Groups:             aws.StringSlice(groupIDs),
	}

	log.Printf("[INFO] Modifying EC2 Network Interface: %s", input)
	_, err = conn.ModifyNetworkInterfaceAttributeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidNetworkInterfaceIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "modifying EC2 Network Interface (%s): %s", networkInterfaceID, err)
	}

	return diags
}

func resourceNetworkInterfaceSGAttachmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "_")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Unexpected format for import: %s. Please use '<NetworkInterfaceId>_<SecurityGroupID>", d.Id())
	}

	networkInterfaceID := parts[0]
	securityGroupID := parts[1]

	log.Printf("[DEBUG] Importing network interface security group association, Interface: %s, Security Group: %s", networkInterfaceID, securityGroupID)

	conn := meta.(*conns.AWSClient).EC2Conn()

	networkInterface, err := FindNetworkInterfaceByID(ctx, conn, networkInterfaceID)

	if err != nil {
		return nil, err
	}

	var associationID string

	for _, attachedSecurityGroup := range networkInterface.Groups {
		if aws.StringValue(attachedSecurityGroup.GroupId) == securityGroupID {
			d.Set("security_group_id", securityGroupID)
			associationID = securityGroupID + "_" + networkInterfaceID

			break
		}
	}

	if associationID == "" {
		return nil, fmt.Errorf("Security Group %s is not attached to Network Interface %s", securityGroupID, networkInterfaceID)
	}

	d.SetId(associationID)
	d.Set("network_interface_id", networkInterfaceID)

	return []*schema.ResourceData{d}, nil
}
