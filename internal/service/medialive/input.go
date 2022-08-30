package medialive

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"input_security_groups": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"media_connect_flows": {
				Type:     schema.TypeSet,
				Optional: true,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameInput = "Input"
)

func resourceInputCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	in := &medialive.CreateInputInput{
		RequestId:           aws.String(resource.UniqueId()),
		InputSecurityGroups: flex.ExpandStringValueList(d.Get("input_security_groups").([]interface{})),
		Name:                aws.String(d.Get("name").(string)),
		Type:                types.InputType(d.Get("type").(string)),
		RoleArn:             aws.String(d.Get("role_arn").(string)),
	}

	if v, ok := d.GetOk("destinations"); ok && v.(*schema.Set).Len() > 0 {
		in.Destinations = expandDestinations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("input_devices"); ok && v.(*schema.Set).Len() > 0 {
		in.InputDevices = expandInputDevices(v.([]interface{}))
	}

	if v, ok := d.GetOk("media_connect_flows"); ok && v.(*schema.Set).Len() > 0 {
		in.MediaConnectFlows = expandMediaConnectFlows(v.([]interface{}))
	}

	if v, ok := d.GetOk("role_arn"); ok {
		in.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("sources"); ok && v.(*schema.Set).Len() > 0 {
		in.Sources = expandSources(v.([]interface{}))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateInput(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameInput, d.Get("name").(string), err)
	}

	if out == nil || out.Input == nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameInput, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Input.Id))

	if _, err := waitInputCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForCreation, ResNameInput, d.Id(), err)
	}

	return resourceInputRead(ctx, d, meta)
}

func resourceInputRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

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
	d.Set("name", out.Name)
	d.Set("input_class", out.InputClass)
	d.Set("input_security_groups", out.SecurityGroups)
	d.Set("role_arn", out.RoleArn)
	d.Set("type", out.Type)

	tags, err := ListTags(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameInput, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameInput, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameInput, d.Id(), err)
	}

	return nil
}

func resourceInputUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	if d.HasChangesExcept("tags", "tags_all") {
		in := &medialive.UpdateInputInput{
			InputId: aws.String(d.Id()),
		}

		if d.HasChange("destinations") {
			in.Destinations = expandDestinations(d.Get("destinations").(*schema.Set).List())
		}

		if d.HasChange("name") {
			in.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("role_arn") {
			in.RoleArn = aws.String(d.Get("role_arn").(string))
		}

		log.Printf("[DEBUG] Updating MediaLive Input (%s): %#v", d.Id(), in)
		out, err := conn.UpdateInput(ctx, in)
		if err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionUpdating, ResNameInput, d.Id(), err)
		}

		if _, err := waitInputUpdated(ctx, conn, aws.ToString(out.Input.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionWaitingForUpdate, ResNameInput, d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return create.DiagError(names.MediaLive, create.ErrActionUpdating, ResNameInput, d.Id(), err)
		}
	}

	return resourceInputRead(ctx, d, meta)
}

func resourceInputDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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

func statusInput(ctx context.Context, conn *medialive.Client, id string) resource.StateRefreshFunc {
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
			return nil, &resource.NotFoundError{
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

//func flattenComplexArgument(apiObject *medialive.ComplexArgument) map[string]interface{} {
//	if apiObject == nil {
//		return nil
//	}
//
//	m := map[string]interface{}{}
//
//	if v := apiObject.SubFieldOne; v != nil {
//		m["sub_field_one"] = aws.ToString(v)
//	}
//
//	if v := apiObject.SubFieldTwo; v != nil {
//		m["sub_field_two"] = aws.ToString(v)
//	}
//
//	return m
//}
//
//func flattenComplexArguments(apiObjects []*medialive.ComplexArgument) []interface{} {
//	if len(apiObjects) == 0 {
//		return nil
//	}
//
//	var l []interface{}
//
//	for _, apiObject := range apiObjects {
//		if apiObject == nil {
//			continue
//		}
//
//		l = append(l, flattenComplexArgument(apiObject))
//	}
//
//	return l
//}

func expandDestinations(tfList []interface{}) []types.InputDestinationRequest {
	if tfList == nil || len(tfList) == 0 {
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

func expandInputDevices(tfList []interface{}) []types.InputDeviceSettings {
	if tfList == nil || len(tfList) == 0 {
		return nil
	}

	var s []types.InputDeviceSettings

	for _, v := range tfList {
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

func expandMediaConnectFlows(tfList []interface{}) []types.MediaConnectFlowRequest {
	if tfList == nil || len(tfList) == 0 {
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
	if tfList == nil || len(tfList) == 0 {
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
