// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_glue_script")
func DataSourceScript() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceScriptRead,
		Schema: map[string]*schema.Schema{
			"dag_edge": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSource: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrTarget: {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_parameter": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"dag_node": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"args": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									"param": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Required: true,
						},
						"line_number": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"node_type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"language": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.LanguagePython,
				ValidateDiagFunc: enum.Validate[awstypes.Language](),
			},
			"python_script": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"scala_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScriptRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	dagEdge := d.Get("dag_edge").([]interface{})
	dagNode := d.Get("dag_node").([]interface{})

	input := &glue.CreateScriptInput{
		DagEdges: expandCodeGenEdges(dagEdge),
		DagNodes: expandCodeGenNodes(dagNode),
	}

	if v, ok := d.GetOk("language"); ok {
		input.Language = awstypes.Language(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Script: %+v", input)
	output, err := conn.CreateScript(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue script: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "script not created")
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("python_script", output.PythonScript)
	d.Set("scala_code", output.ScalaCode)

	return diags
}

func expandCodeGenNodeArgs(l []interface{}) []awstypes.CodeGenNodeArg {
	args := []awstypes.CodeGenNodeArg{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})
		arg := awstypes.CodeGenNodeArg{
			Name:  aws.String(m[names.AttrName].(string)),
			Param: m["param"].(bool),
			Value: aws.String(m[names.AttrValue].(string)),
		}
		args = append(args, arg)
	}

	return args
}

func expandCodeGenEdges(l []interface{}) []awstypes.CodeGenEdge {
	edges := []awstypes.CodeGenEdge{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})
		edge := awstypes.CodeGenEdge{
			Source: aws.String(m[names.AttrSource].(string)),
			Target: aws.String(m[names.AttrTarget].(string)),
		}
		if v, ok := m["target_parameter"]; ok && v.(string) != "" {
			edge.TargetParameter = aws.String(v.(string))
		}
		edges = append(edges, edge)
	}

	return edges
}

func expandCodeGenNodes(l []interface{}) []awstypes.CodeGenNode {
	nodes := []awstypes.CodeGenNode{}

	for _, mRaw := range l {
		m := mRaw.(map[string]interface{})
		node := awstypes.CodeGenNode{
			Args:     expandCodeGenNodeArgs(m["args"].([]interface{})),
			Id:       aws.String(m[names.AttrID].(string)),
			NodeType: aws.String(m["node_type"].(string)),
		}
		if v, ok := m["line_number"]; ok && v.(int) != 0 {
			node.LineNumber = int32(v.(int))
		}
		nodes = append(nodes, node)
	}

	return nodes
}
