/*
Package go-sql is an SQL library that adds Higher Level capabilities to the standard `database/sql` library.

Usage

Example:  

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


For the sake of readability, I removed the error management. DON'T!

As you can see, using actual GO struct types is rather easy now. We do not support the entire set of SQL types or GO types, but we have the basics.  

You can also use the Statement level of using the Database:

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

*/