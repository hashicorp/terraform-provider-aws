// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_medialive_input", name="Input")
// @Tags(identifierAttribute="arn")
func ResourceInput() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInputCreate,
		ReadWithoutTimeout:   resourceInputRead,
		UpdateWithoutTimeout: resourceInputUpdate,
		DeleteWithoutTimeout: resourceInputDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_channels": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"destinations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"input_class": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"input_devices": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"input_partner_ids": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"input_security_groups": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"input_source_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"media_connect_flows": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flow_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
			},
			"sources": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password_param": {
							Type:     schema.TypeString,
							Required: true,
						},
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
						"username": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.InputType](),
			},
			"vpc": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 2,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameInput = "Input"

	propagationTimeout = 2 * time.Minute
)

func resourceInputCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	in := &medialive.CreateInputInput{
		RequestId: aws.String(id.UniqueId()),
		Name:      aws.String(d.Get("name").(string)),
		Tags:      getTagsIn(ctx),
		Type:      types.InputType(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("destinations"); ok && v.(*schema.Set).Len() > 0 {
		in.Destinations = expandDestinations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("input_devices"); ok && v.(*schema.Set).Len() > 0 {
		in.InputDevices = inputDevices(v.(*schema.Set).List()).expandToDeviceSettings()
	}

	if v, ok := d.GetOk("input_security_groups"); ok && len(v.([]interface{})) > 0 {
		in.InputSecurityGroups = flex.ExpandStringValueList(d.Get("input_security_groups").([]interface{}))
	}

	if v, ok := d.GetOk("media_connect_flows"); ok && v.(*schema.Set).Len() > 0 {
		in.MediaConnectFlows = expandMediaConnectFlows(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("role_arn"); ok {
		in.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sources"); ok && v.(*schema.Set).Len() > 0 {
		in.Sources = expandSources(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("vpc"); ok && len(v.([]interface{})) > 0 {
		in.Vpc = expandVPC(v.([]interface{}))
	}

	// IAM propagation
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateInput(ctx, in)
		},
		func(err error) (bool, error) {
			var bre *types.BadRequestException
			if errors.As(err, &bre) {
				return strings.Contains(bre.ErrorMessage(), "Please make sure the role exists and medialive.amazonaws.com is a trusted service"), err
			}
			return false, err
		},
	)

	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameInput, d.Get("name").(string), err)
	}

	if outputRaw == nil || outputRaw.(*medialive.CreateInputOutput).Input == nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameInput, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(outputRaw.(*medialive.CreateInputOutput).Input.Id))

	if _, err := waitInputCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForCreation, ResNameInput, d.Id(), err)
	}

	return resourceInputRead(ctx, d, meta)
}

func resourceInputRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	out, err := FindInputByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaLive Input (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameInput, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("attached_channels", out.AttachedChannels)
	d.Set("media_connect_flows", flattenMediaConnectFlows(out.MediaConnectFlows))
	d.Set("name", out.Name)
	d.Set("input_class", out.InputClass)
	d.Set("input_devices", flattenInputDevices(out.InputDevices))
	d.Set("input_partner_ids", out.InputPartnerIds)
	d.Set("input_security_groups", out.SecurityGroups)
	d.Set("input_source_type", out.InputSourceType)
	d.Set("role_arn", out.RoleArn)
	d.Set("sources", flattenSources(out.Sources))
	d.Set("type", out.Type)

	return nil
}

func resourceInputUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		in := &medialive.UpdateInputInput{
			InputId: aws.String(d.Id()),
		}

		if d.HasChange("destinations") {
			in.Destinations = expandDestinations(d.Get("destinations").(*schema.Set).List())
		}

		if d.HasChange("input_devices") {
			in.InputDevices = inputDevices(d.Get("input_devices").(*schema.Set).List()).expandToDeviceRequest()
		}

		if d.HasChange("media_connect_flows") {
			in.MediaConnectFlows = expandMediaConnectFlows(d.Get("media_connect_flows").(*schema.Set).List())
		}

		if d.HasChange("name") {
			in.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("role_arn") {
			in.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		if d.HasChange("sources") {
			in.Sources = expandSources(d.Get("sources").(*schema.Set).List())
		}

		rawOutput, err := tfresource.RetryWhen(ctx, 2*time.Minute,
			func() (interface{}, error) {
				return conn.UpdateInput(ctx, in)
			},
			func(err error) (bool, error) {
				var bre *types.BadRequestException
				if errors.As(err, &bre) {
					return strings.Contains(bre.ErrorMessage(), "The first input attached to a channel cannot be a dynamic input"), err
				}
				return false, err
			},
		)

		if err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionUpdating, ResNameInput, d.Id(), err)
		}

		out := rawOutput.(*medialive.UpdateInputOutput)

		if _, err := waitInputUpdated(ctx, conn, aws.ToString(out.Input.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionWaitingForUpdate, ResNameInput, d.Id(), err)
		}
	}

	return resourceInputRead(ctx, d, meta)
}

func resourceInputDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	log.Printf("[INFO] Deleting MediaLive Input %s", d.Id())

	_, err := conn.DeleteInput(ctx, &medialive.DeleteInputInput{
		InputId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.MediaLive, create.ErrActionDeleting, ResNameInput, d.Id(), err)
	}

	if _, err := waitInputDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForDeletion, ResNameInput, d.Id(), err)
	}

	return nil
}

func waitInputCreated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeInputOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.InputStateCreating),
		Target:                    enum.Slice(types.InputStateDetached, types.InputStateAttached),
		Refresh:                   statusInput(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		Delay:                     30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeInputOutput); ok {
		return out, err
	}

	return nil, err
}

func waitInputUpdated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeInputOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(types.InputStateDetached, types.InputStateAttached),
		Refresh:                   statusInput(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		Delay:                     30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeInputOutput); ok {
		return out, err
	}

	return nil, err
}

func waitInputDeleted(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeInputOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.InputStateDeleting),
		Target:  enum.Slice(types.InputStateDeleted),
		Refresh: statusInput(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeInputOutput); ok {
		return out, err
	}

	return nil, err
}

func statusInput(ctx context.Context, conn *medialive.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindInputByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindInputByID(ctx context.Context, conn *medialive.Client, id string) (*medialive.DescribeInputOutput, error) {
	in := &medialive.DescribeInputInput{
		InputId: aws.String(id),
	}
	out, err := conn.DescribeInput(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenMediaConnectFlow(apiObject types.MediaConnectFlow) map[string]interface{} {
	if apiObject == (types.MediaConnectFlow{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.FlowArn; v != nil {
		m["flow_arn"] = aws.ToString(v)
	}

	return m
}
func flattenMediaConnectFlows(apiObjects []types.MediaConnectFlow) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (types.MediaConnectFlow{}) {
			continue
		}

		l = append(l, flattenMediaConnectFlow(apiObject))
	}

	return l
}

func flattenInputDevice(apiObject types.InputDeviceSettings) map[string]interface{} {
	if apiObject == (types.InputDeviceSettings{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Id; v != nil {
		m["id"] = aws.ToString(v)
	}

	return m
}

func flattenInputDevices(apiObjects []types.InputDeviceSettings) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (types.InputDeviceSettings{}) {
			continue
		}

		l = append(l, flattenInputDevice(apiObject))
	}

	return l
}

func flattenSource(apiObject types.InputSource) map[string]interface{} {
	if apiObject == (types.InputSource{}) {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Url; v != nil {
		m["url"] = aws.ToString(v)
	}
	if v := apiObject.PasswordParam; v != nil {
		m["password_param"] = aws.ToString(v)
	}
	if v := apiObject.Username; v != nil {
		m["username"] = aws.ToString(v)
	}
	return m
}

func flattenSources(apiObjects []types.InputSource) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == (types.InputSource{}) {
			continue
		}

		l = append(l, flattenSource(apiObject))
	}

	return l
}

func expandDestinations(tfList []interface{}) []types.InputDestinationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.InputDestinationRequest

	for _, v := range tfList {
		m, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		var id types.InputDestinationRequest
		if val, ok := m["stream_name"]; ok {
			id.StreamName = aws.String(val.(string))
			s = append(s, id)
		}
	}
	return s
}

type inputDevices []interface{}

func (i inputDevices) expandToDeviceSettings() []types.InputDeviceSettings {
	if len(i) == 0 {
		return nil
	}

	var s []types.InputDeviceSettings

	for _, v := range i {
		m, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		var id types.InputDeviceSettings
		if val, ok := m["id"]; ok {
			id.Id = aws.String(val.(string))
			s = append(s, id)
		}
	}
	return s
}

func (i inputDevices) expandToDeviceRequest() []types.InputDeviceRequest {
	if len(i) == 0 {
		return nil
	}

	var s []types.InputDeviceRequest

	for _, v := range i {
		m, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		var id types.InputDeviceRequest
		if val, ok := m["id"]; ok {
			id.Id = aws.String(val.(string))
			s = append(s, id)
		}
	}
	return s
}

func expandMediaConnectFlows(tfList []interface{}) []types.MediaConnectFlowRequest {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.MediaConnectFlowRequest

	for _, v := range tfList {
		m, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		var id types.MediaConnectFlowRequest
		if val, ok := m["flow_arn"]; ok {
			id.FlowArn = aws.String(val.(string))
			s = append(s, id)
		}
	}
	return s
}

func expandSources(tfList []interface{}) []types.InputSourceRequest {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.InputSourceRequest

	for _, v := range tfList {
		m, ok := v.(map[string]interface{})

		if !ok {
			continue
		}

		var id types.InputSourceRequest
		if val, ok := m["password_param"]; ok {
			id.PasswordParam = aws.String(val.(string))
		}
		if val, ok := m["url"]; ok {
			id.Url = aws.String(val.(string))
		}
		if val, ok := m["username"]; ok {
			id.Username = aws.String(val.(string))
		}
		s = append(s, id)
	}
	return s
}

func expandVPC(tfList []interface{}) *types.InputVpcRequest {
	if len(tfList) == 0 {
		return nil
	}

	var s types.InputVpcRequest
	vpc := tfList[0].(map[string]interface{})

	if val, ok := vpc["subnet_ids"]; ok {
		s.SubnetIds = flex.ExpandStringValueList(val.([]interface{}))
	}
	if val, ok := vpc["security_group_ids"]; ok {
		s.SecurityGroupIds = flex.ExpandStringValueList(val.([]interface{}))
	}

	return &s
}
