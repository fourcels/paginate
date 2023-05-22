package paginate

import (
	"reflect"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

type WhereType string

const (
	AND = WhereType("AND")
	OR  = WhereType("OR")
)

type Pagination struct {
	Page   int               `query:"page" minimum:"1" default:"1"`
	Size   int               `query:"size" minimum:"1" default:"10"`
	Search string            `query:"search"`
	Filter map[string]string `query:"filter" description:"1. Comparison Operators: **eq**, **ne**, **like**, **gt**, **gte**, **lt**, **lte**, **in**\n2. Conjunction Operators: **AND**, **OR**\n3. Usage: [**op:**]value"`
}

func Paginate(model any, p Pagination) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		setSearch(db, model, p.Search)
		setFilter(db, model, p.Filter)
		return db.Offset((p.Page - 1) * p.Size).Limit(p.Size)
	}
}

func setSearch(db *gorm.DB, model any, search string) *gorm.DB {
	if search = strings.TrimSpace(search); len(search) > 0 {
		fields := getFields(model, "search")
		db2 := db.Session(&gorm.Session{})
		for _, field := range fields {
			db2 = db2.Or(field+" LIKE ?", "%"+search+"%")
		}
		return db.Where(db2)
	}
	return db
}

func setFilter(db *gorm.DB, model any, data map[string]string) *gorm.DB {
	fields := getFields(model, "filter")
	for _, field := range fields {
		value, exists := data[field]
		if exists {
			if value = strings.TrimSpace(value); len(value) > 0 {
				db = addFilter(db, field, value)
			}
		}
	}
	return db
}

func addFilter(db *gorm.DB, field, value string) *gorm.DB {
	pattern := regexp.MustCompile(`(?i)\s+(and|or)\s+`)
	db2 := db.Session(&gorm.Session{})
	values := pattern.Split(value, -1)
	logicals := pattern.FindAllStringSubmatch(value, -1)
	for i, v := range values {
		v = strings.TrimSpace(v)
		if i == 0 {
			db2 = andWhere(db2, field, v)
			continue
		}
		switch strings.ToUpper(logicals[i-1][1]) {
		case "AND":
			db2 = andWhere(db2, field, v)
		case "OR":
			db2 = orWhere(db2, field, v)
		}

	}
	return db.Where(db2)
}

func orWhere(db *gorm.DB, field, value string) *gorm.DB {
	return where(db, field, value, OR)
}

func andWhere(db *gorm.DB, field, value string) *gorm.DB {
	return where(db, field, value, AND)
}

func where(db *gorm.DB, field, value string, t WhereType) *gorm.DB {

	op := "eq"
	if arr := strings.SplitN(value, ":", 2); len(arr) > 1 {
		op = arr[0]
		value = arr[1]
	}
	query := field + " = ?"
	var args any = value
	switch op {
	case "ne":
		query = field + " != ?"
	case "like":
		query = field + " LIKE ?"
	case "gt":
		query = field + " > ?"
	case "gte":
		query = field + " >= ?"
	case "lt":
		query = field + " < ?"
	case "lte":
		query = field + " <= ?"
	case "in":
		query = field + " IN ?"
		args = strings.Split(value, ",")
	}
	if t == OR {
		return db.Or(query, args)
	}
	return db.Where(query, args)
}

func getFields(model any, tag string) []string {
	fields := make([]string, 0)
	if model == nil {
		return fields
	}
	typ := reflect.TypeOf(model).Elem()
	if typ.Kind() != reflect.Struct {
		return fields
	}
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		fieldName := typeField.Tag.Get(tag)
		if len(fieldName) > 0 {
			fields = append(fields, fieldName)
		}
	}
	return fields
}
