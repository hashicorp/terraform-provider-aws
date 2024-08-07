// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
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

// @SDKResource("aws_codedeploy_app", name="App")
// @Tags(identifierAttribute="arn")
func resourceApp() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAppCreate,
		ReadWithoutTimeout:   resourceAppRead,
		UpdateWithoutTimeout: resourceUpdate,
		DeleteWithoutTimeout: resourceAppDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), appResourceIDSeparator)

				if len(idParts) == 2 {
					return []*schema.ResourceData{d}, nil
				}

				applicationName := d.Id()
				conn := meta.(*conns.AWSClient).DeployClient(ctx)

				application, err := findApplicationByName(ctx, conn, applicationName)

				if err != nil {
					return []*schema.ResourceData{}, fmt.Errorf("reading CodeDeploy Application (%s): %w", applicationName, err)
				}

				d.SetId(appCreateResourceID(aws.ToString(application.ApplicationId), applicationName))
				d.Set(names.AttrName, applicationName)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compute_platform": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ComputePlatform](),
				Default:          types.ComputePlatformServer,
			},
			"github_account_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"linked_to_github": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	application := d.Get(names.AttrName).(string)
	input := &codedeploy.CreateApplicationInput{
		ApplicationName: aws.String(application),
		ComputePlatform: types.ComputePlatform(d.Get("compute_platform").(string)),
		Tags:            getTagsIn(ctx),
	}

	output, err := conn.CreateApplication(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeDeploy Application (%s): %s", application, err)
	}

	// Despite giving the application a unique ID, AWS doesn't actually use
	// it in API calls. Use it and the app name to identify the resource in
	// the state file. This allows us to reliably detect both when the TF
	// config file changes and when the user deletes the app without removing
	// it first from the TF config.
	d.SetId(appCreateResourceID(aws.ToString(output.ApplicationId), application))

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	application := appParseResourceID(d.Id())
	name := d.Get(names.AttrName).(string)
	if name != "" && application != name {
		application = name
	}

	app, err := findApplicationByName(ctx, conn, application)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeDeploy Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeDeploy Application (%s): %s", d.Id(), err)
	}

	appName := aws.ToString(app.ApplicationName)

	if !strings.Contains(d.Id(), appName) {
		d.SetId(appCreateResourceID(aws.ToString(app.ApplicationId), appName))
	}

	d.Set(names.AttrApplicationID, app.ApplicationId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "codedeploy",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("application:%s", appName),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("compute_platform", app.ComputePlatform)
	d.Set("github_account_name", app.GitHubAccountName)
	d.Set("linked_to_github", app.LinkedToGitHub)
	d.Set(names.AttrName, appName)

	return diags
}

func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	if d.HasChange(names.AttrName) {
		o, n := d.GetChange(names.AttrName)
		input := &codedeploy.UpdateApplicationInput{
			ApplicationName:    aws.String(o.(string)),
			NewApplicationName: aws.String(n.(string)),
		}

		_, err := conn.UpdateApplication(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeDeploy Application (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceAppDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DeployClient(ctx)

	log.Printf("[INFO] Deleting CodeDeploy Application: %s", d.Id())
	_, err := conn.DeleteApplication(ctx, &codedeploy.DeleteApplicationInput{
		ApplicationName: aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsA[*types.ApplicationDoesNotExistException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeDeploy Application (%s): %s", d.Id(), err)
	}

	return diags
}

const appResourceIDSeparator = ":"

func appCreateResourceID(appID, name string) string {
	parts := []string{appID, name}
	id := strings.Join(parts, appResourceIDSeparator)

	return id
}

func appParseResourceID(id string) string {
	parts := strings.SplitN(id, appResourceIDSeparator, 2)
	// We currently omit the application ID as it is not currently used anywhere
	return parts[1]
}

func findApplicationByName(ctx context.Context, conn *codedeploy.Client, name string) (*types.ApplicationInfo, error) {
	input := &codedeploy.GetApplicationInput{
		ApplicationName: aws.String(name),
	}

	output, err := conn.GetApplication(ctx, input)

	if errs.IsA[*types.ApplicationDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Application == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Application, nil
}
