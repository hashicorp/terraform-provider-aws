// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_instance", name="Instance")
// @Tags
func dataSourceInstance() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstanceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_resolve_best_voices_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"contact_flow_logs_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"contact_lens_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"early_media_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"identity_management_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inbound_calls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"instance_alias": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"instance_alias", names.AttrInstanceID},
			},
			names.AttrInstanceID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrInstanceID, "instance_alias"},
			},
			"multi_party_conference_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"outbound_calls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrServiceRole: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			// "use_custom_tts_voices_enabled": {
			// 	Type:     schema.TypeBool,
			// 	Computed: true,
			// },
		},
	}
}

func dataSourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	var matchedInstance *awstypes.Instance

	if v, ok := d.GetOk(names.AttrInstanceID); ok {
		instanceID := v.(string)
		instance, err := findInstanceByID(ctx, conn, instanceID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Instance (%s): %s", instanceID, err)
		}

		matchedInstance = instance
	} else if v, ok := d.GetOk("instance_alias"); ok {
		instanceAlias := v.(string)
		instanceSummary, err := findInstanceSummaryByAlias(ctx, conn, instanceAlias)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Instance (%s) summary: %s", instanceAlias, err)
		}

		matchedInstance = &awstypes.Instance{
			Arn:                    instanceSummary.Arn,
			CreatedTime:            instanceSummary.CreatedTime,
			Id:                     instanceSummary.Id,
			IdentityManagementType: instanceSummary.IdentityManagementType,
			InboundCallsEnabled:    instanceSummary.InboundCallsEnabled,
			InstanceAlias:          instanceSummary.InstanceAlias,
			InstanceStatus:         instanceSummary.InstanceStatus,
			OutboundCallsEnabled:   instanceSummary.OutboundCallsEnabled,
			ServiceRole:            instanceSummary.ServiceRole,
		}
	}

	d.SetId(aws.ToString(matchedInstance.Id))
	d.Set(names.AttrARN, matchedInstance.Arn)
	if matchedInstance.CreatedTime != nil {
		d.Set(names.AttrCreatedTime, matchedInstance.CreatedTime.Format(time.RFC3339))
	}
	d.Set("identity_management_type", matchedInstance.IdentityManagementType)
	d.Set("inbound_calls_enabled", matchedInstance.InboundCallsEnabled)
	d.Set("instance_alias", matchedInstance.InstanceAlias)
	d.Set("outbound_calls_enabled", matchedInstance.OutboundCallsEnabled)
	d.Set(names.AttrServiceRole, matchedInstance.ServiceRole)
	d.Set(names.AttrStatus, matchedInstance.InstanceStatus)

	if err := readInstanceAttributes(ctx, conn, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	setTagsOut(ctx, matchedInstance.Tags)

	return diags
}

func findInstanceSummaryByAlias(ctx context.Context, conn *connect.Client, alias string) (*awstypes.InstanceSummary, error) {
	const maxResults = 10
	input := &connect.ListInstancesInput{
		MaxResults: aws.Int32(maxResults),
	}

	return findInstanceSummary(ctx, conn, input, func(v *awstypes.InstanceSummary) bool {
		return aws.ToString(v.InstanceAlias) == alias
	})
}

func findInstanceSummary(ctx context.Context, conn *connect.Client, input *connect.ListInstancesInput, filter tfslices.Predicate[*awstypes.InstanceSummary]) (*awstypes.InstanceSummary, error) {
	output, err := findInstanceSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceSummaries(ctx context.Context, conn *connect.Client, input *connect.ListInstancesInput, filter tfslices.Predicate[*awstypes.InstanceSummary]) ([]awstypes.InstanceSummary, error) {
	var output []awstypes.InstanceSummary

	pages := connect.NewListInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.InstanceSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
