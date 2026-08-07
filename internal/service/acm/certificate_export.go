// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package acm

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/crypto/pbkdf2"
)

// @SDKResource("aws_acm_certificate_export", name="Certificate Export")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/acm;acm.ExportCertificateOutput")
func resourceCertificateExport() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCertificateExportCreate,
		ReadWithoutTimeout:   resourceCertificateExportRead,
		UpdateWithoutTimeout: resourceCertificateExportUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Schema: map[string]*schema.Schema{
			names.AttrCertificateARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"passphrase": {
				Type:      schema.TypeString,
				Required:  true,
				ForceNew:  true,
				Sensitive: true,
			},
			"decrypt_private_key": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrCertificate: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			names.AttrCertificateChain: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			names.AttrPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"decrypted_private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(ctx context.Context, d *schema.ResourceDiff, meta any) error {
				if d.HasChange("decrypt_private_key") {
					if err := d.SetNewComputed("decrypted_private_key"); err != nil {
						return err
					}
				}
				return nil
			},
		),
	}
}

func resourceCertificateExportCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	arn := d.Get(names.AttrCertificateARN).(string)
	passphrase := d.Get("passphrase").(string)

	input := &acm.ExportCertificateInput{
		CertificateArn: aws.String(arn),
		Passphrase:     []byte(passphrase),
	}

	output, err := conn.ExportCertificate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "exporting ACM Certificate (%s): %s", arn, err)
	}

	// Use a hash of the certificate ARN and passphrase as the resource ID
	// This ensures a unique ID while being reproducible for the same inputs
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", arn, passphrase)))
	d.SetId(base64.URLEncoding.EncodeToString(hash[:]))

	d.Set(names.AttrCertificate, aws.ToString(output.Certificate))
	d.Set(names.AttrCertificateChain, aws.ToString(output.CertificateChain))
	d.Set(names.AttrPrivateKey, aws.ToString(output.PrivateKey))

	// Decrypt the private key if requested
	if d.Get("decrypt_private_key").(bool) {
		encryptedKey := aws.ToString(output.PrivateKey)
		if decryptedKey, err := decryptPKCS8PrivateKey(encryptedKey, passphrase); err != nil {
			return sdkdiag.AppendErrorf(diags, "decrypting private key: %s", err)
		} else {
			d.Set("decrypted_private_key", decryptedKey)
		}
	}

	return diags
}

func resourceCertificateExportRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ACMClient(ctx)

	arn := d.Get(names.AttrCertificateARN).(string)

	// Verify the certificate still exists
	_, err := findCertificateByARN(ctx, conn, arn)

	if err != nil {
		log.Printf("[WARN] ACM Certificate %s not found, removing from state", arn)
		d.SetId("")
		return diags
	}

	// Note: We cannot re-export the certificate on read without the passphrase stored in state
	// The sensitive values are already stored in state from the create operation
	// We only verify the certificate still exists

	return diags
}

func resourceCertificateExportUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Only decrypt_private_key can change
	if d.HasChange("decrypt_private_key") {
		if d.Get("decrypt_private_key").(bool) {
			// Decrypt the existing encrypted key from state
			encryptedKey := d.Get(names.AttrPrivateKey).(string)
			passphrase := d.Get("passphrase").(string)
			
			log.Printf("[DEBUG] Update: attempting to decrypt with passphrase length=%d, encrypted_key length=%d", 
				len(passphrase), len(encryptedKey))
			
			if passphrase == "" {
				return sdkdiag.AppendErrorf(diags, "passphrase is empty in state, cannot decrypt")
			}
			
			decryptedKey, err := decryptPKCS8PrivateKey(encryptedKey, passphrase)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "decrypting private key: %s", err)
			}
			if err := d.Set("decrypted_private_key", decryptedKey); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting decrypted_private_key: %s", err)
			}
		} else {
			// Clear the decrypted key
			if err := d.Set("decrypted_private_key", ""); err != nil {
				return sdkdiag.AppendErrorf(diags, "clearing decrypted_private_key: %s", err)
			}
		}
	}

	return diags
}

// PKCS#8 encryption algorithm identifiers
var (
	oidPBES2     = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 13}
	oidPBKDF2    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 12}
	oidAES256CBC = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 42}
)

// PKCS#8 structures
type encryptedPrivateKeyInfo struct {
	EncryptionAlgorithm pkix8AlgorithmIdentifier
	EncryptedData       []byte
}

type pkix8AlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue
}

type pbes2Params struct {
	KeyDerivationFunc pkix8AlgorithmIdentifier
	EncryptionScheme  pkix8AlgorithmIdentifier
}

type pbkdf2Params struct {
	Salt           []byte
	IterationCount int
	KeyLength      int                      `asn1:"optional"`
	PrfAlgorithm   pkix8AlgorithmIdentifier `asn1:"optional"`
}

// decryptPKCS8PrivateKey decrypts a PKCS#8 encrypted private key
func decryptPKCS8PrivateKey(encryptedPEM, passphrase string) (string, error) {
	block, _ := pem.Decode([]byte(encryptedPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	if block.Type != "ENCRYPTED PRIVATE KEY" {
		return "", fmt.Errorf("PEM block is not an encrypted private key (type: %s)", block.Type)
	}

	var encryptedKey encryptedPrivateKeyInfo
	if _, err := asn1.Unmarshal(block.Bytes, &encryptedKey); err != nil {
		return "", fmt.Errorf("failed to parse PKCS#8 structure: %w", err)
	}

	if !encryptedKey.EncryptionAlgorithm.Algorithm.Equal(oidPBES2) {
		return "", fmt.Errorf("unsupported encryption algorithm: %v", encryptedKey.EncryptionAlgorithm.Algorithm)
	}

	var pbes2 pbes2Params
	if _, err := asn1.Unmarshal(encryptedKey.EncryptionAlgorithm.Parameters.FullBytes, &pbes2); err != nil {
		return "", fmt.Errorf("failed to parse PBES2 parameters: %w", err)
	}

	if !pbes2.KeyDerivationFunc.Algorithm.Equal(oidPBKDF2) {
		return "", fmt.Errorf("unsupported key derivation function: %v", pbes2.KeyDerivationFunc.Algorithm)
	}

	var pbkdf2Param pbkdf2Params
	if _, err := asn1.Unmarshal(pbes2.KeyDerivationFunc.Parameters.FullBytes, &pbkdf2Param); err != nil {
		return "", fmt.Errorf("failed to parse PBKDF2 parameters: %w", err)
	}

	if !pbes2.EncryptionScheme.Algorithm.Equal(oidAES256CBC) {
		return "", fmt.Errorf("unsupported encryption scheme: %v", pbes2.EncryptionScheme.Algorithm)
	}

	// AES IV is stored directly as an OCTET STRING
	var iv []byte
	if _, err := asn1.Unmarshal(pbes2.EncryptionScheme.Parameters.FullBytes, &iv); err != nil {
		return "", fmt.Errorf("failed to parse AES IV: %w", err)
	}

	keyLen := pbkdf2Param.KeyLength
	if keyLen == 0 {
		keyLen = 32 // AES-256 key size
	}
	
	log.Printf("[DEBUG] PKCS#8 decryption params: salt_len=%d, iterations=%d, key_len=%d, iv_len=%d, encrypted_len=%d",
		len(pbkdf2Param.Salt), pbkdf2Param.IterationCount, keyLen, len(iv), len(encryptedKey.EncryptedData))
	
	// PKCS#8 default PRF for PBKDF2 is HMAC-SHA1 (not SHA256)
	// AWS ACM uses the default unless explicitly specified
	derivedKey := pbkdf2.Key([]byte(passphrase), pbkdf2Param.Salt, pbkdf2Param.IterationCount, keyLen, sha1.New)

	plaintext, err := decryptAES256CBC(encryptedKey.EncryptedData, derivedKey, iv)
	if err != nil {
		return "", fmt.Errorf("decryption failed (salt_len=%d, iter=%d, key_len=%d, iv_len=%d): %w", 
			len(pbkdf2Param.Salt), pbkdf2Param.IterationCount, keyLen, len(iv), err)
	}

	decryptedBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: plaintext,
	}

	return string(pem.EncodeToMemory(decryptedBlock)), nil
}

func decryptAES256CBC(ciphertext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext, err = removePKCS7Padding(plaintext)
	if err != nil {
		return nil, fmt.Errorf("invalid padding: %w", err)
	}

	return plaintext, nil
}

func removePKCS7Padding(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding size: %d", padding)
	}

	for i := 0; i < padding; i++ {
		if data[len(data)-1-i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}

	return data[:len(data)-padding], nil
}
