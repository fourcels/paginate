package paginate

import (
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
	"gorm.io/gorm"
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
	if result := tx.
		Offset((p.Page - 1) * p.Size).Limit(p.Size).Order(parseSort(p.Sort)).
		Find(out); result.Error != nil {
		return count, result.Error
	}
	return count, nil
}

func parseSort(sort string) string {
	sort = strings.TrimSpace(sort)
	if len(sort) == 0 {
		return sort
	}
	sortArr := strings.Split(sort, ",")
	for i := 0; i < len(sortArr); i++ {
		item := sortArr[i]
		if strings.HasPrefix(sort, "-") {
			sortArr[i] = strings.TrimPrefix(item, "-") + " desc"
		}
	}
	return strings.Join(sortArr, ",")
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

	for k, v := range data {
		field, op := k, "eq"
		if arr := strings.SplitN(k, ":", 2); len(arr) > 1 {
			field, op = arr[0], arr[1]
		}
		if slices.Contains(fields, field) && len(v) > 0 {
			db = where(db, field, strings.TrimSpace(v), op)
		}
	}
	return db
}

func where(db *gorm.DB, field, value, op string) *gorm.DB {
	switch op {
	case "eq":
		return db.Where(field+" = ?", value)
	case "ne":
		return db.Where(field+" != ?", value)
	case "contain", "like":
		return db.Where(field+" LIKE ?", "%"+value+"%")
	case "gt":
		return db.Where(field+" > ?", value)
	case "gte":
		return db.Where(field+" >= ?", value)
	case "lt":
		return db.Where(field+" < ?", value)
	case "lte":
		return db.Where(field+" <= ?", value)
	case "in":
		return db.Where(field+" IN ?", strings.Split(value, ","))
	}
	return db
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
