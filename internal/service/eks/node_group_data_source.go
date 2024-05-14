// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_eks_node_group")
func dataSourceNodeGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceNodeGroupRead,

		Schema: map[string]*schema.Schema{
			"ami_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrClusterName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"disk_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"instance_types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrLaunchTemplate: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVersion: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"node_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"node_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"release_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"remote_access": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ec2_ssh_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"source_security_group_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			names.AttrResources: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"autoscaling_groups": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
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
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"min_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"taints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"effect": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceNodeGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EKSClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	clusterName := d.Get(names.AttrClusterName).(string)
	nodeGroupName := d.Get("node_group_name").(string)
	id := NodeGroupCreateResourceID(clusterName, nodeGroupName)
	nodeGroup, err := findNodegroupByTwoPartKey(ctx, conn, clusterName, nodeGroupName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EKS Node Group (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("ami_type", nodeGroup.AmiType)
	d.Set(names.AttrARN, nodeGroup.NodegroupArn)
	d.Set("capacity_type", nodeGroup.CapacityType)
	d.Set(names.AttrClusterName, nodeGroup.ClusterName)
	d.Set("disk_size", nodeGroup.DiskSize)
	d.Set("instance_types", nodeGroup.InstanceTypes)
	d.Set("labels", nodeGroup.Labels)
	if err := d.Set(names.AttrLaunchTemplate, flattenLaunchTemplateSpecification(nodeGroup.LaunchTemplate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting launch_template: %s", err)
	}
	d.Set("node_group_name", nodeGroup.NodegroupName)
	d.Set("node_role_arn", nodeGroup.NodeRole)
	d.Set("release_version", nodeGroup.ReleaseVersion)
	if err := d.Set("remote_access", flattenRemoteAccessConfig(nodeGroup.RemoteAccess)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting remote_access: %s", err)
	}
	if err := d.Set(names.AttrResources, flattenNodeGroupResources(nodeGroup.Resources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resources: %s", err)
	}
	if nodeGroup.ScalingConfig != nil {
		if err := d.Set("scaling_config", []interface{}{flattenNodeGroupScalingConfig(nodeGroup.ScalingConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting scaling_config: %s", err)
		}
	} else {
		d.Set("scaling_config", nil)
	}
	d.Set(names.AttrStatus, nodeGroup.Status)
	d.Set(names.AttrSubnetIDs, nodeGroup.Subnets)
	if err := d.Set("taints", flattenTaints(nodeGroup.Taints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting taints: %s", err)
	}
	d.Set(names.AttrVersion, nodeGroup.Version)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, nodeGroup.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
