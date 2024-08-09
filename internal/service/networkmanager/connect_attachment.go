// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_connect_attachment", name="Connect Attachment")
// @Tags(identifierAttribute="arn")
func ResourceConnectAttachment() *schema.Resource {
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(networkmanager.TunnelProtocol_Values(), false),
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

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	coreNetworkID := d.Get("core_network_id").(string)
	edgeLocation := d.Get("edge_location").(string)
	transportAttachmentID := d.Get("transport_attachment_id").(string)
	options := &networkmanager.ConnectAttachmentOptions{}
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
			return conn.CreateConnectAttachmentWithContext(ctx, input)
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
			if validationExceptionFieldsMessageContains(err, networkmanager.ValidationExceptionReasonFieldValidationFailed, "Transport attachment state is invalid.") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Network Manager Connect Attachment (%s) (%s): %s", transportAttachmentID, coreNetworkID, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*networkmanager.CreateConnectAttachmentOutput).ConnectAttachment.Attachment.AttachmentId))

	if _, err := waitConnectAttachmentCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Attachment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceConnectAttachmentRead(ctx, d, meta)...)
}

func resourceConnectAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	connectAttachment, err := FindConnectAttachmentByID(ctx, conn, d.Id())

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

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	// If ResourceAttachmentAccepter is used, then Connect Attachment state
	// is never updated from StatePendingAttachmentAcceptance and the delete fails
	output, err := FindConnectAttachmentByID(ctx, conn, d.Id())

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Attachment (%s): %s", d.Id(), err)
	}

	if state := aws.StringValue(output.Attachment.State); state == networkmanager.AttachmentStatePendingAttachmentAcceptance || state == networkmanager.AttachmentStatePendingTagAcceptance {
		return sdkdiag.AppendErrorf(diags, "cannot delete Network Manager Connect Attachment (%s) in state: %s", d.Id(), state)
	}

	log.Printf("[DEBUG] Deleting Network Manager Connect Attachment: %s", d.Id())
	_, err = conn.DeleteAttachmentWithContext(ctx, &networkmanager.DeleteAttachmentInput{
		AttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func FindConnectAttachmentByID(ctx context.Context, conn *networkmanager.NetworkManager, id string) (*networkmanager.ConnectAttachment, error) {
	input := &networkmanager.GetConnectAttachmentInput{
		AttachmentId: aws.String(id),
	}

	output, err := conn.GetConnectAttachmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
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

func statusConnectAttachmentState(ctx context.Context, conn *networkmanager.NetworkManager, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConnectAttachmentByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Attachment.State), nil
	}
}

func waitConnectAttachmentCreated(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.ConnectAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateCreating, networkmanager.AttachmentStatePendingNetworkUpdate},
		Target:  []string{networkmanager.AttachmentStateAvailable, networkmanager.AttachmentStatePendingAttachmentAcceptance},
		Timeout: timeout,
		Refresh: statusConnectAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.ConnectAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitConnectAttachmentDeleted(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.ConnectAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{networkmanager.AttachmentStateDeleting},
		Target:         []string{},
		Timeout:        timeout,
		Refresh:        statusConnectAttachmentState(ctx, conn, id),
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.ConnectAttachment); ok {
		return output, err
	}

	return nil, err
}

func waitConnectAttachmentAvailable(ctx context.Context, conn *networkmanager.NetworkManager, id string, timeout time.Duration) (*networkmanager.ConnectAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{networkmanager.AttachmentStateCreating, networkmanager.AttachmentStatePendingNetworkUpdate, networkmanager.AttachmentStatePendingAttachmentAcceptance},
		Target:  []string{networkmanager.AttachmentStateAvailable},
		Timeout: timeout,
		Refresh: statusConnectAttachmentState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*networkmanager.ConnectAttachment); ok {
		return output, err
	}

	return nil, err
}

func expandConnectOptions(o map[string]interface{}) *networkmanager.ConnectAttachmentOptions {
	if o == nil {
		return nil
	}

	object := &networkmanager.ConnectAttachmentOptions{}

	if v, ok := o[names.AttrProtocol].(string); ok {
		object.Protocol = aws.String(v)
	}

	return object
}

func flattenConnectOptions(apiObject *networkmanager.ConnectAttachmentOptions) map[string]interface{} { // nosemgrep:ci.caps5-in-func-name
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Protocol; v != nil {
		tfMap[names.AttrProtocol] = aws.StringValue(v)
	}

	return tfMap
}
