package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsElasticsearchPackage() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsElasticsearchPackageCreate,
		Read:   resourceAwsElasticsearchPackageRead,
		Delete: resourceAwsElasticsearchPackageDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 28),
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					elasticsearch.PackageTypeTxtDictionary,
				}, false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_bucket_name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"s3_key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
		},
	}
}

func resourceAwsElasticsearchPackageCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	name := d.Get("name").(string)
	source := d.Get("source").([]interface{})[0].(map[string]interface{})

	input := &elasticsearch.CreatePackageInput{
		PackageName:        aws.String(name),
		PackageType:        aws.String(d.Get("type").(string)),
		PackageDescription: aws.String(d.Get("description").(string)),
		PackageSource: &elasticsearch.PackageSource{
			S3BucketName: aws.String(source["s3_bucket_name"].(string)),
			S3Key:        aws.String(source["s3_key"].(string)),
		},
	}

	out, err := conn.CreatePackage(input)
	if err != nil {
		return fmt.Errorf("Error creating Elasticsearch Package %s: %s", name, err)
	}

	d.SetId(aws.StringValue(out.PackageDetails.PackageID))

	return resourceAwsElasticsearchPackageRead(d, meta)
}

func resourceAwsElasticsearchPackageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	details, err := getElasticsearchPackage(conn, d.Id())
	if err != nil {
		if isAWSErr(err, "ResourceNotFoundException", "") {
			log.Printf("[WARN] ElasticSearch Package (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error reading ElasticSearch Package: %s", err)
	}

	if details == nil {
		log.Printf("[WARN] ElasticSearch Package (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.SetId(aws.StringValue(details.PackageID))
	d.Set("name", details.PackageName)
	d.Set("description", details.PackageDescription)
	d.Set("type", details.PackageType)

	return nil
}

func resourceAwsElasticsearchPackageDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).esconn

	input := &elasticsearch.DeletePackageInput{PackageID: aws.String(d.Id())}
	_, err := conn.DeletePackage(input)
	if err != nil {
		return fmt.Errorf("Error deleting Elasticsearch Package: %v", err)
	}

	return nil
}

func getElasticsearchPackage(conn *elasticsearch.ElasticsearchService, packageID string) (*elasticsearch.PackageDetails, error) {
	input := &elasticsearch.DescribePackagesInput{
		Filters: []*elasticsearch.DescribePackagesFilter{
			{
				Name:  aws.String("PackageID"),
				Value: []*string{aws.String(packageID)},
			},
		},
	}
	output, err := conn.DescribePackages(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PackageDetailsList) == 0 {
		return nil, nil
	}

	return output.PackageDetailsList[0], nil
}
