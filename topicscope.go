package logger

// topicscope is a topic and scope pair
type topicscope struct {
	Topic string
	Scope string
}

// newTopicscope creates a new topicscope
func newTopicscope(topic, scope string) (ts topicscope) {
	if len(topic) == 0 || topic == "*" {
		ts.Topic = "any"
	} else {
		ts.Topic = topic
	}
	if len(scope) == 0 || scope == "*" {
		ts.Scope = "any"
	} else {
		ts.Scope = scope
	}
	return
}

func (ts topicscope) Match(topic, scope string) bool {
	return (ts.Topic == "any" || ts.Topic == topic) && (ts.Scope == "any" || ts.Scope == scope)
}
