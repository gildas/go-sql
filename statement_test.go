package sql_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-logger"
	"github.com/gildas/go-sql"
	"github.com/stretchr/testify/suite"
)

type StatementSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time
}

func TestStatementSuite(t *testing.T) {
	suite.Run(t, new(StatementSuite))
}

func (suite *StatementSuite) TestCanBuildDelete() {
	queries := sql.Queries{}.Add("id", "abcd1235").Add("age", sql.QueryGreater, 18)
	statement := sql.DeleteStatement{}
	suite.Require().NotNil(statement)
	stmt, parms := statement.Build("person", nil, queries)
	suite.Assert().NotEmpty(stmt)
	suite.Assert().Len(parms, 2)
	suite.T().Logf("Statement: %s, parms: %#v", stmt, parms)
}

func (suite *StatementSuite) TestCanBuildDeleteAll() {
	statement := sql.DeleteStatement{}
	suite.Require().NotNil(statement)
	stmt, parms := statement.Build("person", nil, sql.Queries{})
	suite.Assert().NotEmpty(stmt)
	suite.Assert().Len(parms, 0)
	suite.Assert().Equal("DELETE FROM person", stmt)
	suite.T().Logf("Statement: %s", stmt)
}

func (suite *StatementSuite) TestCanBuildInsert() {
	queries := sql.Queries{}.Add("id", "abcd1235").Add("age", 18)
	statement := sql.InsertStatement{}
	suite.Require().NotNil(statement)
	stmt, parms := statement.Build("person", nil, queries)
	suite.Assert().NotEmpty(stmt)
	suite.Assert().Len(parms, 2)
	suite.T().Logf("Statement: %s, parms: %#v", stmt, parms)
}

func (suite *StatementSuite) TestCanBuildSelect() {
	columns := []string{"id", "name", "age"}
	queries := sql.Queries{}.Add("id", "abcd1235").Add("age", sql.QueryGreater, 18)
	statement := sql.SelectStatement{}
	suite.Require().NotNil(statement)
	stmt, parms := statement.Build("person", columns, queries)
	suite.Assert().NotEmpty(stmt)
	suite.Assert().Len(parms, 2)
	suite.T().Logf("Statement: %s, parms: %#v", stmt, parms)
}

func (suite *StatementSuite) TestCanBuildSelectAll() {
	columns := []string{"id", "name", "age"}
	statement := sql.SelectStatement{}
	suite.Require().NotNil(statement)
	stmt, parms := statement.Build("person", columns, sql.Queries{})
	suite.Assert().NotEmpty(stmt)
	suite.Assert().Len(parms, 0)
	suite.T().Logf("Statement: %s", stmt)
}

func (suite *StatementSuite) TestCanBuildUpdate() {
	queries := sql.Queries{}.Add("id", "abcd1235").Add("age", sql.QueryGreater, 18).Add("age", sql.QuerySet, 25)
	statement := sql.UpdateStatement{}
	suite.Require().NotNil(statement)
	stmt, parms := statement.Build("person", nil, queries)
	suite.Assert().NotEmpty(stmt)
	suite.Assert().Len(parms, 3)
	suite.T().Logf("Statement: %s, parms: %#v", stmt, parms)
}

func (suite *StatementSuite) TestCannotBuildUpdateWithEmptyQueries() {
	queries := sql.Queries{}.Add("age", sql.QuerySet, 25)
	statement := sql.UpdateStatement{}
	suite.Require().NotNil(statement)
	stmt, parms := statement.Build("person", nil, queries)
	suite.Assert().Empty(stmt)
	suite.Assert().Len(parms, 0)
	suite.T().Logf("Statement: %s, parms: %#v", stmt, parms)
}

// Suite Tools

func (suite *StatementSuite) SetupSuite() {
	var err error
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path: fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered: true,
			FilterLevel: logger.TRACE,
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
	suite.Assert().Nil(err)
}

func (suite *StatementSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
}

func (suite *StatementSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *StatementSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}