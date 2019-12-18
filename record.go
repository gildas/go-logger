package logger

import (
	"encoding/json"
	"github.com/gildas/go-errors"
)

// Record is the map that contains all records of a log entry
//   If the value at a key is a func() interface the func will be called when the record is marshaled
type Record map[string]interface{}

// NewRecord creates a new empty record
func NewRecord() Record {
	return Record(make(map[string]interface{}))
}

// Set the key and value if not yet set
func (record Record) Set(key string, value interface{}) Record {
	if value == nil {
		return record
	}
	if _, ok := record[key]; !ok {
		record[key] = value
	}
	return record
}

// Merge merges a source Record into this Record
//   values already set in this record cannot be overriden
func (record Record) Merge(source Record) Record {
	for key, value := range source {
		record.Set(key, value)
	}
	return record
}

// MarshalJSON marshals this into JSON
func (record Record) MarshalJSON() ([]byte, error) {
	if len(record) == 0 {
		return []byte("null"), nil
	}
	buffer := make(map[string]interface{}, len(record))

	for k, v := range record {
		switch v := v.(type) {
		case func() interface{}:
			buffer[k] = v()
		default:
			buffer[k] = v
		}
	}

	return json.Marshal(buffer)
}

// UnmarshalJSON unmarshals JSON into this
func (record *Record) UnmarshalJSON(payload []byte) error {
	var placeholder map[string]interface{}
	if err := json.Unmarshal(payload, &placeholder); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*record = Record(placeholder)
	return nil
}