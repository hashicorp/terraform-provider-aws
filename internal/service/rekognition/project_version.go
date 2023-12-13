package rekognition

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
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
		Attributes: map[string]schema.Attribute{},
	}
}

func (r *resourceProjectVersion) Create(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

}

func (r *resourceProjectVersion) Read(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

}

func (r *resourceProjectVersion) Update(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

}

func (r *resourceProjectVersion) Delete(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {

}

type resourceProjectVersionData struct {
	ID                  types.String  `tfsdk:"id"`
	ARN                 types.String  `tfsdk:"arn"`
	OutputS3            types.Map     `tfsdk:"output_s3"`
	ProjectARN          types.String  `tfsdk:"project_arn"`
	VersionName         types.String  `tfsdk:"version_name"`
	ConfidenceThreshold types.Float64 `tfsdk:"confidence_threshold"`
	KMSKeyID            types.String  `tfsdk:"kms_key_id"`
	Tags                types.Map     `tfsdk:"tags"`
	TagsAll             types.Map     `tfsdk:"tags_all"`
	TestingData         types.Map     `tfsdk:"testing_data"`
	TrainingData        types.Map     `tfsdk:"training_data"`
	VersionDescription  types.String  `tfsdk:"version_description"`
}

type outputS3Data struct {
	S3Bucket    types.String `tfsdk:"s3_bucket"`
	S3KeyPrefix types.String `tfsdk:"s3_key_prefix"`
}

type testingDataData struct {
	Assets     types.List `tfsdk:"assets"`
	AutoCreate types.Bool `tfsdk:"auto_create"`
}

type trainingDataData struct {
}
