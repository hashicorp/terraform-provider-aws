package elasticbeanstalk

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	name := d.Get("name").(string)
	input := &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(name),
		Description:     aws.String(d.Get("description").(string)),
		Tags:            GetTagsIn(ctx),
	}

	output, err := conn.CreateApplicationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Application (%s): %s", name, err)
	}

	d.SetId(name)

	if err = resourceApplicationAppVersionLifecycleUpdate(ctx, conn, d, output.Application); err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Application (%s): %s", name, err)
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	if d.HasChange("description") {
		if err := resourceApplicationDescriptionUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("appversion_lifecycle") {
		if err := resourceApplicationAppVersionLifecycleUpdate(ctx, conn, d, nil); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationRead(ctx, d, meta)...)
}

func resourceApplicationDescriptionUpdate(ctx context.Context, beanstalkConn *elasticbeanstalk.ElasticBeanstalk, d *schema.ResourceData) error {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	log.Printf("[DEBUG] Elastic Beanstalk application: %s, update description: %s", name, description)

	_, err := beanstalkConn.UpdateApplicationWithContext(ctx, &elasticbeanstalk.UpdateApplicationInput{
		ApplicationName: aws.String(name),
		Description:     aws.String(description),
	})

	return err
}

func resourceApplicationAppVersionLifecycleUpdate(ctx context.Context, beanstalkConn *elasticbeanstalk.ElasticBeanstalk, d *schema.ResourceData, app *elasticbeanstalk.ApplicationDescription) error {
	name := d.Get("name").(string)
	appversion_lifecycles := d.Get("appversion_lifecycle").([]interface{})
	var appversion_lifecycle map[string]interface{}
	if len(appversion_lifecycles) == 1 {
		appversion_lifecycle = appversion_lifecycles[0].(map[string]interface{})
	}

	if appversion_lifecycle == nil && app != nil && app.ResourceLifecycleConfig.ServiceRole == nil {
		// We want appversion lifecycle management to be disabled, and it currently is, and there's no way to reproduce
		// this state in a UpdateApplicationResourceLifecycle service call (fails w/ ServiceRole is not a valid arn).  So,
		// in this special case we just do nothing.
		log.Printf("[DEBUG] Elastic Beanstalk application: %s, update appversion_lifecycle is anticipated no-op", name)
		return nil
	}

	rlc := &elasticbeanstalk.ApplicationResourceLifecycleConfig{
		ServiceRole: nil,
		VersionLifecycleConfig: &elasticbeanstalk.ApplicationVersionLifecycleConfig{
			MaxCountRule: &elasticbeanstalk.MaxCountRule{
				Enabled: aws.Bool(false),
			},
			MaxAgeRule: &elasticbeanstalk.MaxAgeRule{
				Enabled: aws.Bool(false),
			},
		},
	}

	if appversion_lifecycle != nil {
		service_role, ok := appversion_lifecycle["service_role"]
		if ok {
			rlc.ServiceRole = aws.String(service_role.(string))
		}

		rlc.VersionLifecycleConfig = &elasticbeanstalk.ApplicationVersionLifecycleConfig{
			MaxCountRule: &elasticbeanstalk.MaxCountRule{
				Enabled: aws.Bool(false),
			},
			MaxAgeRule: &elasticbeanstalk.MaxAgeRule{
				Enabled: aws.Bool(false),
			},
		}

		max_age_in_days, ok := appversion_lifecycle["max_age_in_days"]
		if ok && max_age_in_days != 0 {
			rlc.VersionLifecycleConfig.MaxAgeRule = &elasticbeanstalk.MaxAgeRule{
				Enabled:            aws.Bool(true),
				DeleteSourceFromS3: aws.Bool(appversion_lifecycle["delete_source_from_s3"].(bool)),
				MaxAgeInDays:       aws.Int64(int64(max_age_in_days.(int))),
			}
		}

		max_count, ok := appversion_lifecycle["max_count"]
		if ok && max_count != 0 {
			rlc.VersionLifecycleConfig.MaxCountRule = &elasticbeanstalk.MaxCountRule{
				Enabled:            aws.Bool(true),
				DeleteSourceFromS3: aws.Bool(appversion_lifecycle["delete_source_from_s3"].(bool)),
				MaxCount:           aws.Int64(int64(max_count.(int))),
			}
		}
	}

	_, err := beanstalkConn.UpdateApplicationResourceLifecycleWithContext(ctx, &elasticbeanstalk.UpdateApplicationResourceLifecycleInput{
		ApplicationName:         aws.String(name),
		ResourceLifecycleConfig: rlc,
	})
	if err != nil {
		return fmt.Errorf("updating application resource lifecycle: %w", err)
	}
	return nil
}

func resourceApplicationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	var app *elasticbeanstalk.ApplicationDescription
	err := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
		var err error
		app, err = getApplication(ctx, d.Id(), conn)
		if err != nil {
			return retry.NonRetryableError(err)
		}

		if app == nil {
			err = fmt.Errorf("Elastic Beanstalk Application %q not found", d.Id())
			if d.IsNewResource() {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		app, err = getApplication(ctx, d.Id(), conn)
	}
	if err != nil {
		if app == nil {
			log.Printf("[WARN] %s, removing from state", err)
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Application (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(app.ApplicationArn)
	d.Set("arn", arn)
	d.Set("name", app.ApplicationName)
	d.Set("description", app.Description)

	if app.ResourceLifecycleConfig != nil {
		d.Set("appversion_lifecycle", flattenResourceLifecycleConfig(app.ResourceLifecycleConfig))
	}

	return diags
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

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

func getApplication(ctx context.Context, id string, conn *elasticbeanstalk.ElasticBeanstalk) (*elasticbeanstalk.ApplicationDescription, error) {
	resp, err := conn.DescribeApplicationsWithContext(ctx, &elasticbeanstalk.DescribeApplicationsInput{
		ApplicationNames: []*string{aws.String(id)},
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "InvalidBeanstalkAppID.NotFound") {
			return nil, nil
		}
		return nil, err
	}

	if len(resp.Applications) > 1 {
		return nil, fmt.Errorf("Error %d Applications matched, expected 1", len(resp.Applications))
	}

	if len(resp.Applications) == 0 {
		return nil, nil
	}

	return resp.Applications[0], nil
}
