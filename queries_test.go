package sql_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-logger"
	"github.com/gildas/go-sql"
	"github.com/stretchr/testify/suite"
)

type QueriesTest struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time
}

func TestQueriesTest(t *testing.T) {
	suite.Run(t, new(QueriesTest))
}

func (suite *QueriesTest) TestCanCreateFromURL() {
	u, _ := url.Parse("https://www.acme.com/api/v1/persons?age=18&name=Doe")
	queries := sql.QueriesFromURL(u)
	suite.Assert().NotNil(queries)
	suite.Assert().Len(queries, 2, "There should be 2 sets of values in this Queries")
}

func (suite *QueriesTest) TestCanCreateFromRequest() {
	r := httptest.NewRequest(http.MethodGet, "/api/v1/persons?age=18&name=Doe", nil)
	queries := sql.QueriesFromRequest(r)
	suite.Assert().NotNil(queries)
	suite.Assert().Len(queries, 2, "There should be 2 sets of values in this Queries")
}

func (suite *QueriesTest) TestCanAddValues() {
	queries := sql.Queries{}
	queries.Add("one", 1).Add("two", 2, 2).Add("three", 3, 3, 3)
	suite.Require().Len(queries, 3, "There should be 3 sets of values in this Queries")
	suite.Assert().Len(queries["one"],   2, "There should be 1 value in queries[0] and 1 operator")
	suite.Assert().Equal(sql.QueryEqual, queries["one"][0], "The operator for one should be Equal")
	suite.Assert().Len(queries["two"],   3, "There should be 2 values in queries[1] and 1 operator")
	suite.Assert().Equal(sql.QueryIn, queries["two"][0], "The operator for two should be In")
	suite.Assert().Len(queries["three"], 4, "There should be 3 values in queries[2] and 1 operator")
	suite.Assert().Equal(sql.QueryIn, queries["three"][0], "The operator for three should be In")
}

func (suite *QueriesTest) TestCanAddValuesToSameKey() {
	queries := sql.Queries{}
	queries.Add("one", 1).Add("two", 2).Add("two", 2)
	suite.Require().Len(queries, 2, "There should be 2 sets of values in this Queries")
	suite.Assert().Len(queries["one"],   2, "There should be 1 value in queries[0] and 1 operator")
	suite.Assert().Equal(sql.QueryEqual, queries["one"][0], "The operator for one should be Equal")
	suite.Assert().Len(queries["two"],   3, "There should be 2 values in queries[1] and 1 operator")
	suite.Assert().Equal(sql.QueryIn, queries["two"][0], "The operator for two should be In")
}

func (suite *QueriesTest) TestCanAddValuesWithOperator() {
	queries := sql.Queries{}
	queries.Add("one", sql.QueryEqual, 1).Add("two", sql.QueryGreater, 2)
	suite.Require().Len(queries, 2, "There should be 2 sets of values in this Queries")
	suite.Assert().Len(queries["one"],   2, "There should be 2 values in queries[0] and 1 operator")
	suite.Assert().IsType(sql.QueryOperator{}, queries["one"][0], "The first value of queries[0] should be a QueryOperator")
	suite.Assert().Len(queries["two"],   2, "There should be 2 values in queries[1] and 1 operator")
}

func (suite *QueriesTest) TestIgnoreEmptyValueArrays() {
	queries := sql.Queries{}
	queries.Add("one", 1).Add("empty")
	suite.Require().Len(queries, 1, "There should be 1 set of values in this Queries")
}

func (suite *QueriesTest) TestCanBuildWhereClause() {
	queries := sql.Queries{}
	queries.Add("one", sql.QueryEqual, 1).Add("two", sql.QueryGreater, 2).Add("deux", 2, 2).Add("three", 3, 6, 9).Add("four", sql.QueryDifferent)
	suite.Require().Len(queries, 5, "There should be 5 sets of values in this Queries")
	where, parms := queries.WhereClause()
	components := strings.Split(where, " AND ")
	suite.Assert().Len(components, 4, "There should be 4 clauses")
	suite.T().Log(where)
	suite.Assert().Len(parms, 7, "There should be 7 parameters")
}

// Suite Tools

func (suite *QueriesTest) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path: fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered: true,
			FilterLevel: logger.TRACE,
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
}

func (suite *QueriesTest) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
}

func (suite *QueriesTest) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *QueriesTest) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}