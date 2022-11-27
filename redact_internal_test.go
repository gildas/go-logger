package logger

import "github.com/gildas/go-core"

func (suite *InternalLoggerSuite) TestCanCreateWithRedactors() {
	log := Create(
		"test",
		*(core.Must(NewRedactor("[0-9]{8}"))),
		core.Must(NewRedactor("[a-z]{8}")),
	)
	suite.Require().NotNil(log, "cannot create a Logger with redactors")
	suite.Require().Len(log.redactors, 2, "The Logger should have 2 redactors")
}

func (suite *InternalLoggerSuite) TestCanCreateChildLoggerWithRedactors() {
	log := Create(
		"test",
		core.Must(NewRedactor("[0-9]{8}")),
		core.Must(NewRedactor("[a-z]{8}")),
	)
	suite.Require().NotNil(log, "cannot create a Logger with redactors")
	suite.Require().Len(log.redactors, 2, "The Logger should have 2 redactors")

	child := log.Child(nil, nil, "mylabel", "myvalue", core.Must(NewRedactor("[A-Z]{8}")))
	suite.Require().NotNil(child, "cannot create a child Logger")
	suite.Assert().Len(child.redactors, 3, "The Child Logger should have 3 redactors")
	suite.Assert().Len(log.redactors, 2, "The Parent Logger should have 2 redactors")

	child2 := log.Child(nil, nil, "mylabel", 3, *(core.Must(NewRedactor("[A-Z]{8}"))))
	suite.Require().NotNil(child2, "cannot create a child Logger")
	suite.Assert().Len(child.redactors, 3, "The Child Logger should have 3 redactors")
	suite.Assert().Len(log.redactors, 2, "The Parent Logger should have 2 redactors")
}
