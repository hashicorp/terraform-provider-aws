package aws

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/terraform/helper/schema"
)

var SupportedPlatforms = map[string]bool{
	//"ADM":          true,  // (Amazon Device Messaging)
	"APNS":         true,  // (Apple Push Notification Service)
	"APNS_SANDBOX": true,  // (Apple Push Notification Service)
	"GCM":          false, // (Google Cloud Messaging).
}

// Mutable attributes
// http://docs.aws.amazon.com/sns/latest/api/API_SetPlatformApplicationAttributes.html
var SNSPlatformAppAttributeMap = map[string]string{
	"event_delivery_failure_topic_arn": "EventDeliveryFailure",
	"event_endpoint_created_topic_arn": "EventEndpointCreated",
	"event_endpoint_deleted_topic_arn": "EventEndpointDeleted",
	"event_endpoint_updated_topic_arn": "EventEndpointUpdated",
	"failure_feedback_role_arn":        "FailureFeedbackRoleArn",
	"platform_principal":               "PlatformPrincipal",
	"success_feedback_role_arn":        "SuccessFeedbackRoleArn",
	"success_feedback_sample_rate":     "SuccessFeedbackSampleRate",
}

func resourceAwsSnsApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSnsApplicationCreate,
		Read:   resourceAwsSnsApplicationRead,
		Update: resourceAwsSnsApplicationUpdate,
		Delete: resourceAwsSnsApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"platform_credential": {
				Type:      schema.TypeString,
				Required:  true,
				StateFunc: hashSum,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"event_delivery_failure_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_endpoint_created_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_endpoint_deleted_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_endpoint_updated_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"failure_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"platform_principal": {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: hashSum,
			},
			"success_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"success_feedback_sample_rate": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsSnsApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn

	attributes := make(map[string]*string)
	name := d.Get("name").(string)
	platform := d.Get("platform").(string)
	principal := d.Get("platform_principal").(string)

	attributes["PlatformCredential"] = aws.String(d.Get("platform_credential").(string))

	if _, ok := SupportedPlatforms[platform]; !ok {
		return errors.New(fmt.Sprintf("Platform %s is not supported", platform))
	}

	if value, _ := SupportedPlatforms[platform]; value {
		if principal == "" {
			return errors.New(fmt.Sprintf("Principal is required for %s", platform))
		} else {
			attributes["PlatformPrincipal"] = aws.String(principal)
		}
	}

	log.Printf("[DEBUG] SNS create application: %s", name)

	req := &sns.CreatePlatformApplicationInput{
		Name:       aws.String(name),
		Platform:   aws.String(platform),
		Attributes: attributes,
	}

	output, err := snsconn.CreatePlatformApplication(req)
	if err != nil {
		return fmt.Errorf("Error creating SNS application: %s", err)
	}

	d.SetId(*output.PlatformApplicationArn)

	return resourceAwsSnsApplicationUpdate(d, meta)
}

func resourceAwsSnsApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn

	resource := *resourceAwsSnsApplication()

	attributes := make(map[string]*string)

	for k, _ := range resource.Schema {
		if attrKey, ok := SNSPlatformAppAttributeMap[k]; ok {
			if d.HasChange(k) {
				log.Printf("[DEBUG] Updating %s", attrKey)
				_, n := d.GetChange(k)
				attributes[attrKey] = aws.String(n.(string))
			}
		}
	}

	if d.HasChange("platform_credential") {
		attributes["PlatformCredential"] = aws.String(d.Get("platform_credential").(string))
		// If the platform requires a principal it must also be specified, even if it didn't change
		// since credential is stored as a hash, the only way to update principal is to update both
		// as they must be specified together in the request.
		if v, _ := SupportedPlatforms[d.Get("platform").(string)]; v {
			attributes["PlatformPrincipal"] = aws.String(d.Get("platform_principal").(string))
		}
	}

	// Make API call to update attributes
	req := &sns.SetPlatformApplicationAttributesInput{
		PlatformApplicationArn: aws.String(d.Id()),
		Attributes:             attributes,
	}
	_, err := snsconn.SetPlatformApplicationAttributes(req)

	if err != nil {
		return fmt.Errorf("Error updating SNS application: %s", err)
	}

	return resourceAwsSnsApplicationRead(d, meta)
}

func resourceAwsSnsApplicationRead(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn

	// There is no SNS Describe/GetPlatformApplication to fetch attributes like name and platform
	// We will use the ID, which should be a platform application ARN, to:
	//  * Validate its an appropriate ARN on import
	//  * Parse out the name and platform
	platformApplicationArn, err := arn.Parse(d.Id())
	if err != nil {
		return fmt.Errorf(
			"SNS Platform Application ID must be of the form "+
				"arn:PARTITION:sns:REGION:ACCOUNTID:app/PLATFORM/NAME, "+
				"was provided %q and received error: %s", platformApplicationArn.String(), err)
	}

	platformApplicationArnResourceParts := strings.Split(platformApplicationArn.Resource, "/")
	if len(platformApplicationArnResourceParts) != 3 || platformApplicationArnResourceParts[0] != "app" {
		return fmt.Errorf(
			"SNS Platform Application ID must be of the form "+
				"arn:PARTITION:sns:REGION:ACCOUNTID:app/PLATFORM/NAME, "+
				"was provided: %s", platformApplicationArn.String())
	}

	d.Set("arn", platformApplicationArn.String())
	d.Set("name", platformApplicationArnResourceParts[2])
	d.Set("platform", platformApplicationArnResourceParts[1])

	attributeOutput, err := snsconn.GetPlatformApplicationAttributes(&sns.GetPlatformApplicationAttributesInput{
		PlatformApplicationArn: aws.String(platformApplicationArn.String()),
	})

	if err != nil {
		return err
	}

	if attributeOutput.Attributes != nil && len(attributeOutput.Attributes) > 0 {
		attrmap := attributeOutput.Attributes
		resource := *resourceAwsSnsApplication()
		// iKey = internal struct key, oKey = AWS Attribute Map key
		for iKey, oKey := range SNSPlatformAppAttributeMap {
			log.Printf("[DEBUG] Updating %s => %s", iKey, oKey)

			if attrmap[oKey] != nil {
				// Some of the fetched attributes are stateful properties such as
				// the number of subscriptions, the owner, etc. skip those
				if resource.Schema[iKey] != nil {
					value := *attrmap[oKey]
					log.Printf("[DEBUG] Updating %s => %s -> %s", iKey, oKey, value)
					d.Set(iKey, *attrmap[oKey])
				}
			}
		}
	}

	return nil
}

func resourceAwsSnsApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	snsconn := meta.(*AWSClient).snsconn

	log.Printf("[DEBUG] SNS Delete Application: %s", d.Id())
	_, err := snsconn.DeletePlatformApplication(&sns.DeletePlatformApplicationInput{
		PlatformApplicationArn: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}
	return nil
}

func hashSum(contents interface{}) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(contents.(string))))
}
