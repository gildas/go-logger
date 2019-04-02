package logger

import (
	"github.com/chakrit/go-bunyan"
)

// ThreadIdInfo returns an Info that contains the current thread identifier
func ThreadIdInfo() bunyan.Info {
	return bunyan.InfoFunc("tid", func() interface{} { return Gettid() })
}

// TopicInfo returns an Info that contains the current topic
func TopicInfo(topic string) bunyan.Info {
	return bunyan.SimpleInfo("topic", topic)
}

// ScopeInfo returns an Info that contains the current scope
func ScopeInfo(scope string) bunyan.Info {
	return bunyan.SimpleInfo("scope", scope)
}
