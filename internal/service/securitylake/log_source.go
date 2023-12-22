package securitylake

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securitylake"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securitylake/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Log Source")
func newResourceLogSource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceLogSource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameLogSource = "Log Source"
)

type resourceLogSource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceLogSource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_log_source"
}

func (r *resourceLogSource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"sources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[logSourceSourcesModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"accounts": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"regions": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
						},
						"source_name": schema.StringAttribute{
							Required: true,
						},
						"source_version": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceLogSource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)
	var data resourceLogSourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &securitylake.CreateAwsLogSourceInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.CreateAwsLogSource(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	var logSources []logSourceSourcesModel
	resp.Diagnostics.Append(data.Sources.ElementsAs(ctx, &logSources, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var regions []string
	for _, logSource := range logSources {
		resp.Diagnostics.Append(logSource.Regions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	var id string
	for _, awsLoglogSources := range logSources {
		id = awsLoglogSources.SourceName.ValueString() + "/" + awsLoglogSources.SourceVersion.ValueString()
	}

	data.ID = flex.StringToFramework(ctx, &id)
	out, err := findLogSourceById(ctx, conn, regions, id)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	//Trying to get the informatiom we need from output
	config, err := extractLogSourceConfiguration(out)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	var logSourceSource logSourceSourcesModel
	resp.Diagnostics.Append(flex.Flatten(ctx, config, &logSourceSource)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Sources = fwtypes.NewListNestedObjectValueOfPtr(ctx, &logSourceSource)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *resourceLogSource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data resourceLogSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var logSources []logSourceSourcesModel
	resp.Diagnostics.Append(data.Sources.ElementsAs(ctx, &logSources, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var regions []string
	for _, logSource := range logSources {
		resp.Diagnostics.Append(logSource.Regions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	out, err := findLogSourceById(ctx, conn, regions, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	if tfresource.NotFound(err) || len(out.Sources) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	config, err := extractLogSourceConfiguration(out)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	var logSourceSource logSourceSourcesModel
	resp.Diagnostics.Append(flex.Flatten(ctx, config, &logSourceSource)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Sources = fwtypes.NewListNestedObjectValueOfPtr(ctx, &logSourceSource)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceLogSource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceLogSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceLogSource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data resourceLogSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var logSources []logSourceSourcesModel
	resp.Diagnostics.Append(data.Sources.ElementsAs(ctx, &logSources, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var regions []string
	for _, logSource := range logSources {
		resp.Diagnostics.Append(logSource.Regions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	out, err := findLogSourceById(ctx, conn, regions, data.ID.ValueString())

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	//Trying to get the informatiom we need from output
	var config *awstypes.AwsLogSourceConfiguration

	config, err = extractLogSourceConfiguration(out)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteAwsLogSource(ctx, &securitylake.DeleteAwsLogSourceInput{
		Sources: []awstypes.AwsLogSourceConfiguration{*config},
	})

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func findLogSourceById(ctx context.Context, conn *securitylake.Client, regions []string, id string) (*securitylake.ListLogSourcesOutput, error) {
	parsedID, err := parseARNString(id)
	if err != nil {
		return nil, err
	}

	logSource := awstypes.AwsLogSourceResource{
		SourceVersion: aws.String(parsedID.SourceVersion),
		SourceName:    awstypes.AwsLogSourceName(parsedID.SourceName),
	}

	logSourceResource := &awstypes.LogSourceResourceMemberAwsLogSource{
		Value: logSource,
	}

	input := &securitylake.ListLogSourcesInput{
		Regions: regions,
		Sources: []awstypes.LogSourceResource{logSourceResource},
	}

	output, err := conn.ListLogSources(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func (r *resourceLogSource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceLogSourceModel struct {
	Sources  fwtypes.ListNestedObjectValueOf[logSourceSourcesModel] `tfsdk:"sources"`
	ID       types.String                                           `tfsdk:"id"`
	Timeouts timeouts.Value                                         `tfsdk:"timeouts"`
}

type logSourceSourcesModel struct {
	Accounts      fwtypes.SetValueOf[types.String] `tfsdk:"accounts"`
	Regions       fwtypes.SetValueOf[types.String] `tfsdk:"regions"`
	SourceName    types.String                     `tfsdk:"source_name"`
	SourceVersion types.String                     `tfsdk:"source_version"`
}

type ParsedID struct {
	SourceName    string
	SourceVersion string
}

// parsing id that includes datalake arn.
func parseARNString(s string) (*ParsedID, error) {
	parts := strings.Split(s, "/")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid ID format")
	}

	return &ParsedID{
		SourceName:    parts[0],
		SourceVersion: parts[1],
	}, nil
}

// extractlogSourceConfiguration extracts the configuration from the first log source in the output.
func extractLogSourceConfiguration(out *securitylake.ListLogSourcesOutput) (*awstypes.AwsLogSourceConfiguration, error) {
	if len(out.Sources) == 0 {
		var nfe *awstypes.ResourceNotFoundException
		return nil, nfe
	}

	var logSourceResource awstypes.AwsLogSourceResource
	var accounts []string
	var regions []string
	for _, logSource := range out.Sources {
		if logSource.Account != nil {
			if !contains(accounts, *logSource.Account) {
				accounts = append(accounts, *logSource.Account)
			}
		}
		if logSource.Region != nil {
			regions = append(regions, *logSource.Region)
		}
		// Extracting logSourceResource
		if logSource, ok := logSource.Sources[0].(*awstypes.LogSourceResourceMemberAwsLogSource); ok {
			logSourceResource = logSource.Value
		} else {
			return nil, fmt.Errorf("log source resource is not of type logSourceResource")
		}
	}

	// Creating the configuration
	config := &awstypes.AwsLogSourceConfiguration{
		Accounts:      accounts,
		Regions:       regions,
		SourceName:    logSourceResource.SourceName,
		SourceVersion: logSourceResource.SourceVersion,
	}

	return config, nil
}

// contains checks if a string is present in a slice.
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
