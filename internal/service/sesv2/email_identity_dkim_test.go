package sesv2_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sesv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccSESv2DomainDKIM_basic(t *testing.T) {
	resourceName := "aws_sesv2_email_identity_dkim.test"
	identity := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, sesv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDKIMDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDomainDKIMConfig, identity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainDKIMExists(resourceName),
					testAccCheckDomainDKIMTokens(resourceName),
				),
			},
		},
	})
}

func TestAccSESv2DomainDKIM_byodkim(t *testing.T) {
	resourceName := "aws_sesv2_email_identity_dkim.test"
	identity := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheck(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, sesv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDomainDKIMDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccDomainDKIMbyodkim, identity),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDomainDKIMExists(resourceName),
					testAccCheckDomainBYODKIMToken(resourceName),
				),
			},
		},
	})
}

func testAccCheckDomainDKIMDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sesv2_email_identity_dkim" {
			continue
		}

		identity := rs.Primary.ID
		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(identity),
		}

		resp, err := conn.GetEmailIdentity(params)

		if err != nil {
			if _, ok := err.(*sesv2.NotFoundException); ok {
				return nil
			}
			return err
		}

		if resp.DkimAttributes != nil {
			return fmt.Errorf("SES identity DKIM %s still exists.", identity)
		}
	}

	return nil
}

func testAccCheckDomainDKIMExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("SES Domain Identity not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("SES Domain Identity name not set")
		}

		identity := rs.Primary.ID
		conn := acctest.Provider.Meta().(*conns.AWSClient).SESV2Conn

		params := &sesv2.GetEmailIdentityInput{
			EmailIdentity: aws.String(identity),
		}

		response, err := conn.GetEmailIdentity(params)
		if err != nil {
			return err
		}

		if response.DkimAttributes == nil {
			return fmt.Errorf("SES identity DKIM %s not found in AWS", identity)
		}

		return nil
	}
}

func testAccCheckDomainDKIMTokens(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		expectedNum := 3
		expectedFormat := regexp.MustCompile("[a-z0-9]{32}")

		tokenNum, _ := strconv.Atoi(rs.Primary.Attributes["dkim_tokens.#"])
		if expectedNum != tokenNum {
			return fmt.Errorf("Incorrect number of DKIM tokens, expected: %d, got: %d", expectedNum, tokenNum)
		}
		for i := 0; i < expectedNum; i++ {
			key := fmt.Sprintf("dkim_tokens.%d", i)
			token := rs.Primary.Attributes[key]
			if !expectedFormat.MatchString(token) {
				return fmt.Errorf("Incorrect format of DKIM token: %v", token)
			}
		}

		return nil
	}
}

func testAccCheckDomainBYODKIMToken(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[n]

		expectedNum := 1

		tokenNum, _ := strconv.Atoi(rs.Primary.Attributes["dkim_tokens.#"])
		if expectedNum != tokenNum {
			return fmt.Errorf("Incorrect number of DKIM tokens, expected: %d, got: %d", expectedNum, tokenNum)
		}
		for i := 0; i < expectedNum; i++ {
			key := fmt.Sprintf("dkim_tokens.%d", i)
			token := rs.Primary.Attributes[key]
			if token != "default" {
				return fmt.Errorf("Incorrect format of DKIM token: %v", token)
			}
		}

		return nil
	}
}

const testAccDomainDKIMConfig = `
resource "aws_sesv2_email_identity" "test" {
  identity = %[1]q
}

resource "aws_sesv2_email_identity_dkim" "test" {
  identity = aws_sesv2_email_identity.test.identity
  next_signing_key_length = "RSA_1024_BIT"
}
`

const testAccDomainDKIMbyodkim = `
resource "aws_sesv2_email_identity" "test" {
  identity = %[1]q
}
resource "aws_sesv2_email_identity_dkim" "test" {
  identity = aws_sesv2_email_identity.test.identity
  origin = "EXTERNAL"
  selector = "default"
  private_key = <<EOT
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCqGKukO1De7zhZj6+H0qtjTkVxwTCpvKe4eCZ0FPqri0cb2JZfXJ/DgYSF6vUp
wmJG8wVQZKjeGcjDOL5UlsuusFncCzWBQ7RKNUSesmQRMSGkVb1/3j+skZ6UtW+5u09lHNsj6tQ5
1s1SPrCBkedbNf0Tp0GbMJDyR4e9T04ZZwIDAQABAoGAFijko56+qGyN8M0RVyaRAXz++xTqHBLh
3tx4VgMtrQ+WEgCjhoTwo23KMBAuJGSYnRmoBZM3lMfTKevIkAidPExvYCdm5dYq3XToLkkLv5L2
pIIVOFMDG+KESnAFV7l2c+cnzRMW0+b6f8mR1CJzZuxVLL6Q02fvLi55/mbSYxECQQDeAw6fiIQX
GukBI4eMZZt4nscy2o12KyYner3VpoeE+Np2q+Z3pvAMd/aNzQ/W9WaI+NRfcxUJrmfPwIGm63il
AkEAxCL5HQb2bQr4ByorcMWm/hEP2MZzROV73yF41hPsRC9m66KrheO9HPTJuo3/9s5p+sqGxOlF
L0NDt4SkosjgGwJAFklyR1uZ/wPJjj611cdBcztlPdqoxssQGnh85BzCj/u3WqBpE2vjvyyvyI5k
X6zk7S0ljKtt2jny2+00VsBerQJBAJGC1Mg5Oydo5NwD6BiROrPxGo2bpTbu/fhrT8ebHkTz2epl
U9VQQSQzY1oZMVX8i1m5WUTLPz2yLJIBQVdXqhMCQBGoiuSoSjafUhV7i1cEGpb88h5NBYZzWXGZ
37sJ5QsW+sJyoNde3xH8vdXhzU7eT82D6X/scw9RZz+/6rCJ4p0=
-----END RSA PRIVATE KEY-----
EOT
}
`
