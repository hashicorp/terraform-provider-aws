// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice"
	"github.com/aws/aws-sdk-go-v2/service/vpclattice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @SDKDataSource("aws_vpclattice_listener", name="Listener")
func DataSourceListener() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceListenerRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_action": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_response": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status_code": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"forward": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"target_groups": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_group_identifier": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"weight": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"last_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"listener_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"protocol": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameListener = "Listener Data Source"
)

func dataSourceListenerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).VPCLatticeClient(ctx)

	serviceId := d.Get("service_identifier").(string)
	listenerId := d.Get("listener_identifier").(string)

	out, err := findListenerByListenerIdAndServiceId(ctx, conn, listenerId, serviceId)
	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameListener, d.Id(), err)
	}

	// Set simple arguments
	d.SetId(aws.ToString(out.Id))
	d.Set("arn", out.Arn)
	d.Set("created_at", aws.ToTime(out.CreatedAt).String())
	d.Set("last_updated_at", aws.ToTime(out.LastUpdatedAt).String())
	d.Set("listener_id", out.Id)
	d.Set("name", out.Name)
	d.Set("port", out.Port)
	d.Set("protocol", out.Protocol)
	d.Set("service_arn", out.ServiceArn)
	d.Set("service_id", out.ServiceId)

	// Flatten complex default_action attribute - uses flatteners from listener.go
	if err := d.Set("default_action", flattenListenerRuleActionsDataSource(out.DefaultAction)); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, DSNameListener, d.Id(), err)
	}

	// Set tags
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags, err := listTags(ctx, conn, aws.ToString(out.Arn))

	if err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionReading, DSNameListener, d.Id(), err)
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.DiagError(names.VPCLattice, create.ErrActionSetting, DSNameListener, d.Id(), err)
	}

	return nil
}

func findListenerByListenerIdAndServiceId(ctx context.Context, conn *vpclattice.Client, listener_id string, service_id string) (*vpclattice.GetListenerOutput, error) {
	in := &vpclattice.GetListenerInput{
		ListenerIdentifier: aws.String(listener_id),
		ServiceIdentifier:  aws.String(service_id),
	}

	out, err := conn.GetListener(ctx, in)
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

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenListenerRuleActionsDataSource(config types.RuleAction) []interface{} {
	m := map[string]interface{}{}

	if config == nil {
		return []interface{}{}
	}

	switch v := config.(type) {
	case *types.RuleActionMemberFixedResponse:
		m["fixed_response"] = flattenRuleActionMemberFixedResponseDataSource(&v.Value)
	case *types.RuleActionMemberForward:
		m["forward"] = flattenComplexDefaultActionForwardDataSource(&v.Value)
	}

	return []interface{}{m}
}

// Flatten function for fixed_response action
func flattenRuleActionMemberFixedResponseDataSource(response *types.FixedResponseAction) []interface{} {
	tfMap := map[string]interface{}{}

	if v := response.StatusCode; v != nil {
		tfMap["status_code"] = aws.ToInt32(v)
	}

	return []interface{}{tfMap}
}

// Flatten function for forward action
func flattenComplexDefaultActionForwardDataSource(forwardAction *types.ForwardAction) []interface{} {
	if forwardAction == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"target_groups": flattenDefaultActionForwardTargetGroupsDataSource(forwardAction.TargetGroups),
	}

	return []interface{}{m}
}

// Flatten function for target_groups
func flattenDefaultActionForwardTargetGroupsDataSource(groups []types.WeightedTargetGroup) []interface{} {
	if len(groups) == 0 {
		return []interface{}{}
	}

	var targetGroups []interface{}

	for _, targetGroup := range groups {
		m := map[string]interface{}{
			"target_group_identifier": aws.ToString(targetGroup.TargetGroupIdentifier),
			"weight":                  aws.ToInt32(targetGroup.Weight),
		}
		targetGroups = append(targetGroups, m)
	}

	return targetGroups
}
