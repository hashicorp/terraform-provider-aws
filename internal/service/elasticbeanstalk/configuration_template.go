// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elastic_beanstalk_configuration_template")
func ResourceConfigurationTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationTemplateCreate,
		ReadWithoutTimeout:   resourceConfigurationTemplateRead,
		UpdateWithoutTimeout: resourceConfigurationTemplateUpdate,
		DeleteWithoutTimeout: resourceConfigurationTemplateDelete,

		Schema: map[string]*schema.Schema{
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
				Set:      optionSettingValueHash,
			},
			"solution_stack_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceConfigurationTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticbeanstalk.CreateConfigurationTemplateInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		OptionSettings:  gatherOptionSettings(d),
		TemplateName:    aws.String(name),
	}

	if attr, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("environment_id"); ok {
		input.EnvironmentId = aws.String(attr.(string))
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

func resourceConfigurationTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	settings, err := FindConfigurationSettingsByTwoPartKey(ctx, conn, d.Get("application").(string), d.Id())

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

func resourceConfigurationTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		rm := extractOptionSettings(os.Difference(ns))
		add := extractOptionSettings(ns.Difference(os))

		// Additions and removals of options are done in a single API call, so we
		// can't do our normal "remove these" and then later "add these", re-adding
		// any updated settings.
		// Because of this, we need to remove any settings in the "removable"
		// settings that are also found in the "add" settings, otherwise they
		// conflict. Here we loop through all the initial removables from the set
		// difference, and we build up a slice of settings not found in the "add"
		// set
		var remove []awstypes.ConfigurationOptionSetting
		for _, r := range rm {
			for _, a := range add {
				if aws.ToString(r.Namespace) == aws.ToString(a.Namespace) &&
					aws.ToString(r.OptionName) == aws.ToString(a.OptionName) {
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

		for _, elem := range remove {
			input.OptionsToRemove = append(input.OptionsToRemove, awstypes.OptionSpecification{
				Namespace:  elem.Namespace,
				OptionName: elem.OptionName,
			})
		}

		_, err := conn.UpdateConfigurationTemplate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationTemplateRead(ctx, d, meta)...)
}

func resourceConfigurationTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	log.Printf("[INFO] Deleting Elastic Beanstalk Configuration Template: %s", d.Id())
	_, err := conn.DeleteConfigurationTemplate(ctx, &elasticbeanstalk.DeleteConfigurationTemplateInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		TemplateName:    aws.String(d.Id()),
	})

	errInvalidParameter := &invalidParameterValueError{err}

	if errs.IsAErrorMessageContains[*invalidParameterValueError](errInvalidParameter, "No Configuration Template named") ||
		errs.IsAErrorMessageContains[*invalidParameterValueError](errInvalidParameter, "No Application named") ||
		errs.IsAErrorMessageContains[*invalidParameterValueError](errInvalidParameter, "No Platform named") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
	}

	return diags
}

func FindConfigurationSettingsByTwoPartKey(ctx context.Context, conn *elasticbeanstalk.Client, applicationName, templateName string) (*awstypes.ConfigurationSettingsDescription, error) {
	input := &elasticbeanstalk.DescribeConfigurationSettingsInput{
		ApplicationName: aws.String(applicationName),
		TemplateName:    aws.String(templateName),
	}

	output, err := conn.DescribeConfigurationSettings(ctx, input)

	errInvalidParameter := &invalidParameterValueError{err}

	if errs.IsAErrorMessageContains[*invalidParameterValueError](errInvalidParameter, "No Configuration Template named") ||
		errs.IsAErrorMessageContains[*invalidParameterValueError](errInvalidParameter, "No Application named") ||
		errs.IsAErrorMessageContains[*invalidParameterValueError](errInvalidParameter, "No Platform named") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ConfigurationSettings) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ConfigurationSettings); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.ConfigurationSettings[0], nil
}

func gatherOptionSettings(d *schema.ResourceData) []awstypes.ConfigurationOptionSetting {
	optionSettingsSet, ok := d.Get("setting").(*schema.Set)
	if !ok || optionSettingsSet == nil {
		optionSettingsSet = new(schema.Set)
	}

	return extractOptionSettings(optionSettingsSet)
}

type invalidParameterValueError struct {
	error
}

func (e *invalidParameterValueError) ErrorMessage() string {
	if e == nil || e.error == nil {
		return ""
	}
	return e.Error()
}
