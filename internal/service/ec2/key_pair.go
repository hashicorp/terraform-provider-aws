package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceKeyPair() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceKeyPairCreate,
		Read:   resourceKeyPairRead,
		Update: resourceKeyPairUpdate,
		Delete: resourceKeyPairDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		SchemaVersion: 1,
		MigrateState:  KeyPairMigrateState,

		Schema: map[string]*schema.Schema{
			"key_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"key_name_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 255),
			},
			"key_name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"key_name"},
				ValidateFunc:  validation.StringLenBetween(0, 255-resource.UniqueIDSuffixLength),
			},
			"public_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					switch v := v.(type) {
					case string:
						return strings.TrimSpace(v)
					default:
						return ""
					}
				},
			},
			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_pair_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceKeyPairCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var keyName string
	if v, ok := d.GetOk("key_name"); ok {
		keyName = v.(string)
	} else if v, ok := d.GetOk("key_name_prefix"); ok {
		keyName = resource.PrefixedUniqueId(v.(string))
		d.Set("key_name", keyName)
	} else {
		keyName = resource.UniqueId()
		d.Set("key_name", keyName)
	}

	publicKey := d.Get("public_key").(string)
	req := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(keyName),
		PublicKeyMaterial: []byte(publicKey),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeKeyPair),
	}
	resp, err := conn.ImportKeyPair(req)
	if err != nil {
		return fmt.Errorf("Error import KeyPair: %s", err)
	}

	d.SetId(aws.StringValue(resp.KeyName))

	return resourceKeyPairRead(d, meta)
}

func resourceKeyPairRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeKeyPairsInput{
		KeyNames: []*string{aws.String(d.Id())},
	}
	resp, err := conn.DescribeKeyPairs(req)
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidKeyPair.NotFound", "") {
			log.Printf("[WARN] Key Pair (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving KeyPair: %s", err)
	}

	if len(resp.KeyPairs) == 0 {
		log.Printf("[WARN] Key Pair (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	kp := resp.KeyPairs[0]

	if aws.StringValue(kp.KeyName) != d.Id() {
		return fmt.Errorf("Unable to find key pair within: %#v", resp.KeyPairs)
	}

	d.Set("key_name", kp.KeyName)
	d.Set("fingerprint", kp.KeyFingerprint)
	d.Set("key_pair_id", kp.KeyPairId)
	tags := KeyValueTags(kp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("key-pair/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceKeyPairUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("key_pair_id").(string), o, n); err != nil {
			return fmt.Errorf("error adding tags: %s", err)
		}
	}

	return resourceKeyPairRead(d, meta)
}

func resourceKeyPairDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	_, err := conn.DeleteKeyPair(&ec2.DeleteKeyPairInput{
		KeyName: aws.String(d.Id()),
	})
	return err
}
