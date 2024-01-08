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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_config_configuration_recorder")
func ResourceConfigurationRecorder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationRecorderPut,
		ReadWithoutTimeout:   resourceConfigurationRecorderRead,
		UpdateWithoutTimeout: resourceConfigurationRecorderPut,
		DeleteWithoutTimeout: resourceConfigurationRecorderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.All(
			resourceConfigCustomizeDiff,
		),

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "default",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"recording_group": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all_supported": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"exclusion_by_resource_types": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_types": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"include_global_resource_types": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"recording_strategy": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"use_only": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(configservice.RecordingStrategyType_Values(), false),
									},
								},
							},
						},
						"resource_types": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceConfigurationRecorderPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	name := d.Get("name").(string)
	input := &configservice.PutConfigurationRecorderInput{
		ConfigurationRecorder: &configservice.ConfigurationRecorder{
			Name:    aws.String(name),
			RoleARN: aws.String(d.Get("role_arn").(string)),
		},
	}

	if v, ok := d.GetOk("recording_group"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ConfigurationRecorder.RecordingGroup = expandRecordingGroup(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.PutConfigurationRecorderWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting ConfigService Configuration Recorder (%s): %s", name, err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceConfigurationRecorderRead(ctx, d, meta)...)
}

func resourceConfigurationRecorderRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	recorder, err := FindConfigurationRecorderByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ConfigService Configuration Recorder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ConfigService Configuration Recorder (%s): %s", d.Id(), err)
	}

	d.Set("name", recorder.Name)
	d.Set("role_arn", recorder.RoleARN)

	if recorder.RecordingGroup != nil {
		if err := d.Set("recording_group", flattenRecordingGroup(recorder.RecordingGroup)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting recording_group: %s", err)
		}
	}

	return diags
}

func resourceConfigurationRecorderDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConfigServiceConn(ctx)

	log.Printf("[DEBUG] Deleting ConfigService Configuration Recorder: %s", d.Id())
	_, err := conn.DeleteConfigurationRecorderWithContext(ctx, &configservice.DeleteConfigurationRecorderInput{
		ConfigurationRecorderName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationRecorderException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ConfigService Configuration Recorder (%s): %s", d.Id(), err)
	}

	return diags
}

func FindConfigurationRecorderByName(ctx context.Context, conn *configservice.ConfigService, name string) (*configservice.ConfigurationRecorder, error) {
	input := &configservice.DescribeConfigurationRecordersInput{
		ConfigurationRecorderNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribeConfigurationRecordersWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, configservice.ErrCodeNoSuchConfigurationRecorderException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSinglePtrResult(output.ConfigurationRecorders)
}

func resourceConfigCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" { // New resource.
		if g, ok := diff.GetOk("recording_group"); ok {
			group := g.([]interface{})[0].(map[string]interface{})

			if h, ok := group["all_supported"]; ok {
				if i, ok := group["recording_strategy"]; ok && len(i.([]interface{})) > 0 && i.([]interface{})[0] != nil {
					strategy := i.([]interface{})[0].(map[string]interface{})

					if j, ok := strategy["use_only"].(string); ok {
						if h.(bool) && j != configservice.RecordingStrategyTypeAllSupportedResourceTypes {
							return errors.New(` Invalid record group strategy  , all_supported must be set to true  `)
						}

						if k, ok := group["exclusion_by_resource_types"]; ok && len(k.([]interface{})) > 0 && k.([]interface{})[0] != nil {
							if h.(bool) {
								return errors.New(` Invalid record group , all_supported must be set to false when exclusion_by_resource_types is set `)
							}

							if j != configservice.RecordingStrategyTypeExclusionByResourceTypes {
								return errors.New(` Invalid record group strategy ,  use only must be set to EXCLUSION_BY_RESOURCE_TYPES`)
							}

							if l, ok := group["resource_types"]; ok {
								resourceTypes := flex.ExpandStringSet(l.(*schema.Set))
								if len(resourceTypes) > 0 {
									return errors.New(` Invalid record group , resource_types must not be set when exclusion_by_resource_types is set `)
								}
							}
						}

						if l, ok := group["resource_types"]; ok {
							resourceTypes := flex.ExpandStringSet(l.(*schema.Set))
							if len(resourceTypes) > 0 {
								if h.(bool) {
									return errors.New(` Invalid record group , all_supported must be set to false when resource_types is set `)
								}

								if j != configservice.RecordingStrategyTypeInclusionByResourceTypes {
									return errors.New(` Invalid record group strategy ,  use only must be set to INCLUSION_BY_RESOURCE_TYPES`)
								}

								if m, ok := group["exclusion_by_resource_types"]; ok && len(m.([]interface{})) > 0 && i.([]interface{})[0] != nil {
									return errors.New(` Invalid record group , exclusion_by_resource_types must not be set when resource_types is set `)
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}
