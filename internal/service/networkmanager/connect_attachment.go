// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_connect_attachment", name="Connect Attachment")
// @Tags(identifierAttribute="arn")
func resourceConnectAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConnectAttachmentCreate,
		ReadWithoutTimeout:   resourceConnectAttachmentRead,
		UpdateWithoutTimeout: resourceConnectAttachmentUpdate,
		DeleteWithoutTimeout: resourceConnectAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachment_id": {
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
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 50),
					validation.StringMatch(regexache.MustCompile(`^core-network-([0-9a-f]{8,17})$`), "Must start with core-network and then have 8 to 17 characters"),
				),
			},
			"edge_location": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`[\s\S]*`), "Anything but whitespace"),
				),
			},
			"options": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TunnelProtocol](),
						},
					},
				},
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"transport_attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 50),
					validation.StringMatch(regexache.MustCompile(`^attachment-([0-9a-f]{8,17})$`), "Must start with attachment- and then have 8 to 17 characters"),
				),
			},
		},
	}
}

func resourceConnectAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	coreNetworkID := d.Get("core_network_id").(string)
	edgeLocation := d.Get("edge_location").(string)
	transportAttachmentID := d.Get("transport_attachment_id").(string)
	options := &awstypes.ConnectAttachmentOptions{}
	if v, ok := d.GetOk("options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		options = expandConnectOptions(v.([]interface{})[0].(map[string]interface{}))
	}

	input := &networkmanager.CreateConnectAttachmentInput{
		CoreNetworkId:         aws.String(coreNetworkID),
		EdgeLocation:          aws.String(edgeLocation),
		Options:               options,
		Tags:                  getTagsIn(ctx),
		TransportAttachmentId: aws.String(transportAttachmentID),
	}

	outputRaw, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateConnectAttachment(ctx, input)
		},
		func(err error) (bool, error) {
			// Connect attachment doesn't have direct dependency to VPC attachment state when using Attachment Accepter.
			// Waiting for Create Timeout period for VPC Attachment to come available state.
			// Only needed if depends_on statement is not used in Connect attachment.
			//
			// ValidationException: Incorrect input.
			// {
			//   RespMetadata: {
			//     StatusCode: 400,
			//     RequestID: "0a711cf7-2210-40c9-a170-a4c42134e195"
			//   },
			//   Fields: [{
			//       Message: "Transport attachment state is invalid.",
			//       Name: "transportAttachmentId"
			//     }],
			//   Message_: "Incorrect input.",
			//   Reason: "FieldValidationFailed"
			// }
			if validationExceptionFieldsMessageContains(err, awstypes.ValidationExceptionReasonFieldValidationFailed, "Transport attachment state is invalid.") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Connect Attachment (%s) (%s): %s", transportAttachmentID, coreNetworkID, err)
	}

	d.SetId(aws.ToString(outputRaw.(*networkmanager.CreateConnectAttachmentOutput).ConnectAttachment.Attachment.AttachmentId))

	if _, err := waitConnectAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectAttachmentRead(ctx, d, meta)...)
}

func resourceConnectAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	connectAttachment, err := findConnectAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Connect Attachment %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Attachment (%s): %s", d.Id(), err)
	}

	a := connectAttachment.Attachment
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "networkmanager",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("attachment/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
	d.Set("attachment_id", a.AttachmentId)
	d.Set("attachment_type", a.AttachmentType)
	d.Set("core_network_arn", a.CoreNetworkArn)
	d.Set("core_network_id", a.CoreNetworkId)
	d.Set("edge_location", a.EdgeLocation)
	if connectAttachment.Options != nil {
		if err := d.Set("options", []interface{}{flattenConnectOptions(connectAttachment.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
		}
	} else {
		d.Set("options", nil)
	}
	d.Set(names.AttrOwnerAccountID, a.OwnerAccountId)
	d.Set(names.AttrResourceARN, a.ResourceArn)
	d.Set("segment_name", a.SegmentName)
	d.Set(names.AttrState, a.State)
	d.Set("transport_attachment_id", connectAttachment.TransportAttachmentId)

	setTagsOut(ctx, a.Tags)

	return diags
}

func resourceConnectAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceConnectAttachmentRead(ctx, d, meta)
}

func resourceConnectAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	// If ResourceAttachmentAccepter is used, then Connect Attachment state
	// is never updated from StatePendingAttachmentAcceptance and the delete fails
	output, err := findConnectAttachmentByID(ctx, conn, d.Id())

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Attachment (%s): %s", d.Id(), err)
	}

	if state := output.Attachment.State; state == awstypes.AttachmentStatePendingAttachmentAcceptance || state == awstypes.AttachmentStatePendingTagAcceptance {
		return sdkdiag.AppendErrorf(diags, "cannot delete Network Manager Connect Attachment (%s) in state: %s", d.Id(), state)
	}

	log.Printf("[DEBUG] Deleting Network Manager Connect Attachment: %s", d.Id())
	_, err = conn.DeleteAttachment(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Network Manager Connect Attachment (%s): %s", d.Id(), err)
	}

	if _, err := waitConnectAttachmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findConnectAttachmentByID(ctx context.Context, conn *networkmanager.Client, id string) (*awstypes.ConnectAttachment, error) {
	input := &networkmanager.GetConnectAttachmentInput{
		AttachmentId: aws.String(id),
	}

	output, err := conn.GetConnectAttachment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ConnectAttachment == nil || output.ConnectAttachment.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ConnectAttachment, nil
}

func statusConnectAttachmentState(ctx context.Context, conn *networkmanager.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findConnectAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Attachment.State), nil
	}
}

func waitConnectAttachmentCreated(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.ConnectAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Timeout: timeout,
		Refresh: statusConnectAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ConnectAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitConnectAttachmentDeleted(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.ConnectAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        enum.Slice(awstypes.AttachmentStateDeleting),
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusConnectAttachmentState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ConnectAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitConnectAttachmentAvailable(ctx context.Context, conn *networkmanager.Client, id string, timeout time.Duration) (*awstypes.ConnectAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AttachmentStateCreating, awstypes.AttachmentStatePendingNetworkUpdate, awstypes.AttachmentStatePendingAttachmentAcceptance),
		Target:  enum.Slice(awstypes.AttachmentStateAvailable),
		Timeout: timeout,
		Refresh: statusConnectAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ConnectAttachment); ok {
		return output, err
	}

	return nil, err
}

func expandConnectOptions(o map[string]interface{}) *awstypes.ConnectAttachmentOptions {
	if o == nil {
		return nil
	}

	object := &awstypes.ConnectAttachmentOptions{}

	if v, ok := o[names.AttrProtocol].(string); ok {
		object.Protocol = awstypes.TunnelProtocol(v)
	}

	return object
}

func flattenConnectOptions(apiObject *awstypes.ConnectAttachmentOptions) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrProtocol: apiObject.Protocol,
	}

	return tfMap
}
