// Code generated by internal/generate/servicepackage/main.go; DO NOT EDIT.

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type servicePackage struct{}

func (p *servicePackage) FrameworkDataSources(ctx context.Context) []*itypes.ServicePackageFrameworkDataSource {
	return []*itypes.ServicePackageFrameworkDataSource{}
}

func (p *servicePackage) FrameworkResources(ctx context.Context) []*itypes.ServicePackageFrameworkResource {
	return []*itypes.ServicePackageFrameworkResource{
		{
			Factory:  newResourceGroupPoliciesExclusive,
			TypeName: "aws_iam_group_policies_exclusive",
			Name:     "Group Policies Exclusive",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newResourceGroupPolicyAttachmentsExclusive,
			TypeName: "aws_iam_group_policy_attachments_exclusive",
			Name:     "Group Policy Attachments Exclusive",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newOrganizationsFeaturesResource,
			TypeName: "aws_iam_organizations_features",
			Name:     "Organizations Features",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newResourceRolePoliciesExclusive,
			TypeName: "aws_iam_role_policies_exclusive",
			Name:     "Role Policies Exclusive",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newResourceRolePolicyAttachmentsExclusive,
			TypeName: "aws_iam_role_policy_attachments_exclusive",
			Name:     "Role Policy Attachments Exclusive",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newResourceUserPoliciesExclusive,
			TypeName: "aws_iam_user_policies_exclusive",
			Name:     "User Policies Exclusive",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  newResourceUserPolicyAttachmentsExclusive,
			TypeName: "aws_iam_user_policy_attachments_exclusive",
			Name:     "User Policy Attachments Exclusive",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) SDKDataSources(ctx context.Context) []*itypes.ServicePackageSDKDataSource {
	return []*itypes.ServicePackageSDKDataSource{
		{
			Factory:  dataSourceAccessKeys,
			TypeName: "aws_iam_access_keys",
			Name:     "Access Keys",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceAccountAlias,
			TypeName: "aws_iam_account_alias",
			Name:     "Account Alias",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceGroup,
			TypeName: "aws_iam_group",
			Name:     "Group",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceInstanceProfile,
			TypeName: "aws_iam_instance_profile",
			Name:     "Instance Profile",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceInstanceProfiles,
			TypeName: "aws_iam_instance_profiles",
			Name:     "Instance Profiles",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceOpenIDConnectProvider,
			TypeName: "aws_iam_openid_connect_provider",
			Name:     "OIDC Provider",
			Tags:     &itypes.ServicePackageResourceTags{},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourcePolicy,
			TypeName: "aws_iam_policy",
			Name:     "Policy",
			Tags:     &itypes.ServicePackageResourceTags{},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourcePolicyDocument,
			TypeName: "aws_iam_policy_document",
			Name:     "Policy Document",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourcePrincipalPolicySimulation,
			TypeName: "aws_iam_principal_policy_simulation",
			Name:     "Principal Policy Simulation",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceRole,
			TypeName: "aws_iam_role",
			Name:     "Role",
			Tags:     &itypes.ServicePackageResourceTags{},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceRoles,
			TypeName: "aws_iam_roles",
			Name:     "Roles",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceSAMLProvider,
			TypeName: "aws_iam_saml_provider",
			Name:     "SAML Provider",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceServerCertificate,
			TypeName: "aws_iam_server_certificate",
			Name:     "Server Certificate",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceSessionContext,
			TypeName: "aws_iam_session_context",
			Name:     "Session Context",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceUser,
			TypeName: "aws_iam_user",
			Name:     "User",
			Tags:     &itypes.ServicePackageResourceTags{},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceUserSSHKey,
			TypeName: "aws_iam_user_ssh_key",
			Name:     "User SSH Key",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  dataSourceUsers,
			TypeName: "aws_iam_users",
			Name:     "Users",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) SDKResources(ctx context.Context) []*itypes.ServicePackageSDKResource {
	return []*itypes.ServicePackageSDKResource{
		{
			Factory:  resourceAccessKey,
			TypeName: "aws_iam_access_key",
			Name:     "Access Key",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceAccountAlias,
			TypeName: "aws_iam_account_alias",
			Name:     "Account Alias",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceAccountPasswordPolicy,
			TypeName: "aws_iam_account_password_policy",
			Name:     "Account Password Policy",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceGroup,
			TypeName: "aws_iam_group",
			Name:     "Group",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceGroupMembership,
			TypeName: "aws_iam_group_membership",
			Name:     "Group Membership",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceGroupPolicy,
			TypeName: "aws_iam_group_policy",
			Name:     "Group Policy",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceGroupPolicyAttachment,
			TypeName: "aws_iam_group_policy_attachment",
			Name:     "Group Policy Attachment",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceInstanceProfile,
			TypeName: "aws_iam_instance_profile",
			Name:     "Instance Profile",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "InstanceProfile",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceOpenIDConnectProvider,
			TypeName: "aws_iam_openid_connect_provider",
			Name:     "OIDC Provider",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
				ResourceType:        "OIDCProvider",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourcePolicy,
			TypeName: "aws_iam_policy",
			Name:     "Policy",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrARN,
				ResourceType:        "Policy",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourcePolicyAttachment,
			TypeName: "aws_iam_policy_attachment",
			Name:     "Policy Attachment",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceRole,
			TypeName: "aws_iam_role",
			Name:     "Role",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrName,
				ResourceType:        "Role",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceRolePolicy,
			TypeName: "aws_iam_role_policy",
			Name:     "Role Policy",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceRolePolicyAttachment,
			TypeName: "aws_iam_role_policy_attachment",
			Name:     "Role Policy Attachment",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceSAMLProvider,
			TypeName: "aws_iam_saml_provider",
			Name:     "SAML Provider",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "SAMLProvider",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceSecurityTokenServicePreferences,
			TypeName: "aws_iam_security_token_service_preferences",
			Name:     "Security Token Service Preferences",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceServerCertificate,
			TypeName: "aws_iam_server_certificate",
			Name:     "Server Certificate",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrName,
				ResourceType:        "ServerCertificate",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceServiceLinkedRole,
			TypeName: "aws_iam_service_linked_role",
			Name:     "Service Linked Role",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "ServiceLinkedRole",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceServiceSpecificCredential,
			TypeName: "aws_iam_service_specific_credential",
			Name:     "Service Specific Credential",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceSigningCertificate,
			TypeName: "aws_iam_signing_certificate",
			Name:     "Signing Certificate",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceUser,
			TypeName: "aws_iam_user",
			Name:     "User",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrName,
				ResourceType:        "User",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceUserGroupMembership,
			TypeName: "aws_iam_user_group_membership",
			Name:     "User Group Membership",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceUserLoginProfile,
			TypeName: "aws_iam_user_login_profile",
			Name:     "User Login Profile",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceUserPolicy,
			TypeName: "aws_iam_user_policy",
			Name:     "User Policy",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceUserPolicyAttachment,
			TypeName: "aws_iam_user_policy_attachment",
			Name:     "User Policy Attachment",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceUserSSHKey,
			TypeName: "aws_iam_user_ssh_key",
			Name:     "User SSH Key",
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
		{
			Factory:  resourceVirtualMFADevice,
			TypeName: "aws_iam_virtual_mfa_device",
			Name:     "Virtual MFA Device",
			Tags: &itypes.ServicePackageResourceTags{
				IdentifierAttribute: names.AttrID,
				ResourceType:        "VirtualMFADevice",
			},
			Region: &itypes.ServicePackageResourceRegion{
				IsGlobal:          true,
				IsOverrideEnabled: true,
			},
		},
	}
}

func (p *servicePackage) ServicePackageName() string {
	return names.IAM
}

// NewClient returns a new AWS SDK for Go v2 client for this service package's AWS API.
func (p *servicePackage) NewClient(ctx context.Context, config map[string]any) (*iam.Client, error) {
	cfg := *(config["aws_sdkv2_config"].(*aws.Config))
	optFns := []func(*iam.Options){
		iam.WithEndpointResolverV2(newEndpointResolverV2()),
		withBaseEndpoint(config[names.AttrEndpoint].(string)),
		func(o *iam.Options) {
			if region := config["region"].(string); o.Region != region {
				tflog.Info(ctx, "overriding provider-configured AWS API region", map[string]any{
					"service":         "iam",
					"original_region": o.Region,
					"override_region": region,
				})
				o.Region = region
			}
		},
		withExtraOptions(ctx, p, config),
	}

	return iam.NewFromConfig(cfg, optFns...), nil
}

// withExtraOptions returns a functional option that allows this service package to specify extra API client options.
// This option is always called after any generated options.
func withExtraOptions(ctx context.Context, sp conns.ServicePackage, config map[string]any) func(*iam.Options) {
	if v, ok := sp.(interface {
		withExtraOptions(context.Context, map[string]any) []func(*iam.Options)
	}); ok {
		optFns := v.withExtraOptions(ctx, config)

		return func(o *iam.Options) {
			for _, optFn := range optFns {
				optFn(o)
			}
		}
	}

	return func(*iam.Options) {}
}

func ServicePackage(ctx context.Context) conns.ServicePackage {
	return &servicePackage{}
}
