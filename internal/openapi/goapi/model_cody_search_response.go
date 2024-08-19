/*
Sourcegraph

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 0.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package goapi

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the CodySearchResponse type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &CodySearchResponse{}

// CodySearchResponse struct for CodySearchResponse
type CodySearchResponse struct {
	Results []FileChunkContext `json:"results"`
}

type _CodySearchResponse CodySearchResponse

// NewCodySearchResponse instantiates a new CodySearchResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewCodySearchResponse(results []FileChunkContext) *CodySearchResponse {
	this := CodySearchResponse{}
	this.Results = results
	return &this
}

// NewCodySearchResponseWithDefaults instantiates a new CodySearchResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewCodySearchResponseWithDefaults() *CodySearchResponse {
	this := CodySearchResponse{}
	return &this
}

// GetResults returns the Results field value
func (o *CodySearchResponse) GetResults() []FileChunkContext {
	if o == nil {
		var ret []FileChunkContext
		return ret
	}

	return o.Results
}

// GetResultsOk returns a tuple with the Results field value
// and a boolean to check if the value has been set.
func (o *CodySearchResponse) GetResultsOk() ([]FileChunkContext, bool) {
	if o == nil {
		return nil, false
	}
	return o.Results, true
}

// SetResults sets field value
func (o *CodySearchResponse) SetResults(v []FileChunkContext) {
	o.Results = v
}

func (o CodySearchResponse) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o CodySearchResponse) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["results"] = o.Results
	return toSerialize, nil
}

func (o *CodySearchResponse) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"results",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varCodySearchResponse := _CodySearchResponse{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varCodySearchResponse)

	if err != nil {
		return err
	}

	*o = CodySearchResponse(varCodySearchResponse)

	return err
}

type NullableCodySearchResponse struct {
	value *CodySearchResponse
	isSet bool
}

func (v NullableCodySearchResponse) Get() *CodySearchResponse {
	return v.value
}

func (v *NullableCodySearchResponse) Set(val *CodySearchResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableCodySearchResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableCodySearchResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableCodySearchResponse(val *CodySearchResponse) *NullableCodySearchResponse {
	return &NullableCodySearchResponse{value: val, isSet: true}
}

func (v NullableCodySearchResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableCodySearchResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
