package dynamodb

import (
	"github.com/mitchellh/copystructure"
)

func stripCapacityAttributes(in map[string]interface{}) (map[string]interface{}, error) {
	mapCopy, err := copystructure.Copy(in)
	if err != nil {
		return nil, err
	}

	m := mapCopy.(map[string]interface{})

	delete(m, "write_capacity")
	delete(m, "read_capacity")

	return m, nil
}

func stripNonKeyAttributes(in map[string]interface{}) (map[string]interface{}, error) {
	mapCopy, err := copystructure.Copy(in)
	if err != nil {
		return nil, err
	}

	m := mapCopy.(map[string]interface{})

	delete(m, "non_key_attributes")

	return m, nil
}
