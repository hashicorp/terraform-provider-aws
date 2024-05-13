// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

// @SDKResource("aws_medialive_multiplex", name="Multiplex")
// @Tags(identifierAttribute="arn")
func ResourceMultiplex() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMultiplexCreate,
		ReadWithoutTimeout:   resourceMultiplexRead,
		UpdateWithoutTimeout: resourceMultiplexUpdate,
		DeleteWithoutTimeout: resourceMultiplexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MinItems: 2,
				MaxItems: 2,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"multiplex_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transport_stream_bitrate": {
							Type:             schema.TypeInt,
							Required:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1000000, 100000000)),
						},
						"transport_stream_reserved_bitrate": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"transport_stream_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"maximum_video_buffer_delay_milliseconds": {
							Type:             schema.TypeInt,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1000, 3000)),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"start_multiplex": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameMultiplex = "Multiplex"
)

func resourceMultiplexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	in := &medialive.CreateMultiplexInput{
		RequestId:         aws.String(id.UniqueId()),
		Name:              aws.String(d.Get(names.AttrName).(string)),
		AvailabilityZones: flex.ExpandStringValueList(d.Get(names.AttrAvailabilityZones).([]interface{})),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("multiplex_settings"); ok && len(v.([]interface{})) > 0 {
		in.MultiplexSettings = expandMultiplexSettings(v.([]interface{}))
	}

	out, err := conn.CreateMultiplex(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameMultiplex, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.Multiplex == nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameMultiplex, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Multiplex.Id))

	if _, err := waitMultiplexCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForCreation, ResNameMultiplex, d.Id(), err)
	}

	if d.Get("start_multiplex").(bool) {
		if err := startMultiplex(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionCreating, ResNameMultiplex, d.Id(), err)
		}
	}

	return append(diags, resourceMultiplexRead(ctx, d, meta)...)
}

func resourceMultiplexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	out, err := FindMultiplexByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaLive Multiplex (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionReading, ResNameMultiplex, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrAvailabilityZones, out.AvailabilityZones)
	d.Set(names.AttrName, out.Name)

	if err := d.Set("multiplex_settings", flattenMultiplexSettings(out.MultiplexSettings)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionSetting, ResNameMultiplex, d.Id(), err)
	}

	return diags
}

func resourceMultiplexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "start_multiplex") {
		in := &medialive.UpdateMultiplexInput{
			MultiplexId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrName) {
			in.Name = aws.String(d.Get(names.AttrName).(string))
		}
		if d.HasChange("multiplex_settings") {
			in.MultiplexSettings = expandMultiplexSettings(d.Get("multiplex_settings").([]interface{}))
		}

		log.Printf("[DEBUG] Updating MediaLive Multiplex (%s): %#v", d.Id(), in)
		out, err := conn.UpdateMultiplex(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameMultiplex, d.Id(), err)
		}

		if _, err := waitMultiplexUpdated(ctx, conn, aws.ToString(out.Multiplex.Id), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForUpdate, ResNameMultiplex, d.Id(), err)
		}
	}

	if d.HasChange("start_multiplex") {
		out, err := FindMultiplexByID(ctx, conn, d.Id())
		if err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameMultiplex, d.Id(), err)
		}
		if d.Get("start_multiplex").(bool) {
			if out.State != types.MultiplexStateRunning {
				if err := startMultiplex(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameMultiplex, d.Id(), err)
				}
			}
		} else {
			if out.State == types.MultiplexStateRunning {
				if err := stopMultiplex(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return create.AppendDiagError(diags, names.MediaLive, create.ErrActionUpdating, ResNameMultiplex, d.Id(), err)
				}
			}
		}
	}

	return append(diags, resourceMultiplexRead(ctx, d, meta)...)
}

func resourceMultiplexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaLiveClient(ctx)

	log.Printf("[INFO] Deleting MediaLive Multiplex %s", d.Id())

	out, err := FindMultiplexByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		create.DiagError(names.MediaLive, create.ErrActionDeleting, ResNameMultiplex, d.Id(), err)
	}

	if out.State == types.MultiplexStateRunning {
		if err := stopMultiplex(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return create.AppendDiagError(diags, names.MediaLive, create.ErrActionDeleting, ResNameMultiplex, d.Id(), err)
		}
	}

	_, err = conn.DeleteMultiplex(ctx, &medialive.DeleteMultiplexInput{
		MultiplexId: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionDeleting, ResNameMultiplex, d.Id(), err)
	}

	if _, err := waitMultiplexDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.MediaLive, create.ErrActionWaitingForDeletion, ResNameMultiplex, d.Id(), err)
	}

	return diags
}

func waitMultiplexCreated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeMultiplexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.MultiplexStateCreating),
		Target:                    enum.Slice(types.MultiplexStateIdle),
		Refresh:                   statusMultiplex(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		Delay:                     30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeMultiplexOutput); ok {
		return out, err
	}

	return nil, err
}

func waitMultiplexUpdated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeMultiplexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(types.MultiplexStateIdle),
		Refresh:                   statusMultiplex(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
		Delay:                     30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeMultiplexOutput); ok {
		return out, err
	}

	return nil, err
}

func waitMultiplexDeleted(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeMultiplexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.MultiplexStateDeleting),
		Target:  enum.Slice(types.MultiplexStateDeleted),
		Refresh: statusMultiplex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeMultiplexOutput); ok {
		return out, err
	}

	return nil, err
}

func waitMultiplexRunning(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeMultiplexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.MultiplexStateStarting),
		Target:  enum.Slice(types.MultiplexStateRunning),
		Refresh: statusMultiplex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeMultiplexOutput); ok {
		return out, err
	}

	return nil, err
}

func waitMultiplexStopped(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeMultiplexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.MultiplexStateStopping),
		Target:  enum.Slice(types.MultiplexStateIdle),
		Refresh: statusMultiplex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeMultiplexOutput); ok {
		return out, err
	}

	return nil, err
}

func statusMultiplex(ctx context.Context, conn *medialive.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindMultiplexByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindMultiplexByID(ctx context.Context, conn *medialive.Client, id string) (*medialive.DescribeMultiplexOutput, error) {
	in := &medialive.DescribeMultiplexInput{
		MultiplexId: aws.String(id),
	}
	out, err := conn.DescribeMultiplex(ctx, in)
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

func flattenMultiplexSettings(apiObject *types.MultiplexSettings) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"transport_stream_bitrate":                apiObject.TransportStreamBitrate,
		"transport_stream_id":                     apiObject.TransportStreamId,
		"maximum_video_buffer_delay_milliseconds": apiObject.MaximumVideoBufferDelayMilliseconds,
		"transport_stream_reserved_bitrate":       apiObject.TransportStreamReservedBitrate,
	}

	return []interface{}{m}
}

func expandMultiplexSettings(tfList []interface{}) *types.MultiplexSettings {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]interface{})

	s := types.MultiplexSettings{}

	if v, ok := m["transport_stream_bitrate"]; ok {
		s.TransportStreamBitrate = aws.Int32(int32(v.(int)))
	}
	if v, ok := m["transport_stream_id"]; ok {
		s.TransportStreamId = aws.Int32(int32(v.(int)))
	}
	if val, ok := m["maximum_video_buffer_delay_milliseconds"]; ok {
		s.MaximumVideoBufferDelayMilliseconds = aws.Int32(int32(val.(int)))
	}
	if val, ok := m["transport_stream_reserved_bitrate"]; ok {
		s.TransportStreamReservedBitrate = aws.Int32(int32(val.(int)))
	}

	return &s
}

func startMultiplex(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Starting Medialive Multiplex: (%s)", id)
	_, err := conn.StartMultiplex(ctx, &medialive.StartMultiplexInput{
		MultiplexId: aws.String(id),
	})

	if err != nil {
		return err
	}

	_, err = waitMultiplexRunning(ctx, conn, id, timeout)

	return err
}

func stopMultiplex(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) error {
	log.Printf("[DEBUG] Starting Medialive Multiplex: (%s)", id)
	_, err := conn.StopMultiplex(ctx, &medialive.StopMultiplexInput{
		MultiplexId: aws.String(id),
	})

	if err != nil {
		return err
	}

	_, err = waitMultiplexStopped(ctx, conn, id, timeout)

	return err
}
