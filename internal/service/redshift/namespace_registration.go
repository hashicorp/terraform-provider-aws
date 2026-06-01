// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/aws/aws-sdk-go-v2/service/redshiftserverless"
	redshiftserverlesstypes "github.com/aws/aws-sdk-go-v2/service/redshiftserverless/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

const (
	namespaceRegistrationInvalidClusterStateFaultTimeout = 15 * time.Minute
	namespaceTypeServerless                              = "serverless"
	namespaceTypeProvisioned                             = "provisioned"
)

// @FrameworkResource("aws_redshift_namespace_registration", name="Namespace Registration")
// @IdentityAttribute("consumer_identifier")
// @IdentityAttribute("namespace_type")
// @IdentityAttribute("serverless_namespace_identifier", optional="true")
// @IdentityAttribute("serverless_workgroup_identifier", optional="true")
// @IdentityAttribute("provisioned_cluster_identifier", optional="true")
// @ImportIDHandler("namespaceRegistrationImportID", setIDAttribute=true)
// @Testing(hasNoPreExistingResource=true)
// @Testing(identityTest=false)
func newNamespaceRegistrationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &namespaceRegistrationResource{}, nil
}

type namespaceRegistrationResource struct {
	framework.ResourceWithModel[namespaceRegistrationResourceModel]
	framework.WithNoUpdate
	framework.WithImportByIdentity
}

func (r *namespaceRegistrationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"consumer_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provisioned_cluster_identifier": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"serverless_namespace_identifier": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"serverless_workgroup_identifier": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *namespaceRegistrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data namespaceRegistrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	input := &redshift.RegisterNamespaceInput{
		ConsumerIdentifiers: []string{data.ConsumerIdentifier.ValueString()},
	}

	if data.NamespaceType.ValueString() == namespaceTypeServerless {
		input.NamespaceIdentifier = &awstypes.NamespaceIdentifierUnionMemberServerlessIdentifier{
			Value: awstypes.ServerlessIdentifier{
				NamespaceIdentifier: fwflex.StringFromFramework(ctx, data.ServerlessNamespaceIdentifier),
				WorkgroupIdentifier: fwflex.StringFromFramework(ctx, data.ServerlessWorkgroupIdentifier),
			},
		}
	} else {
		input.NamespaceIdentifier = &awstypes.NamespaceIdentifierUnionMemberProvisionedIdentifier{
			Value: awstypes.ProvisionedIdentifier{
				ClusterIdentifier: fwflex.StringFromFramework(ctx, data.ProvisionedClusterIdentifier),
			},
		}
	}

	_, err := tfresource.RetryWhenIsA[any, *awstypes.InvalidClusterStateFault](ctx, namespaceRegistrationInvalidClusterStateFaultTimeout,
		func(ctx context.Context) (any, error) {
			return conn.RegisterNamespace(ctx, input)
		})
	if err != nil {
		response.Diagnostics.AddError("creating Redshift Namespace Registration", err.Error())
		return
	}

	// Wait for the internal data share to be created
	if data.NamespaceType.ValueString() == namespaceTypeServerless {
		// Get the namespace ID (UUID) for building the data share ARN
		// serverless_namespace_identifier can be either name or ID
		serverlessConn := r.Meta().RedshiftServerlessClient(ctx)
		namespaceIdentifier := data.ServerlessNamespaceIdentifier.ValueString()

		namespace, err := serverlessConn.GetNamespace(ctx, &redshiftserverless.GetNamespaceInput{
			NamespaceName: aws.String(namespaceIdentifier),
		})

		var namespaceID string
		if err != nil {
			// If GetNamespace fails, assume the identifier is already the namespace ID
			namespaceID = namespaceIdentifier
		} else {
			namespaceID = aws.ToString(namespace.Namespace.NamespaceId)
		}

		err = waitInternalDataShareCreated(ctx, conn,
			namespaceID,
			r.Meta().AccountID(ctx),
			r.Meta().Region(ctx),
			r.Meta().Partition(ctx),
			namespaceRegistrationInvalidClusterStateFaultTimeout)
		if err != nil {
			response.Diagnostics.AddError("waiting for Redshift internal data share creation", err.Error())
			return
		}
	} else if data.NamespaceType.ValueString() == namespaceTypeProvisioned {
		// Get the namespace ID from the cluster
		cluster, err := findClusterByID(ctx, conn, data.ProvisionedClusterIdentifier.ValueString())
		if err != nil {
			response.Diagnostics.AddError("reading Redshift Cluster for namespace ID", err.Error())
			return
		}

		// Extract namespace ID from ClusterNamespaceArn
		// Format: arn:aws:redshift:region:account:namespace:namespace-id
		namespaceArn := aws.ToString(cluster.ClusterNamespaceArn)
		parts := strings.Split(namespaceArn, ":")
		if len(parts) < 7 {
			response.Diagnostics.AddError("parsing cluster namespace ARN", fmt.Sprintf("invalid ARN format: %s", namespaceArn))
			return
		}
		namespaceID := parts[6]

		err = waitInternalDataShareCreated(ctx, conn,
			namespaceID,
			r.Meta().AccountID(ctx),
			r.Meta().Region(ctx),
			r.Meta().Partition(ctx),
			namespaceRegistrationInvalidClusterStateFaultTimeout)
		if err != nil {
			response.Diagnostics.AddError("waiting for Redshift internal data share creation", err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.readNamespaceRegistration(ctx, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *namespaceRegistrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data namespaceRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	notFound := r.readNamespaceRegistration(ctx, &data, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	if notFound {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *namespaceRegistrationResource) readNamespaceRegistration(ctx context.Context, data *namespaceRegistrationResourceModel, diags *diag.Diagnostics) (notFound bool) {
	conn := r.Meta().RedshiftClient(ctx)
	serverlessConn := r.Meta().RedshiftServerlessClient(ctx)

	var namespaceID string

	if data.NamespaceType.ValueString() == namespaceTypeServerless {
		identifier := data.ServerlessNamespaceIdentifier.ValueString()
		if len(identifier) == 36 && identifier[8] == '-' && identifier[13] == '-' {
			namespaceID = identifier
		} else {
			namespace, err := serverlessConn.GetNamespace(ctx, &redshiftserverless.GetNamespaceInput{
				NamespaceName: aws.String(identifier),
			})
			if err != nil {
				if errs.IsA[*redshiftserverlesstypes.ResourceNotFoundException](err) {
					return true
				}
				diags.AddError("reading Redshift Serverless Namespace", err.Error())
				return false
			}
			namespaceID = aws.ToString(namespace.Namespace.NamespaceId)
		}
	} else {
		cluster, err := findClusterByID(ctx, conn, data.ProvisionedClusterIdentifier.ValueString())
		if retry.NotFound(err) {
			return true
		}
		if err != nil {
			diags.AddError("reading Redshift Cluster", err.Error())
			return false
		}

		namespaceArn := aws.ToString(cluster.ClusterNamespaceArn)
		parts := strings.Split(namespaceArn, ":")
		if len(parts) < 7 {
			diags.AddError("parsing cluster namespace ARN", fmt.Sprintf("invalid ARN format: %s", namespaceArn))
			return false
		}
		namespaceID = parts[6]
	}

	_, err := findInternalDataShareByNamespaceID(ctx, conn, namespaceID, r.Meta().AccountID(ctx), r.Meta().Region(ctx), r.Meta().Partition(ctx))
	if retry.NotFound(err) {
		return true
	}
	if err != nil {
		diags.AddError("reading Redshift Namespace Registration", err.Error())
	}
	return false
}

func (r *namespaceRegistrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data namespaceRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().RedshiftClient(ctx)

	input := &redshift.DeregisterNamespaceInput{
		ConsumerIdentifiers: []string{data.ConsumerIdentifier.ValueString()},
	}

	if data.NamespaceType.ValueString() == namespaceTypeServerless {
		input.NamespaceIdentifier = &awstypes.NamespaceIdentifierUnionMemberServerlessIdentifier{
			Value: awstypes.ServerlessIdentifier{
				NamespaceIdentifier: fwflex.StringFromFramework(ctx, data.ServerlessNamespaceIdentifier),
				WorkgroupIdentifier: fwflex.StringFromFramework(ctx, data.ServerlessWorkgroupIdentifier),
			},
		}
	} else {
		input.NamespaceIdentifier = &awstypes.NamespaceIdentifierUnionMemberProvisionedIdentifier{
			Value: awstypes.ProvisionedIdentifier{
				ClusterIdentifier: fwflex.StringFromFramework(ctx, data.ProvisionedClusterIdentifier),
			},
		}
	}

	_, err := tfresource.RetryWhenIsA[any, *awstypes.InvalidClusterStateFault](ctx, namespaceRegistrationInvalidClusterStateFaultTimeout,
		func(ctx context.Context) (any, error) {
			return conn.DeregisterNamespace(ctx, input)
		})
	if err != nil {
		if errs.IsA[*awstypes.ClusterNotFoundFault](err) || errs.IsA[*awstypes.InvalidNamespaceFault](err) {
			return
		}
		response.Diagnostics.AddError("deleting Redshift Namespace Registration", err.Error())
		return
	}
}

type namespaceRegistrationResourceModel struct {
	framework.WithRegionModel
	ConsumerIdentifier            types.String `tfsdk:"consumer_identifier"`
	NamespaceType                 types.String `tfsdk:"namespace_type"`
	ProvisionedClusterIdentifier  types.String `tfsdk:"provisioned_cluster_identifier"`
	ServerlessNamespaceIdentifier types.String `tfsdk:"serverless_namespace_identifier"`
	ServerlessWorkgroupIdentifier types.String `tfsdk:"serverless_workgroup_identifier"`
}

var (
	_ inttypes.ImportIDParser           = namespaceRegistrationImportID{}
	_ inttypes.FrameworkImportIDCreator = namespaceRegistrationImportID{}
)

type namespaceRegistrationImportID struct{}

func (namespaceRegistrationImportID) Parse(id string) (string, map[string]any, error) {
	parts := strings.Split(id, ",")

	// Try serverless format first: consumer_id,namespace_id,workgroup_id
	if len(parts) == 3 {
		result := map[string]any{
			"consumer_identifier":             parts[0],
			"namespace_type":                  namespaceTypeServerless,
			"serverless_namespace_identifier": parts[1],
			"serverless_workgroup_identifier": parts[2],
		}
		return id, result, nil
	}

	// Try provisioned format: consumer_id,cluster_id
	if len(parts) == 2 {
		result := map[string]any{
			"consumer_identifier":            parts[0],
			"namespace_type":                 namespaceTypeProvisioned,
			"provisioned_cluster_identifier": parts[1],
		}
		return id, result, nil
	}

	return "", nil, fmt.Errorf("id %q should be in format <consumer-id>,<cluster-id> or <consumer-id>,<namespace-id>,<workgroup-id>", id)
}

func (namespaceRegistrationImportID) Create(ctx context.Context, state tfsdk.State) string {
	var namespaceType types.String
	state.GetAttribute(ctx, path.Root("namespace_type"), &namespaceType)

	if namespaceType.ValueString() == "serverless" {
		parts := make([]string, 3)
		var attrVal types.String

		state.GetAttribute(ctx, path.Root("consumer_identifier"), &attrVal)
		parts[0] = attrVal.ValueString()

		state.GetAttribute(ctx, path.Root("serverless_namespace_identifier"), &attrVal)
		parts[1] = attrVal.ValueString()

		state.GetAttribute(ctx, path.Root("serverless_workgroup_identifier"), &attrVal)
		parts[2] = attrVal.ValueString()

		return strings.Join(parts, ",")
	}

	// Provisioned
	parts := make([]string, 2)
	var attrVal types.String

	state.GetAttribute(ctx, path.Root("consumer_identifier"), &attrVal)
	parts[0] = attrVal.ValueString()

	state.GetAttribute(ctx, path.Root("provisioned_cluster_identifier"), &attrVal)
	parts[1] = attrVal.ValueString()

	return strings.Join(parts, ",")
}
