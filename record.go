package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gildas/go-errors"
)

// Record is the map that contains all records of a log entry
//
// If the value at a key is a func() interface the func will be called when the record is marshaled
type Record struct {
	Data map[string]interface{}
	KeysToRedact []string
}

// NewRecord creates a new empty record
func NewRecord() *Record {
	return &Record {
		Data: make(map[string]interface{}),
		KeysToRedact: nil,
	}
}

// Reset resets the record
func (record *Record) Reset() {
	for key := range record.Data {
		delete(record.Data, key)
	}
	record.KeysToRedact = nil
}

// NewPooledRecord creates a new empty record
func NewPooledRecord() (record *Record, release func()) {
	record = mapPool.Get()
	return record, func() { record.Close() }
}

// Close returns the record to the pool
func (record *Record) Close() {
	mapPool.Put(record)
}

// Get gets the value at a key
func (record *Record) Get(key string) interface{} {
	return record.Data[key]
}

// Set sets the key and value if not yet set
func (record *Record) Set(key string, value interface{}) *Record {
	if value == nil {
		return record
	}
	if _, ok := record.Data[key]; !ok {
		record.Data[key] = value
	}
	return record
}

// AddKeysToRedact adds keys to redact
func (record *Record) AddKeysToRedact(keys ...string) *Record {
	record.KeysToRedact = append(record.KeysToRedact, keys...)
	return record
}

// Merge merges a source Record into this Record
//
// values already set in this record cannot be overridden
func (record *Record) Merge(source *Record) *Record {
	for key, value := range source.Data {
		record.Set(key, value)
	}
	record.KeysToRedact = append(record.KeysToRedact, source.KeysToRedact...)
	return record
}

// MarshalJSON marshals this into JSON
func (record Record) MarshalJSON() ([]byte, error) {
	if len(record.Data) == 0 {
		return []byte("null"), nil
	}

	var (
		buffer = bufferPool.Get()
		comma  = false
	)
	defer bufferPool.Put(buffer)

	buffer.WriteString("{")
	for key, raw := range record.Data {
		if raw == nil {
			continue
		}
		if comma {
			buffer.WriteString(",")
		} else {
			comma = true
		}
		buffer.WriteString(`"`)
		buffer.WriteString(key)
		buffer.WriteString(`":`)
		jsonValue(raw, buffer, record.KeysToRedact...)
	}
	buffer.WriteString("}")
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshals JSON into this
func (record *Record) UnmarshalJSON(payload []byte) error {
	var placeholder map[string]interface{}
	if err := json.Unmarshal(payload, &placeholder); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*record = Record{
		Data: placeholder,
		KeysToRedact: nil,
	}
	return nil
}

func jsonValue(object interface{}, buffer *bytes.Buffer, keyToRedact ...string) {
	switch value := object.(type) {
	case func() interface{}:
		object = value()
	case RedactableWithKeys:
		object = value.Redact(keyToRedact...)
	case Redactable:
		object = value.Redact()
	}
	// This looks ugly, but it goes way faster than reflection (that is used by json.Marshal)
	if errorobject, ok := object.(error); ok {
		payload, err := json.Marshal(errorobject)
		if err != nil {
			buffer.WriteString(`"`)
			buffer.Write([]byte(errorobject.Error()))
			buffer.WriteString(`"`)
		}
		buffer.Write(payload)
		return
	}

	switch value := object.(type) {
	case bool:
		buffer.WriteString(strconv.FormatBool(value))
	case *bool:
		buffer.WriteString(strconv.FormatBool(*value))
	case complex64:
		buffer.WriteString(`"`)
		buffer.WriteString(strconv.FormatComplex(complex128(value), 'g', -1, 64))
		buffer.WriteString(`"`)
	case *complex64:
		buffer.WriteString(`"`)
		buffer.WriteString(strconv.FormatComplex(complex128(*value), 'g', -1, 64))
		buffer.WriteString(`"`)
	case complex128:
		buffer.WriteString(`"`)
		buffer.WriteString(strconv.FormatComplex(value, 'g', -1, 128))
		buffer.WriteString(`"`)
	case *complex128:
		buffer.WriteString(`"`)
		buffer.WriteString(strconv.FormatComplex(*value, 'g', -1, 128))
		buffer.WriteString(`"`)
	case float32:
		buffer.WriteString(strconv.FormatFloat(float64(value), 'g', -1, 32))
	case *float32:
		buffer.WriteString(strconv.FormatFloat(float64(*value), 'g', -1, 32))
	case float64:
		buffer.WriteString(strconv.FormatFloat(value, 'g', -1, 64))
	case *float64:
		buffer.WriteString(strconv.FormatFloat(*value, 'g', -1, 64))
	case Level:
		buffer.WriteString(strconv.FormatInt(int64(value), 10))
	case int:
		buffer.WriteString(strconv.FormatInt(int64(value), 10))
	case *int:
		buffer.WriteString(strconv.FormatInt(int64(*value), 10))
	case int8:
		buffer.WriteString(strconv.FormatInt(int64(value), 10))
	case *int8:
		buffer.WriteString(strconv.FormatInt(int64(*value), 10))
	case int16:
		buffer.WriteString(strconv.FormatInt(int64(value), 10))
	case *int16:
		buffer.WriteString(strconv.FormatInt(int64(*value), 10))
	case int32:
		buffer.WriteString(strconv.FormatInt(int64(value), 10))
	case *int32:
		buffer.WriteString(strconv.FormatInt(int64(*value), 10))
	case int64:
		buffer.WriteString(strconv.FormatInt(value, 10))
	case *int64:
		buffer.WriteString(strconv.FormatInt(*value, 10))
	case string:
		buffer.WriteString(`"`)
		jsonEscape(value, buffer)
		buffer.WriteString(`"`)
	case *string:
		buffer.WriteString(`"`)
		jsonEscape(*value, buffer)
		buffer.WriteString(`"`)
	case uint:
		buffer.WriteString(strconv.FormatUint(uint64(value), 10))
	case *uint:
		buffer.WriteString(strconv.FormatUint(uint64(*value), 10))
	case uint8:
		buffer.WriteString(strconv.FormatUint(uint64(value), 10))
	case *uint8:
		buffer.WriteString(strconv.FormatUint(uint64(*value), 10))
	case uint16:
		buffer.WriteString(strconv.FormatUint(uint64(value), 10))
	case *uint16:
		buffer.WriteString(strconv.FormatUint(uint64(*value), 10))
	case uint32:
		buffer.WriteString(strconv.FormatUint(uint64(value), 10))
	case *uint32:
		buffer.WriteString(strconv.FormatUint(uint64(*value), 10))
	case uint64:
		buffer.WriteString(strconv.FormatUint(value, 10))
	case *uint64:
		buffer.WriteString(strconv.FormatUint(*value, 10))
	default:
		if payload, err := json.Marshal(object); err == nil {
			buffer.Write(payload)
		} else {
			buffer.WriteString(`"`)
			if stringer, ok := object.(fmt.Stringer); ok {
				buffer.WriteString(stringer.String())
			} else {
				buffer.WriteString(fmt.Sprintf("%+#v", object))
			}
			buffer.WriteString(`"`)
		}
	}
}

func jsonEscape(value string, buffer *bytes.Buffer) {
	for _, char := range value {
		switch char {
		case '\\':
			buffer.WriteString(`\\`)
		case '\b':
			buffer.WriteString(`\b`)
		case '\f':
			buffer.WriteString(`\f`)
		case '\n':
			buffer.WriteString(`\n`)
		case '\r':
			buffer.WriteString(`\r`)
		case '\t':
			buffer.WriteString(`\t`)
		case '"':
			buffer.WriteString(`\"`)
		default:
			buffer.WriteRune(char)
		}
	}
}
