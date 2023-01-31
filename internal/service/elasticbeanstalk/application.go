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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"appversion_lifecycle": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"service_role": {
							Type:     schema.TypeString,
							Required: true,
						},
						"max_age_in_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"delete_source_from_s3": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	beanstalkConn := meta.(*conns.AWSClient).ElasticBeanstalkConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Get the name and description
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	log.Printf("[DEBUG] Elastic Beanstalk application create: %s, description: %s", name, description)

	req := &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(name),
		Description:     aws.String(description),
		Tags:            Tags(tags.IgnoreElasticbeanstalk()),
	}

	app, err := beanstalkConn.CreateApplicationWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Application (%s): %s", name, err)
	}

	d.SetId(name)

	if err = resourceApplicationAppVersionLifecycleUpdate(ctx, beanstalkConn, d, app.Application); err != nil {
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

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application (%s): updating tags: %s", d.Id(), err)
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
	var appversion_lifecycle map[string]interface{} = nil
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
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var app *elasticbeanstalk.ApplicationDescription
	err := resource.RetryContext(ctx, 30*time.Second, func() *resource.RetryError {
		var err error
		app, err = getApplication(ctx, d.Id(), conn)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if app == nil {
			err = fmt.Errorf("Elastic Beanstalk Application %q not found", d.Id())
			if d.IsNewResource() {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
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

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Elastic Beanstalk Application (%s): %s", arn, err)
	}

	tags = tags.IgnoreElasticbeanstalk().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceApplicationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	beanstalkConn := meta.(*conns.AWSClient).ElasticBeanstalkConn()

	_, err := beanstalkConn.DeleteApplicationWithContext(ctx, &elasticbeanstalk.DeleteApplicationInput{
		ApplicationName: aws.String(d.Id()),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Application (%s): %s", d.Id(), err)
	}

	var app *elasticbeanstalk.ApplicationDescription
	err = resource.RetryContext(ctx, 10*time.Second, func() *resource.RetryError {
		app, err = getApplication(ctx, d.Id(), meta.(*conns.AWSClient).ElasticBeanstalkConn())
		if err != nil {
			return resource.NonRetryableError(err)
		}

		if app != nil {
			return resource.RetryableError(
				fmt.Errorf("Beanstalk Application (%s) still exists: %s", d.Id(), err))
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		app, err = getApplication(ctx, d.Id(), meta.(*conns.AWSClient).ElasticBeanstalkConn())
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Application (%s): %s", d.Id(), err)
	}
	return diags
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
