package elasticbeanstalk

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceApplicationVersion() *schema.Resource {
	return &schema.Resource{
		Create: resourceApplicationVersionCreate,
		Read:   resourceApplicationVersionRead,
		Update: resourceApplicationVersionUpdate,
		Delete: resourceApplicationVersionDelete,

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"application": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"force_delete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationVersionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	application := d.Get("application").(string)
	description := d.Get("description").(string)
	bucket := d.Get("bucket").(string)
	key := d.Get("key").(string)
	name := d.Get("name").(string)

	s3Location := elasticbeanstalk.S3Location{
		S3Bucket: aws.String(bucket),
		S3Key:    aws.String(key),
	}

	createOpts := elasticbeanstalk.CreateApplicationVersionInput{
		ApplicationName: aws.String(application),
		Description:     aws.String(description),
		SourceBundle:    &s3Location,
		Tags:            Tags(tags.IgnoreElasticbeanstalk()),
		VersionLabel:    aws.String(name),
	}

	log.Printf("[DEBUG] Elastic Beanstalk Application Version create opts: %s", createOpts)
	_, err := conn.CreateApplicationVersion(&createOpts)
	if err != nil {
		return err
	}

	d.SetId(name)
	log.Printf("[INFO] Elastic Beanstalk Application Version Label: %s", name)

	return resourceApplicationVersionRead(d, meta)
}

func resourceApplicationVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := conn.DescribeApplicationVersions(&elasticbeanstalk.DescribeApplicationVersionsInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		VersionLabels:   []*string{aws.String(d.Id())},
	})
	if err != nil {
		return err
	}

	if len(resp.ApplicationVersions) == 0 {
		log.Printf("[DEBUG] Elastic Beanstalk application version read: application version not found")

		d.SetId("")

		return nil
	} else if len(resp.ApplicationVersions) != 1 {
		return fmt.Errorf("Error reading application version properties: found %d versions of label %q, expected 1",
			len(resp.ApplicationVersions), d.Id())
	}

	arn := aws.StringValue(resp.ApplicationVersions[0].ApplicationVersionArn)
	d.Set("arn", arn)
	d.Set("description", resp.ApplicationVersions[0].Description)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for Elastic Beanstalk Application version (%s): %w", arn, err)
	}

	tags = tags.IgnoreElasticbeanstalk().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceApplicationVersionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn

	if d.HasChange("description") {
		if err := resourceApplicationVersionDescriptionUpdate(conn, d); err != nil {
			return err
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating Elastic Beanstalk Application version (%s) tags: %s", arn, err)
		}
	}

	return resourceApplicationVersionRead(d, meta)

}

func resourceApplicationVersionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticBeanstalkConn

	application := d.Get("application").(string)
	name := d.Id()

	if !d.Get("force_delete").(bool) {
		environments, err := versionUsedBy(application, name, conn)
		if err != nil {
			return err
		}

		if len(environments) > 1 {
			return fmt.Errorf("Unable to delete Application Version, it is currently in use by the following environments: %s.", environments)
		}
	}
	_, err := conn.DeleteApplicationVersion(&elasticbeanstalk.DeleteApplicationVersionInput{
		ApplicationName:    aws.String(application),
		VersionLabel:       aws.String(name),
		DeleteSourceBundle: aws.Bool(false),
	})

	// application version is pending delete, or no longer exists.
	if tfawserr.ErrCodeEquals(err, "InvalidParameterValue") {
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func resourceApplicationVersionDescriptionUpdate(conn *elasticbeanstalk.ElasticBeanstalk, d *schema.ResourceData) error {
	application := d.Get("application").(string)
	description := d.Get("description").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Elastic Beanstalk application version: %s, update description: %s", name, description)

	_, err := conn.UpdateApplicationVersion(&elasticbeanstalk.UpdateApplicationVersionInput{
		ApplicationName: aws.String(application),
		Description:     aws.String(description),
		VersionLabel:    aws.String(name),
	})

	return err
}

func versionUsedBy(applicationName, versionLabel string, conn *elasticbeanstalk.ElasticBeanstalk) ([]string, error) {
	now := time.Now()
	resp, err := conn.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{
		ApplicationName:       aws.String(applicationName),
		VersionLabel:          aws.String(versionLabel),
		IncludeDeleted:        aws.Bool(true),
		IncludedDeletedBackTo: aws.Time(now.Add(-1 * time.Minute)),
	})

	if err != nil {
		return nil, err
	}

	var environmentIDs []string
	for _, environment := range resp.Environments {
		environmentIDs = append(environmentIDs, *environment.EnvironmentId)
	}

	return environmentIDs, nil
}
