// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_eks_node_group", name="Node Group")
// @Tags(identifierAttribute="arn")
func resourceNodeGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNodeGroupCreate,
		ReadWithoutTimeout:   resourceNodeGroupRead,
		UpdateWithoutTimeout: resourceNodeGroupUpdate,
		DeleteWithoutTimeout: resourceNodeGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ami_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.AMITypes](),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.CapacityTypes](),
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validClusterName,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"force_update_version": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"instance_types": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"launch_template": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.name"},
							ValidateFunc:  verify.ValidLaunchTemplateID,
						},
						"name": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"launch_template.0.id"},
							ValidateFunc:  verify.ValidLaunchTemplateName,
						},
						"version": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"node_group_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"node_group_name_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 63),
			},
			"node_group_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"node_group_name"},
				ValidateFunc:  validation.StringLenBetween(0, 63-id.UniqueIDSuffixLength),
			},
			"node_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"release_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"remote_access": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ec2_ssh_key": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"source_security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoscaling_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"remote_access_security_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"scaling_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"max_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"min_size": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"taint": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 63),
						},
						"value": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 63),
						},
						"effect": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.TaintEffect](),
						},
					},
				},
			},
			"update_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_unavailable": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 100),
							ExactlyOneOf: []string{
								"update_config.0.max_unavailable",
								"update_config.0.max_unavailable_percentage",
							},
						},
						"max_unavailable_percentage": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 100),
							ExactlyOneOf: []string{
								"update_config.0.max_unavailable",
								"update_config.0.max_unavailable_percentage",
							},
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceNodeGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName := d.Get("cluster_name").(string)
	nodeGroupName := create.Name(d.Get("node_group_name").(string), d.Get("node_group_name_prefix").(string))
	groupID := NodeGroupCreateResourceID(clusterName, nodeGroupName)
	input := &eks.CreateNodegroupInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		ClusterName:        aws.String(clusterName),
		NodegroupName:      aws.String(nodeGroupName),
		NodeRole:           aws.String(d.Get("node_role_arn").(string)),
		Subnets:            flex.ExpandStringValueSet(d.Get("subnet_ids").(*schema.Set)),
		Tags:               getTagsIn(ctx),
	}

	if v, ok := d.GetOk("ami_type"); ok {
		input.AmiType = types.AMITypes(v.(string))
	}

	if v, ok := d.GetOk("capacity_type"); ok {
		input.CapacityType = types.CapacityTypes(v.(string))
	}

	if v, ok := d.GetOk("disk_size"); ok {
		input.DiskSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("instance_types"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InstanceTypes = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v := d.Get("labels").(map[string]interface{}); len(v) > 0 {
		input.Labels = flex.ExpandStringValueMap(v)
	}

	if v := d.Get("launch_template").([]interface{}); len(v) > 0 {
		input.LaunchTemplate = expandLaunchTemplateSpecification(v)
	}

	if v, ok := d.GetOk("release_version"); ok {
		input.ReleaseVersion = aws.String(v.(string))
	}

	if v := d.Get("remote_access").([]interface{}); len(v) > 0 {
		input.RemoteAccess = expandRemoteAccessConfig(v)
	}

	if v, ok := d.GetOk("scaling_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ScalingConfig = expandNodegroupScalingConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("taint"); ok && v.(*schema.Set).Len() > 0 {
		input.Taints = expandTaints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("update_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.UpdateConfig = expandNodegroupUpdateConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("version"); ok {
		input.Version = aws.String(v.(string))
	}

	_, err := conn.CreateNodegroup(ctx, input)

	if err != nil {
		return diag.Errorf("creating EKS Node Group (%s): %s", groupID, err)
	}

	d.SetId(groupID)

	if _, err := waitNodegroupCreated(ctx, conn, clusterName, nodeGroupName, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for EKS Node Group (%s) create: %s", d.Id(), err)
	}

	return resourceNodeGroupRead(ctx, d, meta)
}

func resourceNodeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, nodeGroupName, err := NodeGroupParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	nodeGroup, err := findNodegroupByTwoPartKey(ctx, conn, clusterName, nodeGroupName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EKS Node Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading EKS Node Group (%s): %s", d.Id(), err)
	}

	d.Set("ami_type", nodeGroup.AmiType)
	d.Set("arn", nodeGroup.NodegroupArn)
	d.Set("capacity_type", nodeGroup.CapacityType)
	d.Set("cluster_name", nodeGroup.ClusterName)
	d.Set("disk_size", nodeGroup.DiskSize)
	d.Set("instance_types", nodeGroup.InstanceTypes)
	d.Set("labels", nodeGroup.Labels)
	if err := d.Set("launch_template", flattenLaunchTemplateSpecification(nodeGroup.LaunchTemplate)); err != nil {
		return diag.Errorf("setting launch_template: %s", err)
	}
	d.Set("node_group_name", nodeGroup.NodegroupName)
	d.Set("node_group_name_prefix", create.NamePrefixFromName(aws.ToString(nodeGroup.NodegroupName)))
	d.Set("node_role_arn", nodeGroup.NodeRole)
	d.Set("release_version", nodeGroup.ReleaseVersion)
	if err := d.Set("remote_access", flattenRemoteAccessConfig(nodeGroup.RemoteAccess)); err != nil {
		return diag.Errorf("setting remote_access: %s", err)
	}
	if err := d.Set("resources", flattenNodeGroupResources(nodeGroup.Resources)); err != nil {
		return diag.Errorf("setting resources: %s", err)
	}
	if nodeGroup.ScalingConfig != nil {
		if err := d.Set("scaling_config", []interface{}{flattenNodeGroupScalingConfig(nodeGroup.ScalingConfig)}); err != nil {
			return diag.Errorf("setting scaling_config: %s", err)
		}
	} else {
		d.Set("scaling_config", nil)
	}
	d.Set("status", nodeGroup.Status)
	d.Set("subnet_ids", nodeGroup.Subnets)
	if err := d.Set("taint", flattenTaints(nodeGroup.Taints)); err != nil {
		return diag.Errorf("setting taint: %s", err)
	}
	if nodeGroup.UpdateConfig != nil {
		if err := d.Set("update_config", []interface{}{flattenNodeGroupUpdateConfig(nodeGroup.UpdateConfig)}); err != nil {
			return diag.Errorf("setting update_config: %s", err)
		}
	} else {
		d.Set("update_config", nil)
	}
	d.Set("version", nodeGroup.Version)

	setTagsOut(ctx, nodeGroup.Tags)

	return nil
}

func resourceNodeGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, nodeGroupName, err := NodeGroupParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Do any version update first.
	if d.HasChanges("launch_template", "release_version", "version") {
		input := &eks.UpdateNodegroupVersionInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			ClusterName:        aws.String(clusterName),
			Force:              d.Get("force_update_version").(bool),
			NodegroupName:      aws.String(nodeGroupName),
		}

		if v := d.Get("launch_template").([]interface{}); len(v) > 0 {
			input.LaunchTemplate = expandLaunchTemplateSpecification(v)

			// When returning Launch Template information, the API returns all
			// fields. Since both the id and name are saved to the Terraform
			// state for drift detection and the API returns the following
			// error if both are present during update:
			// InvalidParameterException: Either provide launch template ID or launch template name in the request.

			// Remove the name if there are no changes, to prefer the ID.
			if input.LaunchTemplate.Id != nil && input.LaunchTemplate.Name != nil && !d.HasChange("launch_template.0.name") {
				input.LaunchTemplate.Name = nil
			}

			// Otherwise, remove the ID, but only if both are present still.
			if input.LaunchTemplate.Id != nil && input.LaunchTemplate.Name != nil && !d.HasChange("launch_template.0.id") {
				input.LaunchTemplate.Id = nil
			}
		}

		if v, ok := d.GetOk("release_version"); ok && d.HasChange("release_version") {
			input.ReleaseVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("version"); ok && d.HasChange("version") {
			input.Version = aws.String(v.(string))
		}

		output, err := conn.UpdateNodegroupVersion(ctx, input)

		if err != nil {
			return diag.Errorf("updating EKS Node Group (%s) version: %s", d.Id(), err)
		}

		updateID := aws.ToString(output.Update.Id)

		if _, err := waitNodegroupUpdateSuccessful(ctx, conn, clusterName, nodeGroupName, updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for EKS Node Group (%s) version update (%s): %s", d.Id(), updateID, err)
		}
	}

	if d.HasChanges("labels", "scaling_config", "taint", "update_config") {
		oldLabelsRaw, newLabelsRaw := d.GetChange("labels")
		oldTaintsRaw, newTaintsRaw := d.GetChange("taint")

		input := &eks.UpdateNodegroupConfigInput{
			ClientRequestToken: aws.String(id.UniqueId()),
			ClusterName:        aws.String(clusterName),
			Labels:             expandUpdateLabelsPayload(ctx, oldLabelsRaw, newLabelsRaw),
			NodegroupName:      aws.String(nodeGroupName),
			Taints:             expandUpdateTaintsPayload(oldTaintsRaw.(*schema.Set).List(), newTaintsRaw.(*schema.Set).List()),
		}

		if d.HasChange("scaling_config") {
			if v, ok := d.GetOk("scaling_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.ScalingConfig = expandNodegroupScalingConfig(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("update_config") {
			if v, ok := d.GetOk("update_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.UpdateConfig = expandNodegroupUpdateConfig(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		output, err := conn.UpdateNodegroupConfig(ctx, input)

		if err != nil {
			return diag.Errorf("updating EKS Node Group (%s) config: %s", d.Id(), err)
		}

		updateID := aws.ToString(output.Update.Id)

		if _, err := waitNodegroupUpdateSuccessful(ctx, conn, clusterName, nodeGroupName, updateID, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for EKS Node Group (%s) config update (%s): %s", d.Id(), updateID, err)
		}
	}

	return resourceNodeGroupRead(ctx, d, meta)
}

func resourceNodeGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)

	clusterName, nodeGroupName, err := NodeGroupParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting EKS Node Group: %s", d.Id())
	_, err = conn.DeleteNodegroup(ctx, &eks.DeleteNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting EKS Node Group (%s): %s", d.Id(), err)
	}

	if _, err := waitNodegroupDeleted(ctx, conn, clusterName, nodeGroupName, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for EKS Node Group (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func findNodegroupByTwoPartKey(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName string) (*types.Nodegroup, error) {
	input := &eks.DescribeNodegroupInput{
		ClusterName:   aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
	}

	output, err := conn.DescribeNodegroup(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Nodegroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Nodegroup, nil
}

func findNodegroupUpdateByThreePartKey(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName, id string) (*types.Update, error) {
	input := &eks.DescribeUpdateInput{
		Name:          aws.String(clusterName),
		NodegroupName: aws.String(nodeGroupName),
		UpdateId:      aws.String(id),
	}

	output, err := conn.DescribeUpdate(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Update == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Update, nil
}

func statusNodegroup(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNodegroupByTwoPartKey(ctx, conn, clusterName, nodeGroupName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusNodegroupUpdate(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNodegroupUpdateByThreePartKey(ctx, conn, clusterName, nodeGroupName, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitNodegroupCreated(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName string, timeout time.Duration) (*types.Nodegroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.NodegroupStatusCreating),
		Target:  enum.Slice(types.NodegroupStatusActive),
		Refresh: statusNodegroup(ctx, conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Nodegroup); ok {
		if status, health := output.Status, output.Health; status == types.NodegroupStatusCreateFailed && health != nil {
			tfresource.SetLastError(err, issuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitNodegroupDeleted(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName string, timeout time.Duration) (*types.Nodegroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.NodegroupStatusActive, types.NodegroupStatusDeleting),
		Target:  []string{},
		Refresh: statusNodegroup(ctx, conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Nodegroup); ok {
		if status, health := output.Status, output.Health; status == types.NodegroupStatusDeleteFailed && health != nil {
			tfresource.SetLastError(err, issuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitNodegroupUpdateSuccessful(ctx context.Context, conn *eks.Client, clusterName, nodeGroupName, id string, timeout time.Duration) (*types.Update, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.UpdateStatusInProgress),
		Target:  enum.Slice(types.UpdateStatusSuccessful),
		Refresh: statusNodegroupUpdate(ctx, conn, clusterName, nodeGroupName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Update); ok {
		if status := output.Status; status == types.UpdateStatusCancelled || status == types.UpdateStatusFailed {
			tfresource.SetLastError(err, errorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func issueError(apiObject types.Issue) error {
	return fmt.Errorf("%s: %s", apiObject.Code, aws.ToString(apiObject.Message))
}

func issuesError(apiObjects []types.Issue) error {
	var errs []error

	for _, apiObject := range apiObjects {
		err := issueError(apiObject)

		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", strings.Join(apiObject.ResourceIds, ", "), err))
		}
	}

	return errors.Join(errs...)
}

func expandLaunchTemplateSpecification(l []interface{}) *types.LaunchTemplateSpecification {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &types.LaunchTemplateSpecification{}

	if v, ok := m["id"].(string); ok && v != "" {
		config.Id = aws.String(v)
	}

	if v, ok := m["name"].(string); ok && v != "" {
		config.Name = aws.String(v)
	}

	if v, ok := m["version"].(string); ok && v != "" {
		config.Version = aws.String(v)
	}

	return config
}

func expandNodegroupScalingConfig(tfMap map[string]interface{}) *types.NodegroupScalingConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.NodegroupScalingConfig{}

	if v, ok := tfMap["desired_size"].(int); ok {
		apiObject.DesiredSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_size"].(int); ok && v != 0 {
		apiObject.MaxSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_size"].(int); ok {
		apiObject.MinSize = aws.Int32(int32(v))
	}

	return apiObject
}

func expandTaints(l []interface{}) []types.Taint {
	if len(l) == 0 {
		return nil
	}

	var taints []types.Taint

	for _, raw := range l {
		t, ok := raw.(map[string]interface{})

		if !ok {
			continue
		}

		taint := types.Taint{}

		if k, ok := t["key"].(string); ok {
			taint.Key = aws.String(k)
		}

		if v, ok := t["value"].(string); ok {
			taint.Value = aws.String(v)
		}

		if e, ok := t["effect"].(string); ok {
			taint.Effect = types.TaintEffect(e)
		}

		taints = append(taints, taint)
	}

	return taints
}

func expandUpdateTaintsPayload(oldTaintsRaw, newTaintsRaw []interface{}) *types.UpdateTaintsPayload {
	oldTaints := expandTaints(oldTaintsRaw)
	newTaints := expandTaints(newTaintsRaw)

	var removedTaints []types.Taint
	for _, ot := range oldTaints {
		removed := true
		for _, nt := range newTaints {
			// If both taint.key and taint.effect are the same, we don't need to remove it.
			if aws.ToString(nt.Key) == aws.ToString(ot.Key) && nt.Effect == ot.Effect {
				removed = false
				break
			}
		}

		if removed {
			removedTaints = append(removedTaints, ot)
		}
	}

	var updatedTaints []types.Taint
	for _, nt := range newTaints {
		updated := true
		for _, ot := range oldTaints {
			if reflect.DeepEqual(nt, ot) {
				updated = false
				break
			}
		}

		if updated {
			updatedTaints = append(updatedTaints, nt)
		}
	}

	if len(removedTaints) == 0 && len(updatedTaints) == 0 {
		return nil
	}

	updateTaintsPayload := &types.UpdateTaintsPayload{}

	if len(removedTaints) > 0 {
		updateTaintsPayload.RemoveTaints = removedTaints
	}

	if len(updatedTaints) > 0 {
		updateTaintsPayload.AddOrUpdateTaints = updatedTaints
	}

	return updateTaintsPayload
}

func expandRemoteAccessConfig(l []interface{}) *types.RemoteAccessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &types.RemoteAccessConfig{}

	if v, ok := m["ec2_ssh_key"].(string); ok && v != "" {
		config.Ec2SshKey = aws.String(v)
	}

	if v, ok := m["source_security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.SourceSecurityGroups = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandNodegroupUpdateConfig(tfMap map[string]interface{}) *types.NodegroupUpdateConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.NodegroupUpdateConfig{}

	if v, ok := tfMap["max_unavailable"].(int); ok && v != 0 {
		apiObject.MaxUnavailable = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_unavailable_percentage"].(int); ok && v != 0 {
		apiObject.MaxUnavailablePercentage = aws.Int32(int32(v))
	}

	return apiObject
}

func expandUpdateLabelsPayload(ctx context.Context, oldLabelsMap, newLabelsMap interface{}) *types.UpdateLabelsPayload {
	// EKS Labels operate similarly to keyvaluetags
	oldLabels := tftags.New(ctx, oldLabelsMap)
	newLabels := tftags.New(ctx, newLabelsMap)

	removedLabels := oldLabels.Removed(newLabels)
	updatedLabels := oldLabels.Updated(newLabels)

	if len(removedLabels) == 0 && len(updatedLabels) == 0 {
		return nil
	}

	updateLabelsPayload := &types.UpdateLabelsPayload{}

	if len(removedLabels) > 0 {
		updateLabelsPayload.RemoveLabels = removedLabels.Keys()
	}

	if len(updatedLabels) > 0 {
		updateLabelsPayload.AddOrUpdateLabels = updatedLabels.Map()
	}

	return updateLabelsPayload
}

func flattenAutoScalingGroups(autoScalingGroups []types.AutoScalingGroup) []map[string]interface{} {
	if len(autoScalingGroups) == 0 {
		return []map[string]interface{}{}
	}

	l := make([]map[string]interface{}, 0, len(autoScalingGroups))

	for _, autoScalingGroup := range autoScalingGroups {
		m := map[string]interface{}{
			"name": aws.ToString(autoScalingGroup.Name),
		}

		l = append(l, m)
	}

	return l
}

func flattenLaunchTemplateSpecification(config *types.LaunchTemplateSpecification) []map[string]interface{} {
	if config == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := config.Id; v != nil {
		m["id"] = aws.ToString(v)
	}

	if v := config.Name; v != nil {
		m["name"] = aws.ToString(v)
	}

	if v := config.Version; v != nil {
		m["version"] = aws.ToString(v)
	}

	return []map[string]interface{}{m}
}

func flattenNodeGroupResources(resources *types.NodegroupResources) []map[string]interface{} {
	if resources == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"autoscaling_groups":              flattenAutoScalingGroups(resources.AutoScalingGroups),
		"remote_access_security_group_id": aws.ToString(resources.RemoteAccessSecurityGroup),
	}

	return []map[string]interface{}{m}
}

func flattenNodeGroupScalingConfig(apiObject *types.NodegroupScalingConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DesiredSize; v != nil {
		tfMap["desired_size"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxSize; v != nil {
		tfMap["max_size"] = aws.ToInt32(v)
	}

	if v := apiObject.MinSize; v != nil {
		tfMap["min_size"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenNodeGroupUpdateConfig(apiObject *types.NodegroupUpdateConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MaxUnavailable; v != nil {
		tfMap["max_unavailable"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxUnavailablePercentage; v != nil {
		tfMap["max_unavailable_percentage"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenRemoteAccessConfig(config *types.RemoteAccessConfig) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"ec2_ssh_key":               aws.ToString(config.Ec2SshKey),
		"source_security_group_ids": config.SourceSecurityGroups,
	}

	return []map[string]interface{}{m}
}

func flattenTaints(taints []types.Taint) []interface{} {
	if len(taints) == 0 {
		return nil
	}

	var results []interface{}

	for _, taint := range taints {
		t := make(map[string]interface{})
		t["key"] = aws.ToString(taint.Key)
		t["value"] = aws.ToString(taint.Value)
		t["effect"] = taint.Effect

		results = append(results, t)
	}
	return results
}
