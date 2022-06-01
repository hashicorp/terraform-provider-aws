package waf

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
	conn := meta.(*conns.AWSClient).WAFConn

	log.Printf("[INFO] Creating WAF Regex Pattern Set: %s", d.Get("name").(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateRegexPatternSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}
		return conn.CreateRegexPatternSet(params)
	})
	if err != nil {
		return fmt.Errorf("Failed creating WAF Regex Pattern Set: %s", err)
	}
	resp := out.(*waf.CreateRegexPatternSetOutput)

	d.SetId(aws.StringValue(resp.RegexPatternSet.RegexPatternSetId))

	return resourceRegexPatternSetUpdate(d, meta)
}

func resourceRegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFConn
	log.Printf("[INFO] Reading WAF Regex Pattern Set: %s", d.Get("name").(string))
	params := &waf.GetRegexPatternSetInput{
		RegexPatternSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetRegexPatternSet(params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			log.Printf("[WARN] WAF Regex Pattern Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", resp.RegexPatternSet.Name)
	d.Set("regex_pattern_strings", aws.StringValueSlice(resp.RegexPatternSet.RegexPatternStrings))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("regexpatternset/%s", d.Id()),
	}
	d.Set("arn", arn.String())

	return nil
}

func resourceRegexPatternSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFConn

	log.Printf("[INFO] Updating WAF Regex Pattern Set: %s", d.Get("name").(string))

	if d.HasChange("regex_pattern_strings") {
		o, n := d.GetChange("regex_pattern_strings")
		oldPatterns, newPatterns := o.(*schema.Set).List(), n.(*schema.Set).List()
		err := updateRegexPatternSetPatternStrings(d.Id(), oldPatterns, newPatterns, conn)
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regex Pattern Set: %s", err)
		}
	}

	return resourceRegexPatternSetRead(d, meta)
}

func resourceRegexPatternSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFConn

	oldPatterns := d.Get("regex_pattern_strings").(*schema.Set).List()
	if len(oldPatterns) > 0 {
		noPatterns := []interface{}{}
		err := updateRegexPatternSetPatternStrings(d.Id(), oldPatterns, noPatterns, conn)
		if err != nil {
			return fmt.Errorf("Error updating WAF Regex Pattern Set: %s", err)
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Regex Pattern Set: %s", req)
		return conn.DeleteRegexPatternSet(req)
	})
	if err != nil {
		return fmt.Errorf("Failed deleting WAF Regex Pattern Set: %s", err)
	}

	return nil
}

func updateRegexPatternSetPatternStrings(id string, oldPatterns, newPatterns []interface{}, conn *waf.WAF) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(id),
			Updates:           DiffRegexPatternSetPatternStrings(oldPatterns, newPatterns),
		}

		return conn.UpdateRegexPatternSet(req)
	})
	if err != nil {
		return fmt.Errorf("Failed updating WAF Regex Pattern Set: %s", err)
	}

	return nil
}
