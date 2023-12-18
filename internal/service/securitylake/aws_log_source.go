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
				// CustomType: fwtypes.NewListNestedObjectTypeOf[awsLogSourceSourcesModel](ctx),
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

	// var awsLogSources []awsLogSourceSourcesModel
	// resp.Diagnostics.Append(data.Sources.ElementsAs(ctx, &awsLogSources, false)...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// sourcesModel := expandSources(awsLogSources)

	input := &securitylake.CreateAwsLogSourceInput{
		// Sources: []awstypes.AwsLogSourceConfiguration{sourcesModel},
	}
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
	region := awsLogSources

	//Looking if datalake exists to create a uniqu ID for the Log Resource
	datalakeIn := &securitylake.ListDataLakesInput{
		Regions: []string{region},
	}

	datalake, err := conn.ListDataLakes(ctx, datalakeIn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecurityLake, create.ErrActionCreating, ResNameAwsLogSource, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	//Creating the unique ID
	id := *datalake.DataLakes[0].DataLakeArn + "/" + awsLogSources[0].SourceName.ValueString() + "/" + awsLogSources[0].SourceVersion.ValueString()
	data.ID = flex.StringToFramework(ctx, &id)

	out, err := findAwsLogSourceById(ctx, conn, id)
	//Trying to get the informatiom we need from output
	config := awstypes.AwsLogSourceConfiguration{}
	source := out.Sources[0]
	var awsLogSourcex awstypes.AwsLogSourceResource
	if awsLogSource, ok := source.Sources[0].(*awstypes.LogSourceResourceMemberAwsLogSource); ok {
		awsLogSourcex = awsLogSource.Value
	}
	config.Accounts = []string{*source.Account}
	config.Regions = []string{*source.Region}
	config.SourceName = awsLogSourcex.SourceName
	config.SourceVersion = awsLogSourcex.SourceVersion

	if resp.Diagnostics.HasError() {
		return
	}
	var awsLogSourcesx awsLogSourceSourcesModel
	resp.Diagnostics.Append(flex.Flatten(ctx, config, &awsLogSourcesx)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.Sources = flattenSources(ctx, &config)
	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func extractAwsLogSourceResource(source awstypes.LogSource) (awstypes.AwsLogSourceResource, error) {
	// Implement the logic to extract AwsLogSourceResource from LogSource
	// This is an example and may need to be adjusted based on actual implementation

	return awstypes.AwsLogSourceResource{}, fmt.Errorf("AWS Log Source Resource not found")
}

func (r *resourceAwsLogSource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// conn := r.Meta().SecurityLakeClient(ctx)

	var state resourceAwsLogSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// out := findLogSourceByRegion(ctx, conn, "eu-west-2")
	// fmt.Println(out)
	// if tfresource.NotFound(err) {
	// 	resp.State.RemoveResource(ctx)
	// 	return
	// }
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		create.ProblemStandardMessage(names.SecurityLake, create.ErrActionSetting, ResNameAwsLogSource, state.ID.String(), err),
	// 		err.Error(),
	// 	)
	// 	return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAwsLogSource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourceAwsLogSourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
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

	// out := findLogSourceByRegion(ctx, conn, "eu-west-2", data.Sources)
	// fmt.Println(out.Sources)

	in := &securitylake.DeleteAwsLogSourceInput{
		Sources: []awstypes.AwsLogSourceConfiguration{},
	}

	_, err := conn.DeleteAwsLogSource(ctx, in)

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

// func findLogSource(ctx context.Context, conn *securitylake.Client, input *securitylake.ListLogSourcesInput, filter tfslices.Predicate[*awstypes.LogSourceResource]) (*awstypes.LogSourceResource, error) {
// 	output, err := findLogSources(ctx, conn, input, filter)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return tfresource.AssertSinglePtrResult(output)
// }

// func findLogSources(ctx context.Context, conn *securitylake.Client, input *securitylake.ListLogSourcesInput, filter tfslices.Predicate[*awstypes.LogSourceResource]) ([]*awstypes.LogSourceResource, error) {
// 	var logSourceResources *awstypes.AwsLogSourceResource

// 	output, err := conn.ListLogSources(ctx, input)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if output == nil {
// 		return nil, tfresource.NewEmptyResultError(input)
// 	}

// 	for _, v := range output.Sources {
// 		v := v.Sources
// 		for _, v := range v {
// 			if v := &v; filter(v) {
// 				logSourceResources = append(logSourceResources, v)
// 			}
// 		}
// 	}

// 	return logSourceResources, nil
// }

func (r *resourceAwsLogSource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceAwsLogSourceModel struct {
	Sources  types.List     `tfsdk:"sources"`
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type awsLogSourceSourcesModel struct {
	Accounts      types.List   `tfsdk:"accounts"`
	Regions       types.List   `tfsdk:"regions"`
	SourceName    types.String `tfsdk:"source_name"`
	SourceVersion types.String `tfsdk:"source_version"`
}

type ParsedID struct {
	Accounts      string
	Regions       string
	SourceName    string
	SourceVersion string
}

// func flattenSources(ctx context.Context, apiObject *awstypes.AwsLogSourceConfiguration) types.List {
// 	attributeTypes := fwtypes.AttributeTypesMust[awsLogSourceSourcesModel](ctx)
// 	elemType := types.ObjectType{AttrTypes: attributeTypes}

// 	if apiObject == nil {
// 		return types.ListNull(elemType)
// 	}
// 	account := apiObject.Accounts
// 	region := apiObject.Regions
// 	sourceName := apiObject.SourceName
// 	sourceVersion := apiObject.SourceVersion

// 	attrs := map[string]attr.Value{
// 		"accounts":       flex.FlattenFrameworkStringList(ctx, &account),
// 		"regions":        flex.StringToFramework(ctx, &region),
// 		"source_name":    flex.StringValueToFramework(ctx, sourceName),
// 		"source_version": flex.StringToFramework(ctx, sourceVersion),
// 	}

// 	val := types.ObjectValueMust(attributeTypes, attrs)

// 	return types.ListValueMust(elemType, []attr.Value{val})
// }

func expandSources(tfList []awsLogSourceSourcesModel) awstypes.AwsLogSourceConfiguration {
	// if len(tfList) == 0 {
	// 	return nil
	// }

	tfObj := tfList[0]

	apiObject := awstypes.AwsLogSourceConfiguration{
		Regions:       []string{tfObj.Regions.ValueString()},
		SourceVersion: aws.String(tfObj.SourceVersion.ValueString()),
		SourceName:    awstypes.AwsLogSourceName(tfObj.SourceName.ValueString()),
	}

	if !tfObj.Accounts.IsNull() {
		apiObject.Accounts = []string{tfObj.Accounts.ValueString()}
	}

	return apiObject
}

func expandListSources(sourceVersion, sourceName string) awstypes.AwsLogSourceResource {

	apiObject := awstypes.AwsLogSourceResource{
		SourceVersion: aws.String(sourceVersion),
		SourceName:    awstypes.AwsLogSourceName(sourceName),
	}

	return apiObject
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
