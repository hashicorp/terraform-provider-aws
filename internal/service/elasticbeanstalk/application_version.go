// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
					Type:          schema.TypeString,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"image_configuration.0.source"},
				},
				"build_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"build_configuration": {
					Type:          schema.TypeList,
					Optional:      true,
					ForceNew:      true,
					MaxItems:      1,
					ConflictsWith: []string{"image_configuration"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"artifact_name": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"code_build_service_role": {
								Type:     schema.TypeString,
								Required: true,
							},
							"compute_type": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: enum.Validate[awstypes.ComputeType](),
							},
							"image": {
								Type:     schema.TypeString,
								Required: true,
							},
							"timeout_in_minutes": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IntBetween(5, 480),
							},
						},
					},
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
				"image_configuration": {
					Type:          schema.TypeList,
					Optional:      true,
					ForceNew:      true,
					MaxItems:      1,
					ConflictsWith: []string{"build_configuration"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"build": {
								Type:         schema.TypeList,
								Optional:     true,
								ForceNew:     true,
								MaxItems:     1,
								RequiredWith: []string{names.AttrBucket, names.AttrKey},
								ExactlyOneOf: []string{"image_configuration.0.source", "image_configuration.0.build"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"architecture": {
											Type:             schema.TypeString,
											Optional:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[awstypes.ArchitectureType](),
										},
										"buildpack": {
											Type:     schema.TypeString,
											Optional: true,
											ForceNew: true,
										},
										"code_build_service_role": {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
										"compute_type": {
											Type:             schema.TypeString,
											Optional:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[awstypes.ComputeType](),
										},
										"dockerfile_location": {
											Type:     schema.TypeString,
											Optional: true,
											ForceNew: true,
										},
										"timeout_in_minutes": {
											Type:         schema.TypeInt,
											Optional:     true,
											ForceNew:     true,
											ValidateFunc: validation.IntBetween(5, 480),
										},
										names.AttrType: {
											Type:             schema.TypeString,
											Required:         true,
											ForceNew:         true,
											ValidateDiagFunc: enum.Validate[awstypes.ImageBuildType](),
										},
									},
								},
							},
							names.AttrSource: {
								Type:         schema.TypeList,
								Optional:     true,
								ForceNew:     true,
								MaxItems:     1,
								ExactlyOneOf: []string{"image_configuration.0.source", "image_configuration.0.build"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrURI: {
											Type:     schema.TypeString,
											Required: true,
											ForceNew: true,
										},
									},
								},
							},
						},
					},
				},
				"image_uri": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrKey: {
					Type:          schema.TypeString,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"image_configuration.0.source"},
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
			}
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
		Tags:            getTagsIn(ctx),
		VersionLabel:    aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrBucket); ok {
		input.SourceBundle = &awstypes.S3Location{
			S3Bucket: aws.String(v.(string)),
			S3Key:    aws.String(d.Get(names.AttrKey).(string)),
		}
	}

	if v, ok := d.GetOk("image_configuration"); ok && len(v.([]any)) > 0 {
		input.ImageConfiguration = expandImageConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("build_configuration"); ok && len(v.([]any)) > 0 {
		input.BuildConfiguration = expandBuildConfiguration(v.([]any))
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

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Elastic Beanstalk Application Version (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elastic Beanstalk Application Version (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, applicationVersion.ApplicationVersionArn)
	d.Set("build_arn", applicationVersion.BuildArn)
	d.Set(names.AttrDescription, applicationVersion.Description)
	if applicationVersion.ImageSource != nil {
		d.Set("image_uri", applicationVersion.ImageSource.Uri)
	}

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

// Expand helpers

func expandImageConfiguration(tfList []any) *awstypes.ImageConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)
	result := &awstypes.ImageConfiguration{}

	if v, ok := tfMap[names.AttrSource].([]any); ok && len(v) > 0 {
		result.Source = expandImageSource(v)
	}
	if v, ok := tfMap["build"].([]any); ok && len(v) > 0 {
		result.Build = expandImageBuildConfiguration(v)
	}

	return result
}

func expandImageSource(tfList []any) *awstypes.ImageSource {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)
	return &awstypes.ImageSource{
		Uri: aws.String(tfMap[names.AttrURI].(string)),
	}
}

func expandBuildConfiguration(tfList []any) *awstypes.BuildConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)
	result := &awstypes.BuildConfiguration{}

	if v, ok := tfMap["artifact_name"].(string); ok && v != "" {
		result.ArtifactName = aws.String(v)
	}
	if v, ok := tfMap["code_build_service_role"].(string); ok && v != "" {
		result.CodeBuildServiceRole = aws.String(v)
	}
	if v, ok := tfMap["compute_type"].(string); ok && v != "" {
		result.ComputeType = awstypes.ComputeType(v)
	}
	if v, ok := tfMap["image"].(string); ok && v != "" {
		result.Image = aws.String(v)
	}
	if v, ok := tfMap["timeout_in_minutes"].(int); ok && v > 0 {
		result.TimeoutInMinutes = aws.Int32(int32(v))
	}

	return result
}

func expandImageBuildConfiguration(tfList []any) *awstypes.ImageBuildConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}
	tfMap := tfList[0].(map[string]any)
	result := &awstypes.ImageBuildConfiguration{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		result.Type = awstypes.ImageBuildType(v)
	}
	if v, ok := tfMap["dockerfile_location"].(string); ok && v != "" {
		result.DockerfileLocation = aws.String(v)
	}
	if v, ok := tfMap["buildpack"].(string); ok && v != "" {
		result.Buildpack = aws.String(v)
	}
	if v, ok := tfMap["architecture"].(string); ok && v != "" {
		result.Architecture = awstypes.ArchitectureType(v)
	}
	if v, ok := tfMap["code_build_service_role"].(string); ok && v != "" {
		result.CodeBuildServiceRole = aws.String(v)
	}
	if v, ok := tfMap["compute_type"].(string); ok && v != "" {
		result.ComputeType = awstypes.ComputeType(v)
	}
	if v, ok := tfMap["timeout_in_minutes"].(int); ok && v > 0 {
		result.TimeoutInMinutes = aws.Int32(int32(v))
	}

	return result
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
