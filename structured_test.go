package sql_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/gildas/go-sql"
	"github.com/google/uuid"
	_ "github.com/proullon/ramsql/driver"
	"github.com/stretchr/testify/suite"
)

type StructuredSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time
	DB     *sql.DB
}

type Person struct {
	ID     string         `json:"id"   sql:"key"`
	Name   string         `json:"name" sql:"index"`
	Age    int            `json:"age"`
	Logger *logger.Logger `json:"-"    sql:"-"`
}

func TestStructuredSuite(t *testing.T) {
	suite.Run(t, new(StructuredSuite))
}

func (suite *StructuredSuite) TestCanCreate() {
	type Mammoth struct {
		ID       string `json:"id" sql:"key,varchar(30)"`
		Unsigned uint   `sql:"there"`
	}
	err := suite.DB.CreateTable(Mammoth{})
	suite.Require().Nil(err, "Failed to create table for Mammoth")
	defer func() {
		err := suite.DB.DeleteTable(Mammoth{})
		suite.Assert().Nil(err, "Failed to drop the table for Mammoth")
	}()
}

func (suite *StructuredSuite) TestCanInsert() {
	person := &Person{"5678", "Doe", 58, suite.DB.Logger}
	err := suite.DB.Insert(person)
	suite.Assert().Nil(err)
}

func (suite *StructuredSuite) TestCanInsertAllFieldTypes() {
	type Mammoth struct {
		ID       uuid.UUID `json:"id" sql:"key"`
		Name     string    `sql:"index,varchar(60)"`
		Bool     bool      `sql:"married"`
		Age      uint
		Position int32     `sql:"pos"`
		Duration float64
	}
	err := suite.DB.CreateTable(Mammoth{})
	suite.Require().Nil(err, "Failed to create table for Mammoth")
	defer func() {
		err := suite.DB.DeleteTable(Mammoth{})
		suite.Assert().Nil(err, "Failed to drop the table for Mammoth")
	}()
	id := uuid.New()
	mammoth := Mammoth{ID: id,Name: "Doe", Bool: false, Age: 58, Position: -10, Duration: 3.1415}
	err = suite.DB.Insert(mammoth)
	suite.Assert().Nil(err)
	err = suite.DB.UpdateAll(Mammoth{}, sql.Queries{}.Add("age", 18).Add("pos", sql.QuerySet, 12))
	suite.Assert().Nil(err)
	found, err := suite.DB.Find(Mammoth{}, sql.Queries{}.Add("id", id))
	suite.Assert().Nil(err)
	animal, ok := found.(*Mammoth)
	suite.Require().True(ok, "The found item should be a Mammoth")
	suite.T().Logf("Mammoth: %#v", animal)
	suite.Assert().Equal("Doe", animal.Name)
	suite.Assert().Equal(id, animal.ID)
}

func (suite *StructuredSuite) TestCanFind() {
	found, err := suite.DB.FindAll(Person{}, sql.Queries{}.Add("age", sql.QueryGreater, 15))
	suite.Assert().Nil(err)
	suite.Assert().Len(found, 1)
	person, ok := found[0].(*Person)
	suite.Require().True(ok, "The first found item should be a person")
	suite.T().Logf("Item: %#v", person)
	suite.Assert().Equal("Doe", person.Name)
	suite.Assert().Equal(18, person.Age)
	suite.Assert().NotEmpty(person.ID)
}

func (suite *StructuredSuite) TestCanFindOne() {
	found, err := suite.DB.Find(Person{}, sql.Queries{}.Add("id", "1234"))
	suite.Assert().Nil(err)
	person, ok := found.(*Person)
	suite.Require().True(ok, "The found item should be a person")
	suite.T().Logf("Item: %#v", person)
	suite.Assert().Equal("Doe", person.Name)
	suite.Assert().Equal(18, person.Age)
	suite.Assert().NotEmpty(person.ID)
}

func (suite *StructuredSuite) TestCanUpdate() {
	err := suite.DB.UpdateAll(Person{}, sql.Queries{}.Add("age", 18).Add("age", sql.QuerySet, 25))
	suite.Assert().Nil(err)
}

func (suite *StructuredSuite) TestCanDelete() {
	suite.Require().Nil(suite.DB.Insert(Person{"5678", "Doe", 58, suite.DB.Logger}))
	err := suite.DB.DeleteAll(Person{}, sql.Queries{}.Add("age", sql.QueryGreater, 50))
	suite.Assert().Nil(err)
}

func (suite *StructuredSuite) TestShouldNotCreateWithUnsupportedFields() {
	type Impossible struct {
		ID string
		NoWay struct {
			NotHere string
		}
	}
	err := suite.DB.CreateTable(Impossible{})
	suite.Require().Nil(err, "Failed to create table for Impossible")
	defer func() {
		err := suite.DB.DeleteTable(Impossible{})
		suite.Assert().Nil(err, "Failed to drop the table for Impossible")
	}()
}

func (suite *StructuredSuite) TestShouldNotFindWithUnknownSchema() {
	type Parasite struct {
		ID string
	}
	_, err := suite.DB.FindAll(Parasite{}, sql.Queries{})
	suite.Assert().NotNil(err)
	suite.Logger.Errorf("(Expected) Failed to find any item", err)
}

func (suite *StructuredSuite) TestShouldNotRetrieveWrongData() {
	type BadType struct {
		ID   string `json:"id" sql:"key,varchar(30)"`
		Over int8   `sql:",int8"` // Well, these are not the same types at all! Watch...
	}
	err := suite.DB.CreateTable(BadType{})
	suite.Require().Nil(err, "Failed to create table for Mammoth")
	defer func() {
		err := suite.DB.DeleteTable(BadType{})
		suite.Assert().Nil(err, "Failed to drop the table for Mammoth")
	}()
	statement, parms := sql.InsertStatement{}.With(suite.DB).Build("badtype", nil, sql.Queries{}.Add("id", "1234").Add("over", 512))
	_, err = suite.DB.Exec(statement, parms...)
	suite.Assert().Nil(err, "Failed to insert data manually")

	_, err = suite.DB.FindAll(BadType{}, sql.Queries{}.Add("id", "1234"))
	suite.Assert().NotNil(err)
	suite.Logger.Errorf("(Expected) Failed to find any item", err)

	_, err = suite.DB.Find(BadType{}, sql.Queries{}.Add("id", "1234"))
	suite.Assert().NotNil(err)
	suite.Logger.Errorf("(Expected) Failed to find any item", err)
}

func (suite *StructuredSuite) TestShouldNotInsertUnsupportedTypes() {
	type BadType struct {
		ID       string  `json:"id" sql:"key,varchar(30)"`
		Imagine  complex64
	}
	err := suite.DB.CreateTable(BadType{})
	suite.Require().Nil(err, "Failed to create table for Mammoth")
	defer func() {
		err := suite.DB.DeleteTable(BadType{})
		suite.Assert().Nil(err, "Failed to drop the table for Mammoth")
	}()
	badtype := BadType{ID: "638146", Imagine: complex64(1)}
	err = suite.DB.Insert(badtype)
	suite.Assert().Nil(err)
}

func (suite *StructuredSuite) TestShouldNotFindUnknownData() {
	_, err := suite.DB.Find(Person{}, sql.Queries{}.Add("id", "nothere"))
	suite.Assert().NotNil(err)
	suite.Assert().True(errors.Is(err, errors.NotFound), "The error should be NotFound")
}

// Suite Tools

func (suite *StructuredSuite) SetupSuite() {
	var err error
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:        fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:  true,
			FilterLevel: logger.TRACE,
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
	suite.DB, err = sql.Open("ramsql", "TestGoSQL", suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	err = suite.DB.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
}

func (suite *StructuredSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
		suite.Logger.Infof("Dropping the person table")
		err := suite.DB.DeleteTable(Person{})
		suite.Assert().Nil(err, "Failed to drop the table for Person")
	}
	suite.Logger.Infof("Disconecting from the database")
	err := suite.DB.Close()
	suite.Assert().Nil(err, "Failed to close the database")
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *StructuredSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
	suite.Require().Nil(suite.DB.Insert(Person{"1234", "Doe", 18, suite.DB.Logger}))
}

func (suite *StructuredSuite) AfterTest(suiteName, testName string) {
	suite.Logger.Infof("Deleting data from person table")
	suite.Require().Nil(suite.DB.DeleteAll(Person{}, sql.Queries{}))
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}
