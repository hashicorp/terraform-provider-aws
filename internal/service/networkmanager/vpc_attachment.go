// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_vpc_attachment", name="VPC Attachment")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networkmanager/types;awstypes;awstypes.VpcAttachment")
// @Testing(skipEmptyTags=true)
// @Testing(generator=false)
func resourceVPCAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCAttachmentCreate,
		ReadWithoutTimeout:   resourceVPCAttachmentRead,
		UpdateWithoutTimeout: resourceVPCAttachmentUpdate,
		DeleteWithoutTimeout: resourceVPCAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.All(
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if d.Id() == "" {
					return nil
				}

				if !d.HasChange("options.0.appliance_mode_support") {
					return nil
				}

				if state := awstypes.AttachmentState(d.Get(names.AttrState).(string)); state == awstypes.AttachmentStatePendingAttachmentAcceptance {
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

				if state := awstypes.AttachmentState(d.Get(names.AttrState).(string)); state == awstypes.AttachmentStatePendingAttachmentAcceptance {
					return d.ForceNew("options.0.ipv6_support")
				}
				return nil
			},
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if d.Id() == "" {
					return nil
				}

				if !d.HasChange("options.0.dns_support") {
					return nil
				}

				if state := awstypes.AttachmentState(d.Get(names.AttrState).(string)); state == awstypes.AttachmentStatePendingAttachmentAcceptance {
					return d.ForceNew("options.0.dns_support")
				}
				return nil
			},
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if d.Id() == "" {
					return nil
				}

				if !d.HasChange("options.0.security_group_referencing_support") {
					return nil
				}

				if state := awstypes.AttachmentState(d.Get(names.AttrState).(string)); state == awstypes.AttachmentStatePendingAttachmentAcceptance {
					return d.ForceNew("options.0.security_group_referencing_support")
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
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"appliance_mode_support": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"dns_support": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"ipv6_support": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"security_group_referencing_support": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
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
			"routing_policy_label": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
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

func resourceVPCAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	coreNetworkID := d.Get("core_network_id").(string)
	vpcARN := d.Get("vpc_arn").(string)
	input := networkmanager.CreateVpcAttachmentInput{
		CoreNetworkId: aws.String(coreNetworkID),
		SubnetArns:    flex.ExpandStringValueSet(d.Get("subnet_arns").(*schema.Set)),
		Tags:          getTagsIn(ctx),
		VpcArn:        aws.String(vpcARN),
	}

	if v, ok := d.GetOk("options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Options = expandVpcOptions(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("routing_policy_label"); ok {
		input.RoutingPolicyLabel = aws.String(v.(string))
	}

	output, err := conn.CreateVpcAttachment(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager VPC (%s) Attachment (%s): %s", vpcARN, coreNetworkID, err)
	}

	d.SetId(aws.ToString(output.VpcAttachment.Attachment.AttachmentId))

	if _, err := waitVPCAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCAttachmentRead(ctx, d, meta)...)
}

func resourceVPCAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	vpcAttachment, err := findVPCAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	attachment := vpcAttachment.Attachment
	d.Set(names.AttrARN, attachmentARN(ctx, meta.(*conns.AWSClient), d.Id()))
	d.Set("attachment_policy_rule_number", attachment.AttachmentPolicyRuleNumber)
	d.Set("attachment_type", attachment.AttachmentType)
	d.Set("core_network_arn", attachment.CoreNetworkArn)
	coreNetworkID := aws.ToString(attachment.CoreNetworkId)
	d.Set("core_network_id", coreNetworkID)
	d.Set("edge_location", attachment.EdgeLocation)
	if vpcAttachment.Options != nil {
		if err := d.Set("options", []any{flattenVpcOptions(vpcAttachment.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
		}
	} else {
		d.Set("options", nil)
	}
	d.Set(names.AttrOwnerAccountID, attachment.OwnerAccountId)
	d.Set(names.AttrResourceARN, attachment.ResourceArn)
	if routingPolicyLabel, err := findAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx, conn, coreNetworkID, d.Id()); err != nil && !retry.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s) routing policy label: %s", d.Id(), err)
	} else {
		d.Set("routing_policy_label", routingPolicyLabel)
	}
	d.Set("segment_name", attachment.SegmentName)
	d.Set(names.AttrState, attachment.State)
	d.Set("subnet_arns", vpcAttachment.SubnetArns)
	d.Set("vpc_arn", attachment.ResourceArn)

	setTagsOut(ctx, attachment.Tags)

	return diags
}

func resourceVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "routing_policy_label") {
		input := networkmanager.UpdateVpcAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		}

		if d.HasChange("options") {
			if v, ok := d.GetOk("options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.Options = expandVpcOptions(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange("subnet_arns") {
			o, n := d.GetChange("subnet_arns")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			if add := ns.Difference(os); len(add.List()) > 0 {
				input.AddSubnetArns = flex.ExpandStringValueSet(add)
			}

			if del := os.Difference(ns); len(del.List()) > 0 {
				input.RemoveSubnetArns = flex.ExpandStringValueSet(del)
			}
		}

		_, err := conn.UpdateVpcAttachment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}
	}

	// An update (via transparent tagging) to tags can put the attachment into PENDING_NETWORK_UPDATE state.
	if _, err := waitVPCAttachmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) update: %s", d.Id(), err)
	}

	if d.HasChange("routing_policy_label") {
		if v, ok := d.GetOk("routing_policy_label"); ok {
			// Set or update routing policy label
			input := networkmanager.PutAttachmentRoutingPolicyLabelInput{
				AttachmentId:       aws.String(d.Id()),
				CoreNetworkId:      aws.String(d.Get("core_network_id").(string)),
				RoutingPolicyLabel: aws.String(v.(string)),
			}
			_, err := conn.PutAttachmentRoutingPolicyLabel(ctx, &input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Network Manager VPC Attachment (%s) routing policy label: %s", d.Id(), err)
			}
		} else {
			// Remove routing policy label
			input := networkmanager.RemoveAttachmentRoutingPolicyLabelInput{
				AttachmentId:  aws.String(d.Id()),
				CoreNetworkId: aws.String(d.Get("core_network_id").(string)),
			}
			_, err := conn.RemoveAttachmentRoutingPolicyLabel(ctx, &input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Network Manager VPC Attachment (%s) routing policy label: %s", d.Id(), err)
			}
		}
		if _, err := waitVPCAttachmentUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) routing policy label update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceVPCAttachmentRead(ctx, d, meta)...)
}

func resourceVPCAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	// If ResourceAttachmentAccepter is used, then VPC Attachment state
	// is not updated from StatePendingAttachmentAcceptance and the delete fails if deleted immediately after create
	output, err := findVPCAttachmentByID(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrState, output.Attachment.State)

	if state := awstypes.AttachmentState(d.Get(names.AttrState).(string)); state == awstypes.AttachmentStatePendingAttachmentAcceptance || state == awstypes.AttachmentStatePendingTagAcceptance {
		input := networkmanager.RejectAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		}

		_, err := conn.RejectAttachment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "rejecting Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCAttachmenRejected(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) reject: %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting Network Manager VPC Attachment: %s", d.Id())
	const (
		// Match at least the default value of aws_networkmanager_connect_attachment's Delete timeout.
		timeout = 10 * time.Minute
	)
	input := networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	}
	_, err = tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ValidationException](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.DeleteAttachment(ctx, &input)
	}, "cannot be deleted due to existing Connect attachment")

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findVPCAttachmentByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.VpcAttachment, error) {
	input := networkmanager.GetVpcAttachmentInput{
		AttachmentId: aws.String(id),
	}

	return findVPCAttachment(ctx, conn, &input)
}

func findVPCAttachment(ctx context.Context, conn *networkmanager.Client, input *networkmanager.GetVpcAttachmentInput) (*awstypes.VpcAttachment, error) {
	output, err := conn.GetVpcAttachment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcAttachment == nil || output.VpcAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.VpcAttachment, nil
}

func findAttachmentRoutingPolicyAssociationLabelByTwoPartKey(ctx context.Context, conn *networkmanager.Client, coreNetworkID, attachmentID string) (*string, error) {
	input := networkmanager.ListAttachmentRoutingPolicyAssociationsInput{
		AttachmentId:  aws.String(attachmentID),
		CoreNetworkId: aws.String(coreNetworkID),
	}
	output, err := findAttachmentRoutingPolicyAssociation(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	return output.RoutingPolicyLabel, nil
}

func findAttachmentRoutingPolicyAssociation(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentRoutingPolicyAssociationsInput) (*awstypes.AttachmentRoutingPolicyAssociationSummary, error) {
	output, err := findAttachmentRoutingPolicyAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAttachmentRoutingPolicyAssociations(ctx context.Context, conn *networkmanager.Client, input *networkmanager.ListAttachmentRoutingPolicyAssociationsInput) ([]awstypes.AttachmentRoutingPolicyAssociationSummary, error) {
	var output []awstypes.AttachmentRoutingPolicyAssociationSummary

	pages := networkmanager.NewListAttachmentRoutingPolicyAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.AttachmentRoutingPolicyAssociations...)
	}

	return output, nil
}

func statusVPCAttachment(conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findVPCAttachmentByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Attachment.State), nil
	}
}

func waitVPCAttachmentCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:                    enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Timeout:                   timeout,
		Refresh:                   statusVPCAttachment(conn, id),
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitVPCAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Timeout: timeout,
		Refresh: statusVPCAttachment(conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitVPCAttachmenRejected(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.VpcAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatePendingAttachmentAcceptance, awstypes.AttachmentStateAvailable),
		Target:  enum.Slice(awstypes.AttachmentStateRejected),
		Timeout: timeout,
		Refresh: statusVPCAttachment(conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitVPCAttachmentDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.VpcAttachment, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.AttachmentStateDeleting, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusVPCAttachment(conn, id),
		Delay:          2 * time.Minute,
		PollInterval:   10 * time.Second,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func waitVPCAttachmentUpdated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.VpcAttachment, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStatePendingNetworkUpdate, awstypes.AttachmentStateUpdating),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingTagAcceptance),
		Timeout: timeout,
		Refresh: statusVPCAttachment(conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcAttachment); ok {
		retry.SetLastError(err, attachmentsError(output.Attachment.LastModificationErrors))

		return output, err
	}

	return nil, err
}

func expandVpcOptions(tfMap map[string]any) *awstypes.VpcOptions { // nosemgrep:ci.caps5-in-func-name
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VpcOptions{}

	if v, ok := tfMap["appliance_mode_support"].(bool); ok {
		apiObject.ApplianceModeSupport = aws.Bool(v)
	}

	if v, ok := tfMap["dns_support"].(bool); ok {
		apiObject.DnsSupport = aws.Bool(v)
	}

	if v, ok := tfMap["ipv6_support"].(bool); ok {
		apiObject.Ipv6Support = aws.Bool(v)
	}

	if v, ok := tfMap["security_group_referencing_support"].(bool); ok {
		apiObject.SecurityGroupReferencingSupport = aws.Bool(v)
	}

	return apiObject
}

func flattenVpcOptions(apiObject *awstypes.VpcOptions) map[string]any { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"appliance_mode_support":             aws.ToBool(apiObject.ApplianceModeSupport),
		"dns_support":                        aws.ToBool(apiObject.DnsSupport),
		"ipv6_support":                       aws.ToBool(apiObject.Ipv6Support),
		"security_group_referencing_support": aws.ToBool(apiObject.SecurityGroupReferencingSupport),
	}

	return tfMap
}
