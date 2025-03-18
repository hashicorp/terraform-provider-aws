// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_emr_instance_group", name="Instance Group")
func resourceInstanceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceGroupCreate,
		ReadWithoutTimeout:   resourceInstanceGroupRead,
		UpdateWithoutTimeout: resourceInstanceGroupUpdate,
		DeleteWithoutTimeout: resourceInstanceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected cluster-id/ig-id", d.Id())
				}
				clusterID := idParts[0]
				resourceID := idParts[1]
				d.Set("cluster_id", clusterID)
				d.SetId(resourceID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"autoscaling_policy": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
			},
			"bid_price": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"configurations_json": {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"ebs_config": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						names.AttrSize: {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validEBSVolumeType(),
						},
						"volumes_per_instance": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  1,
						},
					},
				},
				Set: resourceClusterEBSHashConfig,
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrInstanceCount: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"running_instance_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceInstanceGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	name := d.Get(names.AttrName).(string)
	groupConfig := awstypes.InstanceGroupConfig{
		EbsConfiguration: readEBSConfig(d),
		InstanceRole:     awstypes.InstanceRoleTypeTask,
		InstanceType:     aws.String(d.Get(names.AttrInstanceType).(string)),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("autoscaling_policy"); ok {
		var autoScalingPolicy awstypes.AutoScalingPolicy

		if err := tfjson.DecodeFromString(v.(string), &autoScalingPolicy); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		groupConfig.AutoScalingPolicy = &autoScalingPolicy
	}

	groupConfig.Market = awstypes.MarketTypeOnDemand
	if v, ok := d.GetOk("bid_price"); ok {
		groupConfig.BidPrice = aws.String(v.(string))
		groupConfig.Market = awstypes.MarketTypeSpot
	}

	if v, ok := d.GetOk("configurations_json"); ok {
		v, err := structure.NormalizeJsonString(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		groupConfig.Configurations, err = expandConfigurationJSON(v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if v := d.GetRawConfig().GetAttr(names.AttrInstanceCount); v.IsKnown() && !v.IsNull() {
		v, _ := v.AsBigFloat().Int64()
		groupConfig.InstanceCount = aws.Int32(int32(v))
	} else {
		groupConfig.InstanceCount = aws.Int32(1)
	}

	input := &emr.AddInstanceGroupsInput{
		InstanceGroups: []awstypes.InstanceGroupConfig{groupConfig},
		JobFlowId:      aws.String(d.Get("cluster_id").(string)),
	}

	output, err := conn.AddInstanceGroups(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Instance Group (%s): %s", name, err)
	}

	d.SetId(output.InstanceGroupIds[0])

	const (
		timeout = 30 * time.Minute
	)
	if _, err := waitInstanceGroupRunning(ctx, conn, d.Get("cluster_id").(string), d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Instance Group (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceGroupRead(ctx, d, meta)...)
}

func resourceInstanceGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	ig, err := findInstanceGroupByTwoPartKey(ctx, conn, d.Get("cluster_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Instance Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Instance Group (%s): %s", d.Id(), err)
	}

	autoscalingPolicyString, err := flattenAutoScalingPolicyDescription(ig.AutoScalingPolicy)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("autoscaling_policy", autoscalingPolicyString)
	d.Set("bid_price", ig.BidPrice)
	switch {
	case len(ig.Configurations) > 0:
		configOut, err := flattenConfigurationJSON(ig.Configurations)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("configurations_json", configOut)
	default:
		d.Set("configurations_json", "")
	}
	if err := d.Set("ebs_config", flattenEBSConfig(ig.EbsBlockDevices)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_config: %s", err)
	}
	d.Set("ebs_optimized", ig.EbsOptimized)
	d.Set(names.AttrInstanceCount, ig.RequestedInstanceCount)
	d.Set(names.AttrInstanceType, ig.InstanceType)
	d.Set(names.AttrName, ig.Name)
	d.Set("running_instance_count", ig.RunningInstanceCount)
	d.Set(names.AttrStatus, ig.Status.State)

	return diags
}

func resourceInstanceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	if d.HasChanges(names.AttrInstanceCount, "configurations_json") {
		groupConfig := awstypes.InstanceGroupModifyConfig{
			InstanceGroupId: aws.String(d.Id()),
		}

		if d.HasChange(names.AttrInstanceCount) {
			groupConfig.InstanceCount = aws.Int32(int32(d.Get(names.AttrInstanceCount).(int)))
		}

		if d.HasChange("configurations_json") {
			if v, ok := d.GetOk("configurations_json"); ok {
				v, err := structure.NormalizeJsonString(v)
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
				groupConfig.Configurations, err = expandConfigurationJSON(v)
				if err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}

		input := &emr.ModifyInstanceGroupsInput{
			InstanceGroups: []awstypes.InstanceGroupModifyConfig{
				groupConfig,
			},
		}

		_, err := conn.ModifyInstanceGroups(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Instance Group (%s): %s", d.Id(), err)
		}

		const (
			timeout = 30 * time.Minute
		)
		if _, err := waitInstanceGroupRunning(ctx, conn, d.Get("cluster_id").(string), d.Id(), timeout); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EMR Instance Group (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("autoscaling_policy") {
		var autoScalingPolicy awstypes.AutoScalingPolicy

		if err := tfjson.DecodeFromString(d.Get("autoscaling_policy").(string), &autoScalingPolicy); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &emr.PutAutoScalingPolicyInput{
			AutoScalingPolicy: &autoScalingPolicy,
			ClusterId:         aws.String(d.Get("cluster_id").(string)),
			InstanceGroupId:   aws.String(d.Id()),
		}

		_, err := conn.PutAutoScalingPolicy(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EMR Instance Group (%s) AutoScalingPolicy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceInstanceGroupRead(ctx, d, meta)...)
}

func resourceInstanceGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	input := &emr.ModifyInstanceGroupsInput{
		InstanceGroups: []awstypes.InstanceGroupModifyConfig{
			{
				InstanceCount:   aws.Int32(0),
				InstanceGroupId: aws.String(d.Id()),
			},
		},
	}

	_, err := conn.ModifyInstanceGroups(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeValidationException, "instance group may only be modified when the cluster is running or waiting") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Instance Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findInstanceGroupByTwoPartKey(ctx context.Context, conn *emr.Client, clusterID, groupID string) (*awstypes.InstanceGroup, error) {
	input := &emr.ListInstanceGroupsInput{
		ClusterId: aws.String(clusterID),
	}
	output, err := findInstanceGroup(ctx, conn, input, func(v *awstypes.InstanceGroup) bool {
		return aws.ToString(v.Id) == groupID
	})

	if err != nil {
		return nil, err
	}

	if state := output.Status.State; state == awstypes.InstanceGroupStateTerminating || state == awstypes.InstanceGroupStateTerminated {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func findInstanceGroup(ctx context.Context, conn *emr.Client, input *emr.ListInstanceGroupsInput, filter tfslices.Predicate[*awstypes.InstanceGroup]) (*awstypes.InstanceGroup, error) {
	output, err := findInstanceGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceGroups(ctx context.Context, conn *emr.Client, input *emr.ListInstanceGroupsInput, filter tfslices.Predicate[*awstypes.InstanceGroup]) ([]awstypes.InstanceGroup, error) {
	var output []awstypes.InstanceGroup

	pages := emr.NewListInstanceGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "is not valid") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.InstanceGroups {
			if v.Status != nil && filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusInstanceGroup(ctx context.Context, conn *emr.Client, clusterID, groupID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findInstanceGroupByTwoPartKey(ctx, conn, clusterID, groupID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.State), nil
	}
}

func waitInstanceGroupRunning(ctx context.Context, conn *emr.Client, clusterID, groupID string, timeout time.Duration) (*awstypes.InstanceGroup, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.InstanceGroupStateBootstrapping, awstypes.InstanceGroupStateProvisioning, awstypes.InstanceGroupStateReconfiguring, awstypes.InstanceGroupStateResizing),
		Target:     enum.Slice(awstypes.InstanceGroupStateRunning),
		Refresh:    statusInstanceGroup(ctx, conn, clusterID, groupID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InstanceGroup); ok {
		return output, err
	}

	return nil, err
}

// readEBSConfig populates an emr.EbsConfiguration struct
func readEBSConfig(d *schema.ResourceData) *awstypes.EbsConfiguration {
	result := &awstypes.EbsConfiguration{}
	if v, ok := d.GetOk("ebs_optimized"); ok {
		result.EbsOptimized = aws.Bool(v.(bool))
	}

	ebsConfigs := make([]awstypes.EbsBlockDeviceConfig, 0)
	if rawConfig, ok := d.GetOk("ebs_config"); ok {
		configList := rawConfig.(*schema.Set).List()
		for _, config := range configList {
			conf := config.(map[string]any)
			ebs := awstypes.EbsBlockDeviceConfig{}
			volumeSpec := &awstypes.VolumeSpecification{
				SizeInGB:   aws.Int32(int32(conf[names.AttrSize].(int))),
				VolumeType: aws.String(conf[names.AttrType].(string)),
			}
			if v, ok := conf[names.AttrIOPS].(int); ok && v != 0 {
				volumeSpec.Iops = aws.Int32(int32(v))
			}
			if v, ok := conf["volumes_per_instance"].(int); ok && v != 0 {
				ebs.VolumesPerInstance = aws.Int32(int32(v))
			}
			ebs.VolumeSpecification = volumeSpec
			ebsConfigs = append(ebsConfigs, ebs)
		}
	}
	result.EbsBlockDeviceConfigs = ebsConfigs
	return result
}
