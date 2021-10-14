package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceSecretVersion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSecretVersionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_string": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"secret_binary": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"version_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"version_stage": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "AWSCURRENT",
			},
			"version_stages": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSecretVersionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecretsManagerConn
	secretID := d.Get("secret_id").(string)
	var version string

	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretID),
	}

	if v, ok := d.GetOk("version_id"); ok {
		versionID := v.(string)
		input.VersionId = aws.String(versionID)
		version = versionID
	} else {
		versionStage := d.Get("version_stage").(string)
		input.VersionStage = aws.String(versionStage)
		version = versionStage
	}

	log.Printf("[DEBUG] Reading Secrets Manager Secret Version: %s", input)
	output, err := conn.GetSecretValue(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			return fmt.Errorf("Secrets Manager Secret %q Version %q not found", secretID, version)
		}
		if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException, "You can’t perform this operation on the secret because it was deleted") {
			return fmt.Errorf("Secrets Manager Secret %q Version %q not found", secretID, version)
		}
		return fmt.Errorf("error reading Secrets Manager Secret Version: %w", err)
	}

	d.SetId(fmt.Sprintf("%s|%s", secretID, version))
	d.Set("secret_id", secretID)
	d.Set("secret_string", output.SecretString)
	d.Set("version_id", output.VersionId)
	d.Set("secret_binary", string(output.SecretBinary))
	d.Set("arn", output.ARN)

	if err := d.Set("version_stages", flex.FlattenStringList(output.VersionStages)); err != nil {
		return fmt.Errorf("error setting version_stages: %w", err)
	}

	return nil
}
