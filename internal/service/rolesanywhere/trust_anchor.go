// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rolesanywhere

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere"
	"github.com/aws/aws-sdk-go-v2/service/rolesanywhere/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rolesanywhere_trust_anchor", name="Trust Anchor")
// @Tags(identifierAttribute="arn")
func ResourceTrustAnchor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTrustAnchorCreate,
		ReadWithoutTimeout:   resourceTrustAnchorRead,
		UpdateWithoutTimeout: resourceTrustAnchorUpdate,
		DeleteWithoutTimeout: resourceTrustAnchorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"source": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_data": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"acm_pca_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"x509_certificate_data": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"source_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(trustAnchorTypeValues(types.TrustAnchorType("").Values()...), false),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTrustAnchorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	name := d.Get("name").(string)
	input := &rolesanywhere.CreateTrustAnchorInput{
		Enabled: aws.Bool(d.Get("enabled").(bool)),
		Name:    aws.String(name),
		Source:  expandSource(d.Get("source").([]interface{})),
		Tags:    getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Creating RolesAnywhere Trust Anchor (%s): %#v", d.Id(), input)
	output, err := conn.CreateTrustAnchor(ctx, input)

	if err != nil {
		return diag.Errorf("creating RolesAnywhere Trust Anchor (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.TrustAnchor.TrustAnchorId))

	return resourceTrustAnchorRead(ctx, d, meta)
}

func resourceTrustAnchorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	trustAnchor, err := FindTrustAnchorByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RolesAnywhere Trust Anchor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
	}

	d.Set("arn", trustAnchor.TrustAnchorArn)
	d.Set("enabled", trustAnchor.Enabled)
	d.Set("name", trustAnchor.Name)

	if err := d.Set("source", flattenSource(trustAnchor.Source)); err != nil {
		return diag.Errorf("setting source: %s", err)
	}

	return nil
}

func resourceTrustAnchorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &rolesanywhere.UpdateTrustAnchorInput{
			TrustAnchorId: aws.String(d.Id()),
			Name:          aws.String(d.Get("name").(string)),
			Source:        expandSource(d.Get("source").([]interface{})),
		}

		log.Printf("[DEBUG] Updating RolesAnywhere Trust Anchor (%s): %#v", d.Id(), input)
		_, err := conn.UpdateTrustAnchor(ctx, input)

		if err != nil {
			return diag.Errorf("updating RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
		}

		if d.HasChange("enabled") {
			_, n := d.GetChange("enabled")
			if n == true {
				if err := enableTrustAnchor(ctx, d.Id(), meta); err != nil {
					diag.Errorf("enabling RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
				}
			} else {
				if err := disableTrustAnchor(ctx, d.Id(), meta); err != nil {
					diag.Errorf("disabling RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
				}
			}
		}
	}

	return resourceTrustAnchorRead(ctx, d, meta)
}

func resourceTrustAnchorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	log.Printf("[DEBUG] Deleting RolesAnywhere Trust Anchor (%s)", d.Id())
	_, err := conn.DeleteTrustAnchor(ctx, &rolesanywhere.DeleteTrustAnchorInput{
		TrustAnchorId: aws.String(d.Id()),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting RolesAnywhere Trust Anchor (%s): %s", d.Id(), err)
	}

	return nil
}

func flattenSource(apiObject *types.Source) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["source_type"] = apiObject.SourceType
	m["source_data"] = flattenSourceData(apiObject.SourceData)

	return []interface{}{m}
}

func flattenSourceData(apiObject types.SourceData) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	switch v := apiObject.(type) {
	case *types.SourceDataMemberAcmPcaArn:
		m["acm_pca_arn"] = v.Value
	case *types.SourceDataMemberX509CertificateData:
		m["x509_certificate_data"] = v.Value
	case *types.UnknownUnionMember:
		log.Println("unknown tag:", v.Tag)
	default:
		log.Println("union is nil or unknown type")
	}

	return []interface{}{m}
}

func expandSource(tfList []interface{}) *types.Source {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Source{}

	if v, ok := tfMap["source_type"].(string); ok && v != "" {
		result.SourceType = types.TrustAnchorType(v)
	}

	if v, ok := tfMap["source_data"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if result.SourceType == types.TrustAnchorTypeAwsAcmPca {
			result.SourceData = expandSourceDataACMPCA(v[0].(map[string]interface{}))
		} else if result.SourceType == types.TrustAnchorTypeCertificateBundle {
			result.SourceData = expandSourceDataCertificateBundle(v[0].(map[string]interface{}))
		}
	}

	return result
}

func expandSourceDataACMPCA(tfMap map[string]interface{}) *types.SourceDataMemberAcmPcaArn {
	result := &types.SourceDataMemberAcmPcaArn{}

	if v, ok := tfMap["acm_pca_arn"].(string); ok && v != "" {
		result.Value = v
	}

	return result
}

func expandSourceDataCertificateBundle(tfMap map[string]interface{}) *types.SourceDataMemberX509CertificateData {
	result := &types.SourceDataMemberX509CertificateData{}

	if v, ok := tfMap["x509_certificate_data"].(string); ok && v != "" {
		result.Value = v
	}

	return result
}

func disableTrustAnchor(ctx context.Context, trustAnchorId string, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.DisableTrustAnchorInput{
		TrustAnchorId: aws.String(trustAnchorId),
	}

	_, err := conn.DisableTrustAnchor(ctx, input)
	return err
}

func enableTrustAnchor(ctx context.Context, trustAnchorId string, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RolesAnywhereClient(ctx)

	input := &rolesanywhere.EnableTrustAnchorInput{
		TrustAnchorId: aws.String(trustAnchorId),
	}

	_, err := conn.EnableTrustAnchor(ctx, input)
	return err
}

func trustAnchorTypeValues(input ...types.TrustAnchorType) []string {
	var output []string

	for _, v := range input {
		output = append(output, string(v))
	}

	return output
}
