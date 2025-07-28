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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elastic_beanstalk_application_version", name="Application Version")
// @Tags(identifierAttribute="arn")
func resourceApplicationVersion() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationVersionCreate,
		ReadWithoutTimeout:   resourceApplicationVersionRead,
		UpdateWithoutTimeout: resourceApplicationVersionUpdate,
		DeleteWithoutTimeout: resourceApplicationVersionDelete,

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
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrForceDelete: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			"process": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceApplicationVersionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &elasticbeanstalk.CreateApplicationVersionInput{
		ApplicationName: aws.String(d.Get("application").(string)),
		Description:     aws.String(d.Get(names.AttrDescription).(string)),
		Process:         aws.Bool(d.Get("process").(bool)),
		SourceBundle: &awstypes.S3Location{
			S3Bucket: aws.String(d.Get(names.AttrBucket).(string)),
			S3Key:    aws.String(d.Get(names.AttrKey).(string)),
		},
		Tags:         getTagsIn(ctx),
		VersionLabel: aws.String(name),
	}

	_, err := conn.CreateApplicationVersion(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elastic Beanstalk Application Version (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceApplicationVersionRead(ctx, d, meta)...)
}

func resourceApplicationVersionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	applicationVersion, err := findApplicationVersionByTwoPartKey(ctx, conn, d.Get("application").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elastic Beanstalk Application Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Application Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, applicationVersion.ApplicationVersionArn)
	d.Set(names.AttrDescription, applicationVersion.Description)

	return diags
}

func resourceApplicationVersionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	if d.HasChange(names.AttrDescription) {
		input := &elasticbeanstalk.UpdateApplicationVersionInput{
			ApplicationName: aws.String(d.Get("application").(string)),
			Description:     aws.String(d.Get(names.AttrDescription).(string)),
			VersionLabel:    aws.String(d.Id()),
		}

		_, err := conn.UpdateApplicationVersion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elastic Beanstalk Application Version (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceApplicationVersionRead(ctx, d, meta)...)
}

func resourceApplicationVersionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticBeanstalkClient(ctx)

	applicationName := d.Get("application").(string)

	if !d.Get(names.AttrForceDelete).(bool) {
		now := time.Now()
		input := &elasticbeanstalk.DescribeEnvironmentsInput{
			ApplicationName:       aws.String(applicationName),
			IncludeDeleted:        aws.Bool(true),
			IncludedDeletedBackTo: aws.Time(now.Add(-1 * time.Minute)),
			VersionLabel:          aws.String(d.Id()),
		}

		environments, err := findEnvironments(ctx, conn, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Environments: %s", err)
		}

		environmentIDs := tfslices.ApplyToAll(environments, func(v awstypes.EnvironmentDescription) string {
			return aws.ToString(v.EnvironmentId)
		})

		if len(environmentIDs) > 1 {
			return sdkdiag.AppendErrorf(diags, "Elastic Beanstalk Application Version (%s) is currently in use by the following environments: %s", d.Id(), environmentIDs)
		}
	}

	_, err := conn.DeleteApplicationVersion(ctx, &elasticbeanstalk.DeleteApplicationVersionInput{
		ApplicationName:    aws.String(applicationName),
		DeleteSourceBundle: aws.Bool(false),
		VersionLabel:       aws.String(d.Id()),
	})

	// application version is pending delete, or no longer exists.
	if tfawserr.ErrCodeEquals(err, errCodeInvalidParameterValue) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elastic Beanstalk Application Version (%s): %s", d.Id(), err)
	}

	return diags
}

func findApplicationVersionByTwoPartKey(ctx context.Context, conn *elasticbeanstalk.Client, applicationName, versionLabel string) (*awstypes.ApplicationVersionDescription, error) {
	input := &elasticbeanstalk.DescribeApplicationVersionsInput{
		ApplicationName: aws.String(applicationName),
		VersionLabels:   []string{versionLabel},
	}

	return findApplicationVersion(ctx, conn, input)
}

func findApplicationVersion(ctx context.Context, conn *elasticbeanstalk.Client, input *elasticbeanstalk.DescribeApplicationVersionsInput) (*awstypes.ApplicationVersionDescription, error) {
	output, err := findApplicationVersions(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findApplicationVersions(ctx context.Context, conn *elasticbeanstalk.Client, input *elasticbeanstalk.DescribeApplicationVersionsInput) ([]awstypes.ApplicationVersionDescription, error) {
	var output []awstypes.ApplicationVersionDescription

	err := describeApplicationVersionsPages(ctx, conn, input, func(page *elasticbeanstalk.DescribeApplicationVersionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.ApplicationVersions...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
