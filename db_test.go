package sql_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-logger"
	"github.com/gildas/go-sql"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"
	_ "github.com/proullon/ramsql/driver"
)

type DBSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time
}

func TestDBSuite(t *testing.T) {
	suite.Run(t, new(DBSuite))
}

func (suite *DBSuite) TestCanOpen() {
	db, err := sql.Open("ramsql", "", suite.Logger)
	suite.Assert().Nil(err)
	suite.Assert().NotNil(db)
	err = db.Close()
	suite.Assert().Nil(err, "Failed to close the database")
}

func (suite *DBSuite) TestCanPing() {
	db, err := sql.Open("ramsql", "", suite.Logger)
	suite.Assert().Nil(err)
	suite.Assert().NotNil(db)

	err = db.Ping()
	suite.Assert().Nil(err)
	err = db.Close()
	suite.Assert().Nil(err, "Failed to close the database")
}

func (suite *DBSuite) TestCanStoreAndRetrieveInContext() {
	db, err := sql.Open("ramsql", "", suite.Logger)
	suite.Assert().Nil(err)
	suite.Assert().NotNil(db)
	defer db.Close()

	ctx := db.ToContext(context.Background())
	_, err = sql.FromContext(ctx)
	suite.Assert().Nil(err)

	_ = sql.Must(sql.FromContext(ctx))
}

func (suite *DBSuite) TestFailsWhenNotStoredInContext() {
	ctx := context.Background()
	_, err := sql.FromContext(ctx)
	suite.Assert().NotNil(err)

	defer func() {
		if r := recover(); r == nil {
			suite.T().Error("Should have panicked")
		}
	}()
	_ = sql.Must(sql.FromContext(ctx))
}

func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.FromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func (suite *DBSuite) TestCanBePassedViaHttpHandler() {
	db, err := sql.Open("ramsql", "", suite.Logger)
	suite.Assert().Nil(err)
	suite.Assert().NotNil(db)
	defer db.Close()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	suite.Require().Nil(err, "Failed to create an HTTP Request")

	recorder := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Methods("GET").Path("/").Handler(db.HttpHandler()(dummyHandler()))
	router.ServeHTTP(recorder, req)
}

// Suite Tools

func (suite *DBSuite) SetupSuite() {
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

func (suite *DBSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
}

func (suite *DBSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *DBSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}