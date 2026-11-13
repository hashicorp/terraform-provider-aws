// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Tags are intentionally unsupported. CreateSpace accepts a Tags member, but the
// Space response type does not expose it and CloudWatch Omni provides no
// TagResource, UntagResource or ListTagsForResource operations, so tags can be
// neither refreshed nor updated. Modelling them would yield an attribute that
// silently fails to detect drift, cannot be changed in place, cannot survive
// import, and cannot honour provider default_tags. Revisit once the service adds
// tag read APIs.

// @FrameworkResource("aws_cloudwatchomni_space", name="Space")
// @IdentityAttribute("space_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types;awstypes;awstypes.Space")
// @Testing(preCheck="testAccPreCheck")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(importStateIdAttribute="space_id")
func newSpaceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &spaceResource{}

	return r, nil
}

type spaceResource struct {
	framework.ResourceWithModel[spaceResourceModel]
	framework.WithImportByIdentity
}

func (r *spaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_core_evaluation_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_access_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`), "use lowercase letters, numbers, and hyphens. Start and end with a letter or number."),
				},
			},
			"space_arn": framework.ARNAttributeComputedOnly(),
			"space_id":  framework.IDAttribute(),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.SpaceStatus](),
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrEncryptionConfiguration: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"encryption_strategy": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.EncryptionStrategy](),
							Required:   true,
							Validators: []validator.String{
								tfstringvalidator.AlsoRequiresWhenEquals(
									awstypes.EncryptionStrategyCustomerManaged,
									path.MatchRelative().AtParent().AtName(names.AttrKMSKeyARN),
								),
								tfstringvalidator.ConflictsWithWhenEquals(
									awstypes.EncryptionStrategyAwsOwned,
									path.MatchRelative().AtParent().AtName(names.AttrKMSKeyARN),
								),
							},
						},
						names.AttrKMSKeyARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
		},
	}
}

func (r *spaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data spaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input cloudwatchomni.CreateSpaceInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))

	output, err := conn.CreateSpace(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.Name, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.Space, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *spaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data spaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID := fwflex.StringValueFromFramework(ctx, data.SpaceID)
	output, err := findSpaceByID(ctx, conn, spaceID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, spaceID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *spaceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new spaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		spaceID := fwflex.StringValueFromFramework(ctx, new.SpaceID)
		input := cloudwatchomni.UpdateSpaceInput{
			SpaceId: aws.String(spaceID),
		}

		if !new.Name.Equal(old.Name) {
			input.Name = fwflex.StringFromFramework(ctx, new.Name)
		}

		if !new.EncryptionConfiguration.Equal(old.EncryptionConfiguration) {
			var encryptionConfig awstypes.EncryptionConfiguration
			if new.EncryptionConfiguration.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0 {
				encryptionConfig.EncryptionStrategy = awstypes.EncryptionStrategyAwsOwned
			} else {
				smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new.EncryptionConfiguration, &encryptionConfig))
				if response.Diagnostics.HasError() {
					return
				}
			}
			input.EncryptionConfiguration = &encryptionConfig
		}

		output, err := conn.UpdateSpace(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, spaceID)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, r.flatten(ctx, output.Space, &new))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *spaceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data spaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	spaceID := fwflex.StringValueFromFramework(ctx, data.SpaceID)
	input := cloudwatchomni.DeleteSpaceInput{
		SpaceId: aws.String(spaceID),
	}
	if _, err := conn.DeleteSpace(ctx, &input); err != nil {
		// DeleteSpace cannot report "already gone" distinguishably (see
		// findSpaceByID), so confirm absence before surfacing the error.
		_, ferr := findSpaceByID(ctx, conn, spaceID)

		if retry.NotFound(ferr) {
			return
		}

		if ferr != nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.Join(err, ferr), smerr.ID, spaceID)
			return
		}

		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, spaceID)
		return
	}

	const timeout = 5 * time.Minute
	_, err := tfresource.RetryUntilNotFound(ctx, timeout, func(ctx context.Context) (any, error) {
		return findSpaceByID(ctx, conn, spaceID)
	})

	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, spaceID)
		return
	}
}

func (r *spaceResource) flatten(ctx context.Context, space *awstypes.Space, data *spaceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	// Drop the encryption_configuration block when the API reports the implicit service default.
	if data.EncryptionConfiguration.IsNull() && space.EncryptionConfiguration != nil && space.EncryptionConfiguration.EncryptionStrategy == awstypes.EncryptionStrategyAwsOwned {
		space.EncryptionConfiguration = nil
	}
	diags.Append(fwflex.Flatten(ctx, space, data, fwflex.WithIgnoredFieldNamesAppend("Region"))...)
	// Recover the domain ID from the ARN.
	if domainID, ok := domainIDFromARN(aws.ToString(space.DomainArn)); ok {
		data.DomainID = fwflex.StringValueToFramework(ctx, domainID)
	}
	return diags
}

// domainIDFromARN extracts the domain ID from a domain ARN of the form
// arn:<partition>:cloudwatch:<region>:<account>:organization-domain/<domain-id>.
//
// Space.DomainArn is documented as absent when the space has no associated
// domain, so absence is reported rather than treated as an error.
func domainIDFromARN(domainARN string) (string, bool) {
	if domainARN == "" {
		return "", false
	}

	parsed, err := arn.Parse(domainARN)
	if err != nil {
		return "", false
	}

	_, id, found := strings.Cut(parsed.Resource, "/")
	if !found || id == "" {
		return "", false
	}

	return id, true
}

// findSpaceByID resolves a space scoped to the given region.
//
// ListSpaces is account-wide and returns spaces from every region, and unlike
// GetSpace it reliably reports absence, so it is used as the existence oracle.
// Only once the space is known to exist in this region is GetSpace called for
// the full detail, which means a 403 from GetSpace surfaces as a real error
// rather than being mistaken for deletion.
func findSpaceByID(ctx context.Context, conn *cloudwatchomni.Client, id string) (*awstypes.Space, error) {
	var inputLS cloudwatchomni.ListSpacesInput
	spaces, err := findSpaces(ctx, conn, &inputLS, tfslices.WithFilter(func(v awstypes.SpaceSummary) bool {
		return aws.ToString(v.SpaceId) == id && aws.ToString(v.Region) == conn.Options().Region
	}))
	if err != nil {
		return nil, err
	}

	space, err := tfresource.AssertSingleValueResult(spaces)
	if err != nil {
		return nil, err
	}

	inputGS := cloudwatchomni.GetSpaceInput{
		SpaceId: space.SpaceId,
	}
	return findSpace(ctx, conn, &inputGS)
}

// findSpace performs the raw GetSpace call.
//
// It deliberately does NOT translate errors into NotFoundError. GetSpace
// returns AccessDeniedException (403) for a deleted space, a never-existing
// space, a space in another region, AND a genuine permissions failure — these
// are indistinguishable, so treating 403 as "not found" here would silently
// drop resources from state on a real IAM problem. Existence is established by
// findSpaceByID via ListSpaces instead.
func findSpace(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.GetSpaceInput) (*awstypes.Space, error) {
	output, err := conn.GetSpace(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.Space == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Space, nil
}

func findSpaces(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListSpacesInput, optFns ...tfslices.FinderOptionsFunc[awstypes.SpaceSummary]) ([]awstypes.SpaceSummary, error) {
	return tfslices.CollectAndConcatWithError(listSpacePages(ctx, conn, input), optFns...)
}

func listSpacePages(ctx context.Context, conn *cloudwatchomni.Client, input *cloudwatchomni.ListSpacesInput, optFns ...func(*cloudwatchomni.Options)) iter.Seq2[[]awstypes.SpaceSummary, error] {
	return func(yield func([]awstypes.SpaceSummary, error) bool) {
		pages := cloudwatchomni.NewListSpacesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx, optFns...)
			if err != nil {
				yield(nil, fmt.Errorf("listing CloudWatch Omni Spaces: %w", err))
				return
			}

			if !yield(page.Items, nil) {
				return
			}
		}
	}
}

type spaceResourceModel struct {
	framework.WithRegionModel
	AgentCoreEvaluationRoleARN fwtypes.ARN                                                   `tfsdk:"agent_core_evaluation_role_arn"`
	DataAccessRoleARN          fwtypes.ARN                                                   `tfsdk:"data_access_role_arn"`
	DomainARN                  types.String                                                  `tfsdk:"domain_arn"`
	DomainID                   types.String                                                  `tfsdk:"domain_id"`
	EncryptionConfiguration    fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	Name                       types.String                                                  `tfsdk:"name"`
	SpaceARN                   types.String                                                  `tfsdk:"space_arn"`
	SpaceID                    types.String                                                  `tfsdk:"space_id"`
	Status                     fwtypes.StringEnum[awstypes.SpaceStatus]                      `tfsdk:"status"`
}

type encryptionConfigurationModel struct {
	EncryptionStrategy fwtypes.StringEnum[awstypes.EncryptionStrategy] `tfsdk:"encryption_strategy"`
	KMSKeyARN          fwtypes.ARN                                     `tfsdk:"kms_key_arn"`
}
