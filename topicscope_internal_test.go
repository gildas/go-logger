package logger

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type TopicScopeSuite struct {
	suite.Suite
	Name string
}

func TestTopicScopeSuite(t *testing.T) {
	suite.Run(t, new(TopicScopeSuite))
}

func (suite *TopicScopeSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
}

func (suite *TopicScopeSuite) TestCanCreate() {
	ts := newTopicscope("topic", "scope")
	suite.Assert().Equal("topic", ts.Topic, "Topic should be \"topic\"")
	suite.Assert().Equal("scope", ts.Scope, "Scope should be \"scope\"")
}

func (suite *TopicScopeSuite) TestCanMatchTopicsAndScopes() {
	ts := newTopicscope("", "")
	suite.Assert().Equal("any", ts.Topic, `Topic should be "any"`)
	suite.Assert().Equal("any", ts.Scope, `Scope should be "any"`)

	suite.Assert().True(ts.Match("", ""), `Should match no topic and no scope`)
	suite.Assert().True(ts.Match("any", "any"), `Should match topic "any" and scope "any"`)
	suite.Assert().True(ts.Match("topic", "any"), `Should match topic "topic" and scope "any"`)
	suite.Assert().True(ts.Match("any", "scope"), `Should match topic "any" and scope "scope"`)
	suite.Assert().True(ts.Match("topic", "scope"), `Should match topic "topic" and scope "scope"`)

	ts = newTopicscope("topic", "")
	suite.Assert().False(ts.Match("", ""), `Should not match no topic and no scope`)
	suite.Assert().False(ts.Match("any", "any"), `Should not match topic "any" and scope "any"`)
	suite.Assert().True(ts.Match("topic", "any"), `Should match topic "topic" and scope "any"`)
	suite.Assert().False(ts.Match("any", "scope"), `Should not match topic "any" and scope "scope"`)
	suite.Assert().True(ts.Match("topic", "scope"), `Should match topic "topic" and scope "scope"`)
	suite.Assert().False(ts.Match("topic1", "scope1"), `Should not match topic "topic1" and scope "scope1"`)

	ts = newTopicscope("", "scope")
	suite.Assert().False(ts.Match("", ""), `Should not match no topic and no scope`)
	suite.Assert().False(ts.Match("any", "any"), `Should not match topic "any" and scope "any"`)
	suite.Assert().False(ts.Match("topic", "any"), `Should not match topic "topic" and scope "any"`)
	suite.Assert().True(ts.Match("any", "scope"), `Should match topic "any" and scope "scope"`)
	suite.Assert().True(ts.Match("topic", "scope"), `Should match topic "topic" and scope "scope"`)
	suite.Assert().False(ts.Match("topic1", "scope1"), `Should not match topic "topic1" and scope "scope1"`)

	ts = newTopicscope("topic", "scope")
	suite.Assert().False(ts.Match("", ""), `Should not match no topic and no scope`)
	suite.Assert().False(ts.Match("any", "any"), `Should not match topic "any" and scope "any"`)
	suite.Assert().False(ts.Match("topic", "any"), `Should not match topic "topic" and scope "any"`)
	suite.Assert().False(ts.Match("any", "scope"), `Should not match topic "any" and scope "scope"`)
	suite.Assert().True(ts.Match("topic", "scope"), `Should match topic "topic" and scope "scope"`)
	suite.Assert().False(ts.Match("topic1", "scope1"), `Should not match topic "topic1" and scope "scope1"`)
}
