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

func (suite *StructuredSuite) TestCanCreateAndDeleteTable() {
	type Mammoth struct {
		ID       string `json:"id" sql:"key,varchar(30)"`
		Unsigned uint   `sql:"there"`
	}
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Mammoth{})
	suite.Require().Nil(err, "Failed to create table for Mammoth")
	err = db.DeleteTable(Mammoth{})
	suite.Assert().Nil(err, "Failed to drop the table for Mammoth")
}

func (suite *StructuredSuite) TestCanInsert() {
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	person := &Person{"5678", "Doe", 58, db.Logger}
	err = db.Insert(person)
	suite.Assert().Nil(err)
}

func (suite *StructuredSuite) TestCanInsertAllFieldTypes() {
	type Mammoth struct {
		ID       uuid.UUID `json:"id" sql:"key"`
		Name     string    `sql:"index,varchar(60)"`
		Bool     bool      `sql:"married"`
		Age      uint
		Position int32 `sql:"pos"`
		Pointy   *int64
		Distance float64
		Day      time.Weekday
		Duration time.Duration
		Created  *time.Time
		Stamp    time.Time
	}
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Mammoth{})
	suite.Require().Nil(err, "Failed to create table for Mammoth")
	// No need to delete this table with ramsql, it will be dropped when AfterTest disconnects from the DB
	// Plus, there is a deadlock in ramsql, sometimes when a statement fails, the next statement locks on its mutex
	//defer func() {
	//	err := db.DeleteTable(Mammoth{})
	//	suite.Assert().Nil(err, "Failed to drop the table for Mammoth")
	//}()
	id := uuid.New()
	now := time.Now()
	pointy := int64(12)
	mammoth := Mammoth{id, "Doe", false, 58, -10, &pointy, 3.1415, time.Tuesday, 2 * time.Minute / time.Second, &now, now.UTC()}
	err = db.Insert(mammoth)
	suite.Require().Nil(err)
	err = db.UpdateAll(Mammoth{}, sql.Queries{}.Add("age", 18).Add("pos", sql.QuerySet, 12))
	suite.Assert().Nil(err)
	found, err := db.Find(Mammoth{}, sql.Queries{}.Add("id", id))
	suite.Require().Nil(err)
	animal, ok := found.(*Mammoth)
	suite.Require().True(ok, "The found item should be a Mammoth")
	suite.T().Logf("Mammoth: %#v", animal)
	suite.Assert().Equal(id, animal.ID)
	suite.Assert().Equal("Doe", animal.Name)
	suite.Assert().Equal(false, animal.Bool)
	suite.Assert().Equal(uint(58), animal.Age)
	suite.Assert().Equal(int32(-10), animal.Position)
	suite.Assert().Equal(int64(12), *animal.Pointy)
	suite.Assert().Equal(3.1415, animal.Distance)
	suite.Assert().Equal(time.Tuesday, animal.Day)
	suite.Assert().Equal(2*time.Minute / time.Second, animal.Duration)
	suite.Assert().Truef(now.Equal(*animal.Created), "Created \"%s\" should equal \"%s\"", *animal.Created, now)
	suite.Assert().Truef(now.UTC().Equal(animal.Stamp), "Stamp \"%s\" should equal \"%s\"", animal.Stamp, now.UTC())
}

func (suite *StructuredSuite) TestCanFind() {
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	found, err := db.FindAll(Person{}, sql.Queries{}.Add("age", sql.QueryGreater, 15))
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
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	found, err := db.Find(Person{}, sql.Queries{}.Add("id", "1234"))
	suite.Assert().Nil(err)
	person, ok := found.(*Person)
	suite.Require().True(ok, "The found item should be a person")
	suite.T().Logf("Item: %#v", person)
	suite.Assert().Equal("Doe", person.Name)
	suite.Assert().Equal(18, person.Age)
	suite.Assert().NotEmpty(person.ID)
}

func (suite *StructuredSuite) TestCanUpdate() {
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	err = db.UpdateAll(Person{}, sql.Queries{}.Add("age", 18).Add("age", sql.QuerySet, 25))
	suite.Assert().Nil(err)
}

func (suite *StructuredSuite) TestCanDelete() {
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	suite.Require().Nil(db.Insert(Person{"5678", "Doe", 58, db.Logger}))
	err = db.DeleteAll(Person{}, sql.Queries{}.Add("age", sql.QueryGreater, 50))
	suite.Assert().Nil(err)
}

func (suite *StructuredSuite) TestShouldNotCreateWithUnsupportedFields() {
	type Impossible struct {
		ID    string
		NoWay struct {
			NotHere string
		}
	}
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	err = db.CreateTable(Impossible{})
	suite.Require().NotNil(err, "Failed to create table for Impossible")
}

func (suite *StructuredSuite) TestShouldNotFindWithUnknownSchema() {
	type Parasite struct {
		ID string
	}
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	_, err = db.FindAll(Parasite{}, sql.Queries{})
	suite.Assert().NotNil(err)
	suite.Logger.Errorf("(Expected) Failed to find any item", err)
}

func (suite *StructuredSuite) TestShouldNotRetrieveWrongData() {
	type BadType struct {
		ID   string `json:"id" sql:"key,varchar(30)"`
		Over int8   `sql:",int8"` // Well, these are not the same types at all! Watch...
	}
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	err = db.CreateTable(BadType{})
	suite.Require().Nil(err, "Failed to create table for Mammoth")
	// No need to delete this table with ramsql, it will be dropped when AfterTest disconnects from the DB
	// Plus, there is a deadlock in ramsql, sometimes when a statement fails, the next statement locks on its mutex
	//defer func() {
	//	err := db.DeleteTable(BadType{})
	//	suite.Assert().Nil(err, "Failed to drop the table for Mammoth")
	//}()
	statement, parms := sql.InsertStatement{}.With(db).Build("badtype", nil, sql.Queries{}.Add("id", "1234").Add("over", 512))
	_, err = db.Exec(statement, parms...)
	suite.Assert().Nil(err, "Failed to insert data manually")

	_, err = db.FindAll(BadType{}, sql.Queries{}.Add("id", "1234"))
	suite.Assert().NotNil(err)
	suite.Logger.Errorf("(Expected) Failed to find any item", err)

	_, err = db.Find(BadType{}, sql.Queries{}.Add("id", "1234"))
	suite.Assert().NotNil(err)
	suite.Logger.Errorf("(Expected) Failed to find any item", err)
}

func (suite *StructuredSuite) TestShouldNotInsertUnsupportedTypes() {
	type Test struct {
		ID      string `json:"id" sql:"key,varchar(30)"`
		BadType complex64
	}
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	suite.Logger.Infof("Creating test Table")
	_, err = db.Exec(`CREATE TABLE test (id TEXT, badtype INT)`)
	suite.Assert().Nil(err, "Failed to execute statement")
	// No need to delete this table with ramsql, it will be dropped when AfterTest disconnects from the DB
	// Plus, there is a deadlock in ramsql, sometimes when a statement fails, the next statement locks on its mutex
	//defer func() {
	//	db.Exec("DROP TABLE test")
	//}()
	suite.Logger.Infof("Inserting data into the test Table")
	_, err = db.Exec(`INSERT INTO test (id, badtype) VALUES ($1, $2)`, "1234", 12)
	suite.Assert().Nil(err, "Failed to execute statement")
	badtype := Test{"638146", complex64(1)}
	err = db.Insert(badtype)
	suite.Assert().NotNil(err)
	suite.T().Logf("Expected Insert error: %s", err.Error())
}

func (suite *StructuredSuite) TestShouldNotFindUnknownData() {
	db, err := sql.Open("ramsql", suite.T().Name(), suite.Logger)
	suite.Require().Nil(err, "Failed to open Database")
	defer func () {
		err := db.Close()
		suite.Assert().Nil(err, "Failed to close the database")
	}()
	err = db.CreateTable(Person{})
	suite.Require().Nil(err, "Failed to create table")
	suite.Require().Nil(db.Insert(Person{"1234", "Doe", 18, db.Logger}))
	_, err = db.Find(Person{}, sql.Queries{}.Add("id", "nothere"))
	suite.Assert().NotNil(err)
	suite.Assert().True(errors.Is(err, errors.NotFound), "The error should be NotFound")
}

func (suite *StructuredSuite) TestScannerShouldComplainWithInvalidData() {
	now := time.Now()
	payload := []byte(now.Format(time.RFC822))
	scanner := &sql.DBTime{}
	err := scanner.Scan(payload)
	suite.Assert().NotNil(err)
}

// Suite Tools

func (suite *StructuredSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:        fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:  true,
			FilterLevel: logger.TRACE,
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
}

func (suite *StructuredSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *StructuredSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *StructuredSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}
