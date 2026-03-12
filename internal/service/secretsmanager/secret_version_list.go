// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_secretsmanager_secret_version")
func secretVersionResourceAsListResource() inttypes.ListResourceForSDK {
	l := secretVersionListResource{}
	l.SetResourceSchema(resourceSecretVersion())
	return &l
}

type secretVersionListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type secretVersionListResourceModel struct {
	framework.WithRegionModel
	SecretID types.String `tfsdk:"secret_id"`
}

func (l *secretVersionListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"secret_id": listschema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]listschema.Block{},
	}
}

func (l *secretVersionListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	var query secretVersionListResourceModel
	if diags := request.Config.Get(ctx, &query); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	awsClient := l.Meta()
	conn := awsClient.SecretsManagerClient(ctx)

	secretID := query.SecretID.ValueString()

	tflog.Info(ctx, "Listing Secrets Manager secret versions", map[string]any{
		"secret_id": secretID,
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		input := &secretsmanager.ListSecretVersionIdsInput{
			SecretId:          aws.String(secretID),
			IncludeDeprecated: aws.Bool(true),
		}

		paginator := secretsmanager.NewListSecretVersionIdsPaginator(conn, input)

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			for _, version := range page.Versions {
				versionID := aws.ToString(version.VersionId)
				ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), versionID)

				result := request.NewListResult(ctx)

				rd := l.ResourceData()
				rd.SetId(secretVersionCreateResourceID(secretID, versionID))
				rd.Set("secret_id", secretID)
				rd.Set("version_id", versionID)

				diags := resourceSecretVersionRead(ctx, rd, awsClient)
				if diags.HasError() || rd.Id() == "" {
					tflog.Error(ctx, "Reading Secrets Manager secret version", map[string]any{
						names.AttrID: versionID,
						"diags":      sdkdiag.DiagnosticsString(diags),
					})
					continue
				}

				result.DisplayName = versionID

				l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
				if result.Diagnostics.HasError() {
					yield(result)
					return
				}

				if !yield(result) {
					return
				}
			}
		}
	}
}
