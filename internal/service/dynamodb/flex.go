package dynamodb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
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

	bar := make(map[string]foo, len(attrs))
	for k, v := range attrs {
		bar[k] = foo(*v)
	}
	err := encoder.Encode(bar)
	if err != nil {
		return "", fmt.Errorf("Encoding failed: %s", err)
	}

	return buf.String(), nil
}

type foo dynamodb.AttributeValue

func (f foo) MarshalJSON() ([]byte, error) {
	thing := map[string]any{}

	if f.B != nil {
		thing["B"] = f.B
	}
	if f.BS != nil {
		thing["BS"] = f.BS
	}
	if f.BOOL != nil {
		thing["BOOL"] = f.BOOL
	}
	if f.L != nil {
		thing["L"] = slices.ApplyToAll(f.L, func(t *dynamodb.AttributeValue) foo {
			return foo(*t)
		})
	}
	if f.M != nil {
		thing["M"] = maps.ApplyToAll(f.M, func(t *dynamodb.AttributeValue) foo {
			return foo(*t)
		})
	}
	if f.N != nil {
		thing["N"] = f.N
	}
	if f.NS != nil {
		thing["NS"] = f.NS
	}
	if f.NULL != nil {
		thing["NULL"] = f.NULL
	}
	if f.S != nil {
		thing["S"] = f.S
	}
	if f.SS != nil {
		thing["SS"] = f.SS
	}

	return json.Marshal(thing)
}
