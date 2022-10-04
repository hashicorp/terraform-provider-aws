package wafregional

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

func ResourceRegexPatternSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceRegexPatternSetCreate,
		Read:   resourceRegexPatternSetRead,
		Update: resourceRegexPatternSetUpdate,
		Delete: resourceRegexPatternSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"regex_pattern_strings": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceRegexPatternSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Creating WAF Regional Regex Pattern Set: %s", d.Get("name").(string))

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateRegexPatternSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}
		return conn.CreateRegexPatternSet(params)
	})
	if err != nil {
		return fmt.Errorf("failed creating WAF Regional Regex Pattern Set: %w", err)
	}
	resp := out.(*waf.CreateRegexPatternSetOutput)

	d.SetId(aws.StringValue(resp.RegexPatternSet.RegexPatternSetId))

	return resourceRegexPatternSetUpdate(d, meta)
}

func resourceRegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn

	log.Printf("[INFO] Reading WAF Regional Regex Pattern Set: %s", d.Get("name").(string))
	params := &waf.GetRegexPatternSetInput{
		RegexPatternSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetRegexPatternSet(params)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			log.Printf("[WARN] WAF Regional Regex Pattern Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting WAF Regional Regex Pattern Set (%s): %w", d.Id(), err)
	}

	d.Set("name", resp.RegexPatternSet.Name)
	d.Set("regex_pattern_strings", aws.StringValueSlice(resp.RegexPatternSet.RegexPatternStrings))

	return nil
}

func resourceRegexPatternSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Updating WAF Regional Regex Pattern Set: %s", d.Get("name").(string))

	if d.HasChange("regex_pattern_strings") {
		o, n := d.GetChange("regex_pattern_strings")
		oldPatterns, newPatterns := o.(*schema.Set).List(), n.(*schema.Set).List()
		err := updateRegexPatternSetPatternStringsWR(d.Id(), oldPatterns, newPatterns, conn, region)
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regional Regex Pattern Set(%s): %s", d.Id(), err)
		}
	}

	return resourceRegexPatternSetRead(d, meta)
}

func resourceRegexPatternSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	oldPatterns := d.Get("regex_pattern_strings").(*schema.Set).List()
	if len(oldPatterns) > 0 {
		noPatterns := []interface{}{}
		err := updateRegexPatternSetPatternStringsWR(d.Id(), oldPatterns, noPatterns, conn, region)
		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed updating WAF Regional Regex Pattern Set (%s): %w", d.Id(), err)
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(d.Id()),
		}
		return conn.DeleteRegexPatternSet(req)
	})
	if err != nil {
		return fmt.Errorf("failed deleting WAF Regional Regex Pattern Set (%s): %w", d.Id(), err)
	}

	return nil
}

func updateRegexPatternSetPatternStringsWR(id string, oldPatterns, newPatterns []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(id),
			Updates:           tfwaf.DiffRegexPatternSetPatternStrings(oldPatterns, newPatterns),
		}

		return conn.UpdateRegexPatternSet(req)
	})

	return err
}
