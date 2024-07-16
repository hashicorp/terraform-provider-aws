// // Copyright (c) HashiCorp, Inc.
// // SPDX-License-Identifier: MPL-2.0
package finspace

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
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

// @SDKResource("aws_finspace_kx_volume", name="Kx Volume")
// @Tags(identifierAttribute="arn")
func ResourceKxVolume() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceKxVolumeCreate,
		ReadWithoutTimeout:   resourceKxVolumeRead,
		UpdateWithoutTimeout: resourceKxVolumeUpdate,
		DeleteWithoutTimeout: resourceKxVolumeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(45 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attached_clusters": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrClusterName: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(3, 63),
						},
						"cluster_status": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.KxClusterStatus](),
						},
						"cluster_type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.KxClusterType](),
						},
					},
				},
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
				ForceNew: true,
			},
			"az_mode": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.KxAzMode](),
			},
			"created_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 32),
			},
			"last_modified_timestamp": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"nas1_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSize: {
							Type:         schema.TypeInt,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1200, 33600),
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[types.KxNAS1Type](),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusReason: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.KxVolumeType](),
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameKxVolume     = "Kx Volume"
	kxVolumeIDPartCount = 2
)

func resourceKxVolumeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	environmentId := d.Get("environment_id").(string)
	volumeName := d.Get(names.AttrName).(string)
	idParts := []string{
		environmentId,
		volumeName,
	}
	rID, err := flex.FlattenResourceId(idParts, kxVolumeIDPartCount, false)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionFlatteningResourceId, ResNameKxVolume, d.Get(names.AttrName).(string), err)
	}
	d.SetId(rID)

	in := &finspace.CreateKxVolumeInput{
		ClientToken:         aws.String(id.UniqueId()),
		AvailabilityZoneIds: flex.ExpandStringValueList(d.Get(names.AttrAvailabilityZones).([]interface{})),
		EnvironmentId:       aws.String(environmentId),
		VolumeType:          types.KxVolumeType(d.Get(names.AttrType).(string)),
		VolumeName:          aws.String(volumeName),
		AzMode:              types.KxAzMode(d.Get("az_mode").(string)),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("nas1_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		in.Nas1Configuration = expandNas1Configuration(v.([]interface{}))
	}

	out, err := conn.CreateKxVolume(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxVolume, d.Get(names.AttrName).(string), err)
	}

	if out == nil || out.VolumeName == nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxVolume, d.Get(names.AttrName).(string), errors.New("empty output"))
	}

	if _, err := waitKxVolumeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForCreation, ResNameKxVolume, d.Id(), err)
	}

	// The CreateKxVolume API currently fails to tag the Volume when the
	// Tags field is set. Until the API is fixed, tag after creation instead.
	if err := createTags(ctx, conn, aws.ToString(out.VolumeArn), getTagsIn(ctx)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionCreating, ResNameKxVolume, d.Id(), err)
	}

	return append(diags, resourceKxVolumeRead(ctx, d, meta)...)
}

func resourceKxVolumeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	out, err := FindKxVolumeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FinSpace KxVolume (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionReading, ResNameKxVolume, d.Id(), err)
	}

	d.Set(names.AttrARN, out.VolumeArn)
	d.Set(names.AttrName, out.VolumeName)
	d.Set(names.AttrDescription, out.Description)
	d.Set(names.AttrType, out.VolumeType)
	d.Set(names.AttrStatus, out.Status)
	d.Set(names.AttrStatusReason, out.StatusReason)
	d.Set("az_mode", out.AzMode)
	d.Set(names.AttrDescription, out.Description)
	d.Set("created_timestamp", out.CreatedTimestamp.String())
	d.Set("last_modified_timestamp", out.LastModifiedTimestamp.String())
	d.Set(names.AttrAvailabilityZones, aws.StringSlice(out.AvailabilityZoneIds))

	if err := d.Set("nas1_configuration", flattenNas1Configuration(out.Nas1Configuration)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionSetting, ResNameKxVolume, d.Id(), err)
	}

	if err := d.Set("attached_clusters", flattenAttachedClusters(out.AttachedClusters)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionSetting, ResNameKxVolume, d.Id(), err)
	}

	parts, err := flex.ExpandResourceId(d.Id(), kxVolumeIDPartCount, false)
	if err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionSetting, ResNameKxVolume, d.Id(), err)
	}
	d.Set("environment_id", parts[0])

	return diags
}

func resourceKxVolumeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	updateVolume := false

	in := &finspace.UpdateKxVolumeInput{
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
		VolumeName:    aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok && d.HasChanges(names.AttrDescription) {
		in.Description = aws.String(v.(string))
		updateVolume = true
	}

	if v, ok := d.GetOk("nas1_configuration"); ok && len(v.([]interface{})) > 0 && d.HasChanges("nas1_configuration") {
		in.Nas1Configuration = expandNas1Configuration(v.([]interface{}))
		updateVolume = true
	}

	if !updateVolume {
		return diags
	}

	log.Printf("[DEBUG] Updating FinSpace KxVolume (%s): %#v", d.Id(), in)

	if _, err := conn.UpdateKxVolume(ctx, in); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxVolume, d.Id(), err)
	}
	if _, err := waitKxVolumeUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionUpdating, ResNameKxVolume, d.Id(), err)
	}

	return append(diags, resourceKxVolumeRead(ctx, d, meta)...)
}

func resourceKxVolumeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FinSpaceClient(ctx)

	log.Printf("[INFO] Deleting FinSpace Kx Volume: %s", d.Id())
	_, err := conn.DeleteKxVolume(ctx, &finspace.DeleteKxVolumeInput{
		VolumeName:    aws.String(d.Get(names.AttrName).(string)),
		EnvironmentId: aws.String(d.Get("environment_id").(string)),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return diags
		}

		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionDeleting, ResNameKxVolume, d.Id(), err)
	}

	_, err = waitKxVolumeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))
	if err != nil && !tfresource.NotFound(err) {
		return create.AppendDiagError(diags, names.FinSpace, create.ErrActionWaitingForDeletion, ResNameKxVolume, d.Id(), err)
	}

	return diags
}

func waitKxVolumeCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxVolumeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.KxVolumeStatusCreating),
		Target:                    enum.Slice(types.KxVolumeStatusActive),
		Refresh:                   statusKxVolume(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxVolumeOutput); ok {
		return out, err
	}

	return nil, err
}

func waitKxVolumeUpdated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxVolumeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.KxVolumeStatusCreating, types.KxVolumeStatusUpdating),
		Target:                    enum.Slice(types.KxVolumeStatusActive),
		Refresh:                   statusKxVolume(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxVolumeOutput); ok {
		return out, err
	}

	return nil, err
}

func waitKxVolumeDeleted(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxVolumeOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.KxVolumeStatusDeleting),
		Target:  enum.Slice(types.KxVolumeStatusDeleted),
		Refresh: statusKxVolume(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxVolumeOutput); ok {
		return out, err
	}

	return nil, err
}

func statusKxVolume(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindKxVolumeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func FindKxVolumeByID(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxVolumeOutput, error) {
	parts, err := flex.ExpandResourceId(id, kxVolumeIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &finspace.GetKxVolumeInput{
		EnvironmentId: aws.String(parts[0]),
		VolumeName:    aws.String(parts[1]),
	}

	out, err := conn.GetKxVolume(ctx, in)
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

	if out == nil || out.VolumeArn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandNas1Configuration(tfList []interface{}) *types.KxNAS1Configuration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	a := &types.KxNAS1Configuration{}

	if v, ok := tfMap[names.AttrSize].(int); ok && v != 0 {
		a.Size = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		a.Type = types.KxNAS1Type(v)
	}
	return a
}

func flattenNas1Configuration(apiObject *types.KxNAS1Configuration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Size; v != nil {
		m[names.AttrSize] = aws.ToInt32(v)
	}

	if v := apiObject.Type; v != "" {
		m[names.AttrType] = v
	}

	return []interface{}{m}
}

func flattenCluster(apiObject *types.KxAttachedCluster) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.ClusterName; aws.ToString(v) != "" {
		m[names.AttrClusterName] = aws.ToString(v)
	}

	if v := apiObject.ClusterStatus; v != "" {
		m["cluster_status"] = string(v)
	}

	if v := apiObject.ClusterType; v != "" {
		m["cluster_type"] = string(v)
	}

	return m
}

func flattenAttachedClusters(apiObjects []types.KxAttachedCluster) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenCluster(&apiObject))
	}

	return l
}
