package shield

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"log"
	"time"
)

func ResourceAdvancedAutomaticLayerProtection() *schema.Resource {
	return &schema.Resource{
		Create: resourceAdvancedAutomaticLayerProtectionCreate,
		Update: resourceAdvancedAutomaticLayerProtectionUpdate,
		Read:   resourceAdvancedAutomaticLayerProtectionRead,
		Delete: resourceAdvancedAutomaticLayerProtectionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"block", "count"}, false),
			},
		},
	}
}

func resourceAdvancedAutomaticLayerProtectionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	if !d.HasChange("action") {
		return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
	}

	action := &shield.ResponseAction{}
	switch d.Get("action").(string) {
	case "block":
		action.Block = &shield.BlockAction{}
	case "count":
		action.Count = &shield.CountAction{}
	}

	input := &shield.UpdateApplicationLayerAutomaticResponseInput{
		Action:      action,
		ResourceArn: aws.String(d.Id()),
	}

	_, err := conn.UpdateApplicationLayerAutomaticResponse(input)
	if err != nil {
		return fmt.Errorf("error updating Application Layer Automatic Protection: %s", err)
	}

	return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	action := &shield.ResponseAction{}
	switch d.Get("action").(string) {
	case "block":
		action.Block = &shield.BlockAction{}
	case "count":
		action.Count = &shield.CountAction{}
	}

	enableAutomaticResponseInput := &shield.EnableApplicationLayerAutomaticResponseInput{
		Action:      action,
		ResourceArn: aws.String(d.Get("resource_arn").(string)),
	}

	_, err := conn.EnableApplicationLayerAutomaticResponse(enableAutomaticResponseInput)
	if err != nil {
		return fmt.Errorf("error creating Application Layer Automatic Protection: %s", err)
	}

	d.SetId(d.Get("resource_arn").(string))
	return resourceAdvancedAutomaticLayerProtectionRead(d, meta)
}

func resourceAdvancedAutomaticLayerProtectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ShieldConn

	input := &shield.DescribeProtectionInput{
		ResourceArn: aws.String(d.Id()),
	}

	resp, err := conn.DescribeProtection(input)

	if tfawserr.ErrCodeEquals(err, shield.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Shield Protection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Application Layer Automatic Protection (%s): %s", d.Id(), err)
	}

	var action string
	if resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action != nil {
		if resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Block != nil {
			action = "block"
		}
		if resp.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Count != nil {
			action = "count"
		}
	}

	d.Set("action", action)

	return nil
}

func resourceAdvancedAutomaticLayerProtectionDelete(d *schema.ResourceData, meta interface{}) error {
	shieldConn := meta.(*conns.AWSClient).ShieldConn

	/*
		cloudfrontConn := meta.(*conns.AWSClient).CloudFrontConn

		splitResourceArn := strings.Split(d.Get("resource_arn").(string), "/")
		cloudfrontID := splitResourceArn[len(splitResourceArn)-1]


			cloudfrontInformation, err := cloudfrontConn.GetDistribution(&cloudfront.GetDistributionInput{
				Id: aws.String(cloudfrontID),
			})
			if err != nil {
				return fmt.Errorf("error describing Protection (%s): %s", d.Get("resource_arn").(string), err)
			}

				webAclARN, err := arn.Parse(aws.StringValue(cloudfrontInformation.Distribution.DistributionConfig.WebACLId))
				if err != nil {
					return fmt.Errorf("unable to parse WebACLId (%s): %s",
						aws.StringValue(cloudfrontInformation.Distribution.DistributionConfig.WebACLId),
						err,
					)
				}
	*/
	input := &shield.DisableApplicationLayerAutomaticResponseInput{
		ResourceArn: aws.String(d.Id()),
	}

	_, err := shieldConn.DisableApplicationLayerAutomaticResponse(input)
	if err != nil {
		return fmt.Errorf("error deleting Application Layer Automatic Protection (%s): %s", d.Id(), err)
	}

	time.Sleep(15 * time.Second)
	/*
		if err := AdvancedAutomaticLayerProtectionWaitUntilDeAssociated(
			webAclARN,
			meta); err != nil {
			return fmt.Errorf("error waiting for shield advanced waf rule de-association (%s): %s", d.Id(), err)
		}
	*/

	return nil
}

/*
// resourceWAFStateRefreshFunc blocks until AWS has removed AWS Shield WAF rule from webACL
// before considering AdvancedAutomaticLayerProtection removed. This is to avoid WAFOptimisticLockException that
// happens when trying to remove WAF resource immediately after AdvancedAutomaticLayerProtection.
func AdvancedAutomaticLayerProtectionWaitUntilDeAssociated(webACLArn arn.ARN, meta interface{}) error {
	stateConf := &resource.StateChangeConf{
		Refresh:    resourceWAFStateRefreshFunc(webACLArn, meta),
		Pending:    []string{"Associated"},
		Target:     []string{"NotAssociated"},
		Timeout:    30 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func resourceWAFStateRefreshFunc(webACLArn arn.ARN, meta interface{}) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		conn := meta.(*conns.AWSClient).WAFV2Conn

		splitWebACLArn := strings.Split(webACLArn.Resource, "/")
		webACLId := splitWebACLArn[len(splitWebACLArn)-1]
		webACLName := splitWebACLArn[len(splitWebACLArn)-2]

		params := &wafv2.GetWebACLInput{
			Id:    aws.String(webACLId),
			Name:  aws.String(webACLName),
			Scope: aws.String("CLOUDFRONT"),
		}

		resp, err := conn.GetWebACL(params)
		if err != nil {
			log.Printf("[WARN] Error retrieving WEBACL %q details: %s", webACLId, err)
			return nil, "", err
		}

		if resp == nil {
			return nil, "", nil
		}

		//TODO Implement regexp match
		for _, rule := range resp.WebACL.Rules {

		}

		return resp.WebACL, "NotAssociated", nil
	}
}
*/
