// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// @SDKResource("aws_finspace_kx_dataview", name="Kx Dataview")
// @Tags(identifierAttribute="arn")
func ResourceKxDataview() *schema.Resource {

	return &schema.Resource{
		CreateWithoutTimeout: resourceKxDataviewCreate,
		ReadWithoutTimeout:   resourceKxDataviewRead,
		UpdateWithoutTimeout: resourceKxDataviewUpdate,
		DeleteWithoutTimeout: resourceKxDataviewDelete,

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
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"database_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"auto_update": {
				Type:     schema.TypeBool,
				ForceNew: true,
				Required: true,
			},
			"changeset_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"az_mode": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.KxAzMode](),
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"segment_configurations": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"volume_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"db_paths": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Required: true,
						},
					},
				},
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameKxDataview     = "Kx Dataview"
	kxDataviewIdPartCount = 3
)

func resourceKxDataviewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	idParts := []string{
		d.Get("environment_id").(string),
		d.Get("database_name").(string),
		d.Get("name").(string),
	}

	rId, err := flex.FlattenResourceId(idParts, kxDataviewIdPartCount, false)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionFlatteningResourceId, ResNameKxDataview, d.Get("name").(string), err)
	}
	d.SetId(rId)

	in := &finspace.CreateKxDataviewInput{
		DatabaseName:  aws.String(d.Get("database_name").(string)),
		DataviewName:  aws.String(d.Get("name").(string)),
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		AutoUpdate:    d.Get("auto_update").(bool),
		AzMode:        types.KxAzMode(d.Get("az_mode").(string)),
		ClientToken:   aws.String(id.UniqueId()),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("changeset_id"); ok {
		in.ChangesetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("availability_zone_id"); ok {
		in.AvailabilityZoneId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("segment_configurations"); ok && len(v.([]interface{})) > 0 {
		in.SegmentConfigurations = expandSegmentConfigurations(v.([]interface{}))
	}

	out, err := conn.CreateKxDataview(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxDataview, d.Get("name").(string), err)
	}
	if out == nil || out.DataviewName == nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxDataview, d.Get("name").(string), errors.New("empty output"))
	}
	if _, err := waitKxDataviewCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForCreation, ResNameKxDataview, d.Get("name").(string), err)
	}

	return append(diags, resourceKxDataviewRead(ctx, d, meta)...)
}

func resourceKxDataviewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	out, err := findKxDataviewById(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxDataview (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxDataview, d.Id(), err)
	}
	d.Set("name", out.DataviewName)
	d.Set("description", out.Description)
	d.Set("auto_update", out.AutoUpdate)
	d.Set("changeset_id", out.ChangesetId)
	d.Set("availability_zone_id", out.AvailabilityZoneId)
	d.Set("status", out.Status)
	d.Set("created_timestamp", out.CreatedTimestamp.String())
	d.Set("last_modified_timestamp", out.LastModifiedTimestamp.String())
	d.Set("database_name", out.DatabaseName)
	d.Set("environment_id", out.EnvironmentId)
	d.Set("az_mode", out.AzMode)
	if err := d.Set("segment_configurations", flattenSegmentConfigurations(out.SegmentConfigurations)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxDataview, d.Id(), err)
	}

	return diags
}

func resourceKxDataviewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)
	in := &finspace.UpdateKxDataviewInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		DatabaseName:  aws.String(d.Get("database_name").(string)),
		DataviewName:  aws.String(d.Get("name").(string)),
		ClientToken:   aws.String(id.UniqueId()),
	}

	if v, ok := d.GetOk("changeset_id"); ok && d.HasChange("changeset_id") && d.Get("auto_update").(bool) != true {
		in.ChangesetId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("segment_configurations"); ok && len(v.([]interface{})) > 0 && d.HasChange("segment_configurations") {
		in.SegmentConfigurations = expandSegmentConfigurations(v.([]interface{}))
	}

	if _, err := conn.UpdateKxDataview(ctx, in); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxDataview, d.Get("name").(string), err)
	}

	if _, err := waitKxDataviewUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForUpdate, ResNameKxDataview, d.Get("name").(string), err)
	}

	return append(diags, resourceKxDataviewRead(ctx, d, meta)...)
}

func resourceKxDataviewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	_, err := conn.DeleteKxDataview(ctx, &finspace.DeleteKxDataviewInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		DatabaseName:  aws.String(d.Get("database_name").(string)),
		DataviewName:  aws.String(d.Get("name").(string)),
		ClientToken:   aws.String(id.UniqueId()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionDeleting, ResNameKxDataview, d.Get("name").(string), err)
	}

	if _, err := waitKxDataviewDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil && !tfresource.NotFound(err) {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForDeletion, ResNameKxDataview, d.Id(), err)
	}
	return diags
}

func findKxDataviewById(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxDataviewOutput, error) {
	idParts, err := flex.ExpandResourceId(id, kxDataviewIdPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &finspace.GetKxDataviewInput{
		EnvironmentId: aws.String(idParts[0]),
		DatabaseName:  aws.String(idParts[1]),
		DataviewName:  aws.String(idParts[2]),
	}

	out, err := conn.GetKxDataview(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException

		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}

		}
		return nil, err
	}

	if out == nil || out.DataviewName == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	return out, nil
}

func waitKxDataviewCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxDataviewOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.KxDataviewStatusCreating),
		Target:                    enum.Slice(types.KxDataviewStatusActive),
		Refresh:                   statusKxDataview(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxDataviewOutput); ok {
		return out, err
	}
	return nil, err
}

func waitKxDataviewUpdated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxDataviewOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.KxDataviewStatusUpdating),
		Target:                    enum.Slice(types.KxDataviewStatusActive),
		Refresh:                   statusKxDataview(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*finspace.GetKxDataviewOutput); ok {
		return out, err
	}
	return nil, err
}

func waitKxDataviewDeleted(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxDataviewOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.KxDataviewStatusDeleting),
		Target:  []string{},
		Refresh: statusKxDataview(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxDataviewOutput); ok {
		return out, err
	}

	return nil, err
}

func statusKxDataview(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findKxDataviewById(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

func expandDbPath(tfList []interface{}) []string {
	if tfList == nil {
		return nil
	}
	var s []string

	for _, v := range tfList {
		s = append(s, v.(string))
	}
	return s
}

func expandSegmentConfigurations(tfList []interface{}) []types.KxDataviewSegmentConfiguration {
	if tfList == nil {
		return nil
	}
	var s []types.KxDataviewSegmentConfiguration

	for _, v := range tfList {
		m := v.(map[string]interface{})
		s = append(s, types.KxDataviewSegmentConfiguration{
			VolumeName: aws.String(m["volume_name"].(string)),
			DbPaths:    expandDbPath(m["db_paths"].([]interface{})),
		})
	}

	return s
}
func flattenSegmentConfiguration(apiObject *types.KxDataviewSegmentConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}
	m := map[string]interface{}{}
	if v := apiObject.VolumeName; aws.ToString(v) != "" {
		m["volume_name"] = aws.ToString(v)
	}
	if v := apiObject.DbPaths; v != nil {
		m["db_paths"] = v
	}
	return m
}

func flattenSegmentConfigurations(apiObjects []types.KxDataviewSegmentConfiguration) []interface{} {
	if apiObjects == nil {
		return nil
	}
	var l []interface{}
	for _, apiObject := range apiObjects {
		l = append(l, flattenSegmentConfiguration(&apiObject))
	}
	return l
}
