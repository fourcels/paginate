# GORM Paginate

[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/fourcels/paginate)

## Query Params

```go
type PaginationDefault struct {
	Page   int               `query:"page" minimum:"1" default:"1"`
	Size   int               `query:"size" minimum:"1" default:"10"`
	Sort   string            `query:"sort" description:"1. asc: **id**\n2. desc: **-id**\n3. multi: **id,created_at**"`
	Search string            `query:"search"`
	Filter map[string]string `query:"filter" description:"1. Comparison Operators: **eq**, **ne**, **like**, **contain**, **gt**, **gte**, **lt**, **lte**, **in**\n2. Usage: \"field**[:op]**\":value"`
}
```

## Filter

1. Comparison Operators: `eq`, `ne`, `like`, `contain`, `gt`, `gte`, `lt`,
   `lte`, `in`
1. Usage: "field`[:op]`":value

## Sort

1. asc: id
1. desc: -id
1. multi: id,created_at

## Example

[exmaples](./examples/main.go)
