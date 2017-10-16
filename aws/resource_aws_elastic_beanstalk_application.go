package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/terraform/helper/resource"
)

func resourceAwsElasticBeanstalkApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticBeanstalkApplicationCreate,
		Read:   resourceAwsElasticBeanstalkApplicationRead,
		Update: resourceAwsElasticBeanstalkApplicationUpdate,
		Delete: resourceAwsElasticBeanstalkApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
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
		},
	}
}

func resourceAwsElasticBeanstalkApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	beanstalkConn := meta.(*AWSClient).elasticbeanstalkconn

	// Get the name and description
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	log.Printf("[DEBUG] Elastic Beanstalk application create: %s, description: %s", name, description)

	req := &elasticbeanstalk.CreateApplicationInput{
		ApplicationName: aws.String(name),
		Description:     aws.String(description),
	}

	app, err := beanstalkConn.CreateApplication(req)
	if err != nil {
		return err
	}

	d.SetId(name)

	if err = resourceAwsElasticBeanstalkApplicationAppversionLifecycleUpdate(beanstalkConn, d, app.Application); err != nil {
		return err
	}

	return resourceAwsElasticBeanstalkApplicationRead(d, meta)
}

func resourceAwsElasticBeanstalkApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	beanstalkConn := meta.(*AWSClient).elasticbeanstalkconn

	if d.HasChange("description") {
		if err := resourceAwsElasticBeanstalkApplicationDescriptionUpdate(beanstalkConn, d); err != nil {
			return err
		}
	}

	if d.HasChange("appversion_lifecycle") {
		if err := resourceAwsElasticBeanstalkApplicationAppversionLifecycleUpdate(beanstalkConn, d, nil); err != nil {
			return err
		}
	}

	return resourceAwsElasticBeanstalkApplicationRead(d, meta)
}

func resourceAwsElasticBeanstalkApplicationDescriptionUpdate(beanstalkConn *elasticbeanstalk.ElasticBeanstalk, d *schema.ResourceData) error {
	name := d.Get("name").(string)
	description := d.Get("description").(string)

	log.Printf("[DEBUG] Elastic Beanstalk application: %s, update description: %s", name, description)

	_, err := beanstalkConn.UpdateApplication(&elasticbeanstalk.UpdateApplicationInput{
		ApplicationName: aws.String(name),
		Description:     aws.String(description),
	})

	return err
}

func resourceAwsElasticBeanstalkApplicationAppversionLifecycleUpdate(beanstalkConn *elasticbeanstalk.ElasticBeanstalk, d *schema.ResourceData, app *elasticbeanstalk.ApplicationDescription) error {
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

	log.Printf("[DEBUG] Elastic Beanstalk application: %s, update appversion_lifecycle: %v", name, appversion_lifecycle)

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

	_, err := beanstalkConn.UpdateApplicationResourceLifecycle(&elasticbeanstalk.UpdateApplicationResourceLifecycleInput{
		ApplicationName:         aws.String(name),
		ResourceLifecycleConfig: rlc,
	})

	return err
}

func resourceAwsElasticBeanstalkApplicationRead(d *schema.ResourceData, meta interface{}) error {
	a, err := getBeanstalkApplication(d, meta)
	if err != nil {
		return err
	}
	if a == nil {
		return err
	}

	d.Set("name", a.ApplicationName)
	d.Set("description", a.Description)

	if a.ResourceLifecycleConfig != nil {
		d.Set("appversion_lifecycle", flattenResourceLifecycleConfig(a.ResourceLifecycleConfig))
	}

	return nil
}

func resourceAwsElasticBeanstalkApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	beanstalkConn := meta.(*AWSClient).elasticbeanstalkconn

	a, err := getBeanstalkApplication(d, meta)
	if err != nil {
		return err
	}
	_, err = beanstalkConn.DeleteApplication(&elasticbeanstalk.DeleteApplicationInput{
		ApplicationName: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}

	return resource.Retry(10*time.Second, func() *resource.RetryError {
		if a, err = getBeanstalkApplication(d, meta); a != nil {
			return resource.RetryableError(
				fmt.Errorf("Beanstalk Application (%s) still exists: %s", d.Id(), err))
		}
		return nil
	})
}

func getBeanstalkApplication(
	d *schema.ResourceData,
	meta interface{}) (*elasticbeanstalk.ApplicationDescription, error) {
	conn := meta.(*AWSClient).elasticbeanstalkconn

	resp, err := conn.DescribeApplications(&elasticbeanstalk.DescribeApplicationsInput{
		ApplicationNames: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() != "InvalidBeanstalkAppID.NotFound" {
			log.Printf("[Err] Error reading Elastic Beanstalk Application (%s): Application not found", d.Id())
			d.SetId("")
			return nil, nil
		}
		return nil, err
	}

	switch {
	case len(resp.Applications) > 1:
		return nil, fmt.Errorf("Error %d Applications matched, expected 1", len(resp.Applications))
	case len(resp.Applications) == 0:
		d.SetId("")
		return nil, nil
	default:
		return resp.Applications[0], nil
	}
}
