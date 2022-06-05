package secretsmanager

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceSecretRotation() *schema.Resource {
	return &schema.Resource{
		Create: resourceSecretRotationCreate,
		Read:   resourceSecretRotationRead,
		Update: resourceSecretRotationUpdate,
		Delete: resourceSecretRotationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rotation_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"rotation_lambda_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rotation_rules": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automatically_after_days": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func resourceSecretRotationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecretsManagerConn
	secretID := d.Get("secret_id").(string)

	if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
		input := &secretsmanager.RotateSecretInput{
			RotationLambdaARN: aws.String(v.(string)),
			RotationRules:     expandRotationRules(d.Get("rotation_rules").([]interface{})),
			SecretId:          aws.String(secretID),
		}

		log.Printf("[DEBUG] Enabling Secrets Manager Secret rotation: %s", input)
		var output *secretsmanager.RotateSecretOutput
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error
			output, err = conn.RotateSecret(input)
			if err != nil {
				// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
				if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			output, err = conn.RotateSecret(input)
		}

		if err != nil {
			return fmt.Errorf("error enabling Secrets Manager Secret %q rotation: %s", d.Id(), err)
		}

		d.SetId(aws.StringValue(output.ARN))
	}

	return resourceSecretRotationRead(d, meta)
}

func resourceSecretRotationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecretsManagerConn

	input := &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(d.Id()),
	}

	var output *secretsmanager.DescribeSecretOutput

	err := resource.Retry(PropagationTimeout, func() *resource.RetryError {
		var err error

		output, err = conn.DescribeSecret(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.DescribeSecret(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, secretsmanager.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Secrets Manager Secret Rotation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Secrets Manager Secret Rotation (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading Secrets Manager Secret Rotation (%s): empty response", d.Id())
	}

	d.Set("secret_id", d.Id())
	d.Set("rotation_enabled", output.RotationEnabled)

	if aws.BoolValue(output.RotationEnabled) {
		d.Set("rotation_lambda_arn", output.RotationLambdaARN)
		if err := d.Set("rotation_rules", flattenRotationRules(output.RotationRules)); err != nil {
			return fmt.Errorf("error setting rotation_rules: %s", err)
		}
	} else {
		d.Set("rotation_lambda_arn", "")
		d.Set("rotation_rules", []interface{}{})
	}

	return nil
}

func resourceSecretRotationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecretsManagerConn
	secretID := d.Get("secret_id").(string)

	if d.HasChanges("rotation_lambda_arn", "rotation_rules") {
		if v, ok := d.GetOk("rotation_lambda_arn"); ok && v.(string) != "" {
			input := &secretsmanager.RotateSecretInput{
				RotationLambdaARN: aws.String(v.(string)),
				RotationRules:     expandRotationRules(d.Get("rotation_rules").([]interface{})),
				SecretId:          aws.String(secretID),
			}

			log.Printf("[DEBUG] Enabling Secrets Manager Secret Rotation: %s", input)
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.RotateSecret(input)
				if err != nil {
					// AccessDeniedException: Secrets Manager cannot invoke the specified Lambda function.
					if tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})

			if tfresource.TimedOut(err) {
				_, err = conn.RotateSecret(input)
			}

			if err != nil {
				return fmt.Errorf("error updating Secrets Manager Secret Rotation %q : %s", d.Id(), err)
			}
		} else {
			input := &secretsmanager.CancelRotateSecretInput{
				SecretId: aws.String(d.Id()),
			}

			log.Printf("[DEBUG] Cancelling Secrets Manager Secret Rotation: %s", input)
			_, err := conn.CancelRotateSecret(input)
			if err != nil {
				return fmt.Errorf("error cancelling Secret Manager Secret Rotation %q : %s", d.Id(), err)
			}
		}
	}

	return resourceSecretRotationRead(d, meta)
}

func resourceSecretRotationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecretsManagerConn
	secretID := d.Get("secret_id").(string)

	input := &secretsmanager.CancelRotateSecretInput{
		SecretId: aws.String(secretID),
	}

	log.Printf("[DEBUG] Deleting Secrets Manager Rotation: %s", input)
	_, err := conn.CancelRotateSecret(input)
	if err != nil {
		return fmt.Errorf("error cancelling Secret Manager Secret %q rotation: %s", d.Id(), err)
	}

	return nil
}

func expandRotationRules(l []interface{}) *secretsmanager.RotationRulesType {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	rules := &secretsmanager.RotationRulesType{
		AutomaticallyAfterDays: aws.Int64(int64(m["automatically_after_days"].(int))),
	}

	return rules
}

func flattenRotationRules(rules *secretsmanager.RotationRulesType) []interface{} {
	if rules == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"automatically_after_days": int(aws.Int64Value(rules.AutomaticallyAfterDays)),
	}

	return []interface{}{m}
}
