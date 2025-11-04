// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

type CreateCapacityProviderInput struct {
	capacityProviderMock
}

type CreateCapacityProviderOutput struct {
	capacityProviderMock
	CapacityProviderArn string
	State               State
}

type capacityProviderMock struct {
	Name      string
	VpcConfig struct {
		SecurityGroupIds []string
		SubnetIds        []string
	}
	Tags map[string]string
}

type UpdateCapacityProviderInput struct {
	CapacityProviderName string
}

type UpdateCapacityProviderOutput struct {
	capacityProviderMock
	CapacityProviderArn string
	State               State
}

type DeleteCapacityProviderInput struct {
	CapacityProviderName string
}

type DeleteCapacityProviderOutput struct{}

type GetCapacityProviderInput struct {
	CapacityProviderName *string
}

type GetCapacityProviderOutput struct {
	State State
}

type ListCapacityProvidersInput struct{}

type ListCapacityProvidersOutput struct{}

type State string

const (
	StatePending  State = "Pending"
	StateActive   State = "Active"
	StateDeleting State = "Deleting"
)

func (s State) Values() []State {
	return []State{
		"Pending",
		"Active",
		"Deleting",
	}
}
