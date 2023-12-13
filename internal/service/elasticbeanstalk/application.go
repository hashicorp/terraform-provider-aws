// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elastic_beanstalk_application", name="Application")
// @Tags(identifierAttribute="arn")
func ResourceApplication() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationCreate,
		ReadWithoutTimeout:   resourceApplicationRead,
		UpdateWithoutTimeout: resourceApplicationUpdate,
		DeleteWithoutTimeout: resourceApplicationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"appversion_lifecycle": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_source_from_s3": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"max_age_in_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"service_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn(ctx)

	name := d.Get("name").(string)
	input := &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(name),
		Description:     aws.String(d.Get("description").(string)),
		Tags:            getTagsIn(ctx),
	}

	_, err := conn.CreateApplicationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Application (%s): %s", name, err)
	}

	d.SetId(name)

	_, err = tfresource.RetryWhenNotFound(ctx, 30*time.Second, func() (interface{}, error) {
		return FindApplicationByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elastic Beanstalk Application (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("appversion_lifecycle"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := &elasticbeanstalk.UpdateApplicationResourceLifecycleInput{
			ApplicationName:         aws.String(d.Id()),
			ResourceLifecycleConfig: expandApplicationResourceLifecycleConfig(v.([]interface{})[0].(map[string]interface{})),
		}

		_, err := conn.UpdateApplicationResourceLifecycleWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s) resource lifecycle: %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn(ctx)

	app, err := FindApplicationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elastic Beanstalk Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Application (%s): %s", d.Id(), err)
	}

	if err := d.Set("appversion_lifecycle", flattenApplicationResourceLifecycleConfig(app.ResourceLifecycleConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting appversion_lifecycle: %s", err)
	}
	d.Set("arn", app.ApplicationArn)
	d.Set("description", app.Description)
	d.Set("name", app.ApplicationName)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn(ctx)

	if d.HasChange("description") {
		input := &elasticbeanstalk.UpdateApplicationInput{
			ApplicationName: aws.String(d.Id()),
			Description:     aws.String(d.Get("description").(string)),
		}

		_, err := conn.UpdateApplicationWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("appversion_lifecycle") {
		var resourceLifecycleConfig *elasticbeanstalk.ApplicationResourceLifecycleConfig

		if v, ok := d.GetOk("appversion_lifecycle"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			resourceLifecycleConfig = expandApplicationResourceLifecycleConfig(v.([]interface{})[0].(map[string]interface{}))
		} else {
			resourceLifecycleConfig = expandApplicationResourceLifecycleConfig(map[string]interface{}{})
		}

		input := &elasticbeanstalk.UpdateApplicationResourceLifecycleInput{
			ApplicationName:         aws.String(d.Id()),
			ResourceLifecycleConfig: resourceLifecycleConfig,
		}

		_, err := conn.UpdateApplicationResourceLifecycleWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s) resource lifecycle: %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn(ctx)

	log.Printf("[DEBUG] Deleting Elastic Beanstalk Application: %s", d.Id())
	_, err := conn.DeleteApplicationWithContext(ctx, &elasticbeanstalk.DeleteApplicationInput{
		ApplicationName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Application (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 10*time.Second, func() (interface{}, error) {
		return FindApplicationByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elastic Beanstalk Application (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindApplicationByName(ctx context.Context, conn *elasticbeanstalk.ElasticBeanstalk, name string) (*elasticbeanstalk.ApplicationDescription, error) {
	input := &elasticbeanstalk.DescribeApplicationsInput{
		ApplicationNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribeApplicationsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Applications) == 0 || output.Applications[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Applications); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Applications[0], nil
}

func expandApplicationResourceLifecycleConfig(tfMap map[string]interface{}) *elasticbeanstalk.ApplicationResourceLifecycleConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &elasticbeanstalk.ApplicationResourceLifecycleConfig{
		VersionLifecycleConfig: &elasticbeanstalk.ApplicationVersionLifecycleConfig{
			MaxCountRule: &elasticbeanstalk.MaxCountRule{
				Enabled: aws.Bool(false),
			},
			MaxAgeRule: &elasticbeanstalk.MaxAgeRule{
				Enabled: aws.Bool(false),
			},
		},
	}

	if v, ok := tfMap["service_role"].(string); ok && v != "" {
		apiObject.ServiceRole = aws.String(v)
	}

	if v, ok := tfMap["max_age_in_days"].(int); ok && v != 0 {
		apiObject.VersionLifecycleConfig.MaxAgeRule = &elasticbeanstalk.MaxAgeRule{
			DeleteSourceFromS3: aws.Bool(tfMap["delete_source_from_s3"].(bool)),
			Enabled:            aws.Bool(true),
			MaxAgeInDays:       aws.Int64(int64(v)),
		}
	}

	if v, ok := tfMap["max_count"].(int); ok && v != 0 {
		apiObject.VersionLifecycleConfig.MaxCountRule = &elasticbeanstalk.MaxCountRule{
			DeleteSourceFromS3: aws.Bool(tfMap["delete_source_from_s3"].(bool)),
			Enabled:            aws.Bool(true),
			MaxCount:           aws.Int64(int64(v)),
		}
	}

	return apiObject
}

func flattenApplicationResourceLifecycleConfig(apiObject *elasticbeanstalk.ApplicationResourceLifecycleConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject := apiObject.VersionLifecycleConfig; apiObject != nil {
		if apiObject := apiObject.MaxAgeRule; apiObject != nil && aws.BoolValue(apiObject.Enabled) {
			if v := apiObject.DeleteSourceFromS3; v != nil {
				tfMap["delete_source_from_s3"] = aws.BoolValue(v)
			}

			if v := apiObject.MaxAgeInDays; v != nil {
				tfMap["max_age_in_days"] = aws.Int64Value(v)
			}
		}

		if apiObject := apiObject.MaxCountRule; apiObject != nil && aws.BoolValue(apiObject.Enabled) {
			if v := apiObject.DeleteSourceFromS3; v != nil {
				tfMap["delete_source_from_s3"] = aws.BoolValue(v)
			}

			if v := apiObject.MaxCount; v != nil {
				tfMap["max_count"] = aws.Int64Value(v)
			}
		}
	}

	if len(tfMap) == 0 {
		return nil
	}

	if v := apiObject.ServiceRole; v != nil {
		tfMap["service_role"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
