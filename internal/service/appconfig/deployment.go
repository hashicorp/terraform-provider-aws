// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appconfig/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appconfig_deployment", name="Deployment")
// @Tags(identifierAttribute="arn")
// @Testing(checkDestroyNoop=true)
// @Testing(importIgnore="state")
func resourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentCreate,
		ReadWithoutTimeout:   resourceDeploymentRead,
		UpdateWithoutTimeout: resourceDeploymentUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

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
	}
}

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	input := appconfig.StartDeploymentInput{
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
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.ConflictException](ctx, timeout, func() (any, error) {
		return conn.StartDeployment(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "starting AppConfig Deployment: %s", err)
	}

	output := outputRaw.(*appconfig.StartDeploymentOutput)

	d.SetId(deploymentCreateResourceID(aws.ToString(output.ApplicationId), aws.ToString(output.EnvironmentId), output.DeploymentNumber))

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID, environmentID, deploymentNumber, err := deploymentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findDeploymentByThreePartKey(ctx, conn, applicationID, environmentID, deploymentNumber)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppConfig Deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Deployment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrARN, deploymentARN(ctx, meta.(*conns.AWSClient), applicationID, environmentID, deploymentNumber))
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

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

const deploymentResourceIDSeparator = "/"

func deploymentCreateResourceID(applicationID, environmentID string, deploymentNumber int32) string {
	parts := []string{applicationID, environmentID, flex.Int32ValueToStringValue(deploymentNumber)}
	id := strings.Join(parts, deploymentResourceIDSeparator)

	return id
}

func deploymentParseResourceID(id string) (string, string, int32, error) {
	parts := strings.Split(id, deploymentResourceIDSeparator)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", 0, fmt.Errorf("unexpected format for ID (%[1]s), expected ApplicationID%[2]sEnvironmentID%[2]sDeploymentNumber", id, deploymentResourceIDSeparator)
	}

	return parts[0], parts[1], flex.StringValueToInt32Value(parts[2]), nil
}

func findDeploymentByThreePartKey(ctx context.Context, conn *appconfig.Client, applicationID, environmentID string, deploymentNumber int32) (*appconfig.GetDeploymentOutput, error) {
	input := appconfig.GetDeploymentInput{
		ApplicationId:    aws.String(applicationID),
		DeploymentNumber: aws.Int32(deploymentNumber),
		EnvironmentId:    aws.String(environmentID),
	}

	return findDeployment(ctx, conn, &input)
}

func findDeployment(ctx context.Context, conn *appconfig.Client, input *appconfig.GetDeploymentInput) (*appconfig.GetDeploymentOutput, error) {
	output, err := conn.GetDeployment(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

	return output, nil
}

func deploymentARN(ctx context.Context, c *conns.AWSClient, applicationID, environmentID string, deploymentNumber int32) string {
	return c.RegionalARN(ctx, "appconfig", "application/"+applicationID+"/environment/"+environmentID+"/deployment/"+flex.Int32ValueToStringValue(deploymentNumber))
}
