// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_app", name="App")
// @Tags(identifierAttribute="arn")
func resourceApp() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAppCreate,
		ReadWithoutTimeout:   resourceAppRead,
		UpdateWithoutTimeout: resourceAppUpdate,
		DeleteWithoutTimeout: resourceAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,62}`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"app_type": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AppType](),
			},
			"domain_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"resource_spec": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrInstanceType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AppInstanceType](),
						},
						"lifecycle_config_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"sagemaker_image_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidARN,
						},
						"sagemaker_image_version_alias": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"sagemaker_image_version_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"space_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"space_name", "user_profile_name"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"user_profile_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"space_name", "user_profile_name"},
			},
		},
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.CreateAppInput{
		AppName:  aws.String(d.Get("app_name").(string)),
		AppType:  awstypes.AppType(d.Get("app_type").(string)),
		DomainId: aws.String(d.Get("domain_id").(string)),
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("user_profile_name"); ok {
		input.UserProfileName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("space_name"); ok {
		input.SpaceName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_spec"); ok {
		input.ResourceSpec = expandResourceSpec(v.([]any))
	}

	log.Printf("[DEBUG] SageMaker AI App create config: %#v", *input)
	output, err := conn.CreateApp(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI App: %s", err)
	}

	appArn := aws.ToString(output.AppArn)
	domainID, userProfileOrSpaceName, appType, appName, err := decodeAppID(appArn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI App (%s): %s", appArn, err)
	}

	d.SetId(appArn)

	if _, err := waitAppInService(ctx, conn, domainID, userProfileOrSpaceName, appType, appName); err != nil {
		return sdkdiag.AppendErrorf(diags, "create SageMaker AI App (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	domainID, userProfileOrSpaceName, appType, appName, err := decodeAppID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI App (%s): %s", d.Id(), err)
	}

	app, err := findAppByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI App (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI App (%s): %s", d.Id(), err)
	}

	d.Set("app_name", app.AppName)
	d.Set("app_type", app.AppType)
	d.Set(names.AttrARN, app.AppArn)
	d.Set("domain_id", app.DomainId)
	d.Set("space_name", app.SpaceName)
	d.Set("user_profile_name", app.UserProfileName)

	if err := d.Set("resource_spec", flattenResourceSpec(app.ResourceSpec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resource_spec for SageMaker AI App (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	appName := d.Get("app_name").(string)
	appType := d.Get("app_type").(string)
	domainID := d.Get("domain_id").(string)
	userProfileOrSpaceName := ""

	input := &sagemaker.DeleteAppInput{
		AppName:  aws.String(appName),
		AppType:  awstypes.AppType(appType),
		DomainId: aws.String(domainID),
	}

	if v, ok := d.GetOk("user_profile_name"); ok {
		input.UserProfileName = aws.String(v.(string))
		userProfileOrSpaceName = v.(string)
	}

	if v, ok := d.GetOk("space_name"); ok {
		input.SpaceName = aws.String(v.(string))
		userProfileOrSpaceName = v.(string)
	}

	if _, err := conn.DeleteApp(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "has already been deleted") ||
			tfawserr.ErrMessageContains(err, ErrCodeValidationException, "previously failed and was automatically deleted") {
			return diags
		}

		if !errs.IsA[*awstypes.ResourceNotFound](err) {
			return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI App (%s): %s", d.Id(), err)
		}
	}

	if _, err := waitAppDeleted(ctx, conn, domainID, userProfileOrSpaceName, appType, appName); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI App (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findAppByName(ctx context.Context, conn *sagemaker.Client, domainID, userProfileOrSpaceName, appType, appName string) (*sagemaker.DescribeAppOutput, error) {
	foundApp, err := listAppsByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

	if err != nil {
		return nil, err
	}

	input := &sagemaker.DescribeAppInput{
		AppName:  aws.String(appName),
		AppType:  awstypes.AppType(appType),
		DomainId: aws.String(domainID),
	}
	if foundApp.SpaceName != nil {
		input.SpaceName = foundApp.SpaceName
	}
	if foundApp.UserProfileName != nil {
		input.UserProfileName = foundApp.UserProfileName
	}

	output, err := conn.DescribeApp(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "RecordNotFound") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if state := output.Status; state == awstypes.AppStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func listAppsByName(ctx context.Context, conn *sagemaker.Client, domainID, userProfileOrSpaceName, appType, appName string) (*awstypes.AppDetails, error) {
	input := &sagemaker.ListAppsInput{
		DomainIdEquals: aws.String(domainID),
	}
	var output []awstypes.AppDetails

	pages := sagemaker.NewListAppsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Apps...)
	}

	for _, v := range output {
		if aws.ToString(v.AppName) == appName && string(v.AppType) == appType && (aws.ToString(v.SpaceName) == userProfileOrSpaceName || aws.ToString(v.UserProfileName) == userProfileOrSpaceName) {
			return &v, nil
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func decodeAppID(id string) (string, string, string, string, error) {
	appArn, err := arn.Parse(id)
	if err != nil {
		return "", "", "", "", err
	}

	appResourceName := strings.TrimPrefix(appArn.Resource, "app/")
	parts := strings.Split(appResourceName, "/")

	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("unexpected format of ID (%q), expected DOMAIN-ID/USER-PROFILE-NAME OR PROFILE-NAME/APP-TYPE/APP-NAME", appResourceName)
	}

	domainID := parts[0]
	userProfileOrSpaceName := parts[1]
	appType := parts[2]

	for _, appTypeValue := range awstypes.AppType("").Values() {
		if appType == strings.ToLower(string(appTypeValue)) {
			appType = string(appTypeValue)
			break
		}
	}

	appName := parts[3]

	return domainID, userProfileOrSpaceName, appType, appName, nil
}
