package loader

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"

	"github.com/housecanary/gq/ast"
	"github.com/housecanary/gq/schema"
)

// A ModelProvider is a function that can construct a model
// given a name
type ModelProvider func(modelName string) DBModel

// A DBModel is an struct that can be loaded from the database
type DBModel interface {
	PrepareScan(fieldNames []string) []interface{}
	SetJoinedModels(fieldName string, values []interface{})
}

// A DBLoader can fetch a graph of objects from the database
type DBLoader struct {
	db            *sql.DB
	modelProvider ModelProvider
	rootModel     *modelDescriptor
	where         struct {
		col string
		op  string
		arg interface{}
	}
	page struct {
		lastVal interface{}
		count   int
	}
}

// NewDBLoader makes a new DBLoader
func NewDBLoader(db *sql.DB, modelProvider ModelProvider, modelName string, tableName string, idColumn string) *DBLoader {
	return &DBLoader{
		db:            db,
		modelProvider: modelProvider,
		rootModel: &modelDescriptor{
			modelFactory: modelFactory{modelProvider, modelName},
			table:        tableName,
			idColumn:     idColumn,
		}}
}

// Where adds a where clause to a query
func (l *DBLoader) Where(col string, op string, arg interface{}) {
	l.where.col = col
	l.where.op = op
	l.where.arg = arg
}

// Page paginates a query
func (l *DBLoader) Page(from interface{}, count int) {
	l.page.lastVal = from
	l.page.count = count
}

// WalkQuery implements schema.FieldWalkCB
func (l *DBLoader) WalkQuery(selection *ast.Field, field *schema.FieldDescriptor, walker schema.ChildWalker) bool {
	return l.rootModel.queryWalker(l.modelProvider)(selection, field, walker)
}

// Load executes a load from the database
func (l *DBLoader) Load() ([]interface{}, error) {
	assignAliases(0, l.rootModel)
	cols := flatColumns(l.rootModel, nil)

	q := &bytes.Buffer{}
	q.WriteString("SELECT")
	for i, c := range cols {
		if i > 0 {
			q.WriteString(",")
		}
		q.WriteString("\n  ")
		q.WriteString(c)
	}

	if l.page.count != 0 || l.where.col != "" {
		ncols := len(l.rootModel.columns)
		njoins := len(l.rootModel.joins)
		rootcols := make([]string, 0, ncols+njoins+2)
		idIncluded := false
		whereIncluded := false
		if l.where.col == "" {
			whereIncluded = true
		}
		if l.where.col == l.rootModel.idColumn {
			whereIncluded = true
		}
		for _, c := range l.rootModel.columns {
			rootcols = append(rootcols, c.tableName)
			if c.tableName == l.rootModel.idColumn {
				idIncluded = true
			}
		}

		joinCols := make(map[string]bool, njoins)
		for _, j := range l.rootModel.joins {
			if !joinCols[j.sourceColumn] {
				joinCols[j.sourceColumn] = true
			} else {
				continue
			}
			rootcols = append(rootcols, j.sourceColumn)
		}

		if !idIncluded {
			rootcols = append(rootcols, l.rootModel.idColumn)
		}

		if !whereIncluded {
			rootcols = append(rootcols, l.where.col)
		}
		whereClauses := make([]string, 0, 2)
		if l.page.lastVal != nil {
			whereClauses = append(whereClauses, fmt.Sprintf("%s > ?", l.rootModel.idColumn))
		}
		if l.where.col != "" {
			whereClauses = append(whereClauses, fmt.Sprintf("%s %s ?", l.where.col, l.where.op))
		}

		where := strings.Join(whereClauses, " AND ")
		if where != "" {
			where = " WHERE " + where
		}

		var limit string
		if l.page.count != 0 {
			limit = fmt.Sprintf("LIMIT %d", l.page.count)
		}
		q.WriteString(fmt.Sprintf("\nFROM (SELECT %s FROM %s%s ORDER BY %s %s) %s", strings.Join(rootcols, ","), l.rootModel.table, where, l.rootModel.idColumn, limit, l.rootModel.alias))
	} else {
		q.WriteString(fmt.Sprintf("\nFROM %s %s", l.rootModel.table, l.rootModel.alias))
	}
	addJoins(l.rootModel, q)

	args := make([]interface{}, 0, 2)
	if l.page.lastVal != nil {
		args = append(args, l.page.lastVal)
	}

	if l.where.col != "" {
		args = append(args, l.where.arg)
	}

	q.WriteString("\nORDER BY ")
	q.WriteString(strings.Join(idColumns(l.rootModel, nil), ","))

	fmt.Println("Executing query: ", q.String(), args)
	rows, err := l.db.Query(q.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	modelRows := []*resultTreeRow{}
	for rows.Next() {
		resultRow, rowTargets := l.rootModel.prepareScan(nil)
		modelRows = append(modelRows, resultRow)
		rows.Scan(rowTargets...)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	t := tableToTree(modelRows)
	return t, nil
}

type modelFactory struct {
	provider ModelProvider
	name     string
}

type modelDescriptor struct {
	modelFactory modelFactory
	alias        string
	table        string
	idColumn     string
	columns      []columnDescriptor
	joins        []*joinDescriptor
}

type columnDescriptor struct {
	modelName string
	tableName string
}

type joinDescriptor struct {
	m            *modelDescriptor
	fieldName    string
	many         bool
	sourceColumn string
	targetColumn string
}

type resultTreeRow struct {
	model       DBModel
	id          interface{}
	childModels map[string]*resultTreeRow
}

func (r *resultTreeRow) idEquals(or *resultTreeRow) bool {
	if b1, ok1 := r.id.([]byte); ok1 {
		b2, ok2 := or.id.([]byte)
		if !ok2 {
			return false
		}

		return bytes.Equal(b1, b2)
	}

	return r.id == or.id
}

func assignAliases(counter int, d *modelDescriptor) int {
	d.alias = fmt.Sprintf("t_%v", counter)
	counter++

	for _, j := range d.joins {
		counter = assignAliases(counter, j.m)
	}
	return counter
}

func flatColumns(d *modelDescriptor, cols []string) []string {
	cols = append(cols, fmt.Sprintf("%s.%s AS %s__id", d.alias, d.idColumn, d.alias))
	for _, c := range d.columns {
		cols = append(cols, fmt.Sprintf("%s.%s", d.alias, c.tableName))
	}

	for _, j := range d.joins {
		cols = flatColumns(j.m, cols)
	}
	return cols
}

func idColumns(d *modelDescriptor, cols []string) []string {
	cols = append(cols, fmt.Sprintf("%s__id", d.alias))
	for _, j := range d.joins {
		cols = idColumns(j.m, cols)
	}
	return cols
}

func addJoins(d *modelDescriptor, q *bytes.Buffer) {
	for _, j := range d.joins {
		q.WriteString(fmt.Sprintf("\nLEFT JOIN %s %s ON %s.%s = %s.%s", j.m.table, j.m.alias, d.alias, j.sourceColumn, j.m.alias, j.targetColumn))
		addJoins(j.m, q)
	}
}

func tableToTree(rr []*resultTreeRow) []interface{} {
	var childResultRows map[string][]*resultTreeRow
	var groupHead *resultTreeRow
	var results []interface{}
	for _, row := range rr {
		if groupHead == nil || !groupHead.idEquals(row) {
			if groupHead != nil {
				lastModel := groupHead.model
				for k, v := range childResultRows {
					childModels := tableToTree(v)
					lastModel.SetJoinedModels(k, childModels)
				}
				results = append(results, lastModel)
			}
			childResultRows = make(map[string][]*resultTreeRow)
			groupHead = row
		}

		for k, v := range row.childModels {
			childResultRows[k] = append(childResultRows[k], v)
		}
	}

	if groupHead != nil {
		lastModel := groupHead.model
		for k, v := range childResultRows {
			childModels := tableToTree(v)
			lastModel.SetJoinedModels(k, childModels)
		}
		results = append(results, lastModel)

	}
	return results
}

func (d *modelDescriptor) prepareScan(targets []interface{}) (*resultTreeRow, []interface{}) {
	rr := &resultTreeRow{childModels: make(map[string]*resultTreeRow)}
	m := d.modelFactory.createInstance()
	rr.model = m
	targets = append(targets, &rr.id)
	fieldNames := make([]string, len(d.columns))
	for i, cd := range d.columns {
		fieldNames[i] = cd.modelName
	}
	targets = append(targets, m.PrepareScan(fieldNames)...)

	for _, j := range d.joins {
		childResultRow, newTargets := j.m.prepareScan(targets)
		targets = newTargets
		rr.childModels[j.fieldName] = childResultRow
	}

	return rr, targets
}

func (d *modelDescriptor) queryWalker(modelProvider ModelProvider) schema.FieldWalkCB {
	return func(selection *ast.Field, field *schema.FieldDescriptor, walker schema.ChildWalker) bool {
		// Handle @dbColumn directives
		if col := field.GetDirective("dbColumn"); col != nil {
			if nameArg := col.Argument("name"); nameArg != nil {
				val := nameArg.Value()
				if s, ok := val.(schema.LiteralString); ok {
					d.columns = append(d.columns, columnDescriptor{field.Name(), string(s)})
				} else {
					panic(fmt.Sprintf("For field %s a name argument is required to be a string", field.Name()))
				}
			} else {
				d.columns = append(d.columns, columnDescriptor{field.Name(), strings.ToLower(field.Name())})
			}
		}

		// Handle @dbJoin directives
		if join := field.GetDirective("dbJoin"); join != nil {
			var joinedModel *modelDescriptor
			var many bool
			var sourceColumn string
			var targetColumn string

			typ := field.Type()
		loop:
			for {
				switch v := typ.(type) {
				case *schema.ObjectType:
					joinedModel = makeModelDescriptor(modelProvider, field.Name(), v)
					break loop
				case *schema.ListType:
					many = true
					typ = v.Unwrap()
				case *schema.NotNilType:
					typ = v.Unwrap()
				default:
					panic(fmt.Sprintf("For field %s the field should be reference to a type with a @dbTable directive, or a slice of such fields", field.Name()))
				}
			}

			if fromArg := join.Argument("from"); fromArg != nil {
				val := fromArg.Value()
				if s, ok := val.(schema.LiteralString); ok {
					sourceColumn = string(s)
				} else {
					panic(fmt.Sprintf("For field %s the @dbJoin from argument is required to be a string", field.Name()))
				}
			} else {
				panic(fmt.Sprintf("For field %s the @dbJoin directive should have a from argument", field.Name()))
			}

			if toArg := join.Argument("to"); toArg != nil {
				val := toArg.Value()
				if s, ok := val.(schema.LiteralString); ok {
					targetColumn = string(s)
				} else {
					panic(fmt.Sprintf("For field %s the @dbJoin to argument is required to be a string", field.Name()))
				}
			} else {
				panic(fmt.Sprintf("For field %s the @dbJoin directive should have a to argument", field.Name()))
			}

			d.joins = append(d.joins, &joinDescriptor{
				m:            joinedModel,
				fieldName:    field.Name(),
				many:         many,
				sourceColumn: sourceColumn,
				targetColumn: targetColumn,
			})

			walker.WalkChildSelections(joinedModel.queryWalker(modelProvider))
		}

		return false
	}
}

func makeModelDescriptor(modelProvider ModelProvider, fieldName string, ot *schema.ObjectType) *modelDescriptor {
	var table string
	var idColumn = "id"
	if tableDir := ot.GetDirective("dbTable"); tableDir != nil {
		if nameArg := tableDir.Argument("name"); nameArg != nil {
			val := nameArg.Value()
			if s, ok := val.(schema.LiteralString); ok {
				table = string(s)
			}
		}

		if idArg := tableDir.Argument("id"); idArg != nil {
			val := idArg.Value()
			if s, ok := val.(schema.LiteralString); ok {
				idColumn = string(s)
			}
		}
	}

	if table == "" {
		return nil
	}

	return &modelDescriptor{
		modelFactory: modelFactory{modelProvider, ot.Name()},
		table:        table,
		idColumn:     idColumn,
	}
}

func (m modelFactory) createInstance() DBModel {
	return m.provider(m.name)
}
