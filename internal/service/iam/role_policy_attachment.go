// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_iam_role_policy_attachment", name="Role Policy Attachment")
// @IdentityAttribute("role")
// @IdentityAttribute("policy_arn")
// @IdAttrFormat("{role}/{policy_arn}")
// @ImportIDHandler("rolePolicyAttachmentImportID")
// @Testing(preIdentityVersion="6.0.0")
func resourceRolePolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRolePolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceRolePolicyAttachmentRead,
		DeleteWithoutTimeout: resourceRolePolicyAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"policy_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRole: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

// @SDKListResource("aws_iam_role_policy_attachment")
func rolePolicyAttachmentResourceAsListResource() inttypes.ListResourceForSDK {
	l := rolePolicyAttachmentListResource{}
	l.SetResourceSchema(resourceRolePolicyAttachment())

	return &l
}

func resourceRolePolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	role := d.Get(names.AttrRole).(string)
	policyARN := d.Get("policy_arn").(string)

	if err := attachPolicyToRole(ctx, conn, role, policyARN); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(createRolePolicyAttachmentImportID(d))

	return append(diags, resourceRolePolicyAttachmentRead(ctx, d, meta)...)
}

func resourceRolePolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	role := d.Get(names.AttrRole).(string)
	policyARN := d.Get("policy_arn").(string)
	// Human friendly ID for error messages since d.Id() is non-descriptive.
	id := fmt.Sprintf("%s:%s", role, policyARN)

	_, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return findAttachedRolePolicyByTwoPartKey(ctx, conn, role, policyARN)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Role Policy Attachment (%s) not found, removing from state", id)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Role Policy Attachment (%s): %s", id, err)
	}

	return diags
}

func resourceRolePolicyAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	if err := detachPolicyFromRole(ctx, conn, d.Get(names.AttrRole).(string), d.Get("policy_arn").(string)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func attachPolicyToRole(ctx context.Context, conn *iam.Client, role, policyARN string) error {
	var errConcurrentModificationException *awstypes.ConcurrentModificationException
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(role),
		})
	}, errConcurrentModificationException.ErrorCode())

	if err != nil {
		return fmt.Errorf("attaching IAM Policy (%s) to IAM Role (%s): %w", policyARN, role, err)
	}

	return nil
}

func detachPolicyFromRole(ctx context.Context, conn *iam.Client, role, policyARN string) error {
	var errConcurrentModificationException *awstypes.ConcurrentModificationException
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
			PolicyArn: aws.String(policyARN),
			RoleName:  aws.String(role),
		})
	}, errConcurrentModificationException.ErrorCode())

	if errs.IsA[*awstypes.NoSuchEntityException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching IAM Policy (%s) from IAM Role (%s): %w", policyARN, role, err)
	}

	return nil
}

func findAttachedRolePolicyByTwoPartKey(ctx context.Context, conn *iam.Client, roleName, policyARN string) (*awstypes.AttachedPolicy, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}

	return findAttachedRolePolicy(ctx, conn, input, func(v awstypes.AttachedPolicy) bool {
		return aws.ToString(v.PolicyArn) == policyARN
	})
}

func findAttachedRolePolicy(ctx context.Context, conn *iam.Client, input *iam.ListAttachedRolePoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) (*awstypes.AttachedPolicy, error) {
	output, err := findAttachedRolePolicies(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAttachedRolePolicies(ctx context.Context, conn *iam.Client, input *iam.ListAttachedRolePoliciesInput, filter tfslices.Predicate[awstypes.AttachedPolicy]) ([]awstypes.AttachedPolicy, error) {
	var output []awstypes.AttachedPolicy

	pages := iam.NewListAttachedRolePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.NoSuchEntityException](err) {
			return nil, &sdkretry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AttachedPolicies {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func createRolePolicyAttachmentImportID(d *schema.ResourceData) string {
	return (rolePolicyAttachmentImportID{}).Create(d)
}

type rolePolicyAttachmentImportID struct{}

func (v rolePolicyAttachmentImportID) Create(d *schema.ResourceData) string {
	return v.create(d.Get(names.AttrRole).(string), d.Get("policy_arn").(string))
}

func (rolePolicyAttachmentImportID) create(roleName, policyARN string) string {
	return fmt.Sprintf("%s/%s", roleName, policyARN)
}

func (rolePolicyAttachmentImportID) Parse(id string) (string, map[string]string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, fmt.Errorf("unexpected format for Import ID (%q), expected <role-name>/<policy_arn>", id)
	}

	result := map[string]string{
		names.AttrRole: parts[0],
		"policy_arn":   parts[1],
	}
	return id, result, nil
}

var _ list.ListResourceWithRawV5Schemas = &rolePolicyAttachmentListResource{}

type rolePolicyAttachmentListResource struct {
	framework.ResourceWithConfigure
	framework.ListResourceWithSDKv2Resource
}

type rolePolicyAttachmentListResourceModel struct {
}

func (l *rolePolicyAttachmentListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{},
	}
}

func (l *rolePolicyAttachmentListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.IAMClient(ctx)

	var query rolePolicyAttachmentListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input iam.ListRolesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for role, err := range listNonServiceLinkedRoles(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			tflog.Info(ctx, "Listing attached policies for role", map[string]any{
				logging.ResourceAttributeKey(names.AttrRole): aws.ToString(role.RoleName),
			})

			input := iam.ListAttachedRolePoliciesInput{
				RoleName: role.RoleName,
			}
			pages := iam.NewListAttachedRolePoliciesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if errs.IsA[*awstypes.NoSuchEntityException](err) {
					tflog.Warn(ctx, "Resource disappeared during listing, skipping", map[string]any{
						logging.ResourceAttributeKey(names.AttrRole): aws.ToString(role.RoleName),
					})
					continue
				}
				if err != nil {
					result := fwdiag.NewListResultErrorDiagnostic(err)
					yield(result)
					return
				}

				for _, attachedPolicy := range page.AttachedPolicies {
					ctx := resourceRolePolicyAttachmentListItemLoggingContext(ctx, role, attachedPolicy)

					result := request.NewListResult(ctx)

					rd := l.ResourceData()
					rd.SetId(resourceRolePolicyAttachmentImportIDFromData(role, attachedPolicy))

					tflog.Info(ctx, "Reading resource")
					rd.Set(names.AttrRole, aws.ToString(role.RoleName))
					rd.Set("policy_arn", aws.ToString(attachedPolicy.PolicyArn))

					result.DisplayName = resourceRolePolicyAttachmentDisplayName(role, attachedPolicy)

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
}

func resourceRolePolicyAttachmentListItemLoggingContext(ctx context.Context, role awstypes.Role, attachedPolicy awstypes.AttachedPolicy) context.Context {
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrRole), aws.ToString(role.RoleName))
	ctx = tflog.SetField(ctx, logging.ResourceAttributeKey("policy_arn"), aws.ToString(attachedPolicy.PolicyArn))
	return ctx
}

func resourceRolePolicyAttachmentImportIDFromData(role awstypes.Role, attachedPolicy awstypes.AttachedPolicy) string {
	return (rolePolicyAttachmentImportID{}).create(aws.ToString(role.RoleName), aws.ToString(attachedPolicy.PolicyArn))
}

func resourceRolePolicyAttachmentDisplayName(role awstypes.Role, attachedPolicy awstypes.AttachedPolicy) string {
	foo := aws.ToString(attachedPolicy.PolicyArn)
	policyARN, err := arn.Parse(foo)
	if err == nil {
		foo = strings.TrimPrefix(policyARN.Resource, "policy/")
	}

	return fmt.Sprintf("Role: %s - Policy: %s", resourceRoleDisplayName(role), foo)
}
