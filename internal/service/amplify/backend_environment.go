// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/aws/aws-sdk-go-v2/service/amplify/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_amplify_backend_environment", name="Backend Environment")
func resourceBackendEnvironment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBackendEnvironmentCreate,
		ReadWithoutTimeout:   resourceBackendEnvironmentRead,
		DeleteWithoutTimeout: resourceBackendEnvironmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"app_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_artifacts": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]{1,100}$`), "should be not be more than 100 alphanumeric or hyphen characters"),
			},
			"environment_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-z]{2,10}$`), "should be between 2 and 10 characters (only lowercase alphabetic)"),
			},
			"stack_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]{1,100}$`), "should be not be more than 100 alphanumeric or hyphen characters"),
			},
		},
	}
}

func resourceBackendEnvironmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	appID := d.Get("app_id").(string)
	environmentName := d.Get("environment_name").(string)
	id := backendEnvironmentCreateResourceID(appID, environmentName)

	input := &amplify.CreateBackendEnvironmentInput{
		AppId:           aws.String(appID),
		EnvironmentName: aws.String(environmentName),
	}

	if v, ok := d.GetOk("deployment_artifacts"); ok {
		input.DeploymentArtifacts = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stack_name"); ok {
		input.StackName = aws.String(v.(string))
	}

	_, err := conn.CreateBackendEnvironment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Amplify Backend Environment (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceBackendEnvironmentRead(ctx, d, meta)...)
}

func resourceBackendEnvironmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	appID, environmentName, err := backendEnvironmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	backendEnvironment, err := findBackendEnvironmentByTwoPartKey(ctx, conn, appID, environmentName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Amplify Backend Environment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Amplify Backend Environment (%s): %s", d.Id(), err)
	}

	d.Set("app_id", appID)
	d.Set(names.AttrARN, backendEnvironment.BackendEnvironmentArn)
	d.Set("deployment_artifacts", backendEnvironment.DeploymentArtifacts)
	d.Set("environment_name", backendEnvironment.EnvironmentName)
	d.Set("stack_name", backendEnvironment.StackName)

	return diags
}

func resourceBackendEnvironmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AmplifyClient(ctx)

	appID, environmentName, err := backendEnvironmentParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Amplify Backend Environment: %s", d.Id())
	_, err = conn.DeleteBackendEnvironment(ctx, &amplify.DeleteBackendEnvironmentInput{
		AppId:           aws.String(appID),
		EnvironmentName: aws.String(environmentName),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Amplify Backend Environment (%s): %s", d.Id(), err)
	}

	return diags
}

func findBackendEnvironmentByTwoPartKey(ctx context.Context, conn *amplify.Client, appID, environmentName string) (*types.BackendEnvironment, error) {
	input := &amplify.GetBackendEnvironmentInput{
		AppId:           aws.String(appID),
		EnvironmentName: aws.String(environmentName),
	}

	output, err := conn.GetBackendEnvironment(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BackendEnvironment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.BackendEnvironment, nil
}

const backendEnvironmentResourceIDSeparator = "/"

func backendEnvironmentCreateResourceID(appID, environmentName string) string {
	parts := []string{appID, environmentName}
	id := strings.Join(parts, backendEnvironmentResourceIDSeparator)

	return id
}

func backendEnvironmentParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, backendEnvironmentResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPID%[2]sENVIRONMENTNAME", id, backendEnvironmentResourceIDSeparator)
}
