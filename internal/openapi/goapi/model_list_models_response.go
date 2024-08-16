// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * Cody Service
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 0.0.0
 */

package goapi

type ListModelsResponse struct {
	Object string `json:"object"`

	Data []Model `json:"data"`
}

// AssertListModelsResponseRequired checks if the required fields are not zero-ed
func AssertListModelsResponseRequired(obj ListModelsResponse) error {
	elements := map[string]interface{}{
		"object": obj.Object,
		"data":   obj.Data,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.Data {
		if err := AssertModelRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertListModelsResponseConstraints checks if the values respects the defined constraints
func AssertListModelsResponseConstraints(obj ListModelsResponse) error {
	for _, el := range obj.Data {
		if err := AssertModelConstraints(el); err != nil {
			return err
		}
	}
	return nil
}
