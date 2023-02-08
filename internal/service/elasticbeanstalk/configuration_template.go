package elasticbeanstalk

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceConfigurationTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConfigurationTemplateCreate,
		ReadWithoutTimeout:   resourceConfigurationTemplateRead,
		UpdateWithoutTimeout: resourceConfigurationTemplateUpdate,
		DeleteWithoutTimeout: resourceConfigurationTemplateDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"application": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"environment_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"setting": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     resourceOptionSetting(),
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
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	// Get the relevant properties
	name := d.Get("name").(string)
	appName := d.Get("application").(string)

	optionSettings := gatherOptionSettings(d)

	opts := elasticbeanstalk.CreateConfigurationTemplateInput{
		ApplicationName: aws.String(appName),
		TemplateName:    aws.String(name),
		OptionSettings:  optionSettings,
	}

	if attr, ok := d.GetOk("description"); ok {
		opts.Description = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("environment_id"); ok {
		opts.EnvironmentId = aws.String(attr.(string))
	}

	if attr, ok := d.GetOk("solution_stack_name"); ok {
		opts.SolutionStackName = aws.String(attr.(string))
	}

	log.Printf("[DEBUG] Elastic Beanstalk configuration template create opts: %s", opts)
	if _, err := conn.CreateConfigurationTemplateWithContext(ctx, &opts); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk configuration template: %s", err)
	}

	d.SetId(name)

	return append(diags, resourceConfigurationTemplateRead(ctx, d, meta)...)
}

func resourceConfigurationTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	log.Printf("[DEBUG] Elastic Beanstalk configuration template read: %s", d.Get("name").(string))

	resp, err := conn.DescribeConfigurationSettingsWithContext(ctx, &elasticbeanstalk.DescribeConfigurationSettingsInput{
		TemplateName:    aws.String(d.Id()),
		ApplicationName: aws.String(d.Get("application").(string)),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidParameterValue" && strings.Contains(awsErr.Message(), "No Configuration Template named") {
				log.Printf("[WARN] No Configuration Template named (%s) found", d.Id())
				d.SetId("")
				return diags
			} else if awsErr.Code() == "InvalidParameterValue" && strings.Contains(awsErr.Message(), "No Platform named") {
				log.Printf("[WARN] No Platform named (%s) found", d.Get("solution_stack_name").(string))
				d.SetId("")
				return diags
			}
		}
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
	}

	if len(resp.ConfigurationSettings) != 1 {
		return sdkdiag.AppendErrorf(diags, "reading application properties: found %d applications, expected 1", len(resp.ConfigurationSettings))
	}

	d.Set("description", resp.ConfigurationSettings[0].Description)

	return diags
}

func resourceConfigurationTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	log.Printf("[DEBUG] Elastic Beanstalk configuration template update: %s", d.Get("name").(string))

	if d.HasChange("description") {
		if err := resourceConfigurationTemplateDescriptionUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("setting") {
		if err := resourceConfigurationTemplateOptionSettingsUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceConfigurationTemplateRead(ctx, d, meta)...)
}

func resourceConfigurationTemplateDescriptionUpdate(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, d *schema.ResourceData) error {
	_, err := conn.UpdateConfigurationTemplateWithContext(ctx, &elasticbeanstalk.UpdateConfigurationTemplateInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		TemplateName:    aws.String(d.Get("name").(string)),
		Description:     aws.String(d.Get("description").(string)),
	})

	return err
}

func resourceConfigurationTemplateOptionSettingsUpdate(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, d *schema.ResourceData) error {
	if d.HasChange("setting") {
		_, err := conn.ValidateConfigurationSettingsWithContext(ctx, &elasticbeanstalk.ValidateConfigurationSettingsInput{
			ApplicationName: aws.String(d.Get("application").(string)),
			TemplateName:    aws.String(d.Get("name").(string)),
			OptionSettings:  gatherOptionSettings(d),
		})
		if err != nil {
			return err
		}

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
		var remove []*elasticbeanstalk.ConfigurationOptionSetting
		for _, r := range rm {
			for _, a := range add {
				if aws.StringValue(r.Namespace) == aws.StringValue(a.Namespace) &&
					aws.StringValue(r.OptionName) == aws.StringValue(a.OptionName) {
					continue
				}
				remove = append(remove, r)
			}
		}

		req := &elasticbeanstalk.UpdateConfigurationTemplateInput{
			ApplicationName: aws.String(d.Get("application").(string)),
			TemplateName:    aws.String(d.Get("name").(string)),
			OptionSettings:  add,
		}

		for _, elem := range remove {
			req.OptionsToRemove = append(req.OptionsToRemove, &elasticbeanstalk.OptionSpecification{
				Namespace:  elem.Namespace,
				OptionName: elem.OptionName,
			})
		}

		log.Printf("[DEBUG] Update Configuration Template request: %s", req)
		if _, err := conn.UpdateConfigurationTemplateWithContext(ctx, req); err != nil {
			return err
		}
	}

	return nil
}

func resourceConfigurationTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	application := d.Get("application").(string)

	_, err := conn.DeleteConfigurationTemplateWithContext(ctx, &elasticbeanstalk.DeleteConfigurationTemplateInput{
		ApplicationName: aws.String(application),
		TemplateName:    aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Configuration Template (%s): %s", d.Id(), err)
	}
	return diags
}

func gatherOptionSettings(d *schema.ResourceData) []*elasticbeanstalk.ConfigurationOptionSetting {
	optionSettingsSet, ok := d.Get("setting").(*schema.Set)
	if !ok || optionSettingsSet == nil {
		optionSettingsSet = new(schema.Set)
	}

	return extractOptionSettings(optionSettingsSet)
}
