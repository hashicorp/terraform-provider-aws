// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appstream_image_builder", name="Image Builder")
// @Tags(identifierAttribute="arn")
func resourceImageBuilder() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageBuilderCreate,
		ReadWithoutTimeout:   resourceImageBuilderRead,
		UpdateWithoutTimeout: resourceImageBuilderUpdate,
		DeleteWithoutTimeout: resourceImageBuilderDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_endpoint": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 4,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEndpointType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AccessEndpointType](),
						},
						"vpce_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"appstream_agent_version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			names.AttrDisplayName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 100),
			},
			"domain_join_info": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"organizational_unit_distinguished_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"enable_default_internet_access": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrIAMRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"image_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image_arn", "image_name"},
				ValidateFunc: verify.ValidARN,
			},
			"image_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"image_name", "image_arn"},
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceImageBuilderCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := appstream.CreateImageBuilderInput{
		InstanceType: aws.String(d.Get(names.AttrInstanceType).(string)),
		Name:         aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_endpoint"); ok && v.(*schema.Set).Len() > 0 {
		input.AccessEndpoints = expandAccessEndpoints(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("appstream_agent_version"); ok {
		input.AppstreamAgentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDisplayName); ok {
		input.DisplayName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("domain_join_info"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.DomainJoinInfo = expandDomainJoinInfo(v.([]any))
	}

	if v, ok := d.GetOk("enable_default_internet_access"); ok {
		input.EnableDefaultInternetAccess = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrIAMRoleARN); ok {
		input.IamRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_arn"); ok {
		input.ImageArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_name"); ok {
		input.ImageName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.VpcConfig = expandImageBuilderVPCConfig(v.([]any))
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidRoleException](ctx, propagationTimeout, func() (any, error) {
		return conn.CreateImageBuilder(ctx, &input)
	}, "encountered an error because your IAM role")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppStream ImageBuilder (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*appstream.CreateImageBuilderOutput).ImageBuilder.Name))

	if _, err = waitImageBuilderRunning(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream ImageBuilder (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceImageBuilderRead(ctx, d, meta)...)
}

func resourceImageBuilderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	imageBuilder, err := findImageBuilderByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppStream ImageBuilder (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppStream ImageBuilder (%s): %s", d.Id(), err)
	}

	if err = d.Set("access_endpoint", flattenAccessEndpoints(imageBuilder.AccessEndpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_endpoint: %s", err)
	}
	d.Set("appstream_agent_version", imageBuilder.AppstreamAgentVersion)
	arn := aws.ToString(imageBuilder.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedTime, aws.ToTime(imageBuilder.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, imageBuilder.Description)
	d.Set(names.AttrDisplayName, imageBuilder.DisplayName)
	if imageBuilder.DomainJoinInfo != nil {
		if err = d.Set("domain_join_info", []any{flattenDomainInfo(imageBuilder.DomainJoinInfo)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting domain_join_info: %s", err)
		}
	} else {
		d.Set("domain_join_info", nil)
	}
	d.Set("enable_default_internet_access", imageBuilder.EnableDefaultInternetAccess)
	d.Set(names.AttrIAMRoleARN, imageBuilder.IamRoleArn)
	d.Set("image_arn", imageBuilder.ImageArn)
	d.Set(names.AttrInstanceType, imageBuilder.InstanceType)
	d.Set(names.AttrName, imageBuilder.Name)
	d.Set(names.AttrState, imageBuilder.State)
	if imageBuilder.VpcConfig != nil {
		if err = d.Set(names.AttrVPCConfig, []any{flattenVPCConfig(imageBuilder.VpcConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
		}
	} else {
		d.Set(names.AttrVPCConfig, nil)
	}

	return diags
}

func resourceImageBuilderUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Tags only.
	return resourceImageBuilderRead(ctx, d, meta)
}

func resourceImageBuilderDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppStreamClient(ctx)

	log.Printf("[DEBUG] Deleting AppStream ImageBuilder: %s", d.Id())
	input := appstream.DeleteImageBuilderInput{
		Name: aws.String(d.Id()),
	}
	_, err := conn.DeleteImageBuilder(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppStream ImageBuilder (%s): %s", d.Id(), err)
	}

	if _, err = waitImageBuilderDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppStream ImageBuilder (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findImageBuilderByID(ctx context.Context, conn *appstream.Client, id string) (*awstypes.ImageBuilder, error) {
	input := &appstream.DescribeImageBuildersInput{
		Names: []string{id},
	}

	return findImageBuilder(ctx, conn, input)
}

func findImageBuilder(ctx context.Context, conn *appstream.Client, input *appstream.DescribeImageBuildersInput) (*awstypes.ImageBuilder, error) {
	output, err := findImageBuilders(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findImageBuilders(ctx context.Context, conn *appstream.Client, input *appstream.DescribeImageBuildersInput) ([]awstypes.ImageBuilder, error) {
	var output []awstypes.ImageBuilder

	err := describeImageBuildersPages(ctx, conn, input, func(page *appstream.DescribeImageBuildersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.ImageBuilders...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusImageBuilder(ctx context.Context, conn *appstream.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findImageBuilderByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func waitImageBuilderRunning(ctx context.Context, conn *appstream.Client, id string) (*awstypes.ImageBuilder, error) {
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageBuilderStatePending),
		Target:  enum.Slice(awstypes.ImageBuilderStateRunning),
		Refresh: statusImageBuilder(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ImageBuilder); ok {
		tfresource.SetLastError(err, resourcesError(output.ImageBuilderErrors))

		return output, err
	}

	return nil, err
}

func waitImageBuilderDeleted(ctx context.Context, conn *appstream.Client, id string) (*awstypes.ImageBuilder, error) {
	const (
		timeout = 60 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageBuilderStatePending, awstypes.ImageBuilderStateDeleting),
		Target:  []string{},
		Refresh: statusImageBuilder(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ImageBuilder); ok {
		tfresource.SetLastError(err, resourcesError(output.ImageBuilderErrors))

		return output, err
	}

	return nil, err
}

func resourceError(apiObject *awstypes.ResourceError) error {
	if apiObject == nil {
		return nil
	}

	return errs.APIError(apiObject.ErrorCode, aws.ToString(apiObject.ErrorMessage))
}

func resourcesError(apiObjects []awstypes.ResourceError) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := resourceError(&apiObject); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// Differs from expandVPCConfig due to use of TypeSet.
func expandImageBuilderVPCConfig(tfList []any) *awstypes.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)

	if !ok {
		return nil
	}

	apiObject := &awstypes.VpcConfig{}

	if v, ok := tfMap[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}
