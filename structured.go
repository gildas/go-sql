package sql

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gildas/go-errors"
)

// CreateTable creates an SQL Table from a schema
func (db *DB) CreateTable(schema interface{}) error {
	log := db.Logger.Child(nil, "create")
	schemaType, _ := getTypeAndValue(schema)
	table := strings.ToLower(schemaType.Name())

	log = log.Record("table", table)
	log.Tracef("Schema %s => table=%s", schemaType.Name(), table)
	columns := []string{}
	for i := 0; i < schemaType.NumField(); i++ {
		field := schemaType.Field(i)
		options := getOptions(field)
		if getOptions(field).Ignore {
			continue
		}
		log.Tracef("Field: %s, type=%s, kind=%s", field.Name, field.Type.Name(), field.Type.Kind())
		column := strings.Builder{}
		if len(options.ColumnName) > 0 {
			column.WriteString(options.ColumnName)
		} else {
			column.WriteString(strings.ToLower(field.Name))
		}
		if len(options.ForeignKey) > 0 {
			column.WriteString("_")
			column.WriteString(strings.ToLower(options.ForeignKey))
		}
		column.WriteString(" ")
		if len(options.ForeignKey) > 0 {
			log.Debugf("Field should use a foreign key: %s", options.ForeignKey)
			foreignType := field.Type
			if foreignType.Kind() == reflect.Ptr {
				foreignType = foreignType.Elem()
			}
			if foreignType.Kind() != reflect.Struct {
				return errors.ArgumentInvalid.With("typeof", field.Name).WithStack()
			}
			var sqltype string
			for j := 0; j < foreignType.NumField(); j++ {
				subfield := foreignType.Field(j)
				if subfield.Name == options.ForeignKey {
					log.Debugf("SubField: %s, type=%s, kind=%s", subfield.Name, subfield.Type.Name(), subfield.Type.Kind())
					if len(options.ColumnType) > 0 {
						sqltype = strings.ToUpper(options.ColumnType)
					} else {
						switch subfield.Type.Kind() {
						case reflect.Array, reflect.Slice:
							switch subfield.Type.Name() {
							case "UUID":
								sqltype = "UUID"
							default:
								return errors.ArgumentInvalid.With("typeof", subfield.Name).WithStack()
							}
						case reflect.String:
							sqltype = "VARCHAR(80)"
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							sqltype = "INT"
						default:
							log.Errorf("Unsupported Kind: %s", subfield.Type.Kind())
							return errors.ArgumentInvalid.With("typeof", subfield.Name).WithStack()
						}
					}
					log.Debugf("Matched? with %s", sqltype)
					break
				}
			}
			log.Debugf("Foreign Type: %s, kind=%s => %s", foreignType.Name(), foreignType.Kind(), sqltype)
			if len(sqltype) == 0 {
				return errors.ArgumentInvalid.With("foreignkey", options.ForeignKey).WithStack()
			}
			column.WriteString(sqltype)
		} else if len(options.ColumnType) > 0 {
			column.WriteString(strings.ToUpper(options.ColumnType))
		} else {
			sqltype, err := getSQLType(field.Name, field.Type)
			if err != nil {
				log.Warnf("Field details: %#v", field)
				log.Errorf("Unsupported Field Type %s (%s) for %s", field.Type.Name(), field.Type.Kind(), field.Name)
				return err
			}
			column.WriteString(sqltype)
		}
		if options.PrimaryKey {
			column.WriteString(" ")
			column.WriteString("PRIMARY KEY")
		}
		columns = append(columns, column.String())
		// TODO: How do we handle indices?
	}
	statement := fmt.Sprintf("CREATE TABLE %s (%s)", table, strings.Join(columns, ", "))
	parms := []interface{}{}
	log.Tracef("Statement: %s with %d parameters", statement, len(parms))
	_, err := db.db.Exec(statement, parms...)
	return err
}

// DeleteTable deletes (drops) the SQL table that represents the schema
func (db *DB) DeleteTable(schema interface{}) error {
	log := db.Logger.Child(nil, "drop")
	schemaType, _ := getTypeAndValue(schema)
	table := strings.ToLower(schemaType.Name())

	log = log.Record("table", table)
	log.Tracef("Schema %s => table=%s", schemaType.Name(), table)
	statement := fmt.Sprintf("DROP TABLE %s", table)
	log.Tracef("Statement: %s", statement)
	_, err := db.db.Exec(statement)
	return err
}

// Insert insert a blob in its SQL table
func (db *DB) Insert(blob interface{}) error {
	log := db.Logger.Child(nil, "insert")
	blobType, blobValue := getTypeAndValue(blob)
	table := strings.ToLower(blobType.Name())
	queries := Queries{}

	log = log.Record("table", table)
	log.Tracef("Schema %s => table=%s", blobType.Name(), table)
	for i := 0; i < blobType.NumField(); i++ {
		field := blobType.Field(i)
		options := getOptions(field)
		if options.Ignore {
			continue
		}
		log.Tracef("Field: %s, type=%s, kind=%s", field.Name, field.Type.Name(), field.Type.Kind())
		value := blobValue.Field(i)
		column := strings.ToLower(field.Name)
		if len(options.ColumnName) > 0 {
			column = options.ColumnName
		}
		if len(options.ForeignKey) > 0 {
			column = column + "_" + strings.ToLower(options.ForeignKey)
			foreignType := field.Type
			foreignValue := value
			log.Debugf("Foreign Value: %#v", foreignValue)
			if foreignType.Kind() == reflect.Ptr {
				foreignType = foreignType.Elem()
				foreignValue = value.Elem()
			}
			if foreignType.Kind() != reflect.Struct {
				return errors.ArgumentInvalid.With("typeof", field.Name).WithStack()
			}
			found := false
			for j := 0; j < foreignType.NumField(); j++ {
				subfield := foreignType.Field(j)
				if subfield.Name == options.ForeignKey {
					log.Tracef("SubField: %s, type=%s, kind=%s", subfield.Name, subfield.Type.Name(), subfield.Type.Kind())
					found = true
					value = foreignValue.Field(j)
					break
				}
			}
			if !found {
				return errors.ArgumentInvalid.With("foreignkey", options.ForeignKey).WithStack()
			}
		}
		log.Debugf("Adding value: %#v", value)
		queries.Add(column, QuerySet, value.Interface())
	}
	statement, parms := InsertStatement{}.With(db).Build(table, nil, queries)
	log.Tracef("Statement: %s with %d parameters", statement, len(parms))
	_, err := db.db.Exec(statement, parms...)
	return err
}

// FindAll retrieves all objects of a schema that satisfy the queries
func (db *DB) FindAll(schema interface{}, queries Queries) ([]interface{}, error) {
	log := db.Logger.Child(nil, "find_all")
	schemaType, _ := getTypeAndValue(schema)
	table := strings.ToLower(schemaType.Name())

	log = log.Record("table", table)
	log.Tracef("Schema %s => table=%s", schemaType.Name(), table)
	statement, parms := SelectStatement{}.With(db).Build(table, getColumns(schemaType), queries)
	log.Tracef("Statement: %s with %d parameters", statement, len(parms))
	rows, err := db.db.Query(statement, parms...)
	if err != nil {
		return []interface{}{}, err
	}
	defer rows.Close()

	results := []interface{}{}
	for rows.Next() {
		blob := reflect.New(schemaType)
		components := []interface{}{}
		for i := 0; i < schemaType.NumField(); i++ {
			field := schemaType.Field(i)
			if getOptions(field).Ignore {
				continue
			}
			log.Tracef("Field: %s, type=%s, kind=%s", field.Name, field.Type.Name(), field.Type.Kind())
			if field.Type.Kind() == reflect.Ptr {
				log.Tracef("Field: %s, type=%s, kind=%s", field.Name, field.Type.Elem().Name(), field.Type.Elem().Kind())

			}
			placeholder, err := getInterface(field.Name, field.Type, blob.Elem().Field(i))
			if err != nil {
				return results, err
			}
			components = append(components, placeholder)
		}
		err = rows.Scan(components...)
		if err != nil {
			log.Errorf("Failed to scan columns", err)
			return []interface{}{}, err
		}
		results = append(results, blob.Interface())
	}
	log.Tracef("Found %d results", len(results))
	return results, nil
}

// Find retrieves the first object of a schema that satisfies the queries
func (db *DB) Find(schema interface{}, queries Queries) (interface{}, error) {
	blobs, err := db.FindAll(schema, queries)
	if err != nil {
		return nil, err
	}
	if len(blobs) == 0 {
		return nil, errors.NotFound.WithStack()
	}
	return blobs[0], nil
}

// UpdateAll updates all objects of a schema that satisfy the queries
func (db *DB) UpdateAll(schema interface{}, queries Queries) error {
	log := db.Logger.Child(nil, "update")
	schemaType, _ := getTypeAndValue(schema)
	table := strings.ToLower(schemaType.Name())

	log = log.Record("table", table)
	log.Tracef("Schema %s => table=%s", schemaType.Name(), table)
	statement, parms := UpdateStatement{}.With(db).Build(table, getColumns(schemaType), queries)
	log.Tracef("Statement: %s with %d parameters", statement, len(parms))
	_, err := db.db.Exec(statement, parms...)
	return err
}

// DeleteAll deletes all objects of a schema that satisfy the queries
func (db *DB) DeleteAll(schema interface{}, queries Queries) error {
	log := db.Logger.Child(nil, "delete_all")
	schemaType, _ := getTypeAndValue(schema)
	table := strings.ToLower(schemaType.Name())

	log = log.Record("table", table)
	log.Tracef("Schema %s => table=%s", schemaType.Name(), table)
	columns := getColumns(schemaType)
	statement, parms := DeleteStatement{}.With(db).Build(table, columns, queries)
	log.Tracef("Statement: %s with %d parameters", statement, len(parms))
	_, err := db.db.Exec(statement, parms...)
	return err
}

// private methods

func getTypeAndValue(blob interface{}) (reflect.Type, reflect.Value) {
	blobType := reflect.TypeOf(blob)
	if blobType.Kind() == reflect.Ptr {
		return blobType.Elem(), reflect.ValueOf(blob).Elem()
	}
	return blobType, reflect.ValueOf(blob)
}

func getColumns(schemaType reflect.Type) []string {
	columns := []string{}
	for i := 0; i < schemaType.NumField(); i++ {
		field := schemaType.Field(i)
		options := getOptions(field)
		if options.Ignore {
			continue
		}
		column := strings.ToLower(field.Name)
		if len(options.ColumnName) > 0 {
			column = options.ColumnName
		}
		if len(options.ForeignKey) > 0 {
			column = column + "_" + strings.ToLower(options.ForeignKey)

		}
		columns = append(columns, column)
	}
	return columns
}

type fieldOptions struct {
	PrimaryKey bool
	Index      bool
	Ignore     bool
	ColumnName string
	ColumnType string
	ForeignKey string
}

func getOptions(field reflect.StructField) fieldOptions {
	options := fieldOptions{Ignore: false}
	for i, option := range strings.Split(field.Tag.Get("sql"), ",") {
		name := strings.ToLower(strings.TrimSpace(option)) 
		if strings.HasPrefix(name, "foreign=") {
			options.ForeignKey = strings.TrimSpace(strings.Split(option, "=")[1])
		} else {
			switch name {
			case "index":
				options.Index = true
			case "key":
				options.PrimaryKey = true
			case "-":
				options.Ignore = true
			default:
				if i == 0 {
					options.ColumnName = name
				} else {
					options.ColumnType = name
				}
			}
		}
	}
	return options
}

func getSQLType(name string, t reflect.Type) (string, error) {
	switch t.Kind() {
	case reflect.Array, reflect.Slice:
		switch t.Name() {
		case "UUID":
			return "UUID", nil
		default:
			return "", errors.ArgumentInvalid.With("typeof", name).WithStack()
		}
	case reflect.Struct:
		switch t.Name() {
		case "Time":
			return "TIMESTAMP", nil
		default:
			return "", errors.ArgumentInvalid.With("typeof", name).WithStack()
		}
	case reflect.Bool:
		return "BOOL", nil
	case reflect.Float32, reflect.Float64:
		return "FLOAT8", nil
	case reflect.String:
		return "VARCHAR(80)", nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "INT", nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "INT", nil
	case reflect.Ptr:
		return getSQLType(name, t.Elem())
	default:
		return "", errors.ArgumentInvalid.With("typeof", name).WithStack()
	}
}

func getInterface(fieldName string, fieldType reflect.Type, fieldValue reflect.Value) (interface{}, error) {
	switch fieldType.Kind() {
	case reflect.Ptr:
		pvalue := reflect.New(fieldType.Elem())
		fieldValue.Set(pvalue)
		return getInterface(fieldName, fieldType.Elem(), fieldValue.Elem())
	default:
		switch fieldType.Name() {
		case "Time":
			placeholder, ok := fieldValue.Addr().Interface().(*time.Time)
			if !ok {
				return nil, errors.ArgumentInvalid.With("typeof", fieldName).WithStack()
			}
			return (*DBTime)(placeholder), nil
		default:
			return fieldValue.Addr().Interface(), nil
		}
	}
}