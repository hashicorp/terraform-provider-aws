package wafregional

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

func ResourceXSSMatchSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceXSSMatchSetCreate,
		Read:   resourceXSSMatchSetRead,
		Update: resourceXSSMatchSetUpdate,
		Delete: resourceXSSMatchSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"xss_match_tuple": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_to_match": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(wafregional.MatchFieldType_Values(), false),
									},
								},
							},
						},
						"text_transformation": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(wafregional.TextTransformation_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceXSSMatchSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Creating regional WAF XSS Match Set: %s", d.Get("name").(string))

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateXssMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}

		return conn.CreateXssMatchSet(params)
	})
	if err != nil {
		return fmt.Errorf("Failed creating regional WAF XSS Match Set: %w", err)
	}
	resp := out.(*waf.CreateXssMatchSetOutput)

	d.SetId(aws.StringValue(resp.XssMatchSet.XssMatchSetId))

	if v, ok := d.Get("xss_match_tuple").(*schema.Set); ok && v.Len() > 0 {
		err := updateXSSMatchSetResourceWR(d.Id(), nil, v.List(), conn, region)
		if err != nil {
			return fmt.Errorf("Failed updating regional WAF XSS Match Set: %w", err)
		}
	}

	return resourceXSSMatchSetRead(d, meta)
}

func resourceXSSMatchSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	log.Printf("[INFO] Reading regional WAF XSS Match Set: %s", d.Get("name").(string))
	params := &waf.GetXssMatchSetInput{
		XssMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetXssMatchSet(params)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			log.Printf("[WARN] Regional WAF XSS Match Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	set := resp.XssMatchSet

	if err := d.Set("xss_match_tuple", flattenXSSMatchTuples(set.XssMatchTuples)); err != nil {
		return fmt.Errorf("error setting xss_match_tuple: %w", err)
	}
	d.Set("name", set.Name)

	return nil
}

func resourceXSSMatchSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("xss_match_tuple") {
		o, n := d.GetChange("xss_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateXSSMatchSetResourceWR(d.Id(), oldT, newT, conn, region)
		if err != nil {
			return fmt.Errorf("Failed updating regional WAF XSS Match Set: %w", err)
		}
	}

	return resourceXSSMatchSetRead(d, meta)
}

func resourceXSSMatchSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	if v, ok := d.GetOk("xss_match_tuple"); ok {
		oldTuples := v.(*schema.Set).List()
		if len(oldTuples) > 0 {
			noTuples := []interface{}{}
			err := updateXSSMatchSetResourceWR(d.Id(), oldTuples, noTuples, conn, region)
			if err != nil {
				return fmt.Errorf("Error updating regional WAF XSS Match Set: %w", err)
			}
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteXssMatchSet(req)
	})
	if err != nil {
		return fmt.Errorf("Failed deleting regional WAF XSS Match Set: %w", err)
	}

	return nil
}

func updateXSSMatchSetResourceWR(id string, oldT, newT []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(id),
			Updates:       diffXSSMatchSetTuples(oldT, newT),
		}

		log.Printf("[INFO] Updating XSS Match Set tuples: %s", req)
		return conn.UpdateXssMatchSet(req)
	})
	if err != nil {
		return fmt.Errorf("Failed updating regional WAF XSS Match Set: %w", err)
	}

	return nil
}

func flattenXSSMatchTuples(ts []*waf.XssMatchTuple) []interface{} {
	out := make([]interface{}, len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m["field_to_match"] = tfwaf.FlattenFieldToMatch(t.FieldToMatch)
		m["text_transformation"] = aws.StringValue(t.TextTransformation)
		out[i] = m
	}
	return out
}

func diffXSSMatchSetTuples(oldT, newT []interface{}) []*waf.XssMatchSetUpdate {
	updates := make([]*waf.XssMatchSetUpdate, 0)

	for _, od := range oldT {
		tuple := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, &waf.XssMatchSetUpdate{
			Action: aws.String(waf.ChangeActionDelete),
			XssMatchTuple: &waf.XssMatchTuple{
				FieldToMatch:       tfwaf.ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: aws.String(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nd := range newT {
		tuple := nd.(map[string]interface{})

		updates = append(updates, &waf.XssMatchSetUpdate{
			Action: aws.String(waf.ChangeActionInsert),
			XssMatchTuple: &waf.XssMatchTuple{
				FieldToMatch:       tfwaf.ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: aws.String(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
