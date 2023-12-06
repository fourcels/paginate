package paginate

import (
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Pagination struct {
	Page   int               `query:"page" minimum:"1" default:"1"`
	Size   int               `query:"size" minimum:"1" default:"10"`
	Sort   string            `query:"sort" default:"id" description:"1. asc: **id**\n2. desc: **-id**\n3. multi: **id,created_at**"`
	Search string            `query:"search"`
	Filter map[string]string `query:"filter" description:"1. Comparison Operators: **eq**, **ne**, **like**, **contain**, **gt**, **gte**, **lt**, **lte**, **in**\n2. Usage: \"field**[:op]**\":value"`
}

func Paginate[T any](db *gorm.DB, p Pagination, out *[]T, query ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	model := new(T)
	scopes := append(query, func(db *gorm.DB) *gorm.DB {
		setSearch(db, model, p.Search)
		setFilter(db, model, p.Filter)
		return db
	})
	tx := db.Model(model).Scopes(scopes...).Session(&gorm.Session{})
	var count int64
	if result := tx.Count(&count); result.Error != nil {
		return count, result.Error
	}
	if result := tx.Scopes(OrderByScope(p.Sort)).
		Offset((p.Page - 1) * p.Size).Limit(p.Size).
		Find(out); result.Error != nil {
		return count, result.Error
	}
	return count, nil
}

func OrderByScope(sort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		sort = strings.TrimSpace(sort)
		if len(sort) == 0 {
			return db
		}
		sorts := strings.Split(sort, ",")
		for _, v := range sorts {
			table, field, desc := clause.CurrentTable, strings.TrimSpace(v), false
			if strings.Contains(v, ".") {
				table = ""
			}
			if strings.HasPrefix(v, "-") {
				field, desc = strings.TrimPrefix(v, "-"), true
			}
			db.Order(clause.OrderByColumn{
				Column: clause.Column{Table: table, Name: field},
				Desc:   desc,
			})
		}
		return db
	}
}

func setSearch(db *gorm.DB, model any, search string) *gorm.DB {
	if search = strings.TrimSpace(search); len(search) > 0 {
		fields := getFields(model, "search")
		db2 := db.Session(&gorm.Session{})
		for _, field := range fields {
			column := clause.Column{
				Table: clause.CurrentTable,
				Name:  field,
			}
			db2 = db2.Or(clause.Like{Value: "%" + search + "%", Column: column})
		}
		return db.Where(db2)
	}
	return db
}

func setFilter(db *gorm.DB, model any, data map[string]string) *gorm.DB {
	fields := getFields(model, "filter")

	for k, v := range data {
		field, op := k, "eq"
		if arr := strings.SplitN(k, ":", 2); len(arr) > 1 {
			field, op = arr[0], arr[1]
		}
		if checkField(fields, field) && len(v) > 0 {
			db = where(db, field, strings.TrimSpace(v), op)
		}
	}
	return db
}

func checkField(fields []string, field string) bool {
	field = strings.Split(field, ".")[0]
	for _, v := range fields {
		if v == field {
			return true
		}
	}
	return false
}

func where(db *gorm.DB, field, value, op string) *gorm.DB {
	tableName := clause.CurrentTable
	if strings.Contains(field, ".") {
		tableName = ""
	}
	column := clause.Column{
		Table: tableName,
		Name:  field,
	}
	switch op {
	case "eq":
		return db.Where(clause.Eq{Value: value, Column: column})
	case "ne":
		return db.Where(clause.Neq{Value: value, Column: column})
	case "contain", "like":
		return db.Where(clause.Like{Value: "%" + value + "%", Column: column})
	case "gt":
		return db.Where(clause.Gt{Value: value, Column: column})
	case "gte":
		return db.Where(clause.Gte{Value: value, Column: column})
	case "lt":
		return db.Where(clause.Lt{Value: value, Column: column})
	case "lte":
		return db.Where(clause.Lte{Value: value, Column: column})
	case "in":
		return db.Where(clause.IN{Values: toAnyList(strings.Split(value, ",")), Column: column})
	}
	return db
}

func toAnyList[T any](input []T) []any {
	list := make([]any, len(input))
	for i, v := range input {
		list[i] = v
	}
	return list
}

func getFields(model any, tag string) []string {
	fields := make([]string, 0)
	if model == nil {
		return fields
	}
	typ := reflect.TypeOf(model)
	val := reflect.ValueOf(model)
	val = reflect.Indirect(val)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return fields
	}
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		if !typeField.IsExported() {
			continue
		}
		if typeField.Anonymous {
			structField := val.Field(i)
			fieldVal := structField.Interface()
			fields = append(fields, getFields(fieldVal, tag)...)
			continue
		}
		fieldName := typeField.Tag.Get(tag)
		if len(fieldName) > 0 {
			tagSetting := schema.ParseTagSetting(typeField.Tag.Get("gorm"), ";")
			column, ok := tagSetting["COLUMN"]
			if ok {
				fieldName = column
			}
			fields = append(fields, fieldName)
		}
	}
	return fields
}
