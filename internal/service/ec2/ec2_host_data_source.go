// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_host", name="Host")
// @Tags
// @Testing(tagsTest=false)
func dataSourceHost() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHostRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"allocation_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"allows_multiple_instance_types": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"asset_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"auto_placement": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"available_capacity": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"available_instance_capacity": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"available_capacity": {
											Type:     schema.TypeInt,
											Computed: true,
										},
										names.AttrInstanceType: {
											Type:     schema.TypeString,
											Computed: true,
										},
										"total_capacity": {
											Type:     schema.TypeInt,
											Computed: true,
										},
									},
								},
							},
							"available_vcpus": {
								Type:     schema.TypeInt,
								Computed: true,
							},
						},
					},
				},
				names.AttrAvailabilityZone: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"availability_zone_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"cores": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				names.AttrFilter: customFiltersSchema(),
				"host_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"host_maintenance": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"host_recovery": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"host_reservation_id": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"instance_family": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrInstanceType: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"instances": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrInstanceID: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrInstanceType: {
								Type:     schema.TypeString,
								Computed: true,
							},
							names.AttrOwnerID: {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"member_of_service_linked_resource_group": {
					Type:     schema.TypeBool,
					Computed: true,
				},
				names.AttrOutpostARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrOwnerID: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"release_time": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"sockets": {
					Type:     schema.TypeInt,
					Computed: true,
				},
				names.AttrState: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
				"total_vcpus": {
					Type:     schema.TypeInt,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceHostRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.EC2Client(ctx)

	input := ec2.DescribeHostsInput{
		Filter: newCustomFilterList(d.Get(names.AttrFilter).(*schema.Set)),
	}

	if v, ok := d.GetOk("host_id"); ok {
		input.HostIds = []string{v.(string)}
	}

	if len(input.Filter) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filter = nil
	}

	host, err := findHost(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Host", err))
	}

	d.SetId(aws.ToString(host.HostId))

	if host.AllocationTime != nil {
		d.Set("allocation_time", aws.ToTime(host.AllocationTime).Format(time.RFC3339))
	}
	d.Set("allows_multiple_instance_types", host.AllowsMultipleInstanceTypes)
	d.Set(names.AttrARN, hostARN(ctx, c, aws.ToString(host.OwnerId), d.Id()))
	d.Set("asset_id", host.AssetId)
	d.Set("auto_placement", host.AutoPlacement)
	if err := d.Set("available_capacity", flattenAvailableCapacity(host.AvailableCapacity)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting available_capacity: %s", err)
	}
	d.Set(names.AttrAvailabilityZone, host.AvailabilityZone)
	d.Set("availability_zone_id", host.AvailabilityZoneId)
	d.Set("cores", host.HostProperties.Cores)
	d.Set("host_id", host.HostId)
	d.Set("host_maintenance", host.HostMaintenance)
	d.Set("host_recovery", host.HostRecovery)
	d.Set("host_reservation_id", host.HostReservationId)
	d.Set("instance_family", host.HostProperties.InstanceFamily)
	d.Set(names.AttrInstanceType, host.HostProperties.InstanceType)
	if err := d.Set("instances", flattenHostInstances(host.Instances)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instances: %s", err)
	}
	d.Set("member_of_service_linked_resource_group", host.MemberOfServiceLinkedResourceGroup)
	d.Set(names.AttrOutpostARN, host.OutpostArn)
	d.Set(names.AttrOwnerID, host.OwnerId)
	if host.ReleaseTime != nil {
		d.Set("release_time", aws.ToTime(host.ReleaseTime).Format(time.RFC3339))
	}
	d.Set("sockets", host.HostProperties.Sockets)
	d.Set(names.AttrState, host.State)
	d.Set("total_vcpus", host.HostProperties.TotalVCpus)

	setTagsOut(ctx, host.Tags)

	return diags
}

func flattenHostInstances(instances []awstypes.HostInstance) []map[string]any {
	if len(instances) == 0 {
		return nil
	}

	result := make([]map[string]any, len(instances))
	for i, inst := range instances {
		result[i] = map[string]any{
			names.AttrInstanceID:   aws.ToString(inst.InstanceId),
			names.AttrInstanceType: aws.ToString(inst.InstanceType),
			names.AttrOwnerID:      aws.ToString(inst.OwnerId),
		}
	}
	return result
}

func flattenAvailableCapacity(capacity *awstypes.AvailableCapacity) []map[string]any {
	if capacity == nil {
		return nil
	}

	result := map[string]any{
		"available_vcpus":             int(aws.ToInt32(capacity.AvailableVCpus)),
		"available_instance_capacity": flattenInstanceCapacity(capacity.AvailableInstanceCapacity),
	}
	return []map[string]any{result}
}

func flattenInstanceCapacity(capacities []awstypes.InstanceCapacity) []map[string]any {
	if len(capacities) == 0 {
		return nil
	}

	result := make([]map[string]any, len(capacities))
	for i, cap := range capacities {
		result[i] = map[string]any{
			"available_capacity":   int(aws.ToInt32(cap.AvailableCapacity)),
			names.AttrInstanceType: aws.ToString(cap.InstanceType),
			"total_capacity":       int(aws.ToInt32(cap.TotalCapacity)),
		}
	}
	return result
}
