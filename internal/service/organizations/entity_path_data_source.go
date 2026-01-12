// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package organizations

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	awstypes "github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkDataSource("aws_organizations_entity_path", name="Entity Path")
func newEntityPathDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &entityPathDataSource{}, nil
}

type entityPathDataSource struct {
	framework.DataSourceWithModel[entityPathDataSourceModel]
}

func (d *entityPathDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"entity_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^(\d{12})|(ou-[0-9a-z]{4,32}-[a-z0-9]{8,32})$`), "must be an organizational unit (OU) or AWS account ID"),
				},
			},
			"entity_path": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *entityPathDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data entityPathDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().OrganizationsClient(ctx)

	// https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_last-accessed-view-data-orgs.html#access_policies_last-accessed-viewing-orgs-entity-path.
	entityID := fwflex.StringValueFromFramework(ctx, data.EntityID)
	childID := entityID
	var parts []string
	for {
		input := organizations.ListParentsInput{
			ChildId: aws.String(childID),
		}
		parent, err := findParent(ctx, conn, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Organizations entity (%s) parent", childID), err.Error())
			return
		}

		parts = append(parts, childID)
		parentID := aws.ToString(parent.Id)

		if parent.Type == awstypes.ParentTypeRoot {
			organization, err := findOrganization(ctx, conn)
			if err != nil {
				response.Diagnostics.AddError("reading Organizations organization", err.Error())
				return
			}

			parts = append(parts, parentID)                      // Root.
			parts = append(parts, aws.ToString(organization.Id)) // Organization.

			break
		}

		childID = parentID
	}

	slices.Reverse(parts)
	data.EntityPath = fwflex.StringValueToFramework(ctx, strings.Join(parts, "/"))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type entityPathDataSourceModel struct {
	EntityID   types.String `tfsdk:"entity_id"`
	EntityPath types.String `tfsdk:"entity_path"`
}
