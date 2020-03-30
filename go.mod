module github.com/gildas/go-sql

go 1.14

require (
	cloud.google.com/go v0.55.0 // indirect
	github.com/gildas/go-core v0.4.2 // indirect
	github.com/gildas/go-errors v0.1.0
	github.com/gildas/go-logger v1.3.4
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.3
	github.com/lib/pq v1.3.0 // indirect
	github.com/mattn/go-sqlite3 v2.0.3+incompatible // indirect
	github.com/onsi/ginkgo v1.12.0 // indirect
	github.com/onsi/gomega v1.9.0 // indirect
	github.com/proullon/ramsql v0.0.0-20181213202341-817cee58a244
	github.com/stretchr/testify v1.4.0
	github.com/ziutek/mymysql v1.5.4 // indirect
	golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e // indirect
	google.golang.org/genproto v0.0.0-20200326112834-f447254575fd // indirect
)

replace github.com/gildas/ramsql => ../ramsql
