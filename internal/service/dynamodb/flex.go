package dynamodb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func ExpandTableItemAttributes(input string) (map[string]*dynamodb.AttributeValue, error) {
	var attributes map[string]*dynamodb.AttributeValue

	dec := json.NewDecoder(strings.NewReader(input))
	err := dec.Decode(&attributes)
	if err != nil {
		return nil, fmt.Errorf("Decoding failed: %s", err)
	}

	return attributes, nil
}

func flattenTableItemAttributes(attrs map[string]*dynamodb.AttributeValue) (string, error) {
	buf := bytes.NewBufferString("")
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(attrs)
	if err != nil {
		return "", fmt.Errorf("Encoding failed: %s", err)
	}

	var rawData map[string]map[string]interface{}

	// Reserialize so we get rid of the nulls
	decoder := json.NewDecoder(strings.NewReader(buf.String()))
	err = decoder.Decode(&rawData)
	if err != nil {
		return "", fmt.Errorf("Decoding failed: %s", err)
	}

	for _, value := range rawData {
		for typeName, typeVal := range value {
			if typeVal == nil {
				delete(value, typeName)
			}
		}
	}

	rawBuffer := bytes.NewBufferString("")
	rawEncoder := json.NewEncoder(rawBuffer)
	err = rawEncoder.Encode(rawData)
	if err != nil {
		return "", fmt.Errorf("Re-encoding failed: %s", err)
	}

	return rawBuffer.String(), nil
}
