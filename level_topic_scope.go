package logger

// TopicScopeLevels stores the level for a given topic,scope pair
type TopicScopeLevels map[string]Level

// Set sets the level for a given topic,scope pair
func (levels *TopicScopeLevels) Set(topic, scope string, level Level) {
	if *levels == nil {
		*levels = TopicScopeLevels{}
	}
	if len(scope) > 0 {
		(*levels)[topic+"|"+scope] = level
	} else {
		(*levels)[topic] = level
	}
}

// Get returns the level for a given topic,scope pair
//
// If the scope is empty, it will return the level for the topic, if found
func (levels TopicScopeLevels) Get(topic, scope string) (level Level, found bool) {
	if levels == nil || len(topic) == 0 {
		return UNSET, false
	}
	if len(scope) > 0 {
		if level, found = levels[topic + "|" + scope]; found {
			return
		}
	}
	level, found = levels[topic]
	return
}
