// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_config_configuration_aggregator", name="Configuration Aggregator")
// @Tags(identifierAttribute="arn")
func resourceConfigurationAggregator() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
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
						names.AttrRoleARN: {
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
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	if d.IsNewResource() || d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrName).(string)
		input := &configservice.PutConfigurationAggregatorInput{
			ConfigurationAggregatorName: aws.String(name),
			Tags:                        getTagsIn(ctx),
		}

		if v, ok := d.GetOk("account_aggregation_source"); ok && len(v.([]interface{})) > 0 {
			input.AccountAggregationSources = expandAccountAggregationSources(v.([]interface{}))
		}

		if v, ok := d.GetOk("organization_aggregation_source"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.OrganizationAggregationSource = expandOrganizationAggregationSource(v.([]interface{})[0].(map[string]interface{}))
		}

		output, err := conn.PutConfigurationAggregator(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting ConfigService Configuration Aggregator (%s): %s", name, err)
		}

		if d.IsNewResource() {
			d.SetId(aws.ToString(output.ConfigurationAggregator.ConfigurationAggregatorName))
		}
	}

	return append(diags, resourceConfigurationAggregatorRead(ctx, d, meta)...)
}

func resourceConfigurationAggregatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	aggregator, err := findConfigurationAggregatorByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Configuration Aggregator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Configuration Aggregator (%s): %s", d.Id(), err)
	}

	if err := d.Set("account_aggregation_source", flattenAccountAggregationSources(aggregator.AccountAggregationSources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting account_aggregation_source: %s", err)
	}
	d.Set(names.AttrARN, aggregator.ConfigurationAggregatorArn)
	d.Set(names.AttrName, aggregator.ConfigurationAggregatorName)
	if err := d.Set("organization_aggregation_source", flattenOrganizationAggregationSource(aggregator.OrganizationAggregationSource)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting organization_aggregation_source: %s", err)
	}

	return diags
}

func resourceConfigurationAggregatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceClient(ctx)

	log.Printf("[DEBUG] Deleting ConfigService Configuration Aggregator: %s", d.Id())
	_, err := conn.DeleteConfigurationAggregator(ctx, &configservice.DeleteConfigurationAggregatorInput{
		ConfigurationAggregatorName: aws.String(d.Id()),
	})

	if errs.IsA[*types.NoSuchConfigurationAggregatorException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Configuration Aggregator (%s): %s", d.Id(), err)
	}

	return diags
}

func findConfigurationAggregatorByName(ctx context.Context, conn *configservice.Client, name string) (*types.ConfigurationAggregator, error) {
	input := &configservice.DescribeConfigurationAggregatorsInput{
		ConfigurationAggregatorNames: []string{name},
	}

	return findConfigurationAggregator(ctx, conn, input)
}

func findConfigurationAggregator(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigurationAggregatorsInput) (*types.ConfigurationAggregator, error) {
	output, err := findConfigurationAggregators(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigurationAggregators(ctx context.Context, conn *configservice.Client, input *configservice.DescribeConfigurationAggregatorsInput) ([]types.ConfigurationAggregator, error) {
	var output []types.ConfigurationAggregator

	pages := configservice.NewDescribeConfigurationAggregatorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.NoSuchConfigurationAggregatorException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.ConfigurationAggregators...)
	}

	return output, nil
}

func expandAccountAggregationSources(tfList []interface{}) []types.AccountAggregationSource {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.AccountAggregationSource

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := types.AccountAggregationSource{
			AllAwsRegions: tfMap["all_regions"].(bool),
		}

		if v, ok := tfMap["account_ids"].([]interface{}); ok && len(v) > 0 {
			apiObject.AccountIds = flex.ExpandStringValueList(v)
		}

		if v, ok := tfMap["regions"].([]interface{}); ok && len(v) > 0 {
			apiObject.AwsRegions = flex.ExpandStringValueList(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandOrganizationAggregationSource(tfMap map[string]interface{}) *types.OrganizationAggregationSource {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.OrganizationAggregationSource{
		AllAwsRegions: tfMap["all_regions"].(bool),
		RoleArn:       aws.String(tfMap[names.AttrRoleARN].(string)),
	}

	if v, ok := tfMap["regions"].([]interface{}); ok && len(v) > 0 {
		apiObject.AwsRegions = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func flattenAccountAggregationSources(apiObjects []types.AccountAggregationSource) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	apiObject := apiObjects[0]
	tfMap := map[string]interface{}{
		"account_ids": apiObject.AccountIds,
		"all_regions": apiObject.AllAwsRegions,
		"regions":     apiObject.AwsRegions,
	}

	return []interface{}{tfMap}
}

func flattenOrganizationAggregationSource(apiObject *types.OrganizationAggregationSource) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"all_regions":     apiObject.AllAwsRegions,
		"regions":         apiObject.AwsRegions,
		names.AttrRoleARN: aws.ToString(apiObject.RoleArn),
	}

	return []interface{}{tfMap}
}
