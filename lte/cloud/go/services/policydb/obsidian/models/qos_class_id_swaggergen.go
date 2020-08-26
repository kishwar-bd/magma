// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/validate"
)

// QosClassID qos class id
// swagger:model qos_class_id
type QosClassID int32

// for schema
var qosClassIdEnum []interface{}

func init() {
	var res []QosClassID
	if err := json.Unmarshal([]byte(`[0,1,2,3,4,5,6,7,8,9,65,66,67,70,75,79]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		qosClassIdEnum = append(qosClassIdEnum, v)
	}
}

func (m QosClassID) validateQosClassIDEnum(path, location string, value QosClassID) error {
	if err := validate.Enum(path, location, value, qosClassIdEnum); err != nil {
		return err
	}
	return nil
}

// Validate validates this qos class id
func (m QosClassID) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateQosClassIDEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}