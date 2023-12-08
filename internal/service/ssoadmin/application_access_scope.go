// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_ssoadmin_application_access_scope")
func ResourceApplicationAccessScope() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceApplicationAccessScopeCreate,
		ReadWithoutTimeout:   resourceApplicationAccessScopeRead,
		DeleteWithoutTimeout: resourceApplicationAccessScopeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"application_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"authorized_targets": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceApplicationAccessScopeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	applicationARN := d.Get("application_arn").(string)
	scope := d.Get("scope").(string)
	id := ApplicationAccessScopeCreateResourceID(applicationARN, scope)

	input := &ssoadmin.PutApplicationAccessScopeInput{
		ApplicationArn: aws.String(applicationARN),
		Scope:          aws.String(scope),
	}

	if v, ok := d.GetOk("authorized_targets"); ok {
		input.AuthorizedTargets = flex.ExpandStringValueList(v.([]interface{}))
	}

	_, err := conn.PutApplicationAccessScope(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SSO Application Access Scope (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceApplicationAccessScopeRead(ctx, d, meta)...)
}

func resourceApplicationAccessScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	applicationARN, scope, err := ApplicationAccessScopeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := FindApplicationAccessScopeByScopeAndApplicationARN(ctx, conn, applicationARN, scope)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SSO Application Access Scope (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SSO Application Access Scope (%s): %s", d.Id(), err)
	}

	d.Set("application_arn", applicationARN)
	d.Set("scope", output.Scope)
	d.Set("authorized_targets", output.AuthorizedTargets)

	return diags
}

func resourceApplicationAccessScopeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSOAdminClient(ctx)

	applicationARN, scope, err := ApplicationAccessScopeParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting SSO Application Access Scope: %s", d.Id())
	_, err = conn.DeleteApplicationAccessScope(ctx, &ssoadmin.DeleteApplicationAccessScopeInput{
		ApplicationArn: aws.String(applicationARN),
		Scope:          aws.String(scope),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SSO Application Access Scope (%s): %s", d.Id(), err)
	}

	return diags
}

const applicationAccessScopeIDSeparator = ","

func ApplicationAccessScopeCreateResourceID(applicationARN, scope string) string {
	parts := []string{applicationARN, scope}
	id := strings.Join(parts, applicationAccessScopeIDSeparator)

	return id
}

func ApplicationAccessScopeParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, applicationAccessScopeIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPLICATION_ARN%[2]sSCOPE", id, applicationAccessScopeIDSeparator)
}

func FindApplicationAccessScopeByScopeAndApplicationARN(ctx context.Context, conn *ssoadmin.Client, applicationARN, scope string) (*ssoadmin.GetApplicationAccessScopeOutput, error) {
	input := &ssoadmin.GetApplicationAccessScopeInput{
		ApplicationArn: aws.String(applicationARN),
		Scope:          aws.String(scope),
	}

	output, err := conn.GetApplicationAccessScope(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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
