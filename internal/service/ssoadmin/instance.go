// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ssoadmin_instance", name="Instance")
// @ArnIdentity
// @ArnFormat(global=true)
// @Tags
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/ssoadmin;ssoadmin.DescribeInstanceOutput")
// @Testing(serialize=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(identityTest=false)
func newInstanceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &instanceResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)
	return r, nil
}

const ResNameInstance = "Instance"

type instanceResource struct {
	framework.ResourceWithModel[instanceResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *instanceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"client_token": schema.StringAttribute{
				Optional:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 64)},
			},
			names.AttrCreatedDate:             schema.StringAttribute{CustomType: timetypes.RFC3339Type{}, Computed: true},
			names.AttrEncryptionConfiguration: framework.ResourceOptionalComputedSingleNestedObjectAttribute[encryptionConfigurationModel](ctx),
			"identity_store_id":               schema.StringAttribute{Computed: true},
			names.AttrName: schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 32), stringvalidator.RegexMatches(regexache.MustCompile(`^[\w+=,.@-]+$`), "must contain only alphanumeric characters and +=,.@-_")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			names.AttrOwnerAccountID: schema.StringAttribute{Computed: true},
			names.AttrStatus:         schema.StringAttribute{CustomType: fwtypes.StringEnumType[awstypes.InstanceStatus](), Computed: true},
			names.AttrStatusReason:   schema.StringAttribute{Computed: true},
			names.AttrTags:           tftags.TagsAttribute(),
			names.AttrTagsAll:        tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true, Delete: true}),
		},
	}
}

func (r *instanceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data instanceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() || data.EncryptionConfiguration.IsNull() || data.EncryptionConfiguration.IsUnknown() {
		return
	}
	config, diags := data.EncryptionConfiguration.ToPtr(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() || config == nil || config.KeyType.IsUnknown() {
		return
	}
	hasKeyARN := !config.KMSKeyARN.IsNull() && !config.KMSKeyARN.IsUnknown()
	switch config.KeyType.ValueEnum() {
	case awstypes.KmsKeyTypeCustomerManagedKey:
		if !hasKeyARN {
			resp.Diagnostics.AddAttributeError(path.Root(names.AttrEncryptionConfiguration).AtListIndex(0).AtName(names.AttrKMSKeyARN), "Missing KMS Key ARN", "kms_key_arn is required when key_type is CUSTOMER_MANAGED_KEY.")
		}
	case awstypes.KmsKeyTypeAwsOwnedKmsKey:
		if hasKeyARN {
			resp.Diagnostics.AddAttributeError(path.Root(names.AttrEncryptionConfiguration).AtListIndex(0).AtName(names.AttrKMSKeyARN), "Invalid KMS Key ARN", "kms_key_arn must not be specified when key_type is AWS_OWNED_KMS_KEY.")
		}
	}
}

func (r *instanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan instanceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().SSOAdminClient(ctx)
	input := ssoadmin.CreateInstanceInput{Tags: getTagsIn(ctx)}
	if !plan.ClientToken.IsNull() {
		input.ClientToken = plan.ClientToken.ValueStringPointer()
	}
	if !plan.Name.IsNull() {
		input.Name = plan.Name.ValueStringPointer()
	}
	output, err := conn.CreateInstance(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.ValueString())
		return
	}
	plan.ARN = fwtypes.ARNValue(aws.ToString(output.InstanceArn))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	described, err := waitInstanceActive(ctx, conn, plan.ARN.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
		return
	}
	if !plan.EncryptionConfiguration.IsNull() && !plan.EncryptionConfiguration.IsUnknown() {
		if err := updateInstanceEncryption(ctx, conn, plan.ARN.ValueString(), plan.EncryptionConfiguration); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}
		described, err = waitInstanceEncryptionEnabled(ctx, conn, plan.ARN.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}
	}
	resp.Diagnostics.Append(r.flatten(ctx, described, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *instanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state instanceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().SSOAdminClient(ctx)
	output, err := findInstanceByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
	resp.Diagnostics.Append(r.flatten(ctx, output, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tags, err := listTags(ctx, conn, state.ARN.ValueString(), state.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
	setTagsOut(ctx, svcTags(tags))
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *instanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var old, plan instanceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &old))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().SSOAdminClient(ctx)
	if !old.Name.Equal(plan.Name) {
		_, err := conn.UpdateInstance(ctx, &ssoadmin.UpdateInstanceInput{InstanceArn: plan.ARN.ValueStringPointer(), Name: plan.Name.ValueStringPointer()})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}
	}
	if !old.EncryptionConfiguration.Equal(plan.EncryptionConfiguration) {
		if err := updateInstanceEncryption(ctx, conn, plan.ARN.ValueString(), plan.EncryptionConfiguration); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}
		if _, err := waitInstanceEncryptionEnabled(ctx, conn, plan.ARN.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}
	}
	if !old.TagsAll.Equal(plan.TagsAll) {
		if err := updateTags(ctx, conn, plan.ARN.ValueString(), plan.ARN.ValueString(), old.TagsAll, plan.TagsAll); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
			return
		}
	}
	output, err := findInstanceByARN(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ARN.ValueString())
		return
	}
	resp.Diagnostics.Append(r.flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *instanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state instanceResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().SSOAdminClient(ctx)
	_, err := conn.DeleteInstance(ctx, &ssoadmin.DeleteInstanceInput{InstanceArn: state.ARN.ValueStringPointer()})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
		return
	}
	if err := waitInstanceDeleted(ctx, conn, state.ARN.ValueString(), r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.ValueString())
	}
}

func updateInstanceEncryption(ctx context.Context, conn *ssoadmin.Client, arn string, value fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel]) error {
	model, diags := value.ToPtr(ctx)
	if diags.HasError() {
		diagnostic := diags.Errors()[0]
		return fmt.Errorf("%s: %s", diagnostic.Summary(), diagnostic.Detail())
	}
	if model == nil {
		return nil
	}
	input := ssoadmin.UpdateInstanceInput{InstanceArn: aws.String(arn), EncryptionConfiguration: &awstypes.EncryptionConfiguration{KeyType: model.KeyType.ValueEnum()}}
	if !model.KMSKeyARN.IsNull() {
		input.EncryptionConfiguration.KmsKeyArn = model.KMSKeyARN.ValueStringPointer()
	}
	_, err := conn.UpdateInstance(ctx, &input)
	return smarterr.NewError(err)
}

func (r *instanceResource) flatten(ctx context.Context, output *ssoadmin.DescribeInstanceOutput, data *instanceResourceModel) (diags diag.Diagnostics) {
	diags.Append(fwflex.Flatten(ctx, output, data)...)
	if output.EncryptionConfigurationDetails != nil {
		d := output.EncryptionConfigurationDetails
		kmsKeyARN := fwtypes.ARNNull()
		if d.KmsKeyArn != nil {
			kmsKeyARN = fwtypes.ARNValue(aws.ToString(d.KmsKeyArn))
		}
		data.EncryptionConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &encryptionConfigurationModel{
			EncryptionStatus:       fwtypes.StringEnumValue(d.EncryptionStatus),
			EncryptionStatusReason: fwflex.StringToFramework(ctx, d.EncryptionStatusReason),
			KeyType:                fwtypes.StringEnumValue(d.KeyType),
			KMSKeyARN:              kmsKeyARN,
		})
	}
	return diags
}

func findInstanceByARN(ctx context.Context, conn *ssoadmin.Client, arn string) (*ssoadmin.DescribeInstanceOutput, error) {
	output, err := conn.DescribeInstance(ctx, &ssoadmin.DescribeInstanceInput{InstanceArn: aws.String(arn)})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}
	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}
	return output, nil
}

func statusInstance(conn *ssoadmin.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findInstanceByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return output, string(output.Status), nil
	}
}

func waitInstanceActive(ctx context.Context, conn *ssoadmin.Client, arn string, timeout time.Duration) (*ssoadmin.DescribeInstanceOutput, error) {
	conf := retry.StateChangeConf{Pending: enum.Slice(awstypes.InstanceStatusCreateInProgress), Target: enum.Slice(awstypes.InstanceStatusActive), Refresh: statusInstance(conn, arn), Timeout: timeout, Delay: 10 * time.Second, NotFoundChecks: 5}
	output, err := conf.WaitForStateContext(ctx)
	if out, ok := output.(*ssoadmin.DescribeInstanceOutput); ok && out.Status == awstypes.InstanceStatusCreateFailed {
		retry.SetLastError(err, errors.New(aws.ToString(out.StatusReason)))
		return out, err
	}
	if err != nil {
		return nil, err
	}
	return output.(*ssoadmin.DescribeInstanceOutput), nil
}

func waitInstanceEncryptionEnabled(ctx context.Context, conn *ssoadmin.Client, arn string, timeout time.Duration) (*ssoadmin.DescribeInstanceOutput, error) {
	pending := append([]string{""}, enum.Slice(awstypes.KmsKeyStatusUpdating)...)
	conf := retry.StateChangeConf{Pending: pending, Target: enum.Slice(awstypes.KmsKeyStatusEnabled), Refresh: func(ctx context.Context) (any, string, error) {
		output, err := findInstanceByARN(ctx, conn, arn)
		if err != nil {
			return nil, "", err
		}
		if output.EncryptionConfigurationDetails == nil {
			return output, "", nil
		}
		return output, string(output.EncryptionConfigurationDetails.EncryptionStatus), nil
	}, Timeout: timeout, Delay: 10 * time.Second}
	output, err := conf.WaitForStateContext(ctx)
	if out, ok := output.(*ssoadmin.DescribeInstanceOutput); ok && out.EncryptionConfigurationDetails != nil && out.EncryptionConfigurationDetails.EncryptionStatus == awstypes.KmsKeyStatusUpdateFailed {
		retry.SetLastError(err, errors.New(aws.ToString(out.EncryptionConfigurationDetails.EncryptionStatusReason)))
		return out, err
	}
	if err != nil {
		return nil, err
	}
	return output.(*ssoadmin.DescribeInstanceOutput), nil
}

func waitInstanceDeleted(ctx context.Context, conn *ssoadmin.Client, arn string, timeout time.Duration) error {
	conf := retry.StateChangeConf{Pending: enum.Slice(awstypes.InstanceStatusActive, awstypes.InstanceStatusDeleteInProgress), Target: []string{}, Refresh: statusInstance(conn, arn), Timeout: timeout, Delay: 10 * time.Second}
	_, err := conf.WaitForStateContext(ctx)
	return err
}

type instanceResourceModel struct {
	framework.WithRegionModel
	ARN                     fwtypes.ARN                                                   `tfsdk:"arn"`
	ClientToken             types.String                                                  `tfsdk:"client_token"`
	CreatedDate             timetypes.RFC3339                                             `tfsdk:"created_date"`
	EncryptionConfiguration fwtypes.ListNestedObjectValueOf[encryptionConfigurationModel] `tfsdk:"encryption_configuration" autoflex:"-"`
	IdentityStoreID         types.String                                                  `tfsdk:"identity_store_id"`
	Name                    types.String                                                  `tfsdk:"name"`
	OwnerAccountID          types.String                                                  `tfsdk:"owner_account_id"`
	Status                  fwtypes.StringEnum[awstypes.InstanceStatus]                   `tfsdk:"status"`
	StatusReason            types.String                                                  `tfsdk:"status_reason"`
	Tags                    tftags.Map                                                    `tfsdk:"tags"`
	TagsAll                 tftags.Map                                                    `tfsdk:"tags_all"`
	Timeouts                timeouts.Value                                                `tfsdk:"timeouts"`
}

type encryptionConfigurationModel struct {
	EncryptionStatus       fwtypes.StringEnum[awstypes.KmsKeyStatus] `tfsdk:"encryption_status"`
	EncryptionStatusReason types.String                              `tfsdk:"encryption_status_reason"`
	KeyType                fwtypes.StringEnum[awstypes.KmsKeyType]   `tfsdk:"key_type"`
	KMSKeyARN              fwtypes.ARN                               `tfsdk:"kms_key_arn"`
}
