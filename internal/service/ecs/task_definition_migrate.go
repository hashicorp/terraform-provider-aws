// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceTaskDefinitionMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	ctx := context.Background()
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	switch v {
	case 0:
		log.Println("[INFO] Found AWS ECS Task Definition State v0; migrating to v1")
		return migrateTaskDefinitionStateV0toV1(ctx, is, conn)
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateTaskDefinitionStateV0toV1(ctx context.Context, is *terraform.InstanceState, conn *ecs.Client) (*terraform.InstanceState, error) {
	arn := is.Attributes[names.AttrARN]

	// We need to pull definitions from the API b/c they're unrecoverable from the checksum.
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(arn),
	}
	td, _, err := findTaskDefinition(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	s, err := tfjson.EncodeToString(td.ContainerDefinitions)

	if err != nil {
		return nil, err
	}

	is.Attributes["container_definitions"] = s

	return is, nil
}
