package main

import (
	"log"
	"strconv"
	"time"

	"github.com/fourcels/paginate"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
