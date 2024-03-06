package logger_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/gildas/go-core"
	"github.com/gildas/go-logger"
)

type LevelSetSuite struct {
	suite.Suite
	Name string
}

func TestLevelSetSuite(t *testing.T) {
	suite.Run(t, new(LevelSetSuite))
}

func (suite *LevelSetSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
}

func (suite *LevelSetSuite) TestCanCreate() {
	levels := logger.NewLevelSet(logger.DEBUG)
	suite.NotNil(levels, "Levels should not be nil")
	suite.Equal(logger.DEBUG, levels.GetDefault(), "Default Level Should be DEBUG")
}

func (suite *LevelSetSuite) TestCanSetAndGetTopicScopeAndLevel() {
	var levels logger.LevelSet
	suite.Nil(levels, "Levels should be nil")
	suite.Equal(logger.UNSET, levels.Get("topic", "scope"), "Level should be UNSET")

	levels.SetDefault(logger.INFO)
	suite.NotNil(levels, "Levels should not be nil")
	levels.Set(logger.DEBUG, "topic1", "scope1")
	levels.Set(logger.DEBUG, "topic1", "")
	levels.Set(logger.TRACE, "", "scope2")

	suite.Assert().Equal(logger.DEBUG, levels.Get("topic1", "scope1"), "Level for topic1, scope1 should be DEBUG")
	suite.Assert().Equal(logger.DEBUG, levels.Get("topic1", "scope2"), "Level for topic1, scope2 should be DEBUG")
	suite.Assert().Equal(logger.DEBUG, levels.Get("topic1", "any"), "Level for topic1, any should be DEBUG")
	suite.Assert().Equal(logger.DEBUG, levels.Get("topic1", ""), `Level for topic1, "" should be DEBUG`)
	suite.Assert().Equal(logger.INFO, levels.Get("topic2", ""), `Level for topic2, "" should be INFO`)
	suite.Assert().Equal(logger.INFO, levels.Get("topic2", "scope1"), "Level for topic2, scope1 should be INFO")
	suite.Assert().Equal(logger.TRACE, levels.Get("topic2", "scope2"), "Level for topic2, scope2 should be TRACE")
	suite.Assert().Equal(logger.TRACE, levels.Get("topic3", "scope2"), "Level for topic3, scope2 should be TRACE")
}

func (suite *LevelSetSuite) TestCanParseLevels() {
	var levels logger.LevelSet

	levels = logger.ParseLevels("")
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.INFO, levels.Get("any", "any"), "Default Level Should be INFO")

	levels = logger.ParseLevels("INFO")
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.INFO, levels.Get("any", "any"), "Default Level Should be INFO")

	levels = logger.ParseLevels("DEBUG")
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.DEBUG, levels.Get("any", "any"), "Default Level Should be DEBUG")

	levels = logger.ParseLevels("TRACE")
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.TRACE, levels.Get("any", "any"), "Default Level Should be TRACE")

	settings := "INFO;DEBUG:{topic1:scope1,scope2};TRACE:{topic2:scope2};TRACE:{:scope3};WARN:{topic3}"
	levels = logger.ParseLevels(settings)
	suite.Require().Len(levels, 6, "Levels should have 6 entries")
	suite.Assert().Equal(logger.INFO, levels.Get("any", "any"), "Default Level Should be INFO")

	suite.Assert().Equal(logger.DEBUG, levels.Get("topic1", "scope1"), "Level for topic1, scope1 should be DEBUG")
	suite.Assert().Equal(logger.DEBUG, levels.Get("topic1", "scope2"), "Level for topic1, scope2 should be DEBUG")
	suite.Assert().Equal(logger.TRACE, levels.Get("topic1", "scope3"), "Level for topic1, scope3 should be TRACE")
	suite.Assert().Equal(logger.INFO, levels.Get("topic1", "scope4"), "Level for topic1, scope4 should be INFO")

	suite.Assert().Equal(logger.INFO, levels.Get("topic2", "scope1"), "Level for topic2, scope1 should be TRACE")
	suite.Assert().Equal(logger.TRACE, levels.Get("topic2", "scope2"), "Level for topic2, scope2 should be TRACE")
	suite.Assert().Equal(logger.TRACE, levels.Get("topic2", "scope3"), "Level for topic2, scope3 should be TRACE")
	suite.Assert().Equal(logger.INFO, levels.Get("topic2", "scope4"), "Level for topic2, scope4 should be INFO")

	suite.Assert().Equal(logger.WARN, levels.Get("topic3", ""), "Level for topic3, any should be WARN")
	suite.Assert().Equal(logger.WARN, levels.Get("topic3", "scope1"), "Level for topic3, scope1 should be WARN")
	suite.Assert().Equal(logger.WARN, levels.Get("topic3", "scope2"), "Level for topic3, scope2 should be WARN")
	suite.Assert().Equal(logger.WARN, levels.Get("topic3", "scope3"), "Level for topic3, scope3 should be WARN")
	suite.Assert().Equal(logger.WARN, levels.Get("topic3", "scope4"), "Level for topic3, scope4 should be WARN")

	suite.Assert().Equal(logger.INFO, levels.Get("topic5", "scope5"), "Level for topic5, scope5 should be INFO")
}

func (suite *LevelSetSuite) TestCanParseLevelsFromEnvironment() {
	currentLogLevel := core.GetEnvAsString("LOG_LEVEL", "")
	currentDebug := core.GetEnvAsString("DEBUG", "")
	defer func() {
		os.Setenv("LOG_LEVEL", currentLogLevel)
		os.Setenv("DEBUG", currentDebug)
	}()
	var levels logger.LevelSet

	os.Setenv("LOG_LEVEL", "")
	levels = logger.ParseLevelsFromEnvironment()
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.INFO, levels.Get("any", "any"), "Default Level Should be INFO (got: %s)", levels.Get("any", "any"))

	os.Setenv("LOG_LEVEL", "DEBUG")
	levels = logger.ParseLevelsFromEnvironment()
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.DEBUG, levels.Get("any", "any"), "Default Level Should be DEBUG (got: %s)", levels.Get("any", "any"))

	os.Setenv("LOG_LEVEL", "WARN")
	os.Setenv("DEBUG", "true")
	levels = logger.ParseLevelsFromEnvironment()
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.DEBUG, levels.Get("any", "any"), "Default Level Should be DEBUG (got: %s)", levels.Get("any", "any"))

	os.Setenv("LOG_LEVEL", "TRACE")
	os.Setenv("DEBUG", "true")
	levels = logger.ParseLevelsFromEnvironment()
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.TRACE, levels.Get("any", "any"), "Default Level Should be TRACE (got: %s)", levels.Get("any", "any"))
}

func (suite *LevelSetSuite) TestCanGetString() {
	var levels logger.LevelSet

	levels = logger.ParseLevels("")
	suite.Assert().Equal("INFO", levels.String())
	levels = logger.ParseLevels(" DEBUG ")
	suite.Assert().Equal("DEBUG", levels.String())
	levels = logger.ParseLevels("INFO; DEBUG:{topic1:scope1, scope2}; TRACE:{topic2:scope2}; TRACE:{:scope3}; WARN:{topic3}")

	values := strings.Split(levels.String(), ";")
	// expected := []string{"INFO", "DEBUG:{topic1:scope1,scope2}", "TRACE:{topic2:scope2}", "TRACE:{:scope3}", "WARN:{topic3}"}
	expected := []string{"INFO", "DEBUG:{topic1:scope1}", "DEBUG:{topic1:scope2}", "TRACE:{topic2:scope2}", "TRACE:{:scope3}", "WARN:{topic3}"}
	suite.Require().Equal(len(expected), len(values), "Should have same number of values")
	for _, expected := range expected {
		suite.Require().Containsf(values, expected, "Should contain value %s", expected)
	}
}

func (suite *LevelSetSuite) TestCanTellIfLevelIsWritable() {
	settings := "INFO;DEBUG:{topic1:scope1,scope2};TRACE:{topic2:scope2};TRACE:{:scope3};WARN:{topic3}"
	levels := logger.ParseLevels(settings)
	suite.Require().Len(levels, 6, "Levels should have 6 entries")

	suite.Assert().True(levels.ShouldWrite(logger.DEBUG, "topic1", "scope1"), "Should write DEBUG for topic1, scope1")
	suite.Assert().False(levels.ShouldWrite(logger.TRACE, "topic1", ""), "Should not write TRACE for topic1")
	suite.Assert().False(levels.ShouldWrite(logger.TRACE, "topic1", "scope1"), "Should not write TRACE for topic1, scope1")
	suite.Assert().True(levels.ShouldWrite(logger.TRACE, "topic1", "scope3"), "Should write TRACE for topic1, scope3")
	suite.Assert().False(levels.ShouldWrite(logger.INFO, "topic3", "scope1"), "Should not write INFO for topic3, scope1")
	suite.Assert().True(levels.ShouldWrite(logger.INFO, "topic4", "scope4"), "Should write INFO for topic4, scope4")
	suite.Assert().False(levels.ShouldWrite(logger.DEBUG, "topic4", "scope4"), "Should not write DEBUG for topic4, scope4")
}

func (suite *LevelSetSuite) TestCanParseLowerCase() {
	settings := "warn"
	levels := logger.ParseLevels(settings)
	suite.Require().Len(levels, 1, "Levels should have 1 entry")
	suite.Assert().Equal(logger.WARN, levels.Get("any", "any"), "Default Level Should be WARN")
}
