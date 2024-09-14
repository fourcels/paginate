package paginate

import (
	"maps"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Pagination interface {
	GetPage() int
	GetSize() int
	GetSort() string
	GetSearch() string
	GetFilter() map[string]string
}

type DefaultPagination struct {
	Page   int               `query:"page" minimum:"1" default:"1"`
	Size   int               `query:"size" minimum:"1" default:"10"`
	Sort   string            `query:"sort" description:"1. asc: **id**\n2. desc: **-id**\n3. multi: **id,created_at**"`
	Search string            `query:"search"`
	Filter map[string]string `query:"filter" description:"1. Comparison Operators: **eq**, **ne**, **like**, **contain**, **gt**, **gte**, **lt**, **lte**, **in**\n2. Usage: \"field**[:op]**\":value"`
}

func (p *DefaultPagination) SetDefaultSort(sort string) Pagination {
	if len(p.Sort) == 0 {
		p.Sort = sort
	}
	return p
}

func (p *DefaultPagination) GetPage() int {
	return p.Page
}
func (p *DefaultPagination) GetSize() int {
	return p.Size
}
func (p *DefaultPagination) GetSort() string {
	return p.Sort
}
func (p *DefaultPagination) GetSearch() string {
	return p.Search
}
func (p *DefaultPagination) GetFilter() map[string]string {
	return p.Filter
}

func Paginate[T any](db *gorm.DB, p Pagination, out *[]T, query ...func(db *gorm.DB) *gorm.DB) (int64, error) {
	model := new(T)
	scopes := append(query, func(db *gorm.DB) *gorm.DB {
		setSearch(db, model, p.GetSearch())
		setFilter(db, model, p.GetFilter())
		return db
	})
	var count int64
	if result := db.Session(&gorm.Session{}).Model(model).Scopes(scopes...).Count(&count); result.Error != nil {
		return count, result.Error
	}
	if result := db.Model(model).Scopes(scopes...).Scopes(OrderByScope(model, p.GetSort())).
		Offset((p.GetPage() - 1) * p.GetSize()).Limit(p.GetSize()).
		Find(out); result.Error != nil {
		return count, result.Error
	}
	return count, nil
}

func OrderByScope(model any, sort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		orderByColumns := sort2OrderByColumns(model, sort)
		for _, v := range orderByColumns {
			db = db.Order(v)
		}
		return db
	}
}

func sort2OrderByColumns(model any, sort string) []clause.OrderByColumn {
	orderByColumns := make([]clause.OrderByColumn, 0)
	fields := getFields(model, "sort")

	sort = strings.TrimSpace(sort)
	for _, v := range strings.Split(sort, ",") {
		field, desc := strings.TrimSpace(v), false
		if strings.HasPrefix(field, "-") {
			field, desc = strings.TrimPrefix(field, "-"), true
		}
		fieldName, ok := fields[field]
		if ok {
			field = fieldName
		}
		orderByColumns = append(orderByColumns, getOrderByColumn(field, desc))
	}

	return orderByColumns
}
func getOrderByColumn(fieldName string, desc bool) clause.OrderByColumn {
	return clause.OrderByColumn{
		Column: getColumn(fieldName),
		Desc:   desc,
	}
}

func getSearchColumns(model any) []clause.Column {
	fields := getFields(model, "search")
	columns := make([]clause.Column, 0)
	for _, field := range fields {
		columns = append(columns, getColumn(field))
	}
	return columns
}

func setSearch(db *gorm.DB, model any, search string) *gorm.DB {
	if search = strings.TrimSpace(search); len(search) > 0 {
		db2 := db.Session(&gorm.Session{})
		searchColumns := getSearchColumns(model)
		for _, column := range searchColumns {
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
		if fieldName, ok := fields[field]; ok && len(v) > 0 {
			db = where(db, fieldName, strings.TrimSpace(v), op)
		}
	}
	return db
}

func where(db *gorm.DB, fieldName, value, op string) *gorm.DB {
	column := getColumn(fieldName)
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

func getColumn(fieldName string) clause.Column {
	column := clause.Column{}
	if len(fieldName) == 0 {
		return column
	}
	fieldArr := strings.Split(fieldName, ".")
	if len(fieldArr) == 1 {
		column.Table = clause.CurrentTable
		column.Name = fieldArr[0]
	} else {
		column.Table = strings.Join(fieldArr[:len(fieldArr)-1], "__")
		column.Name = fieldArr[len(fieldArr)-1]
	}
	return column
}

func getFields(model any, tag string) map[string]string {
	fields := make(map[string]string)
	if model == nil {
		return fields
	}
	typ := reflect.TypeOf(model)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return fields
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		if !field.IsExported() {
			continue
		}
		if field.Anonymous {
			maps.Copy(fields, getFields(reflect.New(fieldType).Interface(), tag))
			continue
		}
		fieldName := field.Tag.Get(tag)
		if len(fieldName) > 0 {
			dbName := fieldName
			tagSetting := schema.ParseTagSetting(field.Tag.Get("gorm"), ";")
			column, ok := tagSetting["COLUMN"]
			if ok {
				dbName = column
			}
			if fieldType.Kind() == reflect.Struct {
				subFields := getFields(reflect.New(fieldType).Interface(), tag)
				if len(subFields) > 0 {
					for k, v := range subFields {
						fields[fieldName+"."+k] = dbName + "." + v
					}
				} else {
					fields[fieldName] = dbName
				}
			} else {
				fields[fieldName] = dbName
			}
		}
	}
	return fields
}
