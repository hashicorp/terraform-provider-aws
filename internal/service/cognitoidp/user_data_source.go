// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cognitoidp

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_cognito_user", name="User")
func newUserDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &userDataSource{}, nil
}

type userDataSource struct {
	framework.DataSourceWithModel[userDataSourceModel]
}

func (d *userDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAttributes: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrCreationDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrEnabled: schema.BoolAttribute{
				Computed: true,
			},
			names.AttrEmail: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"last_modified_date": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"mfa_setting_list": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"preferred_mfa_setting": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"sub": schema.StringAttribute{
				Computed: true,
			},
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
			},
			names.AttrUsername: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
				},
			},
		},
	}
}

func (d *userDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrEmail),
			path.MatchRoot(names.AttrUsername),
		),
	}
}

func (d *userDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data userDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CognitoIDPClient(ctx)

	userPoolID := data.UserPoolID.ValueString()
	username := data.Username.ValueString()

	if !data.Email.IsNull() {
		email := data.Email.ValueString()
		userSummary, err := tfresource.RetryWhenNotFound(ctx, propagationTimeout, func(ctx context.Context) (*awstypes.UserType, error) {
			return findUserByEmail(ctx, conn, userPoolID, email)
		})

		if err != nil {
			title := fmt.Sprintf("reading Cognito User by email (%s)", email)
			switch {
			case errors.Is(err, tfresource.ErrTooManyResults):
				response.Diagnostics.AddError(title, fmt.Sprintf("multiple Cognito Users found with email %q; email must match exactly one user", email))
			case retry.NotFound(err):
				response.Diagnostics.AddError(title, fmt.Sprintf("no Cognito User found with email %q", email))
			default:
				response.Diagnostics.AddError(title, err.Error())
			}

			return
		}

		username = aws.ToString(userSummary.Username)
		if username == "" {
			response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User by email (%s)", email), "email lookup returned a Cognito User without a username")
			return
		}
		data.Username = types.StringValue(username)
	}

	id := userCreateResourceID(userPoolID, username)
	user, err := findUserByTwoPartKey(ctx, conn, userPoolID, username)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User (%s)", id), err.Error())

		return
	}

	data.Attributes = fwflex.FlattenFrameworkStringValueMapOfString(ctx, flattenAttributeTypes(user.UserAttributes))
	data.CreationDate = timetypes.NewRFC3339TimePointerValue(user.UserCreateDate)
	data.Enabled = types.BoolValue(user.Enabled)
	data.ID = fwflex.StringValueToFramework(ctx, id)
	data.LastModifiedDate = timetypes.NewRFC3339TimePointerValue(user.UserLastModifiedDate)
	data.MFASettingList = fwflex.FlattenFrameworkStringValueSetOfStringLegacy(ctx, user.UserMFASettingList)
	data.PreferredMFASetting = fwflex.StringToFramework(ctx, user.PreferredMfaSetting)
	data.Status = fwflex.StringValueToFramework(ctx, user.UserStatus)
	data.Sub = fwflex.StringToFramework(ctx, flattenUserSub(user.UserAttributes))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findUserByEmail(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID, email string) (*awstypes.UserType, error) {
	input := &cognitoidentityprovider.ListUsersInput{
		Filter:     aws.String(userEmailFilter(email)),
		Limit:      aws.Int32(2),
		UserPoolId: aws.String(userPoolID),
	}
	var users []awstypes.UserType

	pages := cognitoidentityprovider.NewListUsersPaginator(conn, input, func(o *cognitoidentityprovider.ListUsersPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, user := range page.Users {
			users = append(users, user)
			if len(users) == 2 {
				return nil, tfresource.NewTooManyResultsError(2, input)
			}
		}
	}

	return tfresource.AssertSingleValueResult(users)
}

func userEmailFilter(email string) string {
	email = strings.ReplaceAll(email, `\`, `\\`)
	email = strings.ReplaceAll(email, `"`, `\"`)
	return fmt.Sprintf(`email = "%s"`, email)
}

type userDataSourceModel struct {
	framework.WithRegionModel
	Attributes          fwtypes.MapOfString `tfsdk:"attributes"`
	CreationDate        timetypes.RFC3339   `tfsdk:"creation_date"`
	Enabled             types.Bool          `tfsdk:"enabled"`
	Email               types.String        `tfsdk:"email"`
	ID                  types.String        `tfsdk:"id"`
	LastModifiedDate    timetypes.RFC3339   `tfsdk:"last_modified_date"`
	MFASettingList      fwtypes.SetOfString `tfsdk:"mfa_setting_list"`
	PreferredMFASetting types.String        `tfsdk:"preferred_mfa_setting"`
	Status              types.String        `tfsdk:"status"`
	Sub                 types.String        `tfsdk:"sub"`
	UserPoolID          types.String        `tfsdk:"user_pool_id"`
	Username            types.String        `tfsdk:"username"`
}
