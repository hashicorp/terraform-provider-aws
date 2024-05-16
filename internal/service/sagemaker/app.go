// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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

// @SDKResource("aws_sagemaker_app", name="App")
// @Tags(identifierAttribute="arn")
func ResourceApp() *schema.Resource {
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
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.AppType_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(sagemaker.AppInstanceType_Values(), false),
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	input := &sagemaker.CreateAppInput{
		AppName:  aws.String(d.Get("app_name").(string)),
		AppType:  aws.String(d.Get("app_type").(string)),
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
		input.ResourceSpec = expandResourceSpec(v.([]interface{}))
	}

	log.Printf("[DEBUG] SageMaker App create config: %#v", *input)
	output, err := conn.CreateAppWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker App: %s", err)
	}

	appArn := aws.StringValue(output.AppArn)
	domainID, userProfileOrSpaceName, appType, appName, err := decodeAppID(appArn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker App (%s): %s", appArn, err)
	}

	d.SetId(appArn)

	if _, err := WaitAppInService(ctx, conn, domainID, userProfileOrSpaceName, appType, appName); err != nil {
		return sdkdiag.AppendErrorf(diags, "create SageMaker App (%s): waiting for completion: %s", d.Id(), err)
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	domainID, userProfileOrSpaceName, appType, appName, err := decodeAppID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker App (%s): %s", d.Id(), err)
	}

	app, err := FindAppByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)
	if err != nil {
		if !d.IsNewResource() && tfresource.NotFound(err) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker App (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker App (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(app.AppArn)
	d.Set("app_name", app.AppName)
	d.Set("app_type", app.AppType)
	d.Set(names.AttrARN, arn)
	d.Set("domain_id", app.DomainId)
	d.Set("space_name", app.SpaceName)
	d.Set("user_profile_name", app.UserProfileName)

	if err := d.Set("resource_spec", flattenResourceSpec(app.ResourceSpec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting resource_spec for SageMaker App (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceAppUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	appName := d.Get("app_name").(string)
	appType := d.Get("app_type").(string)
	domainID := d.Get("domain_id").(string)
	userProfileOrSpaceName := ""

	input := &sagemaker.DeleteAppInput{
		AppName:  aws.String(appName),
		AppType:  aws.String(appType),
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

	if _, err := conn.DeleteAppWithContext(ctx, input); err != nil {
		if tfawserr.ErrMessageContains(err, "ValidationException", "has already been deleted") ||
			tfawserr.ErrMessageContains(err, "ValidationException", "previously failed and was automatically deleted") {
			return diags
		}

		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return sdkdiag.AppendErrorf(diags, "deleting SageMaker App (%s): %s", d.Id(), err)
		}
	}

	if _, err := WaitAppDeleted(ctx, conn, domainID, userProfileOrSpaceName, appType, appName); err != nil {
		if !tfawserr.ErrCodeEquals(err, sagemaker.ErrCodeResourceNotFound) {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker App (%s) to delete: %s", d.Id(), err)
		}
	}

	return diags
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

	if appType == "jupyterserver" {
		appType = sagemaker.AppTypeJupyterServer
	} else if appType == "kernelgateway" {
		appType = sagemaker.AppTypeKernelGateway
	} else if appType == "tensorboard" {
		appType = sagemaker.AppTypeTensorBoard
	} else if appType == "rstudioserverpro" {
		appType = sagemaker.AppTypeRstudioServerPro
	} else if appType == "rsessiongateway" {
		appType = sagemaker.AppTypeRsessionGateway
	}

	appName := parts[3]

	return domainID, userProfileOrSpaceName, appType, appName, nil
}
