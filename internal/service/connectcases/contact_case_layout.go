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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_connectcases_contact_case_layout", name="Connect Cases Contact Case Layout")
func ResourceContactCaseLayout() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactCaseLayoutCreate,
		ReadWithoutTimeout:   resourceContactCaseDomainRead,
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
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
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
																			Type:     schema.TypeList,
																			Optional: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
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
																			Type:     schema.TypeList,
																			Optional: true,
																			Elem:     &schema.Schema{Type: schema.TypeString},
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
							},
						},
					},
				},
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceContactCaseLayoutCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)
	log.Print("[DEBUG] Creating Connect Case Layout")

	name := d.Get("name").(string)

	params := &connectcases.CreateLayoutInput{
		Content:  expandContactCaseLayoutContent(d.Get("content").([]interface{})[0].(map[string]interface{})),
		DomainId: aws.String(d.Get("domain_id").(string)),
		Name:     aws.String(name),
	}

	output, err := conn.CreateLayout(ctx, params)
	if err != nil {
		return diag.Errorf("creating Connect Case Layout (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.LayoutId))

	return append(diags, resourceContactCaseLayoutRead(ctx, d, meta)...)
}

func resourceContactCaseLayoutRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	output, err := FindConnectCasesDomainById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Case Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Case Domain (%s): %s", d.Id(), err)
	}

	d.Set("name", output.Name)
	d.Set("domain_arn", output.DomainArn)
	d.Set("domain_id", output.DomainId)
	d.Set("domain_status", output.DomainStatus)

	return diags
}

func expandContactCaseLayoutContent(tfMap map[string]interface{}) *types.LayoutContentMemberBasic {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.LayoutContentMemberBasic{}

	apiObject.Value.TopPanel.Sections = expandLayoutContentSections(tfMap["top_panel"].([]interface{}))
	apiObject.Value.MoreInfo.Sections = expandLayoutContentSections(tfMap["more_info"].([]interface{}))

	return apiObject
}

func expandLayoutContentSections(tfList []interface{}) []types.Section {
	if len(tfList) == 0 && tfList[0] == nil {
		return nil
	}

	apiObject := []types.Section{}

	for i := 0; i < len(tfList); i++ {
		apiObject = append(apiObject, expandSectionFieldGroup(tfList[i].([]interface{})))
	}

	return apiObject
}

func expandSectionFieldGroup(tfList []interface{}) *types.SectionMemberFieldGroup {
	if len(tfList) == 0 && tfList[0] == nil {
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

	if v, ok := tfMap["fields"].([]interface{}); ok && len(v) > 0 {
		apiObject.Value.Fields = expandFieldGroupFields(v)
	}

	return apiObject
}

func expandFieldGroupFields(tfList []interface{}) []types.FieldItem {
	if len(tfList) == 0 && tfList[0] == nil {
		return nil
	}

	apiObject := []types.FieldItem{}

	for i := 0; i < len(tfList); i++ {
		object := tfList[i].(map[string]interface{})

		if v, ok := object["id"].(string); ok && len(v) > 0 {
			apiObject = append(apiObject, types.FieldItem{
				Id: aws.String(v),
			})
		}
	}

	return apiObject
}
