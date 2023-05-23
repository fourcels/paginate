# GORM Paginate

[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/fourcels/paginate)

## Query Params

```go
type Pagination struct {
	Page   int               `query:"page" minimum:"1" default:"1"`
	Size   int               `query:"size" minimum:"1" default:"10"`
	Sort   string            `query:"sort" default:"id" description:"1. asc: **id**\n2. desc: **-id**\n3. multi: **id,created_at**"`
	Search string            `query:"search"`
	Filter map[string]string `query:"filter" description:"1. Comparison Operators: **eq**, **ne**, **like**, **gt**, **gte**, **lt**, **lte**, **in**\n2. Conjunction Operators: **AND**, **OR**\n3. Usage: [**op:**]value"`
}
```

## Filter

1. Comparison Operators: `eq`, `ne`, `like`, `gt`, `gte`, `lt`, `lte`, `in`
1. Conjunction Operators: `AND`, `OR`
1. Usage: [`op:`]value

## Sort

1. asc: id
1. desc: -id
1. multi: id,created_at

## Example

[exmaples](./examples/main.go)

```go
type Post struct {
	ID        uint      `json:"id,omitempty"`
	Title     string    `json:"title,omitempty" search:"title" filter:"title"`
	Content   string    `json:"content,omitempty" search:"content"`
	CreatedAt time.Time `json:"created_at,omitempty" filter:"created_at"`
}

func main() {
	db, err := gorm.Open(sqlite.Open("./test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Post{})
	p := paginate.Pagination{
		Page: 3,
		Size: 10,
		Sort: "-id",
	}
	seedIfNeeded(db)
	var posts []Post
	var count int64
	if err := paginate.Paginate(db, p, &count, &posts); err != nil {
		panic(err)
	}
	log.Println(count, posts)
}

func seedIfNeeded(db *gorm.DB) {
	var count int64
	db.Model(&Post{}).Count(&count)
	if count > 0 {
		return
	}
	posts := make([]Post, 0)
	for i := 0; i < 100; i++ {
		posts = append(posts, Post{
			Title:   "title " + strconv.Itoa(i),
			Content: "content " + strconv.Itoa(i),
		})
	}
	db.Create(posts)
}
```
