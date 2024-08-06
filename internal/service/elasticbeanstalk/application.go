// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticbeanstalk/types"
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
						names.AttrServiceRole: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
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
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(name),
		Description:     aws.String(d.Get(names.AttrDescription).(string)),
		Tags:            getTagsIn(ctx),
	}

	_, err := conn.CreateApplication(ctx, input)

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

		_, err := conn.UpdateApplicationResourceLifecycle(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s) resource lifecycle: %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

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
	d.Set(names.AttrARN, app.ApplicationArn)
	d.Set(names.AttrDescription, app.Description)
	d.Set(names.AttrName, app.ApplicationName)

	return diags
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &elasticbeanstalk.UpdateApplicationInput{
			ApplicationName: aws.String(d.Id()),
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
		}

		_, err := conn.UpdateApplication(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("appversion_lifecycle") {
		var resourceLifecycleConfig *awstypes.ApplicationResourceLifecycleConfig

		if v, ok := d.GetOk("appversion_lifecycle"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			resourceLifecycleConfig = expandApplicationResourceLifecycleConfig(v.([]interface{})[0].(map[string]interface{}))
		} else {
			resourceLifecycleConfig = expandApplicationResourceLifecycleConfig(map[string]interface{}{})
		}

		input := &elasticbeanstalk.UpdateApplicationResourceLifecycleInput{
			ApplicationName:         aws.String(d.Id()),
			ResourceLifecycleConfig: resourceLifecycleConfig,
		}

		_, err := conn.UpdateApplicationResourceLifecycle(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s) resource lifecycle: %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	log.Printf("[DEBUG] Deleting Elastic Beanstalk Application: %s", d.Id())
	_, err := conn.DeleteApplication(ctx, &elasticbeanstalk.DeleteApplicationInput{
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

func FindApplicationByName(ctx context.Context, conn *elasticbeanstalk.Client, name string) (*awstypes.ApplicationDescription, error) {
	input := &elasticbeanstalk.DescribeApplicationsInput{
		ApplicationNames: []string{name},
	}

	output, err := conn.DescribeApplications(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Applications) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Applications); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.Applications[0], nil
}

func expandApplicationResourceLifecycleConfig(tfMap map[string]interface{}) *awstypes.ApplicationResourceLifecycleConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ApplicationResourceLifecycleConfig{
		VersionLifecycleConfig: &awstypes.ApplicationVersionLifecycleConfig{
			MaxCountRule: &awstypes.MaxCountRule{
				Enabled: aws.Bool(false),
			},
			MaxAgeRule: &awstypes.MaxAgeRule{
				Enabled: aws.Bool(false),
			},
		},
	}

	if v, ok := tfMap[names.AttrServiceRole].(string); ok && v != "" {
		apiObject.ServiceRole = aws.String(v)
	}

	if v, ok := tfMap["max_age_in_days"].(int); ok && v != 0 {
		apiObject.VersionLifecycleConfig.MaxAgeRule = &awstypes.MaxAgeRule{
			DeleteSourceFromS3: aws.Bool(tfMap["delete_source_from_s3"].(bool)),
			Enabled:            aws.Bool(true),
			MaxAgeInDays:       aws.Int32(int32(v)),
		}
	}

	if v, ok := tfMap["max_count"].(int); ok && v != 0 {
		apiObject.VersionLifecycleConfig.MaxCountRule = &awstypes.MaxCountRule{
			DeleteSourceFromS3: aws.Bool(tfMap["delete_source_from_s3"].(bool)),
			Enabled:            aws.Bool(true),
			MaxCount:           aws.Int32(int32(v)),
		}
	}

	return apiObject
}

func flattenApplicationResourceLifecycleConfig(apiObject *awstypes.ApplicationResourceLifecycleConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject := apiObject.VersionLifecycleConfig; apiObject != nil {
		if apiObject := apiObject.MaxAgeRule; apiObject != nil && aws.ToBool(apiObject.Enabled) {
			if v := apiObject.DeleteSourceFromS3; v != nil {
				tfMap["delete_source_from_s3"] = aws.ToBool(v)
			}

			if v := apiObject.MaxAgeInDays; v != nil {
				tfMap["max_age_in_days"] = aws.ToInt32(v)
			}
		}

		if apiObject := apiObject.MaxCountRule; apiObject != nil && aws.ToBool(apiObject.Enabled) {
			if v := apiObject.DeleteSourceFromS3; v != nil {
				tfMap["delete_source_from_s3"] = aws.ToBool(v)
			}

			if v := apiObject.MaxCount; v != nil {
				tfMap["max_count"] = aws.ToInt32(v)
			}
		}
	}

	if len(tfMap) == 0 {
		return nil
	}

	if v := apiObject.ServiceRole; v != nil {
		tfMap[names.AttrServiceRole] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}
