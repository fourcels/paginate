package main

import (
	"log"
	"strconv"
	"time"

	"github.com/fourcels/paginate"
	"github.com/fourcels/rest"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint   `json:"id,omitempty"`
	Name string `json:"name,omitempty" filter:"name"`
}

type Post struct {
	ID        uint      `json:"id,omitempty" filter:"id"`
	Title     string    `json:"title,omitempty" search:"title" filter:"title"`
	Content   string    `json:"content,omitempty" search:"content"`
	CreatedAt time.Time `json:"created_at,omitempty" filter:"created_at"`
}

var DB *gorm.DB

func initDB() {
	db, err := gorm.Open(sqlite.Open("./test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&User{}, &Post{})
	DB = db
}

func main() {
	initDB()
	s := rest.NewService()
	s.OpenAPI.Info.WithTitle("Basic Example")
	post := s.Group("/posts", rest.WithTags("Posts"))
	post.POST("", CreatePost())
	post.GET("", GetPosts())

	user := s.Group("/users", rest.WithTags("Users"))
	user.POST("", CreateUser())

	// Swagger UI endpoint at /docs.
	s.Docs("/docs")

	// Start server.
	log.Println("http://localhost:1323/docs")
	s.Start(":1323")
}

func CreateUser() rest.Interactor {
	type input struct {
		Name string `json:"name,omitempty"`
	}

	return rest.NewHandler(func(c echo.Context, in input, out *User) error {
		user := User{
			Name: in.Name,
		}
		if result := DB.Create(&user); result.Error != nil {
			return result.Error
		}
		*out = user
		return nil
	})
}
func CreatePost() rest.Interactor {
	type input struct {
		Title   string `json:"title,omitempty"`
		Content string `json:"content,omitempty"`
	}

	return rest.NewHandler(func(c echo.Context, in input, out *Post) error {
		post := Post{
			Title:   in.Title,
			Content: in.Content,
		}
		if result := DB.Create(&post); result.Error != nil {
			return result.Error
		}
		*out = post
		return nil
	})
}

func GetPosts() rest.Interactor {
	type input struct {
		Title string `query:"title"`
		paginate.PaginationDefault
	}

	return rest.NewHandler(func(c echo.Context, in input, out *[]Post) error {
		err := setupPaginate(c, in.SetDefaultSort("id"), out)
		if err != nil {
			return err
		}
		return nil
	})
}

func setupPaginate[T any](c echo.Context, p paginate.Pagination, out *[]T, query ...func(db *gorm.DB) *gorm.DB) error {
	count, err := paginate.Paginate(DB, p, out, query...)
	if err != nil {
		return err
	}
	c.Response().Header().Set("X-Total", strconv.FormatInt(count, 10))
	return nil
}
