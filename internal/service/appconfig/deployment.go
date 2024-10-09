// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_deployment", name="Deployment")
// @Tags(identifierAttribute="arn")
// @Testing(checkDestroyNoop=true)
// @Testing(importIgnore="state")
func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentCreate,
		ReadWithoutTimeout:   resourceDeploymentRead,
		UpdateWithoutTimeout: resourceDeploymentUpdate,
		DeleteWithoutTimeout: resourceDeploymentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9a-z]{4,7}`), ""),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9a-z]{4,7}`), ""),
			},
			"configuration_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"deployment_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"deployment_strategy_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`(^[0-9a-z]{4,7}$|^AppConfig\.[0-9A-Za-z]{9,40}$)`), ""),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"environment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[0-9a-z]{4,7}`), ""),
			},
			names.AttrKMSKeyARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					verify.ValidARN,
					validation.StringLenBetween(1, 256)),
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	input := &appconfig.StartDeploymentInput{
		ApplicationId:          aws.String(d.Get(names.AttrApplicationID).(string)),
		EnvironmentId:          aws.String(d.Get("environment_id").(string)),
		ConfigurationProfileId: aws.String(d.Get("configuration_profile_id").(string)),
		ConfigurationVersion:   aws.String(d.Get("configuration_version").(string)),
		DeploymentStrategyId:   aws.String(d.Get("deployment_strategy_id").(string)),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("kms_key_identifier"); ok {
		input.KmsKeyIdentifier = aws.String(v.(string))
	}

	const (
		timeout = 30 * time.Minute // AWS SDK for Go v1 compatibility.
	)
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, timeout, func() (interface{}, error) {
		return conn.StartDeployment(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting AppConfig Deployment: %s", err)
	}

	output := outputRaw.(*appconfig.StartDeploymentOutput)
	appID := aws.ToString(output.ApplicationId)
	envID := aws.ToString(output.EnvironmentId)

	d.SetId(fmt.Sprintf("%s/%s/%d", appID, envID, output.DeploymentNumber))

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	appID, envID, deploymentNum, err := DeploymentParseID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Deployment (%s): %s", d.Id(), err)
	}

	input := &appconfig.GetDeploymentInput{
		ApplicationId:    aws.String(appID),
		DeploymentNumber: aws.Int32(deploymentNum),
		EnvironmentId:    aws.String(envID),
	}

	output, err := conn.GetDeployment(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Appconfig Deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Deployment (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Deployment (%s): empty response", d.Id())
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("application/%s/environment/%s/deployment/%d", aws.ToString(output.ApplicationId), aws.ToString(output.EnvironmentId), output.DeploymentNumber),
		Service:   "appconfig",
	}.String()

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrARN, arn)
	d.Set("configuration_profile_id", output.ConfigurationProfileId)
	d.Set("configuration_version", output.ConfigurationVersion)
	d.Set("deployment_number", output.DeploymentNumber)
	d.Set("deployment_strategy_id", output.DeploymentStrategyId)
	d.Set(names.AttrDescription, output.Description)
	d.Set("environment_id", output.EnvironmentId)
	d.Set(names.AttrKMSKeyARN, output.KmsKeyArn)
	d.Set("kms_key_identifier", output.KmsKeyIdentifier)
	d.Set(names.AttrState, output.State)

	return diags
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentDelete(ctx context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[WARN] Cannot destroy AppConfig Deployment. Terraform will remove this resource from the state file, however this resource remains.")
	return diags
}

func DeploymentParseID(id string) (string, string, int32, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, fmt.Errorf("unexpected format of ID (%q), expected ApplicationID:EnvironmentID:DeploymentNumber", id)
	}

	num, err := strconv.ParseInt(parts[2], 0, 32)
	if err != nil {
		return "", "", 0, fmt.Errorf("parsing AppConfig Deployment resource ID deployment_number: %w", err)
	}

	return parts[0], parts[1], int32(num), nil
}
