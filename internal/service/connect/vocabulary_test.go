package connect_test

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
)

func testAccCheckVocabularyExists(resourceName string, function *connect.DescribeVocabularyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Vocabulary not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Vocabulary ID not set")
		}
		instanceID, vocabularyID, err := tfconnect.VocabularyParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		params := &connect.DescribeVocabularyInput{
			InstanceId:   aws.String(instanceID),
			VocabularyId: aws.String(vocabularyID),
		}

		getFunction, err := conn.DescribeVocabulary(params)
		if err != nil {
			return err
		}

		*function = *getFunction

		return nil
	}
}

func testAccCheckVocabularyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_vocabulary" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		instanceID, vocabularyID, err := tfconnect.VocabularyParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		params := &connect.DescribeVocabularyInput{
			InstanceId:   aws.String(instanceID),
			VocabularyId: aws.String(vocabularyID),
		}

		resp, err := conn.DescribeVocabulary(params)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		// API returns an empty list for Vocabulary if there are none
		if resp.Vocabulary == nil {
			continue
		}
	}

	return nil
}
