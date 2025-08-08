// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var (
	instanceAttributeMap = map[string]awstypes.InstanceAttributeType{
		"auto_resolve_best_voices_enabled": awstypes.InstanceAttributeTypeAutoResolveBestVoices,
		"contact_flow_logs_enabled":        awstypes.InstanceAttributeTypeContactflowLogs,
		"contact_lens_enabled":             awstypes.InstanceAttributeTypeContactLens,
		"early_media_enabled":              awstypes.InstanceAttributeTypeEarlyMedia,
		"inbound_calls_enabled":            awstypes.InstanceAttributeTypeInboundCalls,
		"multi_party_conference_enabled":   awstypes.InstanceAttributeTypeMultiPartyConference,
		"outbound_calls_enabled":           awstypes.InstanceAttributeTypeOutboundCalls,
		// Pre-release feature requiring allow-list from AWS. Removing all functionality until feature is GA
		//"use_custom_tts_voices_enabled": awstypes.InstanceAttributeTypeUseCustomTtsVoices,
	}
)

// @SDKResource("aws_connect_instance", name="Instance")
// @Tags(identifierAttribute="arn")
func resourceInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceCreate,
		ReadWithoutTimeout:   resourceInstanceRead,
		UpdateWithoutTimeout: resourceInstanceUpdate,
		DeleteWithoutTimeout: resourceInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_resolve_best_voices_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"contact_flow_logs_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false, //verified default result from ListInstanceAttributes()
			},
			"contact_lens_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(12, 12),
				AtLeastOneOf: []string{"directory_id", "instance_alias"},
			},
			"early_media_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true, //verified default result from ListInstanceAttributes()
			},
			"identity_management_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.DirectoryType](),
			},
			"inbound_calls_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"instance_alias": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				AtLeastOneOf: []string{"directory_id", "instance_alias"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^([0-9A-Za-z]+)([0-9A-Za-z-]+)$`), "must contain only alphanumeric or hyphen characters"),
					validation.StringDoesNotMatch(regexache.MustCompile(`^(d-).+$`), "can not start with d-"),
				),
			},
			"multi_party_conference_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false, //verified default result from ListInstanceAttributes()
			},
			"outbound_calls_enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			names.AttrServiceRole: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Pre-release feature requiring allow-list from AWS. Removing all functionality until feature is GA
			// "use_custom_tts_voices_enabled": {
			// 	Type:     schema.TypeBool,
			// 	Optional: true,
			// 	Default:  false, //verified default result from ListInstanceAttributes()
			// },
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	input := &connect.CreateInstanceInput{
		ClientToken:            aws.String(id.UniqueId()),
		IdentityManagementType: awstypes.DirectoryType(d.Get("identity_management_type").(string)),
		InboundCallsEnabled:    aws.Bool(d.Get("inbound_calls_enabled").(bool)),
		OutboundCallsEnabled:   aws.Bool(d.Get("outbound_calls_enabled").(bool)),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("directory_id"); ok {
		input.DirectoryId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_alias"); ok {
		input.InstanceAlias = aws.String(v.(string))
	}

	output, err := conn.CreateInstance(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Instance: %s", err)
	}

	d.SetId(aws.ToString(output.Id))

	if _, err := waitInstanceCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Connect Instance (%s) create: %s", d.Id(), err)
	}

	if err := createInstanceAttributes(ctx, conn, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instance, err := findInstanceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Instance (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, instance.Arn)
	if instance.CreatedTime != nil {
		d.Set(names.AttrCreatedTime, instance.CreatedTime.Format(time.RFC3339))
	}
	d.Set("identity_management_type", instance.IdentityManagementType)
	d.Set("inbound_calls_enabled", instance.InboundCallsEnabled)
	d.Set("instance_alias", instance.InstanceAlias)
	d.Set("outbound_calls_enabled", instance.OutboundCallsEnabled)
	d.Set(names.AttrServiceRole, instance.ServiceRole)
	d.Set(names.AttrStatus, instance.InstanceStatus)

	if err := readInstanceAttributes(ctx, conn, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	setTagsOut(ctx, instance.Tags)

	return diags
}

func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	if err := updateInstanceAttributes(ctx, conn, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return append(diags, resourceInstanceRead(ctx, d, meta)...)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	log.Printf("[DEBUG] Deleting Connect Instance: %s", d.Id())
	input := connect.DeleteInstanceInput{
		InstanceId: aws.String(d.Id()),
	}
	_, err := conn.DeleteInstance(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Connect Instance (%s): %s", d.Id(), err)
	}

	if _, err := waitInstanceDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Connect Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func createInstanceAttributes(ctx context.Context, conn *connect.Client, d *schema.ResourceData) error {
	for key, attributeType := range instanceAttributeMap {
		if err := updateInstanceAttribute(ctx, conn, d.Id(), attributeType, strconv.FormatBool(d.Get(key).(bool))); err != nil {
			return err
		}
	}

	return nil
}

func readInstanceAttributes(ctx context.Context, conn *connect.Client, d *schema.ResourceData) error {
	for key, attributeType := range instanceAttributeMap {
		attribute, err := findInstanceAttributeByTwoPartKey(ctx, conn, d.Id(), attributeType)

		if err != nil {
			return fmt.Errorf("reading Connect Instance (%s) attribute (%s): %w", d.Id(), attributeType, err)
		}

		v, err := strconv.ParseBool(aws.ToString(attribute.Value))
		if err != nil {
			return err
		}

		d.Set(key, v)
	}

	return nil
}

func updateInstanceAttributes(ctx context.Context, conn *connect.Client, d *schema.ResourceData) error {
	for key, attributeType := range instanceAttributeMap {
		if !d.HasChange(key) {
			continue
		}

		if err := updateInstanceAttribute(ctx, conn, d.Id(), attributeType, strconv.FormatBool(d.Get(key).(bool))); err != nil {
			return err
		}
	}

	return nil
}

func updateInstanceAttribute(ctx context.Context, conn *connect.Client, instanceID string, attributeType awstypes.InstanceAttributeType, value string) error {
	input := &connect.UpdateInstanceAttributeInput{
		AttributeType: attributeType,
		InstanceId:    aws.String(instanceID),
		Value:         aws.String(value),
	}

	_, err := conn.UpdateInstanceAttribute(ctx, input)

	if errs.IsA[*awstypes.AccessDeniedException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("updating Connect Instance (%s) attribute (%s): %w", instanceID, attributeType, err)
	}

	return nil
}

func findInstanceByID(ctx context.Context, conn *connect.Client, id string) (*awstypes.Instance, error) {
	input := &connect.DescribeInstanceInput{
		InstanceId: aws.String(id),
	}

	return findInstance(ctx, conn, input)
}

func findInstance(ctx context.Context, conn *connect.Client, input *connect.DescribeInstanceInput) (*awstypes.Instance, error) {
	output, err := conn.DescribeInstance(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Instance == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Instance, nil
}

func findInstanceAttributeByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID string, attributeType awstypes.InstanceAttributeType) (*awstypes.Attribute, error) {
	input := &connect.DescribeInstanceAttributeInput{
		AttributeType: attributeType,
		InstanceId:    aws.String(instanceID),
	}

	return findInstanceAttribute(ctx, conn, input)
}

func findInstanceAttribute(ctx context.Context, conn *connect.Client, input *connect.DescribeInstanceAttributeInput) (*awstypes.Attribute, error) {
	output, err := conn.DescribeInstanceAttribute(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Attribute == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Attribute, nil
}

func statusInstance(ctx context.Context, conn *connect.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.InstanceStatus), nil
	}
}

func waitInstanceCreated(ctx context.Context, conn *connect.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InstanceStatusCreationInProgress),
		Target:  enum.Slice(awstypes.InstanceStatusActive),
		Refresh: statusInstance(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if statusReason := output.StatusReason; statusReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(statusReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitInstanceDeleted(ctx context.Context, conn *connect.Client, id string, timeout time.Duration) (*awstypes.Instance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.InstanceStatusActive),
		Target:  []string{},
		Refresh: statusInstance(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Instance); ok {
		if statusReason := output.StatusReason; statusReason != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(statusReason.Message)))
		}

		return output, err
	}

	return nil, err
}
