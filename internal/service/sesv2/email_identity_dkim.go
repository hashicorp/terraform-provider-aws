package sesv2

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"golang.org/x/crypto/ssh"
)

func ResourceEmailIdentityDKIM() *schema.Resource {
	return &schema.Resource{
		Create: resourceEmailIdentityDKIMCreate,
		Read:   resourceEmailIdentityDKIMRead,
		Update: resourceEmailIdentityDKIMUpdate,
		Delete: resourceEmailIdentityDKIMDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"origin": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(sesv2.DkimSigningAttributesOrigin_Values(), false),
			},
			"selector": {
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"next_signing_key_length",
				},
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				ConflictsWith: []string{
					"next_signing_key_length",
				},
			},
			"next_signing_key_length": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(sesv2.DkimSigningKeyLength_Values(), false),
				ConflictsWith: []string{
					"selector",
					"private_key",
				},
			},
			"signing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"current_signing_key_length": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dkim_tokens": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceEmailIdentityDKIMCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	identity := d.Get("identity").(string)
	privateKey := d.Get("private_key").(string)
	selector := d.Get("selector").(string)

	createOpts := &sesv2.PutEmailIdentityDkimSigningAttributesInput{
		EmailIdentity: aws.String(identity),
	}

	if v, ok := d.GetOk("origin"); ok {
		createOpts.SigningAttributesOrigin = aws.String(v.(string))
	} else {
		// use the default
		createOpts.SigningAttributesOrigin = aws.String(sesv2.DkimSigningAttributesOriginAwsSes)
	}

	if privateKey != "" || selector != "" {
		rsaKey, err := readRSAPrivateKey(privateKey)
		if err != nil {
			return fmt.Errorf("failed to read RSA private key: %s", err)
		}
		key := base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(rsaKey))
		createOpts.SigningAttributes = &sesv2.DkimSigningAttributes{
			DomainSigningPrivateKey: aws.String(key),
			DomainSigningSelector:   aws.String(selector),
		}
	}

	if v, ok := d.GetOk("next_signing_key_length"); ok {
		// override BYODKIM settings
		createOpts.SigningAttributes = &sesv2.DkimSigningAttributes{
			NextSigningKeyLength: aws.String(v.(string)),
		}
	}

	_, err := conn.PutEmailIdentityDkimSigningAttributes(createOpts)
	if err != nil {
		return fmt.Errorf("Error requesting SES identity identity verification: %s", err)
	}

	if v, ok := d.GetOkExists("signing_enabled"); ok {
		signingOpts := &sesv2.PutEmailIdentityDkimAttributesInput{
			EmailIdentity:  aws.String(identity),
			SigningEnabled: aws.Bool(v.(bool)),
		}
		_, err := conn.PutEmailIdentityDkimAttributes(signingOpts)
		if err != nil {
			return fmt.Errorf("Error setting SES identity signing status: %s", err)
		}
	}

	d.SetId(identity)

	return resourceEmailIdentityDKIMRead(d, meta)
}

func resourceEmailIdentityDKIMRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	identity := d.Id()
	d.Set("identity", identity)

	readOpts := &sesv2.GetEmailIdentityInput{
		EmailIdentity: aws.String(identity),
	}

	response, err := conn.GetEmailIdentity(readOpts)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	if response.DkimAttributes == nil {
		log.Printf("[WARN] Domain not listed in response when fetching verification attributes for %s", d.Id())
		d.SetId("")
		return nil
	}

	origin := aws.StringValue(response.DkimAttributes.SigningAttributesOrigin)
	if origin == sesv2.DkimSigningAttributesOriginAwsSes {
		// unset the BYODKIM attributes
		d.Set("selector", "")
		d.Set("private_key", "")
		d.Set("public_key", "")
	}

	if v := d.Get("private_key").(string); v != "" {
		// set public_key attribute
		rsaKey, err := readRSAPrivateKey(v)
		if err != nil {
			return fmt.Errorf("failed to read RSA private key: %s", err)
		}
		pubKey, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to marshal RSA public key: %s", err)
		}
		d.Set("public_key", base64.StdEncoding.EncodeToString(pubKey))
		d.Set("current_signing_key_length", fmt.Sprintf("RSA_%d_BIT", rsaKey.N.BitLen()))
	} else {
		d.Set("current_signing_key_length", aws.StringValue(response.DkimAttributes.CurrentSigningKeyLength))
	}

	d.Set("origin", origin)
	d.Set("dkim_tokens", aws.StringValueSlice(response.DkimAttributes.Tokens))
	d.Set("next_signing_key_length", aws.StringValue(response.DkimAttributes.NextSigningKeyLength))
	d.Set("signing_enabled", aws.BoolValue(response.DkimAttributes.SigningEnabled))
	d.Set("status", aws.StringValue(response.DkimAttributes.Status))
	return nil
}

func resourceEmailIdentityDKIMUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESV2Conn

	identity := d.Id()

	if d.HasChange("signing_enabled") {
		v := d.Get("signing_enabled").(bool)
		signingOpts := &sesv2.PutEmailIdentityDkimAttributesInput{
			EmailIdentity:  aws.String(identity),
			SigningEnabled: aws.Bool(v),
		}
		log.Printf("[DEBUG] Updating SES identity signing status: %t", v)
		_, err := conn.PutEmailIdentityDkimAttributes(signingOpts)
		if err != nil {
			return fmt.Errorf("Error updating SES identity signing status: %s", err)
		}
	}

	var update bool
	updateOpts := &sesv2.PutEmailIdentityDkimSigningAttributesInput{
		EmailIdentity:           aws.String(identity),
		SigningAttributesOrigin: aws.String(d.Get("origin").(string)),
	}

	if d.HasChange("origin") {
		update = true
	}

	if d.HasChange("private_key") || d.HasChange("selector") {
		privateKey := d.Get("private_key").(string)
		selector := d.Get("selector").(string)
		if privateKey != "" && selector != "" {
			update = true
			rsaKey, err := readRSAPrivateKey(d.Get("private_key").(string))
			if err != nil {
				return fmt.Errorf("failed to read RSA private key: %s", err)
			}
			key := base64.StdEncoding.EncodeToString(x509.MarshalPKCS1PrivateKey(rsaKey))
			updateOpts.SigningAttributes = &sesv2.DkimSigningAttributes{
				DomainSigningPrivateKey: aws.String(key),
				DomainSigningSelector:   aws.String(d.Get("selector").(string)),
			}
		}
	}

	if d.HasChange("next_signing_key_length") {
		// override BYODKIM settings
		update = true
		updateOpts.SigningAttributes = &sesv2.DkimSigningAttributes{
			NextSigningKeyLength: aws.String(d.Get("next_signing_key_length").(string)),
		}
	}

	if !update {
		// nothing to update
		return nil
	}

	log.Printf("[DEBUG] Updating SES identity DKIM attributes: %s", updateOpts)
	_, err := conn.PutEmailIdentityDkimSigningAttributes(updateOpts)
	if err != nil {
		return fmt.Errorf("Error updating SES identity DKIM attributes: %s", err)
	}

	return resourceEmailIdentityDKIMRead(d, meta)
}

func resourceEmailIdentityDKIMDelete(d *schema.ResourceData, meta interface{}) error {
	// deleting of the DKIM configuration is not supported
	return nil
}

func readRSAPrivateKey(key string) (*rsa.PrivateKey, error) {
	// parse the pem key
	privateKey, err := ssh.ParseRawPrivateKey([]byte(key))
	if err != nil {
		return nil, err
	}

	switch v := privateKey.(type) {
	// only RSA PKCS #1 v1.5 is supported
	case *rsa.PrivateKey:
		return v, nil
	default:
		return nil, fmt.Errorf("unsupported key type %T, only RSA PKCS #1 v1.5 is supported", v)
	}
}
