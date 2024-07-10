// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/oam"
	"github.com/aws/aws-sdk-go-v2/service/oam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_oam_link", name="Link")
// @Tags(identifierAttribute="id")
func ResourceLink() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLinkCreate,
		ReadWithoutTimeout:   resourceLinkRead,
		UpdateWithoutTimeout: resourceLinkUpdate,
		DeleteWithoutTimeout: resourceLinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
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
					ValidateDiagFunc: enum.Validate[types.ResourceType](),
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	ResNameLink = "Link"
)

func resourceLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	in := &oam.CreateLinkInput{
		LabelTemplate:     aws.String(d.Get("label_template").(string)),
		LinkConfiguration: expandLinkConfiguration(d.Get("link_configuration").([]interface{})),
		ResourceTypes:     flex.ExpandStringyValueSet[types.ResourceType](d.Get("resource_types").(*schema.Set)),
		SinkIdentifier:    aws.String(d.Get("sink_identifier").(string)),
		Tags:              getTagsIn(ctx),
	}

	out, err := conn.CreateLink(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionCreating, ResNameLink, d.Get("sink_identifier").(string), err)
	}

	if out == nil || out.Id == nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionCreating, ResNameLink, d.Get("sink_identifier").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Arn))

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	out, err := findLinkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ObservabilityAccessManager Link (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, ResNameLink, d.Id(), err)
	}

	d.Set(names.AttrARN, out.Arn)
	d.Set("label", out.Label)
	d.Set("label_template", out.LabelTemplate)
	d.Set("link_configuration", flattenLinkConfiguration(out.LinkConfiguration))
	d.Set("link_id", out.Id)
	d.Set("resource_types", flex.FlattenStringValueList(out.ResourceTypes))
	d.Set("sink_arn", out.SinkArn)
	d.Set("sink_identifier", out.SinkArn)

	return nil
}

func resourceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	update := false

	in := &oam.UpdateLinkInput{
		Identifier: aws.String(d.Id()),
	}

	if d.HasChanges("resource_types", "link_configuration") {
		in.ResourceTypes = flex.ExpandStringyValueSet[types.ResourceType](d.Get("resource_types").(*schema.Set))

		if d.HasChanges("link_configuration") {
			in.LinkConfiguration = expandLinkConfiguration(d.Get("link_configuration").([]interface{}))
		}

		update = true
	}

	if update {
		log.Printf("[DEBUG] Updating ObservabilityAccessManager Link (%s): %#v", d.Id(), in)
		_, err := conn.UpdateLink(ctx, in)
		if err != nil {
			return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionUpdating, ResNameLink, d.Id(), err)
		}
	}

	return append(diags, resourceLinkRead(ctx, d, meta)...)
}

func resourceLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	log.Printf("[INFO] Deleting ObservabilityAccessManager Link %s", d.Id())

	_, err := conn.DeleteLink(ctx, &oam.DeleteLinkInput{
		Identifier: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionDeleting, ResNameLink, d.Id(), err)
	}

	return nil
}

func findLinkByID(ctx context.Context, conn *oam.Client, id string) (*oam.GetLinkOutput, error) {
	in := &oam.GetLinkInput{
		Identifier: aws.String(id),
	}
	out, err := conn.GetLink(ctx, in)
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

	if out == nil || out.Arn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func expandLinkConfiguration(l []interface{}) *types.LinkConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	config := &types.LinkConfiguration{}

	m := l[0].(map[string]interface{})
	if v, ok := m["log_group_configuration"]; ok {
		config.LogGroupConfiguration = expandLogGroupConfiguration(v.([]interface{}))
	}
	if v, ok := m["metric_configuration"]; ok {
		config.MetricConfiguration = expandMetricConfiguration(v.([]interface{}))
	}

	return config
}

func expandLogGroupConfiguration(l []interface{}) *types.LogGroupConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	config := &types.LogGroupConfiguration{}

	m := l[0].(map[string]interface{})
	if v, ok := m[names.AttrFilter]; ok && v != "" {
		config.Filter = aws.String(v.(string))
	}

	return config
}

func expandMetricConfiguration(l []interface{}) *types.MetricConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	config := &types.MetricConfiguration{}

	m := l[0].(map[string]interface{})
	if v, ok := m[names.AttrFilter]; ok && v != "" {
		config.Filter = aws.String(v.(string))
	}

	return config
}

func flattenLinkConfiguration(a *types.LinkConfiguration) []interface{} {
	if a == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}

	if a.LogGroupConfiguration != nil {
		m["log_group_configuration"] = flattenLogGroupConfiguration(a.LogGroupConfiguration)
	}
	if a.MetricConfiguration != nil {
		m["metric_configuration"] = flattenMetricConfiguration(a.MetricConfiguration)
	}

	return []interface{}{m}
}

func flattenLogGroupConfiguration(a *types.LogGroupConfiguration) []interface{} {
	if a == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}

	if a.Filter != nil {
		m[names.AttrFilter] = aws.ToString(a.Filter)
	}

	return []interface{}{m}
}

func flattenMetricConfiguration(a *types.MetricConfiguration) []interface{} {
	if a == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{}

	if a.Filter != nil {
		m[names.AttrFilter] = aws.ToString(a.Filter)
	}

	return []interface{}{m}
}
