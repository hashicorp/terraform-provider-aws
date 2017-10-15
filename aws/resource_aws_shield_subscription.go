package aws

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsShieldSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsShieldSubscriptionCreate,
		Read:   resourceAwsShieldSubscriptionRead,
		Delete: resourceAwsShieldSubscriptionDelete,

		Schema: map[string]*schema.Schema{},
	}
}

func resourceAwsShieldSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.CreateSubscriptionInput{}

	_, err := conn.CreateSubscription(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case shield.ErrCodeResourceAlreadyExistsException:
				return resourceAwsShieldSubscriptionRead(d, meta)
			default:
				return err
			}
		}
		return err
	}
	return resourceAwsShieldSubscriptionRead(d, meta)
}

func resourceAwsShieldSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DescribeSubscriptionInput{}

	_, err := conn.DescribeSubscription(input)
	if err != nil {
		return err
	}
	return nil
}

func resourceAwsShieldSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DeleteSubscriptionInput{}

	_, err := conn.DeleteSubscription(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case shield.ErrCodeResourceNotFoundException:
				return nil
			default:
				return err
			}
		}
		return err
	}
	return nil
}
