// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package oam

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/oam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_oam_link", name="Link")
// @Tags(identifierAttribute="arn")
func resourceLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLinkCreate,
		ReadWithoutTimeout:   resourceLinkRead,
		UpdateWithoutTimeout: resourceLinkUpdate,
		DeleteWithoutTimeout: resourceLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			// These aren't used but are retained for backwards compatibility.
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"label_template": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"link_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrFilter: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 2000),
									},
								},
							},
						},
						"metric_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrFilter: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 2000),
									},
								},
							},
						},
					},
				},
			},
			"link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_types": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.ResourceType](),
				},
			},
			"sink_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sink_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceLinkCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	in := oam.CreateLinkInput{
		LabelTemplate:     aws.String(d.Get("label_template").(string)),
		LinkConfiguration: expandLinkConfiguration(d.Get("link_configuration").([]any)),
		ResourceTypes:     flex.ExpandStringyValueSet[awstypes.ResourceType](d.Get("resource_types").(*schema.Set)),
		SinkIdentifier:    aws.String(d.Get("sink_identifier").(string)),
		Tags:              getTagsIn(ctx),
	}

	out, err := conn.CreateLink(ctx, &in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ObservabilityAccessManager Link: %s", err)
	}

	d.SetId(aws.ToString(out.Arn))

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	out, err := findLinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ObservabilityAccessManager Link (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("label", out.Label)
	d.Set("label_template", out.LabelTemplate)
	d.Set("link_configuration", flattenLinkConfiguration(out.LinkConfiguration))
	d.Set("link_id", out.Id)
	d.Set("resource_types", out.ResourceTypes)
	d.Set("sink_arn", out.SinkArn)
	d.Set("sink_identifier", out.SinkArn)

	return diags
}

func resourceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		in := oam.UpdateLinkInput{
			Identifier:    aws.String(d.Id()),
			ResourceTypes: flex.ExpandStringyValueSet[awstypes.ResourceType](d.Get("resource_types").(*schema.Set)),
		}

		if d.HasChanges("link_configuration") {
			in.LinkConfiguration = expandLinkConfiguration(d.Get("link_configuration").([]any))
		}

		_, err := conn.UpdateLink(ctx, &in)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating ObservabilityAccessManager Link (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	log.Printf("[INFO] Deleting ObservabilityAccessManager Link: %s", d.Id())
	in := oam.DeleteLinkInput{
		Identifier: aws.String(d.Id()),
	}
	_, err := conn.DeleteLink(ctx, &in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ObservabilityAccessManager Link (%s): %s", d.Id(), err)
	}

	return diags
}

func findLinkByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetLinkOutput, error) {
	in := oam.GetLinkInput{
		Identifier: aws.String(id),
	}

	return findLink(ctx, conn, &in)
}

func findLink(ctx context.Context, conn *oam.Client, input *oam.GetLinkInput) (*oam.GetLinkOutput, error) {
	output, err := conn.GetLink(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Arn == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func expandLinkConfiguration(tfList []any) *awstypes.LinkConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.LinkConfiguration{}
	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap["log_group_configuration"]; ok {
		apiObject.LogGroupConfiguration = expandLogGroupConfiguration(v.([]any))
	}
	if v, ok := tfMap["metric_configuration"]; ok {
		apiObject.MetricConfiguration = expandMetricConfiguration(v.([]any))
	}

	return apiObject
}

func expandLogGroupConfiguration(tfList []any) *awstypes.LogGroupConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.LogGroupConfiguration{}
	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap[names.AttrFilter]; ok && v != "" {
		apiObject.Filter = aws.String(v.(string))
	}

	return apiObject
}

func expandMetricConfiguration(tfList []any) *awstypes.MetricConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &awstypes.MetricConfiguration{}
	tfMap := tfList[0].(map[string]any)

	if v, ok := tfMap[names.AttrFilter]; ok && v != "" {
		apiObject.Filter = aws.String(v.(string))
	}

	return apiObject
}

func flattenLinkConfiguration(apiObject *awstypes.LinkConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.LogGroupConfiguration != nil {
		tfMap["log_group_configuration"] = flattenLogGroupConfiguration(apiObject.LogGroupConfiguration)
	}
	if apiObject.MetricConfiguration != nil {
		tfMap["metric_configuration"] = flattenMetricConfiguration(apiObject.MetricConfiguration)
	}

	return []any{tfMap}
}

func flattenLogGroupConfiguration(apiObject *awstypes.LogGroupConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Filter != nil {
		tfMap[names.AttrFilter] = aws.ToString(apiObject.Filter)
	}

	return []any{tfMap}
}

func flattenMetricConfiguration(apiObject *awstypes.MetricConfiguration) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Filter != nil {
		tfMap[names.AttrFilter] = aws.ToString(apiObject.Filter)
	}

	return []any{tfMap}
}
