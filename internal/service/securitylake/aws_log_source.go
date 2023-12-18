package securitylake

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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

// @FrameworkResource(name="Aws Log Source")
func newResourceAwsLogSource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAwsLogSource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAwsLogSource = "Aws Log Source"
)

type resourceAwsLogSource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceAwsLogSource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securitylake_aws_log_source"
}

func (r *resourceAwsLogSource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"sources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[awsLogSourceSourcesModel](ctx),
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

func (r *resourceAwsLogSource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)
	var data resourceAwsLogSourceModel
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
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameAwsLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	var awsLogSources []awsLogSourceSourcesModel
	resp.Diagnostics.Append(data.Sources.ElementsAs(ctx, &awsLogSources, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var regions []string
	for _, awsLogSource := range awsLogSources {
		resp.Diagnostics.Append(awsLogSource.Regions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	//Looking if datalake exists to create a uniqu ID for the Log Resource
	datalakeIn := &securitylake.ListDataLakesInput{
		Regions: regions,
	}

	datalakes, err := conn.ListDataLakes(ctx, datalakeIn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameAwsLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	//Creating the unique ID
	datalake := datalakes.DataLakes[0]
	var id string
	for _, awsLogawsLogSources := range awsLogSources {
		id = *datalake.DataLakeArn + "/" + awsLogawsLogSources.SourceName.ValueString() + "/" + awsLogawsLogSources.SourceVersion.ValueString()
	}

	data.ID = flex.StringToFramework(ctx, &id)
	out, err := findAwsLogSourceById(ctx, conn, id)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameAwsLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	//Trying to get the informatiom we need from output
	config, err := extractAwsLogSourceConfiguration(out)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameAwsLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	var awsLogSourceSource awsLogSourceSourcesModel
	resp.Diagnostics.Append(flex.Flatten(ctx, config, &awsLogSourceSource)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Sources = fwtypes.NewListNestedObjectValueOfPtr(ctx, &awsLogSourceSource)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *resourceAwsLogSource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data resourceAwsLogSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAwsLogSourceById(ctx, conn, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameAwsLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	if len(out.Sources) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	if tfresource.NotFound(err) || out.Sources == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	config, err := extractAwsLogSourceConfiguration(out)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameAwsLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	var awsLogSourceSource awsLogSourceSourcesModel
	resp.Diagnostics.Append(flex.Flatten(ctx, config, &awsLogSourceSource)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Sources = fwtypes.NewListNestedObjectValueOfPtr(ctx, &awsLogSourceSource)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceAwsLogSource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceAwsLogSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceAwsLogSource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SecurityLakeClient(ctx)

	var data resourceAwsLogSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAwsLogSourceById(ctx, conn, data.ID.ValueString())

	//Trying to get the informatiom we need from output
	config, err := extractAwsLogSourceConfiguration(out)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameAwsLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = conn.DeleteAwsLogSource(ctx, &securitylake.DeleteAwsLogSourceInput{
		Sources: []awstypes.AwsLogSourceConfiguration{*config},
	})

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionDeleting, ResNameAwsLogSource, data.ID.String(), err),
			err.Error(),
		)
		return
	}

}

func findAwsLogSourceById(ctx context.Context, conn *securitylake.Client, id string) (*securitylake.ListLogSourcesOutput, error) {

	parsedID, err := parseARNString(id)
	if err != nil {

		return nil, err
	}

	awsLogSource := awstypes.AwsLogSourceResource{
		SourceVersion: aws.String(parsedID.SourceVersion),
		SourceName:    awstypes.AwsLogSourceName(parsedID.SourceName),
	}

	logSourceResource := &awstypes.LogSourceResourceMemberAwsLogSource{
		Value: awsLogSource,
	}

	input := &securitylake.ListLogSourcesInput{
		Regions: []string{parsedID.Regions},
		Sources: []awstypes.LogSourceResource{logSourceResource},
	}

	output, err := conn.ListLogSources(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error listing log sources for region %s: %w", parsedID.Regions, err)
	}

	return output, nil
}

func (r *resourceAwsLogSource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceAwsLogSourceModel struct {
	Sources  fwtypes.ListNestedObjectValueOf[awsLogSourceSourcesModel] `tfsdk:"sources"`
	ID       types.String                                              `tfsdk:"id"`
	Timeouts timeouts.Value                                            `tfsdk:"timeouts"`
}

type awsLogSourceSourcesModel struct {
	Accounts      fwtypes.SetValueOf[types.String] `tfsdk:"accounts"`
	Regions       fwtypes.SetValueOf[types.String] `tfsdk:"regions"`
	SourceName    types.String                     `tfsdk:"source_name"`
	SourceVersion types.String                     `tfsdk:"source_version"`
}

type ParsedID struct {
	Accounts      string
	Regions       string
	SourceName    string
	SourceVersion string
}

// parsing id that includes datalake arn.
func parseARNString(s string) (*ParsedID, error) {
	v, err := arn.Parse(s)

	if err != nil {
		return nil, err
	}

	resource := v.Resource

	parts := strings.Split(resource, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid ID format")
	}

	return &ParsedID{
		Accounts:      v.AccountID,
		Regions:       v.Region,
		SourceName:    parts[len(parts)-2],
		SourceVersion: parts[len(parts)-1],
	}, nil
}

// extractAwsLogSourceConfiguration extracts the configuration from the first log source in the output.
func extractAwsLogSourceConfiguration(out *securitylake.ListLogSourcesOutput) (*awstypes.AwsLogSourceConfiguration, error) {
	if len(out.Sources) == 0 {
		return nil, fmt.Errorf("no log sources found in the output")
	}

	logSource := out.Sources[0]
	var awsLogSourceResource awstypes.AwsLogSourceResource

	// Assuming the log source has at least one source.
	if len(logSource.Sources) == 0 {
		return nil, fmt.Errorf("no log source resources found")
	}

	// Extracting AwsLogSourceResource
	if awsLogSource, ok := logSource.Sources[0].(*awstypes.LogSourceResourceMemberAwsLogSource); ok {
		awsLogSourceResource = awsLogSource.Value
	} else {
		return nil, fmt.Errorf("log source resource is not of type AwsLogSourceResource")
	}

	// Creating the configuration
	config := &awstypes.AwsLogSourceConfiguration{
		Accounts:      []string{*logSource.Account},
		Regions:       []string{*logSource.Region},
		SourceName:    awsLogSourceResource.SourceName,
		SourceVersion: awsLogSourceResource.SourceVersion,
	}

	return config, nil
}
