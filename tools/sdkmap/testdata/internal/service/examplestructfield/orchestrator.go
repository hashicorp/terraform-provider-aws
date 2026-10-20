// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplestructfield

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/examplestructfield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplestructfield/types"
)

type deploymentOrchestrator struct {
	conn          *examplestructfield.Client
	deploymentIDs []string
}

func newDeploymentOrchestrator(conn *examplestructfield.Client) *deploymentOrchestrator {
	return &deploymentOrchestrator{
		conn: conn,
	}
}

func (o *deploymentOrchestrator) CleanUp(ctx context.Context) {
	for _, id := range o.deploymentIDs {
		input := examplestructfield.DeleteDeploymentInput{
			DeploymentId: aws.String(id),
		}

		_, _ = o.conn.DeleteDeployment(ctx, &input)
	}
}

func (o *deploymentOrchestrator) CreateDeployment(ctx context.Context, input *examplestructfield.CreateDeploymentInput) (*awstypes.Deployment, error) {
	output, err := o.conn.CreateDeployment(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("creating Deployment: %w", err)
	}

	o.deploymentIDs = append(o.deploymentIDs, aws.ToString(output.Deployment.DeploymentId))

	return output.Deployment, nil
}

func (o *deploymentOrchestrator) Switchover(ctx context.Context, deploymentID string) error {
	input := examplestructfield.SwitchoverDeploymentInput{
		DeploymentId: aws.String(deploymentID),
	}

	if _, err := o.conn.SwitchoverDeployment(ctx, &input); err != nil {
		return fmt.Errorf("switching over Deployment (%s): %w", deploymentID, err)
	}

	return nil
}
