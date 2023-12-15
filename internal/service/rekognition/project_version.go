package rekognition

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	awstypes "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkResource(name="Project")
func newResourceProjectVersion(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceProject{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultReadTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type resourceProjectVersion struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

const (
	ResNameProjectVersion = "Project Version"
)

func (r *resourceProjectVersion) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_rekognition_project_version"
}

func (r *resourceProjectVersion) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceProjectVersion) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":  framework.IDAttribute(),
			"arn": framework.ARNAttributeComputedOnly(),
			"project_arn": schema.StringAttribute{
				Required: true,
			},
			"version_name": schema.StringAttribute{
				Required: true,
			},
			"kms_key_id": schema.StringAttribute{
				Optional: true,
			},
			"version_description": schema.StringAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"output_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[outputConfigData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"bucket": schema.StringAttribute{
							Required: true,
						},
						"key_prefix": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"testing_data": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[trainingDataData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"assets": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[assetsData](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"ground_truth_manifest": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[groundTruthManifestData](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"s3_object": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ObjectData](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bucket": schema.StringAttribute{
																Optional: true,
															},
															"key_name": schema.StringAttribute{
																Optional: true,
															},
															"version": schema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Attributes: map[string]schema.Attribute{
						"auto_create": schema.BoolAttribute{
							Optional: true,
						},
					},
				},
			},
			"training_data": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[trainingDataData](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"assets": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[assetsData](ctx),
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"ground_truth_manifest": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[groundTruthManifestData](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"s3_object": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[s3ObjectData](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"bucket":   schema.StringAttribute{},
															"key_name": schema.StringAttribute{},
															"version":  schema.StringAttribute{},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceProjectVersion) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().RekognitionClient(ctx)

	var plan resourceProjectVersionData

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ocd []outputConfigData
	resp.Diagnostics.Append(plan.OutputConfig.ElementsAs(ctx, &ocd, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var testData []testingDataData
	resp.Diagnostics.Append(plan.TestingData.ElementsAs(ctx, &testData, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var trainData []trainingDataData
	resp.Diagnostics.Append(plan.TrainingData.ElementsAs(ctx, &trainData, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ocdInput := expandOutputConfig(ctx, ocd)
	testDataInput := expandTestData(ctx, testData)
	trainDataInput := expandTrainingData(ctx, trainData)

	in := rekognition.CreateProjectVersionInput{}
}

func (r *resourceProjectVersion) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (r *resourceProjectVersion) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *resourceProjectVersion) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

}

func expandOutputConfig(ctx context.Context, tfList []outputConfigData) *awstypes.OutputConfig {
	if len(tfList) == 0 {
		return nil
	}

	oc := tfList[0]

	return &awstypes.OutputConfig{
		S3Bucket:    flex.StringFromFramework(ctx, oc.S3Bucket),
		S3KeyPrefix: flex.StringFromFramework(ctx, oc.S3KeyPrefix),
	}
}

func flattenOutputConfig(apiObject *awstypes.OutputConfig) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	elemType := types.ObjectType{AttrTypes: outputConfigDataTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{}), diags
	}

	obj := map[string]attr.Value{
		"bucket":     types.StringValue(aws.ToString(apiObject.S3Bucket)),
		"key_prefix": types.StringValue(aws.ToString(apiObject.S3KeyPrefix)),
	}
	objVal, d := types.ObjectValue(outputConfigDataTypes, obj)
	diags.Append(d...)

	listVal, d := types.ListValue(elemType, []attr.Value{objVal})
	diags.Append(d...)

	return listVal, diags
}

func expandTestData(ctx context.Context, tfList []testingDataData) (*awstypes.TestingData, diag.Diagnostics) {
	diagnostics := diag.Diagnostics{}
	if len(tfList) == 0 {
		diagnostics.AddError("testing_data is empty", "")
		return nil, diagnostics
	}

	td := tfList[0]

	var assets []assetsData
	_ = td.Assets.ElementsAs(ctx, &assets, false)

	return &awstypes.TestingData{
		AutoCreate: *flex.BoolFromFramework(ctx, td.AutoCreate),
		Assets:     expandAssets(ctx, td.Assets),
	}, nil
}

func flattenTestData() {

}

func expandAssets(ctx context.Context, a []assetsData) []awstypes.Asset {
	if td.Assets.IsNull() {
		return nil
	}

	td.Assets.Elements()

	var assets []awstypes.Asset

}

func flattenAssets() {

}

type resourceProjectVersionData struct {
	ID                  types.String  `tfsdk:"id"`
	ARN                 types.String  `tfsdk:"arn"`
	OutputConfig        types.List    `tfsdk:"output_config"`
	ProjectARN          types.String  `tfsdk:"project_arn"`
	VersionName         types.String  `tfsdk:"version_name"`
	ConfidenceThreshold types.Float64 `tfsdk:"confidence_threshold"`
	KMSKeyID            types.String  `tfsdk:"kms_key_id"`
	Tags                types.Map     `tfsdk:"tags"`
	TagsAll             types.Map     `tfsdk:"tags_all"`
	TestingData         types.List    `tfsdk:"testing_data"`
	TrainingData        types.List    `tfsdk:"training_data"`
	VersionDescription  types.String  `tfsdk:"version_description"`
}

type outputConfigData struct {
	S3Bucket    types.String `tfsdk:"bucket"`
	S3KeyPrefix types.String `tfsdk:"key_prefix"`
}

type testingDataData struct {
	Assets     types.List `tfsdk:"assets"`
	AutoCreate types.Bool `tfsdk:"auto_create"`
}

type trainingDataData struct {
	Assets types.List `tfsdk:"assets"`
}

type assetsData struct {
	GroundTruthManifest types.List `tfsdk:"ground_truth_manifest"`
}

type groundTruthManifestData struct {
	S3Object types.List `tfsdk:"s3_object"`
}

type s3ObjectData struct {
	Bucket  types.String `tfsdk:"bucket"`
	Name    types.String `tfsdk:"key_name"`
	Version types.String `tfsdk:"version"`
}

var outputConfigDataTypes = map[string]attr.Type{
	"bucket":     types.StringType,
	"key_prefix": types.StringType,
}
