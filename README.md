# go-sql

[![GoDoc](https://godoc.org/github.com/gildas/go-sql?status.svg)](https://godoc.org/github.com/gildas/go-sql)
go-sql is an SQL library that adds Higher Level capabilities to the standard `database/sql` library.

|  |   |   |   |
---|---|---|---|
master | [![Build Status](https://dev.azure.com/keltiek/gildas/_apis/build/status/gildas.go-sql?branchName=master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=5&branchName=master) | [![Tests](https://img.shields.io/azure-devops/tests/keltiek/gildas/5/master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=5&branchName=master) | [![coverage](https://img.shields.io/azure-devops/coverage/keltiek/gildas/5/master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=5&branchName=master&view=codecoverage-tab)  
dev | [![Build Status](https://dev.azure.com/keltiek/gildas/_apis/build/status/gildas.go-sql?branchName=dev)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=5&branchName=dev) | [![Tests](https://img.shields.io/azure-devops/tests/keltiek/gildas/5/dev)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=5&branchName=dev) | [![coverage](https://img.shields.io/azure-devops/coverage/keltiek/gildas/5/dev)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=5&branchName=dev&view=codecoverage-tab)  

## Usage

Example:  
```go
package main

import (
    "github.com/gildas/g-logger"
    "github.com/gildas/g-sql"
)

type Person struct {
    ID        string `json:"id"        sql:"key"`
    Lastname  string `json:"lastName"  sql:"index,varchar(80)"`
    Firstname string `json:"firstName" sql:"index,varchar(80)"`
    Age       int    `json:"age"`
    NotNeeded int    `                 sql:"-"`
}

func main() {
    db, err := sql.Open("postgres", "connectionstring", logger.Create("MYAPP"))
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Create the SQL table "person" after the type Person
    err = db.CreateTable(Person{})

    // Insert some data
    err = db.Insert(Person{"1234", "Doe", "John", 34, 314})
    err = db.Insert(Person{"4567", "Doe", "Jane", 15, 314})

    // Find data!
    results, err := db.FindAll(Person{}, sql.Queries{}.Add("lastname", "Doe"))

    for _, result := range results {
        if person, ok := result.(*Person); ok {
            fmt.Printf("%s %s, age=%d\n", person.Lastname, person.Firstname, person.Age)
        }
    }

    // Find Single data
    result, err := db.Find(Person{}, sql.Queries{}.Add("lastname", "Doe").Add("age", sql.QueryGreater, 18))
    fmt.Printf("%s %s, age=%d\n", person.Lastname, person.Firstname, person.Age)

    // Update data
	err := suite.DB.UpdateAll(Person{}, sql.Queries{}.Add("age", 18).Add("age", sql.QuerySet, 25))

    // Delete data
	err := suite.DB.DeleteAll(Person{}, sql.Queries{}.Add("age", sql.QueryGreater, 50))

    // Drop the table
    err := suite.DB.DeleteTable(Person{})
}
```

For the sake of readability, I removed the error management. DON'T!

As you can see, using actual GO struct types is rather easy now. We do not support the entire set of SQL types or GO types, but we have the basics.  

We also support `time.Time`, [`uuid.UUID`](https://pkg.go.dev/github.com/google/uuid), pointers for simple types, and foreign keys, this is how you would use these:  
```go
import "github.com/google/uuid"

type Manager struct {
    ID   uuid.UUID `sql:"key"`
    Name string
}

type TeamMember struct {
    ID        uuid.UUID      `sql:"key"`
    Manager   *Manager       `sql:foreign=ID"`
    Joined    time.Time
    Shift     time.Duration
    WeekStart *time.Weekday
}
```

Note that for the foreign keys to work,  
- The target field must be a pointer to a `struct`,
- the target `struct` must implement the `database/sql` [Scanner](https://pkg.go.dev/database/sql?tab=doc#Scanner) interface,
- the target `struct` key must be a `uuid.UUID`, string, or int (any type of int)

You can also use the `Statement` object level of using the Database:

```go
package main

import (
    "github.com/gildas/g-logger"
    "github.com/gildas/g-sql"
)

func main() {
    db, err := sql.Open("postgres", "connectionstring", logger.Create("MYAPP"))
    if err != nil {
        panic(err)
    }
    defer db.Close()

    // Insert some data
    statement, parms = sql.InsertStatement{}.With(db).Build("person", nil, sql.Queries{}.Add("lastname", "Doe").Add("age", 34))
    _, err := db.Exec(statement, parms...)

    // Find data!
    statement, parms = sql.SelectStatement{}.With(db).Build("person", []string("id", "age"), sql.Queries{}.Add("lastname", "Doe"))
    rows, err := db.Query(statement, parms...)

    // Then, scan the rows as usual from "database/sql"

    // Update data
    statement, parms = sql.UpdateStatement{}.With(db).Build("person", nil, sql.Queries{}.Add("lastname", "Doe").Add("age", sql.QuerySet, 25))
    _, err := db.Exec(statement, parms...)

    // Delete data
    statement, parms = sqlDeleteStatement{}.With(db).Build("person", nil, sql.Queries{}.Add("age", sql.Greater, 50))
    _, err := db.Exec(statement, parms...)
}
```

## TODO

There is still a lot to do....

- Support for 'NULL'able columns, until now all columns have to be non null,
- It would be nice to not force the target `struct` of a foreign key to impletement the Scanner interface,