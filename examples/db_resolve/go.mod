module github.com/housecanary/gq/examples/db_resolve

require (
	github.com/housecanary/gq v0.0.0-20181202231151-33d3ccbca4ef
	github.com/housecanary/gq/examples/db_walk_query v0.0.0-20181204063100-be19c2e2057e
	github.com/jmoiron/sqlx v1.2.0
	github.com/mattn/go-sqlite3 v1.10.0
)

replace github.com/housecanary/gq => ../..
