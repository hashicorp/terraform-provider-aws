package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsEcrRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcrRepositoryRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"registry_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEcrRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecrconn

	repositoryName := d.Get("name").(string)
	log.Printf("[DEBUG] Reading repository %s", repositoryName)
	out, err := conn.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		RepositoryNames: []*string{aws.String(repositoryName)},
	})
	if err != nil {
		if ecrerr, ok := err.(awserr.Error); ok && ecrerr.Code() == "RepositoryNotFoundException" {
			d.SetId("")
			return nil
		}
		return err
	}

	repository := out.Repositories[0]

	log.Printf("[DEBUG] Received repository %s", out)

	d.SetId(*repository.RepositoryName)
	d.Set("arn", repository.RepositoryArn)
	d.Set("registry_id", repository.RegistryId)
	d.Set("name", repository.RepositoryName)
	d.Set("repository_url", repository.RepositoryUri)

	return nil
}
