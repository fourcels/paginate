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
	Title2    string    `json:"title2,omitempty" gorm:"column:title" search:"title" filter:"title2"`
	Content   string    `json:"content,omitempty" search:"content"`
	CreatedAt time.Time `json:"created_at,omitempty" filter:"created_at"`
	UserID    uint      `json:"user_id,omitempty"`
	User      *User     `json:"user,omitempty" filter:"user"`
}

type PostLabel struct {
	Post  `json:"post,omitempty"`
	Label string `json:"label,omitempty"`
}

var DB *gorm.DB

func initDB() {
	db, err := gorm.Open(sqlite.Open("./test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Post{})
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
		UserID  uint   `json:"user_id,omitempty"`
	}

	return rest.NewHandler(func(c echo.Context, in input, out *Post) error {
		post := Post{
			Title2:  in.Title,
			Content: in.Content,
			UserID:  in.UserID,
		}
		if result := DB.Create(&post); result.Error != nil {
			return result.Error
		}
		*out = post
		return nil
	})
}

func GetPosts() rest.Interactor {
	return rest.NewHandler(func(c echo.Context, in paginate.Pagination, out *[]Post) error {
		err := setupPaginate(c, in, out, func(db *gorm.DB) *gorm.DB {
			return db.Joins("User")
		})
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
