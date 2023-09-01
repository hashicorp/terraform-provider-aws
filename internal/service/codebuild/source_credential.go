// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_codebuild_source_credential")
func ResourceSourceCredential() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSourceCredentialCreate,
		ReadWithoutTimeout:   resourceSourceCredentialRead,
		DeleteWithoutTimeout: resourceSourceCredentialDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codebuild.AuthType_Values(), false),
			},
			"server_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(codebuild.ServerType_Values(), false),
			},
			"token": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSourceCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	authType := d.Get("auth_type").(string)

	createOpts := &codebuild.ImportSourceCredentialsInput{
		AuthType:   aws.String(authType),
		ServerType: aws.String(d.Get("server_type").(string)),
		Token:      aws.String(d.Get("token").(string)),
	}

	if attr, ok := d.GetOk("user_name"); ok && authType == codebuild.AuthTypeBasicAuth {
		createOpts.Username = aws.String(attr.(string))
	}

	resp, err := conn.ImportSourceCredentialsWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "importing source credentials: %s", err)
	}

	d.SetId(aws.StringValue(resp.Arn))

	return append(diags, resourceSourceCredentialRead(ctx, d, meta)...)
}

func resourceSourceCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	resp, err := FindSourceCredentialByARN(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeBuild Source Credential (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeBuild Source Credential (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.Arn)
	d.Set("auth_type", resp.AuthType)
	d.Set("server_type", resp.ServerType)

	return diags
}

func resourceSourceCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	deleteOpts := &codebuild.DeleteSourceCredentialsInput{
		Arn: aws.String(d.Id()),
	}

	if _, err := conn.DeleteSourceCredentialsWithContext(ctx, deleteOpts); err != nil {
		if tfawserr.ErrCodeEquals(err, codebuild.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Source Credentials(%s): %s", d.Id(), err)
	}

	return diags
}
