package logger

// Redactable can be used by structs that want to redact their fields
//
// When the Logger writes the Record contains a Redactable, it will call Redact
type Redactable interface {
	Redact() interface{}
}
