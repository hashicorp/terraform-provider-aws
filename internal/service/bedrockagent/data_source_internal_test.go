// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagent/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
)

func TestManagedKnowledgeBaseConnectorConfigurationModelFlattenCanonicalizesConnectorParameters(t *testing.T) {
	t.Parallel()

	input := `{"version":"1","type":"CONFLUENCE","filterConfiguration":{"maxFileSizeInMegaBytes":"500","inclusionSpaceKeys":["EXAMPLE"]},"connectionConfiguration":{"type":"SAAS","secretArn":"arn:aws:secretsmanager:us-east-1:111122223333:secret:example","hostUrl":"https://confluence.example.com","authType":"BASIC"},"aclEnabled":false,"largeNumber":9007199254740993}`
	want := `{"aclEnabled":false,"connectionConfiguration":{"authType":"BASIC","hostUrl":"https://confluence.example.com","secretArn":"arn:aws:secretsmanager:us-east-1:111122223333:secret:example","type":"SAAS"},"filterConfiguration":{"inclusionSpaceKeys":["EXAMPLE"],"maxFileSizeInMegaBytes":"500"},"largeNumber":9007199254740993,"type":"CONFLUENCE","version":"1"}`

	var got managedKnowledgeBaseConnectorConfigurationModel
	diags := got.Flatten(t.Context(), awstypes.ManagedKnowledgeBaseConnectorConfiguration{
		ConnectorParameters: document.NewLazyDocument(input),
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %s", diags.Errors())
	}

	if got := got.ConnectorParameters.ValueString(); got != want {
		t.Errorf("connector parameters = %q, want %q", got, want)
	}
}
