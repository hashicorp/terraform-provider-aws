// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/continuum"
	awstypes "github.com/aws/aws-sdk-go-v2/service/continuum/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

const testRoleARN = "arn:aws:iam::123456789012:role/test" //lintignore:AWSAT005

func testProviderConfiguration(ctx context.Context, github *githubConfigurationModel, gitlab *gitlabConfigurationModel, s3 *s3ConfigurationModel) fwtypes.ListNestedObjectValueOf[providerConfigurationModel] {
	m := providerConfigurationModel{
		GitHub: fwtypes.NewListNestedObjectValueOfNull[githubConfigurationModel](ctx),
		GitLab: fwtypes.NewListNestedObjectValueOfNull[gitlabConfigurationModel](ctx),
		S3:     fwtypes.NewListNestedObjectValueOfNull[s3ConfigurationModel](ctx),
	}
	if github != nil {
		m.GitHub = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, github)
	}
	if gitlab != nil {
		m.GitLab = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, gitlab)
	}
	if s3 != nil {
		m.S3 = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, s3)
	}

	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &m)
}

func testS3Integration(ctx context.Context) integrationResourceModel {
	return integrationResourceModel{
		ID:   types.StringValue("abc123"),
		Name: types.StringValue("test"),
		ProviderConfiguration: testProviderConfiguration(ctx, nil, nil, &s3ConfigurationModel{
			Bucket:     types.StringValue("test-bucket"),
			ExternalID: types.StringValue("test-external-id"),
			RoleARN:    fwtypes.ARNValue(testRoleARN),
		}),
	}
}

func testGitHubIntegration(ctx context.Context) integrationResourceModel {
	return integrationResourceModel{
		Name: types.StringValue("test"),
		ProviderConfiguration: testProviderConfiguration(ctx, &githubConfigurationModel{
			OrganizationName: types.StringValue("test-org"),
			Authorization: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &githubAuthorizationConfigurationModel{
				AuthorizationCode: types.StringValue("test-code"),
				RedirectURL:       types.StringNull(),
				State:             types.StringValue("test-state"),
			}),
		}, nil, nil),
	}
}

func testGetIntegrationOutput(integrationType awstypes.IntegrationType, details awstypes.ProviderDetails) *continuum.GetIntegrationOutput {
	now := time.Now()
	return &continuum.GetIntegrationOutput{
		CreatedAt:       aws.Time(now),
		IntegrationArn:  aws.String("arn:aws:continuum:us-west-2:123456789012:integrations/abc123"), //lintignore:AWSAT003,AWSAT005
		IntegrationId:   aws.String("abc123"),
		IntegrationType: integrationType,
		Name:            aws.String("test"),
		ProviderDetails: details,
		UpdatedAt:       aws.Time(now),
	}
}

func providerConfigurationOf(ctx context.Context, t *testing.T, m integrationResourceModel) *providerConfigurationModel {
	t.Helper()

	v, diags := m.ProviderConfiguration.ToPtr(ctx)
	if diags.HasError() || v == nil {
		t.Fatalf("provider_configuration: %v", diags)
	}

	return v
}

func TestIntegrationExpand(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("create S3", func(t *testing.T) {
		t.Parallel()

		var input continuum.CreateIntegrationInput
		if diags := fwflex.Expand(ctx, testS3Integration(ctx), &input); diags.HasError() {
			t.Fatal(diags)
		}

		s3, ok := input.ProviderConfiguration.(*awstypes.ProviderConfigurationMemberS3)
		if !ok {
			t.Fatalf("provider configuration type: %T", input.ProviderConfiguration)
		}
		if got := aws.ToString(input.Name); got != "test" {
			t.Errorf("name: %q", got)
		}
		if got := aws.ToString(s3.Value.Bucket); got != "test-bucket" {
			t.Errorf("bucket: %q", got)
		}
		if got := aws.ToString(s3.Value.ExternalId); got != "test-external-id" {
			t.Errorf("external ID: %q", got)
		}
		if got := aws.ToString(s3.Value.RoleArn); got != testRoleARN {
			t.Errorf("role ARN: %q", got)
		}
	})

	t.Run("create GitHub", func(t *testing.T) {
		t.Parallel()

		var input continuum.CreateIntegrationInput
		if diags := fwflex.Expand(ctx, testGitHubIntegration(ctx), &input); diags.HasError() {
			t.Fatal(diags)
		}

		github, ok := input.ProviderConfiguration.(*awstypes.ProviderConfigurationMemberGithub)
		if !ok {
			t.Fatalf("provider configuration type: %T", input.ProviderConfiguration)
		}
		if got := aws.ToString(github.Value.OrganizationName); got != "test-org" {
			t.Errorf("organization name: %q", got)
		}
		authorization := github.Value.Authorization
		if authorization == nil {
			t.Fatal("authorization is nil")
		}
		if got := aws.ToString(authorization.AuthorizationCode); got != "test-code" {
			t.Errorf("authorization code: %q", got)
		}
		if got := aws.ToString(authorization.State); got != "test-state" {
			t.Errorf("state: %q", got)
		}
		if authorization.RedirectUrl != nil {
			t.Errorf("redirect URL: %q", aws.ToString(authorization.RedirectUrl))
		}
	})

	t.Run("update S3", func(t *testing.T) {
		t.Parallel()

		var input continuum.UpdateIntegrationInput
		if diags := fwflex.Expand(ctx, testS3Integration(ctx), &input, fwflex.WithFieldNamePrefix("Integration")); diags.HasError() {
			t.Fatal(diags)
		}

		s3, ok := input.ProviderConfiguration.(*awstypes.UpdateProviderConfigurationMemberS3)
		if !ok {
			t.Fatalf("provider configuration type: %T", input.ProviderConfiguration)
		}
		if got := aws.ToString(input.IntegrationId); got != "abc123" {
			t.Errorf("integration ID: %q", got)
		}
		if got := aws.ToString(s3.Value.ExternalId); got != "test-external-id" {
			t.Errorf("external ID: %q", got)
		}
		if got := aws.ToString(s3.Value.RoleArn); got != testRoleARN {
			t.Errorf("role ARN: %q", got)
		}
	})
}

func TestIntegrationFlatten(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s3Output := testGetIntegrationOutput(awstypes.IntegrationTypeS3, &awstypes.ProviderDetailsMemberS3{
		Value: awstypes.S3Details{
			Bucket:  aws.String("test-bucket"),
			RoleArn: aws.String(testRoleARN + "2"),
		},
	})
	githubOutput := testGetIntegrationOutput(awstypes.IntegrationTypeGithub, &awstypes.ProviderDetailsMemberGithub{})

	t.Run("S3 keeps external ID", func(t *testing.T) {
		t.Parallel()

		m := testS3Integration(ctx)
		if diags := m.flatten(ctx, s3Output); diags.HasError() {
			t.Fatal(diags)
		}

		if got := m.ARN.ValueString(); got == "" {
			t.Error("ARN is empty")
		}
		if got := m.IntegrationType.ValueEnum(); got != awstypes.IntegrationTypeS3 {
			t.Errorf("integration type: %q", got)
		}
		if m.CreatedAt.IsNull() || m.UpdatedAt.IsNull() {
			t.Error("timestamps are null")
		}

		pc := providerConfigurationOf(ctx, t, m)
		if !pc.GitHub.IsNull() || !pc.GitLab.IsNull() {
			t.Error("other providers are not null")
		}
		s3, _ := pc.S3.ToPtr(ctx)
		if got := s3.ExternalID.ValueString(); got != "test-external-id" {
			t.Errorf("external ID: %q", got)
		}
		if got := s3.RoleARN.ValueString(); got != testRoleARN+"2" {
			t.Errorf("role ARN: %q", got)
		}
	})

	t.Run("S3 import", func(t *testing.T) {
		t.Parallel()

		m := integrationResourceModel{
			ID:                    types.StringValue("abc123"),
			ProviderConfiguration: fwtypes.NewListNestedObjectValueOfNull[providerConfigurationModel](ctx),
		}
		if diags := m.flatten(ctx, s3Output); diags.HasError() {
			t.Fatal(diags)
		}

		s3, _ := providerConfigurationOf(ctx, t, m).S3.ToPtr(ctx)
		if got := s3.Bucket.ValueString(); got != "test-bucket" {
			t.Errorf("bucket: %q", got)
		}
		if !s3.ExternalID.IsNull() {
			t.Errorf("external ID: %q", s3.ExternalID.ValueString())
		}
	})

	t.Run("GitLab keeps access token", func(t *testing.T) {
		t.Parallel()

		m := integrationResourceModel{
			ProviderConfiguration: testProviderConfiguration(ctx, nil, &gitlabConfigurationModel{
				AccessToken: types.StringValue("test-token"),
			}, nil),
		}
		output := testGetIntegrationOutput(awstypes.IntegrationTypeGitlab, &awstypes.ProviderDetailsMemberGitlab{
			Value: awstypes.GitLabDetails{
				ExpiresAt:     aws.Time(time.Now()),
				InstanceUrl:   aws.String("https://gitlab.example.com"),
				TargetPath:    aws.String("group/project"),
				TargetType:    awstypes.GitLabTargetTypeProject,
				TokenRotation: awstypes.GitLabTokenRotationNone,
			},
		})
		if diags := m.flatten(ctx, output); diags.HasError() {
			t.Fatal(diags)
		}

		gitlab, _ := providerConfigurationOf(ctx, t, m).GitLab.ToPtr(ctx)
		if got := gitlab.AccessToken.ValueString(); got != "test-token" {
			t.Errorf("access token: %q", got)
		}
		if got := gitlab.InstanceURL.ValueString(); got != "https://gitlab.example.com" {
			t.Errorf("instance URL: %q", got)
		}
		if got := gitlab.TargetType.ValueEnum(); got != awstypes.GitLabTargetTypeProject {
			t.Errorf("target type: %q", got)
		}
		if got := gitlab.TokenRotation.ValueEnum(); got != awstypes.GitLabTokenRotationNone {
			t.Errorf("token rotation: %q", got)
		}
		if gitlab.ExpiresAt.IsNull() {
			t.Error("expires at is null")
		}
	})

	t.Run("GitHub keeps authorization", func(t *testing.T) {
		t.Parallel()

		m := testGitHubIntegration(ctx)
		if diags := m.flatten(ctx, githubOutput); diags.HasError() {
			t.Fatal(diags)
		}

		github, _ := providerConfigurationOf(ctx, t, m).GitHub.ToPtr(ctx)
		authorization, _ := github.Authorization.ToPtr(ctx)
		if authorization == nil {
			t.Fatal("authorization is nil")
		}
		if got := authorization.AuthorizationCode.ValueString(); got != "test-code" {
			t.Errorf("authorization code: %q", got)
		}
		if !github.OrganizationName.IsNull() {
			t.Errorf("organization name: %q", github.OrganizationName.ValueString())
		}
	})

	t.Run("GitHub import", func(t *testing.T) {
		t.Parallel()

		m := integrationResourceModel{
			ProviderConfiguration: fwtypes.NewListNestedObjectValueOfNull[providerConfigurationModel](ctx),
		}
		if diags := m.flatten(ctx, githubOutput); diags.HasError() {
			t.Fatal(diags)
		}

		github, _ := providerConfigurationOf(ctx, t, m).GitHub.ToPtr(ctx)
		if !github.Authorization.IsNull() {
			t.Error("authorization is not null")
		}
		if _, diags := m.ProviderConfiguration.ToListValue(ctx); diags.HasError() {
			t.Fatal(diags)
		}
	})
}
