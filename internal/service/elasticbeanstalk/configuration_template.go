// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elastic_beanstalk_configuration_template", name="Configuration Template")
func resourceConfigurationTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationTemplateCreate,
		ReadWithoutTimeout:   resourceConfigurationTemplateRead,
		UpdateWithoutTimeout: resourceConfigurationTemplateUpdate,
		DeleteWithoutTimeout: resourceConfigurationTemplateDelete,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"application": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Optional: true,
				},
				"environment_id": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				"setting": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     settingSchema(),
					Set:      hashSettingsValue,
				},
				"solution_stack_name": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
			}
		},
	}
}

func resourceConfigurationTemplateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticbeanstalk.CreateConfigurationTemplateInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		TemplateName:    aws.String(name),
	}

	if attr, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("environment_id"); ok {
		input.EnvironmentId = aws.String(attr.(string))
	}

	if v, ok := d.GetOk("setting"); ok && v.(*schema.Set).Len() > 0 {
		input.OptionSettings = expandConfigurationOptionSettings(v.(*schema.Set).List())
	}

	if attr, ok := d.GetOk("solution_stack_name"); ok {
		input.SolutionStackName = aws.String(attr.(string))
	}

	output, err := conn.CreateConfigurationTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Configuration Template (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TemplateName))

	return append(diags, resourceConfigurationTemplateRead(ctx, d, meta)...)
}

func resourceConfigurationTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	settings, err := findConfigurationSettingsByTwoPartKey(ctx, conn, d.Get("application").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elastic Beanstalk Configuration Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
	}

	d.Set("application", settings.ApplicationName)
	d.Set(names.AttrDescription, settings.Description)
	d.Set(names.AttrName, settings.TemplateName)
	d.Set("solution_stack_name", settings.SolutionStackName)

	return diags
}

func resourceConfigurationTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &elasticbeanstalk.UpdateConfigurationTemplateInput{
			ApplicationName: aws.String(d.Get("application").(string)),
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			TemplateName:    aws.String(d.Id()),
		}

		_, err := conn.UpdateConfigurationTemplate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("setting") {
		o, n := d.GetChange("setting")
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := expandConfigurationOptionSettings(ns.Difference(os).List()), expandConfigurationOptionSettings(os.Difference(ns).List())

		// Additions and removals of options are done in a single API call, so we
		// can't do our normal "remove these" and then later "add these", re-adding
		// any updated settings.
		// Because of this, we need to remove any settings in the "removable"
		// settings that are also found in the "add" settings, otherwise they
		// conflict. Here we loop through all the initial removables from the set
		// difference, and we build up a slice of settings not found in the "add"
		// set
		var remove []awstypes.ConfigurationOptionSetting
		for _, r := range del {
			for _, a := range add {
				if aws.ToString(r.Namespace) == aws.ToString(a.Namespace) && aws.ToString(r.OptionName) == aws.ToString(a.OptionName) {
					continue
				}
				remove = append(remove, r)
			}
		}

		input := &elasticbeanstalk.UpdateConfigurationTemplateInput{
			ApplicationName: aws.String(d.Get("application").(string)),
			OptionSettings:  add,
			TemplateName:    aws.String(d.Id()),
		}

		for _, v := range remove {
			input.OptionsToRemove = append(input.OptionsToRemove, awstypes.OptionSpecification{
				Namespace:  v.Namespace,
				OptionName: v.OptionName,
			})
		}

		_, err := conn.UpdateConfigurationTemplate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationTemplateRead(ctx, d, meta)...)
}

func resourceConfigurationTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	log.Printf("[INFO] Deleting Elastic Beanstalk Configuration Template: %s", d.Id())
	_, err := conn.DeleteConfigurationTemplate(ctx, &elasticbeanstalk.DeleteConfigurationTemplateInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		TemplateName:    aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "No Configuration Template named") ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "No Application named") ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "No Platform named") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
	}

	return diags
}

func findConfigurationSettingsByTwoPartKey(ctx context.Context, conn *elasticbeanstalk.Client, applicationName, templateName string) (*awstypes.ConfigurationSettingsDescription, error) {
	input := &elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: aws.String(applicationName),
		TemplateName:    aws.String(templateName),
	}

	return findConfigurationSettings(ctx, conn, input)
}

func findConfigurationSettings(ctx context.Context, conn *elasticbeanstalk.Client, input *elasticbeanstalk.DescribeConfigurationSettingsInput) (*awstypes.ConfigurationSettingsDescription, error) {
	output, err := findConfigurationSettingses(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findConfigurationSettingses(ctx context.Context, conn *elasticbeanstalk.Client, input *elasticbeanstalk.DescribeConfigurationSettingsInput) ([]awstypes.ConfigurationSettingsDescription, error) {
	output, err := conn.DescribeConfigurationSettings(ctx, input)

	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "No Configuration Template named") ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "No Application named") ||
		tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "No Platform named") {
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

	return output.ConfigurationSettings, nil
}
