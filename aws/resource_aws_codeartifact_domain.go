package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsCodeArtifactDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCodeArtifactDomainCreate,
		Read:   resourceAwsCodeArtifactDomainRead,
		Delete: resourceAwsCodeArtifactDomainDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"encryption_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_size_bytes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"repository_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAwsCodeArtifactDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Print("[DEBUG] Creating CodeArtifact Domain")

	params := &codeartifact.CreateDomainInput{
		Domain:        aws.String(d.Get("domain").(string)),
		EncryptionKey: aws.String(d.Get("encryption_key").(string)),
	}

	domain, err := conn.CreateDomain(params)
	if err != nil {
		return fmt.Errorf("error creating CodeArtifact Domain: %s", err)
	}

	d.SetId(aws.StringValue(domain.Domain.Name))

	return resourceAwsCodeArtifactDomainRead(d, meta)
}

func resourceAwsCodeArtifactDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn

	log.Printf("[DEBUG] Reading CodeArtifact Domain: %s", d.Id())

	sm, err := conn.DescribeDomain(&codeartifact.DescribeDomainInput{
		Domain: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] CodeArtifact Domain %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("domain", sm.Domain.Name)
	d.Set("arn", sm.Domain.Arn)
	d.Set("encryption_key", sm.Domain.EncryptionKey)
	d.Set("owner", sm.Domain.Owner)
	d.Set("asset_size_bytes", sm.Domain.AssetSizeBytes)
	d.Set("repository_count", sm.Domain.RepositoryCount)

	if err := d.Set("created_time", sm.Domain.CreatedTime.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting created_time: %s", err)
	}

	return nil
}

func resourceAwsCodeArtifactDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	log.Printf("[DEBUG] Deleting CodeArtifact Domain: %s", d.Id())

	input := &codeartifact.DeleteDomainInput{
		Domain: aws.String(d.Id()),
	}

	_, err := conn.DeleteDomain(input)

	if err != nil {
		return fmt.Errorf("error deleting CodeArtifact Domain: %s", err)
	}

	return nil
}
