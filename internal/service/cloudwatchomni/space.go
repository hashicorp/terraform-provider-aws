// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatchomni

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchomni"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchomni/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
//
// @FrameworkResource("aws_cloudwatchomni_space", name="Space")
func newSpaceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &spaceResource{}

	return r, nil
}

type spaceResource struct {
	framework.ResourceWithModel[spaceResourceModel]
}

func (*spaceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cloudwatchomni_space"
}

// ImportState imports by space ID. Note this passes through to "space_id" rather
// than "id": the resource has no "id" attribute, and Read looks the space up by
// "space_id".
func (*spaceResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("space_id"), request, response)
}

func (r *spaceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required:    true,
				Description: "Name of the space (3-64 characters).",
			},
			"domain_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ID of the domain to create the space in.",
			},
			"data_access_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ARN of the IAM role for data access.",
			},
			"agent_core_evaluation_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ARN of the IAM role for AgentCore online evaluation.",
			},
			"space_id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique ID of the space.",
			},
			"space_arn": schema.StringAttribute{
				Computed:    true,
				Description: "ARN of the space.",
			},
			"domain_arn": schema.StringAttribute{
				Computed:    true,
				Description: "ARN of the domain.",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: "Status of the space.",
			},
		},
		Blocks: map[string]schema.Block{
			"encryption_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[encryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"encryption_strategy": schema.StringAttribute{
							Required:    true,
							Description: "Encryption strategy (AWS_OWNED, CUSTOMER_MANAGED).",
						},
						"kms_key_arn": schema.StringAttribute{
							CustomType:  fwtypes.ARNType,
							Optional:    true,
							Description: "ARN of the KMS key for CUSTOMER_MANAGED encryption.",
						},
					},
				},
				Description: "Encryption configuration for the space's data at rest.",
			},
		},
		Description: "Manages a CloudWatch Omni Space.",
	}
}

type spaceResourceModel struct {
	framework.WithRegionModel
	Name                       types.String                                                  `tfsdk:"name"`
	DomainID                   types.String                                                  `tfsdk:"domain_id"`
	DataAccessRoleARN          fwtypes.ARN                                                   `tfsdk:"data_access_role_arn"`
	AgentCoreEvaluationRoleARN fwtypes.ARN                                                   `tfsdk:"agent_core_evaluation_role_arn"`
	EncryptionConfiguration    fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration"`
	SpaceID                    types.String                                                  `tfsdk:"space_id"`
	SpaceARN                   types.String                                                  `tfsdk:"space_arn"`
	DomainARN                  types.String                                                  `tfsdk:"domain_arn"`
	Status                     types.String                                                  `tfsdk:"status"`
}

type encryptionConfigurationModel struct {
	EncryptionStrategy types.String `tfsdk:"encryption_strategy"`
	KMSKeyARN          fwtypes.ARN  `tfsdk:"kms_key_arn"`
}

func (r *spaceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data spaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	var input cloudwatchomni.CreateSpaceInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateSpace(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output.Space, &data))
	if response.Diagnostics.HasError() {
		return
	}

	normalizeEncryptionConfiguration(ctx, output.Space, &data)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

// normalizeEncryptionConfiguration drops the encryption_configuration block when
// the API reports the implicit service default.
//
// CreateSpace/GetSpace always return an encryptionConfiguration, defaulting to
// AWS_OWNED with no KMS key even when the caller omitted the block entirely.
// Flattening that verbatim turns an omitted block into a one-element list,
// which Terraform rejects as "Provider produced inconsistent result after
// apply" and would otherwise cause a perpetual diff. Omitting the block and
// explicitly specifying AWS_OWNED are equivalent to the service, so the
// absent form is canonical.
func normalizeEncryptionConfiguration(ctx context.Context, space *awstypes.Space, data *spaceResourceModel) {
	if space == nil || space.EncryptionConfiguration == nil {
		return
	}
	if space.EncryptionConfiguration.EncryptionStrategy == awstypes.EncryptionStrategyAwsOwned &&
		aws.ToString(space.EncryptionConfiguration.KmsKeyArn) == "" {
		data.EncryptionConfiguration = fwtypes.NewListNestedObjectValueOfEmpty[encryptionConfigurationModel](ctx)
	}
}

func (r *spaceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data spaceResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().CloudWatchOmniClient(ctx)

	output, err := findSpaceByID(ctx, conn, data.SpaceID.ValueString(), r.Meta().Region(ctx))
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.SpaceID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}

	normalizeEncryptionConfiguration(ctx, output, &data)

	// GetSpace returns domainArn but not domainId, so recover the ID from the ARN.
	// Without this, domain_id would be null after import (the configured value is
	// only otherwise preserved because it is already present in prior state).
	if domainID, ok := domainIDFromARN(aws.ToString(output.DomainArn)); ok {
		data.DomainID = fwflex.StringValueToFramework(ctx, domainID)
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
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

	if !new.Name.Equal(old.Name) || !new.EncryptionConfiguration.Equal(old.EncryptionConfiguration) {
		var input cloudwatchomni.UpdateSpaceInput
		input.SpaceId = fwflex.StringFromFramework(ctx, new.SpaceID)

		if !new.Name.Equal(old.Name) {
			input.Name = fwflex.StringFromFramework(ctx, new.Name)
		}

		if !new.EncryptionConfiguration.Equal(old.EncryptionConfiguration) {
			var encryptionConfig awstypes.EncryptionConfiguration
			smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new.EncryptionConfiguration, &encryptionConfig))
			if response.Diagnostics.HasError() {
				return
			}
			input.EncryptionConfiguration = &encryptionConfig
		}

		output, err := conn.UpdateSpace(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.SpaceID.ValueString())
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output.Space, &new))
		if response.Diagnostics.HasError() {
			return
		}

		normalizeEncryptionConfiguration(ctx, output.Space, &new)
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
	region := r.Meta().Region(ctx)
	id := data.SpaceID.ValueString()

	input := cloudwatchomni.DeleteSpaceInput{
		SpaceId: fwflex.StringFromFramework(ctx, data.SpaceID),
	}

	if _, err := conn.DeleteSpace(ctx, &input); err != nil {
		// DeleteSpace cannot report "already gone" distinguishably (see
		// findSpaceByID), so confirm absence via ListSpaces before surfacing.
		if _, ferr := findSpaceByID(ctx, conn, id, region); retry.NotFound(ferr) {
			return
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitSpaceDeleted(ctx, conn, id, region, spaceDeletedTimeout); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)
		return
	}
}

const spaceDeletedTimeout = 5 * time.Minute

// waitSpaceDeleted blocks until the space is absent from ListSpaces.
//
// There is no observable DELETING state: a space disappears from ListSpaces
// once deleted. Deletion has been observed to be effectively synchronous, so
// this normally returns on the first poll; it exists to absorb eventual
// consistency.
//
//nolint:unparam // Mirrors the (*output, error) waiter convention; output is always nil.
func waitSpaceDeleted(ctx context.Context, conn *cloudwatchomni.Client, id, region string, timeout time.Duration) (*awstypes.Space, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"EXISTS"},
		Target:  []string{},
		Refresh: func(ctx context.Context) (any, string, error) {
			out, err := findSpaceByID(ctx, conn, id, region)
			if retry.NotFound(err) {
				return nil, "", nil
			}
			if err != nil {
				return nil, "", err
			}
			return out, "EXISTS", nil
		},
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return nil, err
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

// findSpaceByID resolves a space scoped to the given region.
//
// ListSpaces is account-wide and returns spaces from every region, and unlike
// GetSpace it reliably reports absence, so it is used as the existence oracle.
// Only once the space is known to exist in this region is GetSpace called for
// the full detail, which means a 403 from GetSpace surfaces as a real error
// rather than being mistaken for deletion.
func findSpaceByID(ctx context.Context, conn *cloudwatchomni.Client, id, region string) (*awstypes.Space, error) {
	var found bool

	pages := cloudwatchomni.NewListSpacesPaginator(conn, &cloudwatchomni.ListSpacesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, item := range page.Items {
			if aws.ToString(item.SpaceId) == id && aws.ToString(item.Region) == region {
				found = true
				break
			}
		}

		if found {
			break
		}
	}

	if !found {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf("CloudWatch Omni Space (%s) not found in region %s", id, region),
		}
	}

	return findSpace(ctx, conn, &cloudwatchomni.GetSpaceInput{SpaceId: aws.String(id)})
}
