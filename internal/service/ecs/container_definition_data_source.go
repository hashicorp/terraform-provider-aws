// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecs_container_definition")
func DataSourceContainerDefinition() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContainerDefinitionRead,

		Schema: map[string]*schema.Schema{
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			// Computed values.
			"image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_digest": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory_reservation": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"disable_networking": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"docker_labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrEnvironment: {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceContainerDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}
	log.Printf("[DEBUG] Reading ECS Container Definition: %+v", params)
	desc, err := conn.DescribeTaskDefinition(ctx, params)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition: %s", err)
	}

	if desc == nil || desc.TaskDefinition == nil {
		return sdkdiag.AppendErrorf(diags, "reading ECS Task Definition: empty response")
	}

	taskDefinition := desc.TaskDefinition
	for _, def := range taskDefinition.ContainerDefinitions {
		if aws.ToString(def.Name) != d.Get("container_name").(string) {
			continue
		}

		d.SetId(fmt.Sprintf("%s/%s", aws.ToString(taskDefinition.TaskDefinitionArn), d.Get("container_name").(string)))
		d.Set("image", def.Image)
		image := aws.ToString(def.Image)
		if strings.Contains(image, ":") {
			d.Set("image_digest", strings.Split(image, ":")[1])
		}
		d.Set("cpu", def.Cpu)
		d.Set("memory", def.Memory)
		d.Set("memory_reservation", def.MemoryReservation)
		d.Set("disable_networking", def.DisableNetworking)
		d.Set("docker_labels", def.DockerLabels)

		var environment = map[string]string{}
		for _, keyValuePair := range def.Environment {
			environment[aws.ToString(keyValuePair.Name)] = aws.ToString(keyValuePair.Value)
		}
		d.Set(names.AttrEnvironment, environment)
	}

	if d.Id() == "" {
		return sdkdiag.AppendErrorf(diags, "container with name %q not found in task definition %q", d.Get("container_name").(string), d.Get("task_definition").(string))
	}

	return diags
}
