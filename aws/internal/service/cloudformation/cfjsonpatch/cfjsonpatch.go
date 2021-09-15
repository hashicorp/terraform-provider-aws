// cfjsonpatch implements CloudFormation customizations for RFC 6902 JSON Patch operations.
//
// This functionality is temporary, as the API will be updated to support standard
// RFC 6902 JSON Patch Value field.
package cfjsonpatch

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/mattbaird/jsonpatch"
)

const (
	EmptyDocument = "{}"
)

func PatchOperations(oldRaw interface{}, newRaw interface{}) ([]*cloudformation.PatchOperation, error) {
	old, ok := oldRaw.(string)

	if !ok {
		old = EmptyDocument
	}

	new, ok := newRaw.(string)

	if !ok {
		new = EmptyDocument
	}

	rfc6902patch, err := createRFC6902Patch([]byte(old), []byte(new))

	if err != nil {
		return nil, err
	}

	return convertRFC6902Patch(rfc6902patch)
}

func createRFC6902Patch(a []byte, b []byte) ([]jsonpatch.JsonPatchOperation, error) {
	patch, err := jsonpatch.CreatePatch(a, b)

	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] RFC 6902 JSON Patch: %s", patch)

	return patch, nil
}

func convertRFC6902Patch(patches []jsonpatch.JsonPatchOperation) ([]*cloudformation.PatchOperation, error) {
	patchOperations := make([]*cloudformation.PatchOperation, len(patches))

	for idx, patch := range patches {
		patchOperation := &cloudformation.PatchOperation{
			Op:   aws.String(patch.Operation),
			Path: aws.String(patch.Path),
		}

		if patch.Value == nil {
			patchOperations[idx] = patchOperation
			continue
		}

		switch value := patch.Value.(type) {
		default:
			log.Printf("[DEBUG] Object JSON Patch value type: %T", value)

			v, err := json.Marshal(value)

			if err != nil {
				return nil, fmt.Errorf("unable to marshal JSON Patch value: %w", err)
			}

			patchOperation.ObjectValue = aws.String(string(v))
		case bool:
			patchOperation.BooleanValue = aws.Bool(value)
		case float64:
			// jsonpatch does not differentiate between integer versus number types
			// and proper typing would require the CloudFormation resource schema.
			// To keep things simple for now since most properties are integer,
			// fuzzy match round numbers to integer values.
			if value == float64(int64(value)) {
				patchOperation.IntegerValue = aws.Int64(int64(value))
			} else {
				patchOperation.NumberValue = aws.Float64(value)
			}
		case string:
			patchOperation.StringValue = aws.String(value)
		}

		patchOperations[idx] = patchOperation
	}

	log.Printf("[DEBUG] CloudFormation JSON Patch: %s", patchOperations)

	return patchOperations, nil
}
