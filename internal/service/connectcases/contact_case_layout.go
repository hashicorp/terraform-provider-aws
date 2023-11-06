package connectcases

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connectcases"
	"github.com/aws/aws-sdk-go-v2/service/connectcases/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_connectcases_layout", name="Connect Cases Layout")
func ResourceLayout() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLayoutCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"layout_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"more_info": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sections": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"field_group": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"fields": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"id": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"top_panel": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sections": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"field_group": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"fields": {
																Type:     schema.TypeList,
																Required: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"id": {
																			Type:     schema.TypeString,
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceLayoutCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)
	log.Print("[DEBUG] Creating Connect Case Layout")

	name := d.Get("name").(string)

	params := &connectcases.CreateLayoutInput{
		Content:  expandLayoutContent(d.Get("content").([]interface{})),
		DomainId: aws.String(d.Get("domain_id").(string)),
		Name:     aws.String(name),
	}

	output, err := conn.CreateLayout(ctx, params)
	if err != nil {
		return diag.Errorf("creating Connect Case Layout (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.LayoutId))
	d.Set("layout_arn", aws.ToString(output.LayoutArn))

	return append(diags, resourceLayoutRead(ctx, d, meta)...)
}

func resourceLayoutRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	domainId := d.Get("domain_id").(string)
	output, err := FindLayoutById(ctx, conn, d.Id(), domainId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Case Layout %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Case Layout (%s): %s", d.Id(), err)
	}

	d.Set("name", output.Name)

	return diags
}

func expandLayoutContent(tfMap []interface{}) *types.LayoutContentMemberBasic {
	if tfMap == nil || tfMap[0] == nil {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.LayoutContentMemberBasic{}
	apiObject.Value.MoreInfo = expandLayoutContentSections(tfList["more_info"].([]interface{}))
	apiObject.Value.TopPanel = expandLayoutContentSections(tfList["top_panel"].([]interface{}))

	return apiObject
}

func expandLayoutContentSections(tfList []interface{}) *types.LayoutSections {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := &types.LayoutSections{}
	apiArray := make([]types.Section, 0, len(tfList))

	for i := 0; i < len(tfList); i++ {
		apiArray = append(apiArray, expandSectionFieldGroup(tfList[i].(map[string]interface{})["sections"].([]interface{})))
	}

	apiObject.Sections = apiArray

	return apiObject
}

func expandSectionFieldGroup(tfList []interface{}) *types.SectionMemberFieldGroup {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.SectionMemberFieldGroup{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Value.Name = aws.String(v)
	}

	if v, ok := tfMap["field_group"].([]interface{}); ok && len(v) > 0 {
		apiObject.Value.Fields = expandFieldGroupFields(v)
	}

	return apiObject
}

func expandFieldGroupFields(tfList []interface{}) []types.FieldItem {
	if len(tfList) == 0 {
		return nil
	}

	apiResult := make([]types.FieldItem, 0, len(tfList))

	for i := 0; i < len(tfList); i++ {
		field, ok := tfList[i].(map[string]interface{})["fields"].([]interface{})
		if !ok {
			return nil
		}

		if v, ok := field[0].(map[string]interface{}); ok && len(v) > 0 {
			apiObject := types.FieldItem{
				Id: aws.String(v["id"].(string)),
			}
			apiResult = append(apiResult, apiObject)
		}
	}

	return apiResult
}
