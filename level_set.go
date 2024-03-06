package logger

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gildas/go-core"
)

// LevelSet is a set of levels organized by topic and scope
type LevelSet map[topicscope]Level

// NewLevelSet creates a new LevelSet with a default level
func NewLevelSet(defaultLevel Level) (levels LevelSet) {
	levels = LevelSet{}
	levels.SetDefault(defaultLevel)
	return
}

// ParseLevelsFromEnvironment parses the levels from the environment variable LOG_LEVEL
//
// If LOG_LEVEL is not set, it will return a LevelSet with the default level (INFO)
//
// If the environment variable DEBUG is set to 1 (or yes, on, true), the default level in the LevelSet will be DEBUG
func ParseLevelsFromEnvironment() (levels LevelSet) {
	levels = ParseLevels(core.GetEnvAsString("LOG_LEVEL", "INFO"))
	if core.GetEnvAsBool("DEBUG", false) {
		if levels.Get("any", "any") > DEBUG {
			levels.Set(DEBUG, "any", "any")
		}
	}
	return
}

// ParseLevels parses a string and returns a LevelSet
func ParseLevels(settings string) (levels LevelSet) {
	levels = LevelSet{}
	if len(settings) == 0 {
		levels.Set(INFO, "any", "any")
		return
	}

	// Pattern to match a list of topic/scope levels
	// See: https://regex101.com/r/GFjOje/1
	var (
		pattern     = regexp.MustCompile(`\s*(?P<LEVEL>[a-zA-Z]+)(?::\{(?P<TOPIC>\w+)?(?::(?P<SCOPES>\w+\s*(?:,\s*\w+)?))?\})?\s*(?:;\s*|$)`)
		matches     = pattern.FindAllStringSubmatch(settings, -1)
		levelGroup  = pattern.SubexpIndex("LEVEL")
		topicGroup  = pattern.SubexpIndex("TOPIC")
		scopesGroup = pattern.SubexpIndex("SCOPES")
	)

	for _, match := range matches {
		level := ParseLevel(match[levelGroup])
		topic := match[topicGroup]
		scopes := strings.Split(match[scopesGroup], ",")
		for _, scope := range scopes {
			levels.Set(level, topic, strings.TrimSpace(scope))
		}
	}
	return
}

// GetDefault gets the default level
func (levels LevelSet) GetDefault() Level {
	return levels.Get("any", "any")
}

// SetDefault sets the level for any topic and any scope
func (set *LevelSet) SetDefault(level Level) {
	set.Set(level, "any", "any")
}

// Set sets the level for a given topic,scope pair
func (set *LevelSet) Set(level Level, topic, scope string) {
	if *set == nil {
		*set = LevelSet{}
	}
	(*set)[newTopicscope(topic, scope)] = level
}

// Get returns the level for a given topic,scope pair
func (set LevelSet) Get(topic, scope string) (level Level) {
	if len(set) == 0 {
		return UNSET
	}
	if level, found := set[newTopicscope(topic, scope)]; found {
		return level
	}
	// We search topic:any first, as topics are higher on the foodchain
	if level, found := set[newTopicscope(topic, "any")]; found {
		return level
	}
	if level, found := set[newTopicscope("any", scope)]; found {
		return level
	}
	return set[newTopicscope("any", "any")]
}

// ShouldWrite Tells if the given Level, Topic, and Scope should be written
func (set LevelSet) ShouldWrite(level Level, topic, scope string) bool {
	return level.ShouldWrite(set.Get(topic, scope))
}

func (set LevelSet) String() string {
	values := []string{}
	for topicscope, level := range set {
		if topicscope.Topic == "any" && topicscope.Scope == "any" {
			values = append(values, level.String())
		} else if topicscope.Topic == "any" {
			values = append(values, fmt.Sprintf("%s:{:%s}", level.String(), topicscope.Scope))
		} else if topicscope.Scope == "any" {
			values = append(values, fmt.Sprintf("%s:{%s}", level.String(), topicscope.Topic))
		} else {
			values = append(values, fmt.Sprintf("%s:{%s:%s}", level.String(), topicscope.Topic, topicscope.Scope))
		}
	}
	return strings.Join(values, ";")
}
