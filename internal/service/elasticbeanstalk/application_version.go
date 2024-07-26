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
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elastic_beanstalk_application_version", name="Application Version")
// @Tags(identifierAttribute="arn")
func ResourceApplicationVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationVersionCreate,
		ReadWithoutTimeout:   resourceApplicationVersionRead,
		UpdateWithoutTimeout: resourceApplicationVersionUpdate,
		DeleteWithoutTimeout: resourceApplicationVersionDelete,

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"application": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrKey: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrForceDelete: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationVersionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	application := d.Get("application").(string)
	description := d.Get(names.AttrDescription).(string)
	bucket := d.Get(names.AttrBucket).(string)
	key := d.Get(names.AttrKey).(string)
	name := d.Get(names.AttrName).(string)

	s3Location := awstypes.S3Location{
		S3Bucket: aws.String(bucket),
		S3Key:    aws.String(key),
	}

	createOpts := elasticbeanstalk.CreateApplicationVersionInput{
		ApplicationName: aws.String(application),
		Description:     aws.String(description),
		SourceBundle:    &s3Location,
		Tags:            getTagsIn(ctx),
		VersionLabel:    aws.String(name),
	}

	_, err := conn.CreateApplicationVersion(ctx, &createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Application Version (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceApplicationVersionRead(ctx, d, meta)...)
}

func resourceApplicationVersionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	resp, err := conn.DescribeApplicationVersions(ctx, &elasticbeanstalk.DescribeApplicationVersionsInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		VersionLabels:   []string{d.Id()},
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Application Version (%s): %s", d.Id(), err)
	}

	if len(resp.ApplicationVersions) == 0 {
		log.Printf("[DEBUG] Elastic Beanstalk application version read: application version not found")

		d.SetId("")

		return diags
	} else if len(resp.ApplicationVersions) != 1 {
		return sdkdiag.AppendErrorf(diags, "reading application version properties: found %d versions of label %q, expected 1",
			len(resp.ApplicationVersions), d.Id())
	}

	arn := aws.ToString(resp.ApplicationVersions[0].ApplicationVersionArn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, resp.ApplicationVersions[0].Description)

	return diags
}

func resourceApplicationVersionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	if d.HasChange(names.AttrDescription) {
		if err := resourceApplicationVersionDescriptionUpdate(ctx, conn, d); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application Version (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationVersionRead(ctx, d, meta)...)
}

func resourceApplicationVersionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	application := d.Get("application").(string)
	name := d.Id()

	if !d.Get(names.AttrForceDelete).(bool) {
		environments, err := versionUsedBy(ctx, application, name, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Application Version (%s): %s", d.Id(), err)
		}

		if len(environments) > 1 {
			return sdkdiag.AppendErrorf(diags, "Unable to delete Application Version, it is currently in use by the following environments: %s.", environments)
		}
	}
	_, err := conn.DeleteApplicationVersion(ctx, &elasticbeanstalk.DeleteApplicationVersionInput{
		ApplicationName:    aws.String(application),
		VersionLabel:       aws.String(name),
		DeleteSourceBundle: aws.Bool(false),
	})

	// application version is pending delete, or no longer exists.
	if tfawserr.ErrCodeEquals(err, "InvalidParameterValue") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Application version (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceApplicationVersionDescriptionUpdate(ctx context.Context, conn *elasticbeanstalk.Client, d *schema.ResourceData) error {
	application := d.Get("application").(string)
	description := d.Get(names.AttrDescription).(string)
	name := d.Get(names.AttrName).(string)

	_, err := conn.UpdateApplicationVersion(ctx, &elasticbeanstalk.UpdateApplicationVersionInput{
		ApplicationName: aws.String(application),
		Description:     aws.String(description),
		VersionLabel:    aws.String(name),
	})

	return err
}

func versionUsedBy(ctx context.Context, applicationName, versionLabel string, conn *elasticbeanstalk.Client) ([]string, error) {
	now := time.Now()
	resp, err := conn.DescribeEnvironments(ctx, &elasticbeanstalk.DescribeEnvironmentsInput{
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
