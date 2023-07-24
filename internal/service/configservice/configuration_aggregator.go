// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_configuration_aggregator", name="Configuration Aggregator")
// @Tags(identifierAttribute="arn")
func ResourceConfigurationAggregator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationAggregatorPut,
		ReadWithoutTimeout:   resourceConfigurationAggregatorRead,
		UpdateWithoutTimeout: resourceConfigurationAggregatorPut,
		DeleteWithoutTimeout: resourceConfigurationAggregatorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			// This is to prevent this error:
			// All fields are ForceNew or Computed w/out Optional, Update is superfluous
			customdiff.ForceNewIfChange("account_aggregation_source", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 0 && len(new.([]interface{})) > 0
			}),
			customdiff.ForceNewIfChange("organization_aggregation_source", func(_ context.Context, old, new, meta interface{}) bool {
				return len(old.([]interface{})) == 0 && len(new.([]interface{})) > 0
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"account_aggregation_source": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"organization_aggregation_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_ids": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidAccountID,
							},
						},
						"all_regions": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"organization_aggregation_source": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"account_aggregation_source"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_regions": {
							Type:     schema.TypeBool,
							Default:  false,
							Optional: true,
						},
						"regions": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceConfigurationAggregatorPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	input := &configservice.PutConfigurationAggregatorInput{
		ConfigurationAggregatorName: aws.String(d.Get("name").(string)),
		Tags:                        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("account_aggregation_source"); ok && len(v.([]interface{})) > 0 {
		input.AccountAggregationSources = expandAccountAggregationSources(v.([]interface{}))
	}

	if v, ok := d.GetOk("organization_aggregation_source"); ok && len(v.([]interface{})) > 0 {
		input.OrganizationAggregationSource = expandOrganizationAggregationSource(v.([]interface{})[0].(map[string]interface{}))
	}

	resp, err := conn.PutConfigurationAggregatorWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating aggregator: %s", err)
	}

	configAgg := resp.ConfigurationAggregator
	d.SetId(aws.StringValue(configAgg.ConfigurationAggregatorName))

	return append(diags, resourceConfigurationAggregatorRead(ctx, d, meta)...)
}

func resourceConfigurationAggregatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	req := &configservice.DescribeConfigurationAggregatorsInput{
		ConfigurationAggregatorNames: []*string{aws.String(d.Id())},
	}

	res, err := conn.DescribeConfigurationAggregatorsWithContext(ctx, req)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationAggregatorException) {
		create.LogNotFoundRemoveState(names.ConfigService, create.ErrActionReading, ResNameConfigurationAggregator, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameConfigurationAggregator, d.Id(), err)
	}

	if !d.IsNewResource() && (res == nil || len(res.ConfigurationAggregators) == 0) {
		log.Printf("[WARN] No aggregators returned (%s), removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if d.IsNewResource() && (res == nil || len(res.ConfigurationAggregators) == 0) {
		return create.DiagError(names.ConfigService, create.ErrActionReading, ResNameConfigurationAggregator, d.Id(), errors.New("not found after creation"))
	}

	aggregator := res.ConfigurationAggregators[0]
	arn := aws.StringValue(aggregator.ConfigurationAggregatorArn)
	d.Set("arn", arn)
	d.Set("name", aggregator.ConfigurationAggregatorName)

	if err := d.Set("account_aggregation_source", flattenAccountAggregationSources(aggregator.AccountAggregationSources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting account_aggregation_source: %s", err)
	}

	if err := d.Set("organization_aggregation_source", flattenOrganizationAggregationSource(aggregator.OrganizationAggregationSource)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting organization_aggregation_source: %s", err)
	}

	return diags
}

func resourceConfigurationAggregatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	req := &configservice.DeleteConfigurationAggregatorInput{
		ConfigurationAggregatorName: aws.String(d.Id()),
	}
	_, err := conn.DeleteConfigurationAggregatorWithContext(ctx, req)

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationAggregatorException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Config Configuration Aggregator (%s): %s", d.Id(), err)
	}

	return diags
}
