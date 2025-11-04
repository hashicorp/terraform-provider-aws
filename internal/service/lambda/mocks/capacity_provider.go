// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mocks

import (
	"context"

	"github.com/hashicorp/terraform-provider-aws/internal/service/lambda/mocks/types"
)

type Client struct{}

func LambdaClient(ctx context.Context) *Client {
	return &Client{}
}

func (c *Client) CreateCapacityProvider(ctx context.Context, input *types.CreateCapacityProviderInput) (*types.CreateCapacityProviderOutput, error) {
	return &types.CreateCapacityProviderOutput{}, nil
}

func (c *Client) DeleteCapacityProvider(ctx context.Context, input *types.DeleteCapacityProviderInput) (*types.DeleteCapacityProviderOutput, error) {
	return &types.DeleteCapacityProviderOutput{}, nil
}

func (c *Client) UpdateCapacityProvider(ctx context.Context, input *types.UpdateCapacityProviderInput) (*types.UpdateCapacityProviderOutput, error) {
	return &types.UpdateCapacityProviderOutput{}, nil
}

func (c *Client) GetCapacityProvider(ctx context.Context, input *types.GetCapacityProviderInput) (*types.GetCapacityProviderOutput, error) {
	return &types.GetCapacityProviderOutput{}, nil
}

func (c *Client) ListCapacityProviders(ctx context.Context, input *types.ListCapacityProvidersInput) (*types.ListCapacityProvidersOutput, error) {
	return &types.ListCapacityProvidersOutput{}, nil
}
