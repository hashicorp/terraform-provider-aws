// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_option_group", name="DB Option Group")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceOptionGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOptionGroupCreate,
		ReadWithoutTimeout:   resourceOptionGroupRead,
		UpdateWithoutTimeout: resourceOptionGroupUpdate,
		DeleteWithoutTimeout: resourceOptionGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"major_engine_version": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validOptionGroupName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validOptionGroupNamePrefix,
			},
			"option": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_security_group_memberships": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"option_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"option_settings": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrVersion: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_security_group_memberships": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"option_group_description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			names.AttrSkipDestroy: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceOptionGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &rds.CreateOptionGroupInput{
		EngineName:             aws.String(d.Get("engine_name").(string)),
		MajorEngineVersion:     aws.String(d.Get("major_engine_version").(string)),
		OptionGroupDescription: aws.String(d.Get("option_group_description").(string)),
		OptionGroupName:        aws.String(name),
		Tags:                   getTagsIn(ctx),
	}

	_, err := conn.CreateOptionGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Option Group (%s): %s", name, err)
	}

	d.SetId(strings.ToLower(name))

	return append(diags, resourceOptionGroupUpdate(ctx, d, meta)...)
}

func resourceOptionGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	option, err := findOptionGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Option Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Option Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, option.OptionGroupArn)
	d.Set("engine_name", option.EngineName)
	d.Set("major_engine_version", option.MajorEngineVersion)
	d.Set(names.AttrName, option.OptionGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(option.OptionGroupName)))
	if err := d.Set("option", flattenOptions(option.Options, expandOptionConfigurations(d.Get("option").(*schema.Set).List()))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting option: %s", err)
	}
	d.Set("option_group_description", option.OptionGroupDescription)
	// Support in-place update of non-refreshable attribute.
	d.Set(names.AttrSkipDestroy, d.Get(names.AttrSkipDestroy))

	return diags
}

func resourceOptionGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChange("option") {
		o, n := d.GetChange("option")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		optionsToInclude := expandOptionConfigurations(ns.Difference(os).List())
		optionsToIncludeNames := flattenOptionNames(ns.Difference(os).List())
		optionsToRemove := []string{}
		optionsToRemoveNames := flattenOptionNames(os.Difference(ns).List())

		for _, optionToRemoveName := range optionsToRemoveNames {
			if slices.Contains(optionsToIncludeNames, optionToRemoveName) {
				continue
			}
			optionsToRemove = append(optionsToRemove, optionToRemoveName)
		}

		// Ensure there is actually something to update.
		// InvalidParameterValue: At least one option must be added, modified, or removed.
		if len(optionsToInclude) > 0 || len(optionsToRemove) > 0 {
			input := &rds.ModifyOptionGroupInput{
				ApplyImmediately: aws.Bool(true),
				OptionGroupName:  aws.String(d.Id()),
			}

			if len(optionsToInclude) > 0 {
				input.OptionsToInclude = optionsToInclude
			}

			if len(optionsToRemove) > 0 {
				input.OptionsToRemove = optionsToRemove
			}

			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (any, error) {
				return conn.ModifyOptionGroup(ctx, input)
			}, errCodeInvalidParameterValue, "IAM role ARN value is invalid or does not include the required permissions")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying RDS DB Option Group (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceOptionGroupRead(ctx, d, meta)...)
}

func resourceOptionGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if _, ok := d.GetOk(names.AttrSkipDestroy); ok {
		log.Printf("[DEBUG] Retaining RDS DB Option Group: %s", d.Id())
		return diags
	}

	log.Printf("[DEBUG] Deleting RDS DB Option Group: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*types.InvalidOptionGroupStateFault](ctx, d.Timeout(schema.TimeoutDelete), func() (any, error) {
		return conn.DeleteOptionGroup(ctx, &rds.DeleteOptionGroupInput{
			OptionGroupName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.OptionGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Option Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findOptionGroupByName(ctx context.Context, conn *rds.Client, name string) (*types.OptionGroup, error) {
	input := &rds.DescribeOptionGroupsInput{
		OptionGroupName: aws.String(name),
	}
	output, err := findOptionGroup(ctx, conn, input, tfslices.PredicateTrue[*types.OptionGroup]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.OptionGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findOptionGroup(ctx context.Context, conn *rds.Client, input *rds.DescribeOptionGroupsInput, filter tfslices.Predicate[*types.OptionGroup]) (*types.OptionGroup, error) {
	output, err := findOptionGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findOptionGroups(ctx context.Context, conn *rds.Client, input *rds.DescribeOptionGroupsInput, filter tfslices.Predicate[*types.OptionGroup]) ([]types.OptionGroup, error) {
	var output []types.OptionGroup

	pages := rds.NewDescribeOptionGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.OptionGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.OptionGroupsList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func flattenOptionNames(tfList []any) []string {
	return tfslices.ApplyToAll(tfList, func(v any) string {
		return v.(map[string]any)["option_name"].(string)
	})
}

func expandOptionConfigurations(tfList []any) []types.OptionConfiguration {
	var apiObjects []types.OptionConfiguration

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := types.OptionConfiguration{
			OptionName: aws.String(tfMap["option_name"].(string)),
		}

		if v, ok := tfMap["db_security_group_memberships"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.DBSecurityGroupMemberships = flex.ExpandStringValueSet(v)
		}

		if v, ok := tfMap["option_settings"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.OptionSettings = expandOptionSettings(v.List())
		}

		if v, ok := tfMap[names.AttrPort].(int); ok && v != 0 {
			apiObject.Port = aws.Int32(int32(v))
		}

		if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
			apiObject.OptionVersion = aws.String(v)
		}

		if v, ok := tfMap["vpc_security_group_memberships"].(*schema.Set); ok && v.Len() > 0 {
			apiObject.VpcSecurityGroupMemberships = flex.ExpandStringValueSet(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenOptions(apiObjects []types.Option, configuredObjects []types.OptionConfiguration) []any {
	tfList := make([]any, 0)

	for _, apiObject := range apiObjects {
		if apiObject.OptionName == nil {
			continue
		}

		optionName := aws.ToString(apiObject.OptionName)
		var configuredOption *types.OptionConfiguration
		if v := tfslices.Filter(configuredObjects, func(v types.OptionConfiguration) bool {
			return aws.ToString(v.OptionName) == optionName
		}); len(v) > 0 {
			configuredOption = &v[0]
		}

		optionSettings := make([]any, 0)
		for _, apiOptionSetting := range apiObject.OptionSettings {
			// The RDS API responds with all settings. Omit settings that match default value,
			// but only if unconfigured. This is to prevent operators from continually needing
			// to continually update their Terraform configurations to match new option settings
			// when added by the API.
			optionSettingName := aws.ToString(apiOptionSetting.Name)
			var configuredOptionSetting *types.OptionSetting

			if configuredOption != nil {
				if v := tfslices.Filter(configuredOption.OptionSettings, func(v types.OptionSetting) bool {
					return aws.ToString(v.Name) == optionSettingName
				}); len(v) > 0 {
					configuredOptionSetting = &v[0]
				}
			}

			optionSettingValue := aws.ToString(apiOptionSetting.Value)
			if configuredOptionSetting == nil && optionSettingValue == aws.ToString(apiOptionSetting.DefaultValue) {
				continue
			}

			optionSetting := map[string]any{
				names.AttrName:  optionSettingName,
				names.AttrValue: optionSettingValue,
			}

			// Some values, like passwords, are sent back from the API as ****.
			// Set the response to match the configuration to prevent an unexpected difference.
			if configuredOptionSetting != nil && optionSettingValue == "****" {
				optionSetting[names.AttrValue] = aws.ToString(configuredOptionSetting.Value)
			}

			optionSettings = append(optionSettings, optionSetting)
		}

		tfMap := map[string]any{
			"db_security_group_memberships": tfslices.ApplyToAll(apiObject.DBSecurityGroupMemberships, func(v types.DBSecurityGroupMembership) string {
				return aws.ToString(v.DBSecurityGroupName)
			}),
			"option_name":     optionName,
			"option_settings": optionSettings,
			"vpc_security_group_memberships": tfslices.ApplyToAll(apiObject.VpcSecurityGroupMemberships, func(v types.VpcSecurityGroupMembership) string {
				return aws.ToString(v.VpcSecurityGroupId)
			}),
		}

		if apiObject.OptionVersion != nil && configuredOption != nil && configuredOption.OptionVersion != nil {
			tfMap[names.AttrVersion] = aws.ToString(apiObject.OptionVersion)
		}

		if apiObject.Port != nil && configuredOption != nil && configuredOption.Port != nil {
			tfMap[names.AttrPort] = aws.ToInt32(apiObject.Port)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandOptionSettings(tfList []any) []types.OptionSetting {
	apiObjects := make([]types.OptionSetting, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		apiObject := types.OptionSetting{
			Name:  aws.String(tfMap[names.AttrName].(string)),
			Value: aws.String(tfMap[names.AttrValue].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
