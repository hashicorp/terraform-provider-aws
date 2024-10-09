// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_source_credential", name="Source Credential")
func resourceSourceCredential() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSourceCredentialCreate,
		ReadWithoutTimeout:   resourceSourceCredentialRead,
		DeleteWithoutTimeout: resourceSourceCredentialDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auth_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.AuthType](),
			},
			"server_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.ServerType](),
			},
			"token": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			names.AttrUserName: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceSourceCredentialCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	authType := types.AuthType(d.Get("auth_type").(string))
	input := &codebuild.ImportSourceCredentialsInput{
		AuthType:   authType,
		ServerType: types.ServerType(d.Get("server_type").(string)),
		Token:      aws.String(d.Get("token").(string)),
	}

	if attr, ok := d.GetOk(names.AttrUserName); ok && authType == types.AuthTypeBasicAuth {
		input.Username = aws.String(attr.(string))
	}

	output, err := conn.ImportSourceCredentials(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeBuild Source Credential: %s", err)
	}

	d.SetId(aws.ToString(output.Arn))

	return append(diags, resourceSourceCredentialRead(ctx, d, meta)...)
}

func resourceSourceCredentialRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	credentials, err := findSourceCredentialsByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeBuild Source Credential (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeBuild Source Credential (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, credentials.Arn)
	d.Set("auth_type", credentials.AuthType)
	d.Set("server_type", credentials.ServerType)

	return diags
}

func resourceSourceCredentialDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	log.Printf("[INFO] Deleting CodeBuild Source Credential: %s", d.Id())
	_, err := conn.DeleteSourceCredentials(ctx, &codebuild.DeleteSourceCredentialsInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Source Credential (%s): %s", d.Id(), err)
	}

	return diags
}

func findSourceCredentialsByARN(ctx context.Context, conn *codebuild.Client, arn string) (*types.SourceCredentialsInfo, error) {
	input := &codebuild.ListSourceCredentialsInput{}
	output, err := findSourceCredentials(ctx, conn, input, func(v *types.SourceCredentialsInfo) bool {
		return aws.ToString(v.Arn) == arn
	})

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSourceCredentials(ctx context.Context, conn *codebuild.Client, input *codebuild.ListSourceCredentialsInput, filter tfslices.Predicate[*types.SourceCredentialsInfo]) ([]types.SourceCredentialsInfo, error) {
	var sourceCredentials []types.SourceCredentialsInfo
	output, err := conn.ListSourceCredentials(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, v := range output.SourceCredentialsInfos {
		if filter(&v) {
			sourceCredentials = append(sourceCredentials, v)
		}
	}

	return sourceCredentials, nil
}
