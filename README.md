# GORM Paginate

## Pagination Filter

1. Comparison Operators: `eq`, `ne`, `like`, `gt`, `gte`, `lt`, `lte`, `in`
1. Conjunction Operators: `AND`, `OR`
1. Usage: [`op:`]value

```go
type Pagination struct {
	Page   int               `query:"page" minimum:"1" default:"1"`
	Size   int               `query:"size" minimum:"1" default:"10"`
	Search string            `query:"search"` // field tag `search`
	Filter map[string]string `query:"filter" description:"1. Comparison Operators: **eq**, **ne**, **like**, **gt**, **gte**, **lt**, **lte**, **in**\n2. Conjunction Operators: **AND**, **OR**\n3. Usage: [**op:**]value"` // field tag `filter`
}
```

## Example

[exmaples](./examples/main.go)

```go
p := paginate.Pagination{
  Page:   1,
  Size:   10,
}
var posts []Post
if result := db.Scopes(paginate.Paginate(&Post{}, p)).Find(&posts); result.Error != nil {
  panic(result.Error)
}
```
