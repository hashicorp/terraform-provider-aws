// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package pinpointsmsvoicev2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

const (
	keywordDeletePropagationTimeout = 30 * time.Second

	// AWS-managed mandatory keywords exist on every
	// phone and pool origination identity.
	keywordMandatoryHELP = "HELP"
	keywordMandatorySTOP = "STOP"
)

// @FrameworkResource("aws_pinpointsmsvoicev2_keyword", name="Keyword")
// @IdentityAttribute("origination_identity")
// @IdentityAttribute("keyword")
// @ImportIDHandler("keywordImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheckKeyword")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pinpointsmsvoicev2/types;awstypes.KeywordInformation")
// @Testing(importStateIdAttributes="origination_identity;keyword", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(generator="randomKeywordName(t)")
func newKeywordResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &keywordResource{}, nil
}

type keywordResource struct {
	framework.ResourceWithModel[keywordResourceModel]
	framework.WithImportByIdentity
}

var _ resource.ResourceWithValidateConfig = (*keywordResource)(nil)

func (r *keywordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"keyword": schema.StringAttribute{
				Description: "Keyword to configure. 1-30 characters, case-insensitive, and cannot start or end with a space.",
				CustomType:  fwtypes.CaseInsensitiveStringType,
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 30),
					stringvalidator.RegexMatches(regexache.MustCompile(`^\S([ \S]*\S)?$`), "must not start or end with a space, or contain tabs or newlines"),
				},
			},
			"keyword_action": schema.StringAttribute{
				Description: "Action to perform when the keyword is received.",
				CustomType:  fwtypes.StringEnumType[awstypes.KeywordAction](),
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"keyword_message": schema.StringAttribute{
				Description: "Message to send when the keyword is received.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1600),
					stringvalidator.RegexMatches(regexache.MustCompile(`\S`), "must not be entirely whitespace"),
				},
			},
			"origination_identity": schema.StringAttribute{
				Description: "Origination identity to attach the keyword to. Value is the ID or ARN of a phone number or pool.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[A-Za-z0-9_:/-]+$`), "must contain only alphanumeric characters, underscores, hyphens, colons, and slashes"),
				},
			},
			"origination_identity_arn": schema.StringAttribute{
				Description: "ARN of the origination identity the keyword is attached to.",
				CustomType:  fwtypes.ARNType,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *keywordResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config keywordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &config))
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Keyword.IsUnknown() {
		return
	}

	if isMandatoryKeyword(config.Keyword.ValueString()) && !config.KeywordAction.IsNull() && !config.KeywordAction.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("keyword_action"),
			"Invalid Attribute Combination",
			fmt.Sprintf("keyword_action is managed by AWS for the mandatory keyword %q and cannot be set; remove it from your configuration.", config.Keyword.ValueString()),
		)
	}
}

func (r *keywordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan keywordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input pinpointsmsvoicev2.PutKeywordInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutKeyword(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Keyword.ValueString())
		return
	}

	// PutKeyword response: keyword uppercased and the origination identity as an ID.
	// keep the user-configured values so the resource identity remains stable.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan,
		fwflex.WithIgnoredFieldNamesAppend("Keyword"),
		fwflex.WithIgnoredFieldNamesAppend("OriginationIdentity"),
	))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *keywordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state keywordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, originationIdentityARN, err := findKeywordByTwoPartKey(ctx, conn, state.OriginationIdentity.ValueString(), state.Keyword.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Keyword.ValueString())
		return
	}

	// Preserve the user-configured value for Keyword so the resource identity remains stable.
	// KeywordInformation does not included the origination identity ARN, so we need to get it from the DescribeKeywords response.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state, fwflex.WithIgnoredFieldNamesAppend("Keyword")))
	if resp.Diagnostics.HasError() {
		return
	}
	state.OriginationIdentityARN = fwtypes.ARNValue(aws.ToString(originationIdentityARN))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *keywordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var plan keywordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input pinpointsmsvoicev2.PutKeywordInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutKeyword(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Keyword.ValueString())
		return
	}

	// Preserve the user-configured values for Keyword and OriginationIdentity so the resource identity remains stable.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &plan,
		fwflex.WithIgnoredFieldNamesAppend("Keyword"),
		fwflex.WithIgnoredFieldNamesAppend("OriginationIdentity"),
	))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *keywordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PinpointSMSVoiceV2Client(ctx)

	var state keywordResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Mandatory keywords cannot be deleted independently of their origination identity
	if isMandatoryKeyword(state.Keyword.ValueString()) {
		return
	}

	input := pinpointsmsvoicev2.DeleteKeywordInput{
		Keyword:             state.Keyword.ValueStringPointer(),
		OriginationIdentity: state.OriginationIdentity.ValueStringPointer(),
	}

	_, err := conn.DeleteKeyword(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Keyword.ValueString())
		return
	}

	// Deletion propagation lag; poll until DescribeKeywords no longer sees it.
	_, err = tfresource.RetryUntilNotFound(ctx, keywordDeletePropagationTimeout, func(ctx context.Context) (any, error) {
		out, _, err := findKeywordByTwoPartKey(ctx, conn, state.OriginationIdentity.ValueString(), state.Keyword.ValueString())
		return out, err
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Keyword.ValueString())
	}
}

func findKeywordByTwoPartKey(ctx context.Context, conn *pinpointsmsvoicev2.Client, originationIdentity, keyword string) (*awstypes.KeywordInformation, *string, error) {
	input := pinpointsmsvoicev2.DescribeKeywordsInput{
		OriginationIdentity: aws.String(originationIdentity),
		Keywords:            []string{keyword},
	}

	keywords, originationIdentityARN, err := findKeywords(ctx, conn, &input)
	if err != nil {
		return nil, nil, err
	}

	for i := range keywords {
		if strings.EqualFold(aws.ToString(keywords[i].Keyword), keyword) {
			return &keywords[i], originationIdentityARN, nil
		}
	}

	return nil, nil, &retry.NotFoundError{
		Message: fmt.Sprintf("keyword %q not found on origination identity %q", keyword, originationIdentity),
	}
}

func findKeywords(ctx context.Context, conn *pinpointsmsvoicev2.Client, input *pinpointsmsvoicev2.DescribeKeywordsInput) ([]awstypes.KeywordInformation, *string, error) {
	var keywords []awstypes.KeywordInformation
	var originationIdentityARN *string

	pages := pinpointsmsvoicev2.NewDescribeKeywordsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		if err != nil {
			return nil, nil, smarterr.NewError(err)
		}

		originationIdentityARN = page.OriginationIdentityArn
		keywords = append(keywords, page.Keywords...)
	}

	return keywords, originationIdentityARN, nil
}

// isMandatoryKeyword reports whether keyword is an AWS-managed mandatory keyword
func isMandatoryKeyword(keyword string) bool {
	return strings.EqualFold(keyword, keywordMandatoryHELP) ||
		strings.EqualFold(keyword, keywordMandatorySTOP)
}

type keywordResourceModel struct {
	framework.WithRegionModel
	Keyword                fwtypes.CaseInsensitiveString              `tfsdk:"keyword"`
	KeywordAction          fwtypes.StringEnum[awstypes.KeywordAction] `tfsdk:"keyword_action"`
	KeywordMessage         types.String                               `tfsdk:"keyword_message"`
	OriginationIdentity    types.String                               `tfsdk:"origination_identity"`
	OriginationIdentityARN fwtypes.ARN                                `tfsdk:"origination_identity_arn"`
}

var _ inttypes.ImportIDParser = keywordImportID{}

type keywordImportID struct{}

func (keywordImportID) Parse(id string) (string, map[string]any, error) {
	originationIdentity, keyword, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found {
		return "", nil, fmt.Errorf("id %q should be in the format <origination-identity>%s<keyword>", id, intflex.ResourceIdSeparator)
	}

	return id, map[string]any{
		"origination_identity": originationIdentity,
		"keyword":              keyword,
	}, nil
}
