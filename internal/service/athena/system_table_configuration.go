// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package athena

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	awstypes "github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

// @FrameworkResource("aws_athena_system_table_configuration", name="System Table Configuration")
// @IdentityAttribute("table_name")
// @ArnFormat("system-table/{table_name}", attribute="arn")
// @Tags(identifierAttribute="arn")
// @Testing(generator=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttribute="table_name")
// @Testing(preCheckWithRegion="testAccSystemTableConfigurationPreCheck")
// @Testing(serialize=true)
func newSystemTableConfigurationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &systemTableConfigurationResource{}, nil
}

type systemTableConfigurationResource struct {
	framework.ResourceWithModel[systemTableConfigurationResourceModel]
	framework.WithImportByIdentity
}

const (
	systemTablesBucketName = "aws-athena"
	systemTablesNamespace  = "metadata"
	retainedSystemTable    = `the S3 Tables table "query_executions" in namespace "` + systemTablesNamespace + `" of table bucket "` + systemTablesBucketName + `"`
)

var kmsKeyARNRegex = regexache.MustCompile(`^arn:aws[a-z-]*:kms:[a-z0-9-]+:\d{12}:key/(mrk-)?[0-9a-f-]+$`)

func (r *systemTableConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	tableNameType := fwtypes.StringEnumType[awstypes.SystemTableName]()

	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(kmsKeyARNRegex, "must be a KMS key ARN, not an alias ARN. Athena stores the key that an alias points to, not the alias"),
				},
			},
			names.AttrTableName: schema.StringAttribute{
				CustomType: tableNameType,
				Optional:   true,
				Computed:   true,
				Default:    tableNameType.AttributeDefault(awstypes.SystemTableNameAll),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *systemTableConfigurationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	var planKey, stateKey fwtypes.ARN
	if !request.Plan.Raw.IsNull() {
		smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.GetAttribute(ctx, path.Root(names.AttrKMSKeyARN), &planKey))
	}
	if !request.State.Raw.IsNull() {
		smerr.AddEnrich(ctx, &response.Diagnostics, request.State.GetAttribute(ctx, path.Root(names.AttrKMSKeyARN), &stateKey))
	}
	if response.Diagnostics.HasError() {
		return
	}

	switch {
	case request.State.Raw.IsNull():
	case request.Plan.Raw.IsNull():
		response.Diagnostics.AddWarning(
			"Resource Destruction Considerations",
			"Destroying this resource turns off Athena system tables in this Region and removes its tags. "+
				"Athena keeps "+retainedSystemTable+" and its data. "+
				"The table keeps its encryption, so a later configuration must use the same kms_key_arn unless the table is deleted.",
		)
	case planKey.Equal(stateKey):
	case planKey.IsNull():
		response.RequiresReplace = append(response.RequiresReplace, path.Root(names.AttrKMSKeyARN))
		response.Diagnostics.AddAttributeWarning(
			path.Root(names.AttrKMSKeyARN),
			"Removing the KMS key replaces the configuration",
			"Athena cannot remove a key in place, so Terraform deletes the configuration and then creates it. "+
				"While "+retainedSystemTable+" exists, the create fails and system tables stay off until the table is deleted and the plan is applied again. "+
				"Deleting the table deletes the query history it holds.",
		)
	default:
		response.Diagnostics.AddAttributeWarning(
			path.Root(names.AttrKMSKeyARN),
			"Changing the KMS key requires deleting the system table",
			"Athena rejects a key change while "+retainedSystemTable+" exists. The apply then fails and changes nothing. "+
				"Delete the table first. Deleting the table deletes the query history it holds.",
		)
	}
}

func (r *systemTableConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data systemTableConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AthenaClient(ctx)

	name := data.TableName.ValueEnum()
	var input athena.CreateSystemTableConfigurationInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateSystemTableConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, systemTableConfigurationError(err, name), smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.ARN = fwflex.StringValueToFramework(ctx, systemTableConfigurationARN(ctx, r.Meta(), name))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *systemTableConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data systemTableConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AthenaClient(ctx)

	name := data.TableName.ValueEnum()
	output, err := findSystemTableConfigurationByName(ctx, conn, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.ARN = fwflex.StringValueToFramework(ctx, systemTableConfigurationARN(ctx, r.Meta(), name))

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *systemTableConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old systemTableConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AthenaClient(ctx)

	if !new.KMSKeyARN.Equal(old.KMSKeyARN) {
		name := new.TableName.ValueEnum()
		input := athena.UpdateSystemTableConfigurationInput{
			KmsKeyArn: fwflex.StringFromFramework(ctx, new.KMSKeyARN),
			TableName: name,
		}
		if _, err := conn.UpdateSystemTableConfiguration(ctx, &input); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, systemTableConfigurationError(err, name), smerr.ID, name)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *systemTableConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data systemTableConfigurationResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AthenaClient(ctx)

	name := data.TableName.ValueEnum()
	input := athena.DeleteSystemTableConfigurationInput{
		TableName: name,
	}
	if _, err := conn.DeleteSystemTableConfiguration(ctx, &input); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
	}
}

type systemTableConfigurationResourceModel struct {
	framework.WithRegionModel
	ARN       types.String                                 `tfsdk:"arn"`
	KMSKeyARN fwtypes.ARN                                  `tfsdk:"kms_key_arn"`
	TableName fwtypes.StringEnum[awstypes.SystemTableName] `tfsdk:"table_name"`
	Tags      tftags.Map                                   `tfsdk:"tags"`
	TagsAll   tftags.Map                                   `tfsdk:"tags_all"`
}

func systemTableConfigurationError(err error, name awstypes.SystemTableName) error {
	switch {
	case errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "already created"):
		return fmt.Errorf("%w. A system table configuration already exists in this Region. To manage it with Terraform, import it with the ID %q", err, name)
	case errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "different encryption"):
		return fmt.Errorf("%w. Athena keeps %s when a configuration is deleted, and the table keeps the encryption it was created with. Use the same kms_key_arn, or delete the table and apply again. Deleting the table deletes the query history it holds", err, retainedSystemTable)
	}

	return err
}

func systemTableConfigurationARN(ctx context.Context, c *conns.AWSClient, name awstypes.SystemTableName) string {
	return c.RegionalARN(ctx, names.Athena, "system-table/"+string(name))
}

func findSystemTableConfigurationByName(ctx context.Context, conn *athena.Client, name awstypes.SystemTableName) (*athena.GetSystemTableConfigurationOutput, error) {
	input := athena.GetSystemTableConfigurationInput{
		TableName: name,
	}
	output, err := conn.GetSystemTableConfiguration(ctx, &input)
	switch {
	case errs.IsA[*awstypes.ResourceNotFoundException](err):
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	case err != nil:
		return nil, smarterr.NewError(err)
	case output == nil:
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}
