package ecr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceRepository() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRepositoryRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"kms_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"scan_on_push": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"image_tag_mutability": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"registry_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"repository_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRepositoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECRConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)
	params := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{name}),
	}

	if v, ok := d.GetOk("registry_id"); ok {
		params.RegistryId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Reading ECR repository: %#v", params)
	out, err := conn.DescribeRepositories(params)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodeRepositoryNotFoundException) {
			return fmt.Errorf("ECR Repository (%s) not found", name)
		}
		return fmt.Errorf("error reading ECR repository: %w", err)
	}

	repository := out.Repositories[0]
	arn := aws.StringValue(repository.RepositoryArn)

	d.SetId(aws.StringValue(repository.RepositoryName))
	d.Set("arn", arn)
	d.Set("name", repository.RepositoryName)
	d.Set("registry_id", repository.RegistryId)
	d.Set("repository_url", repository.RepositoryUri)
	d.Set("image_tag_mutability", repository.ImageTagMutability)

	if err := d.Set("image_scanning_configuration", flattenImageScanningConfiguration(repository.ImageScanningConfiguration)); err != nil {
		return fmt.Errorf("error setting image_scanning_configuration for ECR Repository (%s): %w", arn, err)
	}

	if err := d.Set("encryption_configuration", flattenRepositoryEncryptionConfiguration(repository.EncryptionConfiguration)); err != nil {
		return fmt.Errorf("error setting encryption_configuration for ECR Repository (%s): %w", arn, err)
	}

	tags, err := ListTags(conn, arn)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if meta.(*conns.AWSClient).Partition != endpoints.AwsPartitionID && verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] failed listing tags for ECR Repository (%s): %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed listing tags for ECR Repository (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags for ECR Repository (%s): %w", arn, err)
	}

	return nil
}
