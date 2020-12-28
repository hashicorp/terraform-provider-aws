package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsSesIdentityFeedbackForwardingEnabled() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSesIdentityFeedbackForwardingEnabledSet,
		Read:   resourceAwsSesIdentityFeedbackForwardingEnabledRead,
		Update: resourceAwsSesIdentityFeedbackForwardingEnabledSet,
		Delete: resourceAwsSesIdentityFeedbackForwardingEnabledDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceAwsSesIdentityFeedbackForwardingEnabledSet(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("resourceAwsSesIdentityFeedbackForwardingEnabledSet")
	conn := meta.(*AWSClient).sesconn

	identity := d.Get("identity").(string)
	enabled := d.Get("enabled").(bool)

	input := &ses.SetIdentityFeedbackForwardingEnabledInput{
		Identity:          aws.String(identity),
		ForwardingEnabled: aws.Bool(enabled),
	}
	fmt.Printf("input: %v\n", input)
	res, err := conn.SetIdentityFeedbackForwardingEnabled(input)
	if err != nil {
		return fmt.Errorf("Error setting Feedback Forwarding identity: %s", err)
	}
	fmt.Printf("res: %v\n", res)
	fmt.Printf("res: %v\n", res.GoString())
	fmt.Printf("res: %s\n", res.String())

	d.SetId(identity)

	return resourceAwsSesIdentityFeedbackForwardingEnabledRead(d, meta)
}

func resourceAwsSesIdentityFeedbackForwardingEnabledRead(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("resourceAwsSesIdentityFeedbackForwardingEnabledRead")
	//conn := meta.(*AWSClient).sesconn
	//
	//email := d.Id()
	//d.Set("identity", domain)

	//readOpts := &ses.GetIdentityVerificationAttributesInput{
	//	Identities: []*string{
	//		aws.String(email),
	//	},
	//}
	//
	//response, err := conn.GetIdentityVerificationAttributes(readOpts)
	//if err != nil {
	//	log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
	//	return err
	//}

	//_, ok := response.VerificationAttributes[email]
	//if !ok {
	//	log.Printf("[WARN] Email not listed in response when fetching verification attributes for %s", d.Id())
	//	d.SetId("")
	//	return nil
	//}

	//arn := arn.ARN{
	//	AccountID: meta.(*AWSClient).accountid,
	//	Partition: meta.(*AWSClient).partition,
	//	Region:    meta.(*AWSClient).region,
	//	Resource:  fmt.Sprintf("identity/%s", d.Id()),
	//	Service:   "ses",
	//}.String()
	//d.Set("arn", arn)
	//d.Set("")
	return nil
}

func resourceAwsSesIdentityFeedbackForwardingEnabledDelete(d *schema.ResourceData, meta interface{}) error {
	fmt.Println("resourceAwsSesIdentityFeedbackForwardingEnabledDelete")
	conn := meta.(*AWSClient).sesconn
	identity := d.Id()

	input := &ses.SetIdentityFeedbackForwardingEnabledInput{
		Identity:          aws.String(identity),
		ForwardingEnabled: aws.Bool(true),
	}
	fmt.Printf("input: %v\n", input)
	res, err := conn.SetIdentityFeedbackForwardingEnabled(input)
	if err != nil {
		return err
	}
	fmt.Printf("res: %v\n", res.String())
	return nil
}
