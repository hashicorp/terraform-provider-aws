// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsrds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwactions "github.com/hashicorp/terraform-provider-aws/internal/framework/actions"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_rds_cluster_instance_reboot, name="ClusterInstanceReboot")
func newClusterInstanceRebootAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &clusterInstanceRebootAction{}, nil
}

var (
	_ action.Action = (*clusterInstanceRebootAction)(nil)
)

type clusterInstanceRebootAction struct {
	framework.ActionWithModel[clusterInstanceRebootActionModel]
}

type clusterInstanceRebootActionModel struct {
	framework.WithRegionModel
	Id types.String `tfsdk:"id"`
}

func (a *clusterInstanceRebootAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Initiates a reboot of an RDS Cluster Instance.",
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Description: "The DB instance identifier. This can be either the instance name or the ARN.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (a *clusterInstanceRebootAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config clusterInstanceRebootActionModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()

	tflog.Info(ctx, "Starting RDS cluster instance reboot action", map[string]any{
		names.AttrID: id,
	})

	// Send initial progress update to the Terraform CLI UI
	cb := fwactions.NewSendProgressFunc(resp)
	cb(ctx, "Initiating reboot for RDS Cluster Instance %s...", id)

	// Get AWS client from the internal framework meta
	// Note: Depending on the current provider version, this might be RDSClientV2(ctx).
	// The compiler will tell us if we need to adjust this.
	conn := a.Meta().RDSClient(ctx)

	input := &awsrds.RebootDBInstanceInput{
		DBInstanceIdentifier: aws.String(id),
	}

	_, err := conn.RebootDBInstance(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Reboot RDS Cluster Instance",
			fmt.Sprintf("Could not reboot instance %s: %s", id, err),
		)
		return
	}

	cb(ctx, "Successfully initiated reboot for RDS Cluster Instance %s", id)

	tflog.Info(ctx, "RDS cluster instance reboot action completed successfully", map[string]any{
		names.AttrID: id,
	})
}
