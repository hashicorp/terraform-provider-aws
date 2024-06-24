// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_vpc_attachment", name="VPC Attachment")
// @Tags(identifierAttribute="arn")
func ResourceVPCAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCAttachmentCreate,
		ReadWithoutTimeout:   resourceVPCAttachmentRead,
		UpdateWithoutTimeout: resourceVPCAttachmentUpdate,
		DeleteWithoutTimeout: resourceVPCAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if d.Id() == "" {
					return nil
				}

				if !d.HasChange("options.0.appliance_mode_support") {
					return nil
				}

				if state := d.Get(names.AttrState).(string); state == networkmanager.AttachmentStatePendingAttachmentAcceptance {
					return d.ForceNew("options.0.appliance_mode_support")
				}
				return nil
			},
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if d.Id() == "" {
					return nil
				}

				if !d.HasChange("options.0.ipv6_support") {
					return nil
				}

				if state := d.Get(names.AttrState).(string); state == networkmanager.AttachmentStatePendingAttachmentAcceptance {
					return d.ForceNew("options.0.ipv6_support")
				}
				return nil
			},
		),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment_policy_rule_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"attachment_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"appliance_mode_support": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"ipv6_support": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_arns": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceVPCAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	coreNetworkID := d.Get("core_network_id").(string)
	vpcARN := d.Get("vpc_arn").(string)
	input := &networkmanager.CreateVpcAttachmentInput{
		CoreNetworkId: aws.String(coreNetworkID),
		SubnetArns:    flex.ExpandStringSet(d.Get("subnet_arns").(*schema.Set)),
		Tags:          getTagsIn(ctx),
		VpcArn:        aws.String(vpcARN),
	}

	if v, ok := d.GetOk("options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Options = expandVpcOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating Network Manager VPC Attachment: %s", input)
	output, err := conn.CreateVpcAttachmentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager VPC (%s) Attachment (%s): %s", vpcARN, coreNetworkID, err)
	}

	d.SetId(aws.StringValue(output.VpcAttachment.Attachment.AttachmentId))

	if _, err := waitVPCAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCAttachmentRead(ctx, d, meta)...)
}

func resourceVPCAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	a := vpcAttachment.Attachment
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "networkmanager",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("attachment/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", a.AttachmentType)
	d.Set("core_network_arn", a.CoreNetworkArn)
	d.Set("core_network_id", a.CoreNetworkId)
	d.Set("edge_location", a.EdgeLocation)
	if vpcAttachment.Options != nil {
		if err := d.Set("options", []interface{}{flattenVpcOptions(vpcAttachment.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
		}
	} else {
		d.Set("options", nil)
	}
	d.Set(names.AttrOwnerAccountID, a.OwnerAccountId)
	d.Set(names.AttrResourceARN, a.ResourceArn)
	d.Set("segment_name", a.SegmentName)
	d.Set(names.AttrState, a.State)
	d.Set("subnet_arns", aws.StringValueSlice(vpcAttachment.SubnetArns))
	d.Set("vpc_arn", a.ResourceArn)

	setTagsOut(ctx, a.Tags)

	return diags
}

func resourceVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &networkmanager.UpdateVpcAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		}

		if d.HasChange("options") {
			if v, ok := d.GetOk("options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.Options = expandVpcOptions(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("subnet_arns") {
			o, n := d.GetChange("subnet_arns")
			if o == nil {
				o = new(schema.Set)
			}
			if n == nil {
				n = new(schema.Set)
			}
			os := o.(*schema.Set)
			ns := n.(*schema.Set)

			if add := ns.Difference(os); len(add.List()) > 0 {
				input.AddSubnetArns = flex.ExpandStringSet(add)
			}

			if del := os.Difference(ns); len(del.List()) > 0 {
				input.RemoveSubnetArns = flex.ExpandStringSet(del)
			}
		}

		_, err := conn.UpdateVpcAttachmentWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCAttachmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCAttachmentRead(ctx, d, meta)...)
}

func resourceVPCAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	// If ResourceAttachmentAccepter is used, then VPC Attachment state
	// is not updated from StatePendingAttachmentAcceptance and the delete fails if deleted immediately after create
	output, sErr := FindVPCAttachmentByID(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(sErr, networkmanager.ErrCodeResourceNotFoundException) {
		return diags
	}

	if sErr != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", d.Id(), sErr)
	}

	d.Set(names.AttrState, output.Attachment.State)

	if state := d.Get(names.AttrState).(string); state == networkmanager.AttachmentStatePendingAttachmentAcceptance || state == networkmanager.AttachmentStatePendingTagAcceptance {
		_, err := conn.RejectAttachmentWithContext(ctx, &networkmanager.RejectAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "detaching Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCAttachmenRejected(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) to be detached: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Network Manager VPC Attachment: %s", d.Id())
	_, err := conn.DeleteAttachmentWithContext(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCAttachmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindVPCAttachmentByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.VpcAttachment, error) {
	input := &networkmanager.GetVpcAttachmentInput{
		AttachmentId: aws.String(id),
	}

	output, err := conn.GetVpcAttachmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcAttachment == nil || output.VpcAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VpcAttachment, nil
}

func StatusVPCAttachmentState(ctx context.Context, conn *networkmanager.NetworkManager, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindVPCAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Attachment.State), nil
	}
}

func waitVPCAttachmentCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			networkmanager.AttachmentStateCreating,
			networkmanager.AttachmentStatePendingNetworkUpdate,
		},
		Target: []string{
			networkmanager.AttachmentStateAvailable,
			networkmanager.AttachmentStatePendingAttachmentAcceptance,
		},
		Timeout: timeout,
		Refresh: StatusVPCAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPCAttachmentAvailable(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			networkmanager.AttachmentStateCreating,
			networkmanager.AttachmentStatePendingNetworkUpdate,
			networkmanager.AttachmentStatePendingAttachmentAcceptance,
		},
		Target: []string{
			networkmanager.AttachmentStateAvailable,
		},
		Timeout: timeout,
		Refresh: StatusVPCAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPCAttachmenRejected(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStatePendingAttachmentAcceptance, networkmanager.AttachmentStateAvailable},
		Target:  []string{networkmanager.AttachmentStateRejected},
		Timeout: timeout,
		Refresh: StatusVPCAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPCAttachmentDeleted(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:        []string{networkmanager.AttachmentStateDeleting},
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        StatusVPCAttachmentState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitVPCAttachmentUpdated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateUpdating},
		Target:  []string{networkmanager.AttachmentStateAvailable, networkmanager.AttachmentStatePendingTagAcceptance},
		Timeout: timeout,
		Refresh: StatusVPCAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func expandVpcOptions(tfMap map[string]interface{}) *networkmanager.VpcOptions { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &networkmanager.VpcOptions{}

	if v, ok := tfMap["appliance_mode_support"].(bool); ok {
		apiObject.ApplianceModeSupport = aws.Bool(v)
	}

	if v, ok := tfMap["ipv6_support"].(bool); ok {
		apiObject.Ipv6Support = aws.Bool(v)
	}

	return apiObject
}

func flattenVpcOptions(apiObject *networkmanager.VpcOptions) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ApplianceModeSupport; v != nil {
		tfMap["appliance_mode_support"] = aws.BoolValue(v)
	}

	if v := apiObject.Ipv6Support; v != nil {
		tfMap["ipv6_support"] = aws.BoolValue(v)
	}

	return tfMap
}
