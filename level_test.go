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
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
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
