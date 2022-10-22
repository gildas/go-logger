package logger_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/gildas/go-logger"
)

type LevelSuite struct {
	suite.Suite
	Name string
}

func TestLevelSuite(t *testing.T) {
	suite.Run(t, new(LevelSuite))
}

func (suite *LevelSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
}

func (suite *LevelSuite) TestCanParseLevel() {
	suite.Assert().Equal(logger.INFO, logger.ParseLevel("INFO"), "Failed to parse level \"INFO\"")
	suite.Assert().Equal(logger.INFO, logger.ParseLevel("info"), "Failed to parse level \"INFO\"")
	suite.Assert().Equal(logger.UNSET, logger.ParseLevel("UNSET"), "Failed to parse level \"UNSET\"")
	suite.Assert().Equal(logger.NEVER, logger.ParseLevel("whatever"), "Failed to parse level \"whatever\" => NEVER")
}

func (suite *LevelSuite) TestCanGetLevelFromEnvironment() {
	os.Unsetenv("LOG_LEVEL")
	level := logger.GetLevelFromEnvironment()
	suite.Assert().Equal(logger.INFO, level, "Level should be INFO before being set in the environment")
	os.Setenv("DEBUG", "1")
	level = logger.GetLevelFromEnvironment()
	suite.Assert().Equal(logger.DEBUG, level, "Level should be INFO before being set in the environment and DEBUG is set")
	os.Setenv("LOG_LEVEL", "WARN")
	level = logger.GetLevelFromEnvironment()
	suite.Assert().Equal(logger.WARN, level, "Level should be WARN after being set in the environment")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DEBUG")
}

func (suite *LevelSuite) TestCanSetTopicScopeLevel() {
	var levels logger.TopicScopeLevels
	suite.Require().Nil(levels) // Just to show that it is not created

	_, found := levels.Get("any", "any")
	suite.Assert().False(found, "The level for any:any should not be found in a nil TopicScopeLevels")

	levels.Set("topic1", "scope1", logger.DEBUG)
	suite.Require().NotNil(levels, "Levels should not be nil after setting a level")
	levels.Set("topic1", "", logger.DEBUG)

	level, found := levels.Get("topic1", "scope1")
	suite.Assert().True(found, "The level for topic1:scope1 should be found")
	suite.Assert().Equal(logger.DEBUG, level, "The level for topic1:scope1 should be DEBUG")

	level, found = levels.Get("topic1", "any")
	suite.Assert().True(found, "The level for topic1:any should be found")
	suite.Assert().Equal(logger.DEBUG, level, "The level for topic1:any should be DEBUG")

	level, found = levels.Get("topic1", "")
	suite.Assert().True(found, "The level for topic1 should be found")
	suite.Assert().Equal(logger.DEBUG, level, "The level for topic1 should be DEBUG")
}

func (suite *LevelSuite) TestCanGetLevelsFromString() {
	level, levels := logger.GetLevelsFromString("")
	suite.Assert().Equal(logger.INFO, level, "The default level should be INFO")
	suite.Assert().Len(levels, 0, "There should be no levels")

	level, levels = logger.GetLevelsFromString("WARN")
	suite.Assert().Equal(logger.WARN, level, "The default level should be WARN")
	suite.Assert().Len(levels, 0, "There should be no levels")

	level, levels = logger.GetLevelsFromString("DEBUG;TRACE:{topic1}")
	suite.Assert().Equal(logger.DEBUG, level, "The default level should be DEBUG")
	suite.Assert().Len(levels, 1, "There should be 1 item in levels")
	for _, topic_scope := range [][]string{{"topic1", "", "TRACE"}, {"topic1", "any", "TRACE"}} {
		topic := topic_scope[0]
		scope := topic_scope[1]
		expected := logger.ParseLevel(topic_scope[2])
		level, found := levels.Get(topic, scope)
		suite.Assert().Truef(found, "The level for %s:%s should be found", topic, scope)
		suite.Assert().Equalf(expected, level, "The level for %s:%s should be %s", topic, scope, expected)
	}
	for _, topic_scope := range [][]string{{"topic2", ""}, {"topic2", "any"}} {
		topic := topic_scope[0]
		scope := topic_scope[1]
		_, found := levels.Get(topic, scope)
		suite.Assert().Falsef(found, "The level for %s:%s should not be found", topic, scope)
	}

	level, levels = logger.GetLevelsFromString("DEBUG;TRACE:{topic1:scope1}")
	suite.Assert().Equal(logger.DEBUG, level, "The default level should be DEBUG")
	suite.Assert().Len(levels, 1, "There should be 1 item in levels")
	for _, topic_scope := range [][]string{{"topic1", "scope1", "TRACE"}} {
		topic := topic_scope[0]
		scope := topic_scope[1]
		expected := logger.ParseLevel(topic_scope[2])
		level, found := levels.Get(topic, scope)
		suite.Assert().Truef(found, "The level for %s:%s should be found", topic, scope)
		suite.Assert().Equalf(expected, level, "The level for %s:%s should be %s", topic, scope, expected)
	}
	for _, topic_scope := range [][]string{{"topic1", ""}, {"topic1", "any"},{"topic2", ""}, {"topic2", "any"}} {
		topic := topic_scope[0]
		scope := topic_scope[1]
		_, found := levels.Get(topic, scope)
		suite.Assert().Falsef(found, "The level for %s:%s should not be found", topic, scope)
	}
}

func (suite *LevelSuite) TestCanGetLevelsFromEnvironment() {
	if current, ok := os.LookupEnv("LOG_LEVEL"); ok {
		defer os.Setenv("LOG_LEVEL", current)
	} else {
		defer os.Unsetenv("LOG_LEVEL")
	}
	if current, ok := os.LookupEnv("DEBUG"); ok {
		defer os.Setenv("DEBUG", current)
	} else {
		defer os.Unsetenv("DEBUG")
	}
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DEBUG")
	level, levels := logger.GetLevelsFromEnvironment()
	suite.Assert().Equal(logger.INFO, level, "The default level should be INFO")
	suite.Assert().Len(levels, 0, "There should be no levels")

	os.Setenv("LOG_LEVEL", "DEBUG:{topic1:scope1,scope2};WARN:{topic2:scope2};TRACE:{topic3};WARN")
	level, levels = logger.GetLevelsFromEnvironment()

	suite.Assert().Equal(logger.WARN, level, "The default level should be WARN")
	suite.Assert().Len(levels, 4, "There should be 4 levels")

	for _, topic_scope := range [][]string{{"topic1", "scope1", "DEBUG"}, {"topic1", "scope2", "DEBUG"}, {"topic2", "scope2", "WARN"}, {"topic3", "", "TRACE"}, {"topic3", "any", "TRACE"}} {
		topic := topic_scope[0]
		scope := topic_scope[1]
		expected := logger.ParseLevel(topic_scope[2])
		level, found := levels.Get(topic, scope)
		suite.Assert().Truef(found, "The level for %s:%s should be found", topic, scope)
		suite.Assert().Equalf(expected, level, "The level for %s:%s should be %s", topic, scope, expected)
	}

	for _, topic_scope := range [][]string{{"topic1", ""}, {"topic1", "any"}, {"topic2", ""}, {"topic2", "any"}, {"topic4", ""}} {
		topic := topic_scope[0]
		scope := topic_scope[1]
		_, found := levels.Get(topic, scope)
		suite.Assert().Falsef(found, "The level for %s:%s should not be found", topic, scope)
	}

	// The environment variable DEBUG should override the environment LOG_LEVEL
	os.Setenv("DEBUG", "1")
	os.Setenv("LOG_LEVEL", "DEBUG:{topic1:scope1,scope2};WARN:{topic2:scope2};TRACE:{topic3};WARN")
	level, levels = logger.GetLevelsFromEnvironment()

	suite.Assert().Equal(logger.DEBUG, level, "The default level should be DEBUG")
	suite.Assert().Len(levels, 4, "There should be 4 levels")

	os.Setenv("LOG_LEVEL", "DEBUG:{topic1:scope1,scope2};WARN:{topic2:scope2};TRACE:{topic3}")
	level, levels = logger.GetLevelsFromEnvironment()

	suite.Assert().Equal(logger.DEBUG, level, "The default level should be DEBUG")
	suite.Assert().Len(levels, 4, "There should be 4 levels")
}

func (suite *LevelSuite) TestCanStringLevel() {
	suite.Assert().Equal("INFO", logger.INFO.String())
	suite.Assert().Equal("ERROR", logger.ERROR.String())
	bigLevel := logger.Level(220)
	suite.Assert().Equal("ALWAYS", bigLevel.String())
	weirdLevel := logger.Level(47)
	suite.Assert().Equal("WARN", weirdLevel.String())
	neverLevel := logger.UNSET
	suite.Assert().Equal("UNSET", neverLevel.String())
}

func (suite *LevelSuite) TestCanCompareLevels() {
	levels := []logger.Level{logger.UNSET, logger.TRACE, logger.DEBUG, logger.INFO, logger.WARN, logger.ERROR, logger.FATAL, logger.ALWAYS}
	var filter logger.Level

	for index, filter := range levels[1:6] {
		for _, level := range levels[:index] {
			suite.Assert().Falsef(level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
		}
		for _, level := range levels[index+1:] {
			suite.Assert().Truef(level.ShouldWrite(filter), "%s should be writable with %s filter", level, filter)
		}
	}

	filter = logger.NEVER
	for _, level := range levels {
		suite.Assert().Falsef(level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
	}

	filter = logger.ALWAYS
	for _, level := range levels[:0] {
		suite.Assert().Falsef(level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
	}
	for _, level := range levels[1:] {
		suite.Assert().Truef(level.ShouldWrite(filter), "%s should be writable with %s filter", level, filter)
	}

	filter = logger.UNSET
	for _, level := range levels[:0] {
		suite.Assert().Falsef(level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
	}
	for _, level := range levels[1:] {
		suite.Assert().Truef(level.ShouldWrite(filter), "%s should be writable with %s filter", level, filter)
	}
}

func (suite *LevelSuite) TestCanGetNextLevel() {
	level := logger.DEBUG
	suite.Assert().Equal(logger.INFO, level.Next())
	level = logger.ALWAYS
	suite.Assert().Equal(logger.ALWAYS, level.Next())
}

func (suite *LevelSuite) TestCanGetPreviousLevel() {
	level := logger.DEBUG
	suite.Assert().Equal(logger.TRACE, level.Previous())
	level = logger.NEVER
	suite.Assert().Equal(logger.NEVER, level.Previous())
}
