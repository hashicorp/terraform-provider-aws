// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecs_container_definition", name="Container Definition")
func dataSourceContainerDefinition() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContainerDefinitionRead,

		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cpu": {
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
			"image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_digest": {
				Type:     schema.TypeString,
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
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceContainerDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	taskDefinition := d.Get("task_definition").(string)
	def, err := findContainerDefinitionByTwoPartKey(ctx, conn, taskDefinition, d.Get("container_name").(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ECS Container Definition", err))
	}

	d.SetId(fmt.Sprintf("%s/%s", taskDefinition, aws.ToString(def.Name)))
	d.Set("cpu", def.Cpu)
	d.Set("disable_networking", def.DisableNetworking)
	d.Set("docker_labels", def.DockerLabels)
	var environment = map[string]string{}
	for _, v := range def.Environment {
		environment[aws.ToString(v.Name)] = aws.ToString(v.Value)
	}
	d.Set(names.AttrEnvironment, environment)
	image := aws.ToString(def.Image)
	d.Set("image", image)
	if strings.Contains(image, ":") {
		d.Set("image_digest", strings.Split(image, ":")[1])
	}
	d.Set("memory", def.Memory)
	d.Set("memory_reservation", def.MemoryReservation)

	return diags
}

func findContainerDefinitionByTwoPartKey(ctx context.Context, conn *ecs.Client, taskDefinitionName, containerName string) (*awstypes.ContainerDefinition, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	}

	taskDefinition, _, err := findTaskDefinition(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(taskDefinition.ContainerDefinitions, func(v awstypes.ContainerDefinition) bool {
		return aws.ToString(v.Name) == containerName
	}))
}
