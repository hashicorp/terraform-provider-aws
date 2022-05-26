package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceKeyPair() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceKeyPairRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_pair_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceKeyPairRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeKeyPairsInput{}

	if v, ok := d.GetOk("filter"); ok {
		input.Filters = BuildFiltersDataSource(v.(*schema.Set))
	}

	if v, ok := d.GetOk("key_name"); ok {
		input.KeyNames = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("key_pair_id"); ok {
		input.KeyPairIds = aws.StringSlice([]string{v.(string)})
	}

	keyPair, err := FindKeyPair(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Key Pair", err)
	}

	d.SetId(aws.StringValue(keyPair.KeyPairId))

	keyName := aws.StringValue(keyPair.KeyName)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("key-pair/%s", keyName),
	}.String()
	d.Set("arn", arn)
	d.Set("fingerprint", keyPair.KeyFingerprint)
	d.Set("key_name", keyName)
	d.Set("key_pair_id", keyPair.KeyPairId)

	if err := d.Set("tags", KeyValueTags(keyPair.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
