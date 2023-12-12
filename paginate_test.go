package paginate

import (
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm/clause"
)

func Test_getFields(t *testing.T) {
	type Post struct {
		ID    uint   `json:"id,omitempty"`
		Title string `json:"title,omitempty" filter:"title"`
	}

	type args struct {
		model any
		tag   string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		// TODO: Add test cases.
		{
			name: "anonymous field",
			args: args{
				model: struct {
					Post
					ID uint `json:"id,omitempty"`
				}{},
				tag: "filter",
			},
			want: map[string]string{
				"title": "title",
			},
		},

		{
			name: "nested field",
			args: args{
				model: struct {
					ID   uint `json:"id,omitempty"`
					Post *struct {
						ID    uint   `json:"id,omitempty"`
						Title string `json:"title,omitempty" filter:"title"`
						User  *struct {
							ID   uint   `json:"id,omitempty"`
							Name string `json:"name,omitempty" filter:"name"`
						} `json:"user,omitempty" filter:"User"`
					} `json:"post,omitempty" filter:"Post"`
				}{},
				tag: "filter",
			},
			want: map[string]string{
				"Post.title":     "Post.title",
				"Post.User.name": "Post.User.name",
			},
		},
		{
			name: "gorm column name",
			args: args{
				model: struct {
					Title string `json:"title,omitempty" gorm:"column:title2" filter:"title"`
					ID    uint   `json:"id,omitempty"`
				}{},
				tag: "filter",
			},
			want: map[string]string{
				"title": "title2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFields(tt.args.model, tt.args.tag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFields() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getColumn(t *testing.T) {
	type args struct {
		fieldName string
	}
	tests := []struct {
		name string
		args args
		want clause.Column
	}{
		// TODO: Add test cases.
		{
			name: "current table field",
			args: args{
				fieldName: "Title",
			},
			want: clause.Column{
				Table: clause.CurrentTable,
				Name:  "Title",
			},
		},
		{
			name: "nested table field",
			args: args{
				fieldName: "Post.Title",
			},
			want: clause.Column{
				Table: "Post",
				Name:  "Title",
			},
		},
		{
			name: "nested table field 2",
			args: args{
				fieldName: "Post.User.Name",
			},
			want: clause.Column{
				Table: "Post__User",
				Name:  "Name",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getColumn(tt.args.fieldName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getColumn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sort2OrderByColumns(t *testing.T) {
	type args struct {
		model any
		sort  string
	}
	tests := []struct {
		name string
		args args
		want []clause.OrderByColumn
	}{
		// TODO: Add test cases.
		{
			name: "asc sort",
			args: args{
				model: struct {
					ID uint `json:"id,omitempty" sort:"id"`
				}{},
				sort: "id",
			},
			want: []clause.OrderByColumn{
				getOrderByColumn("id", false),
			},
		},
		{
			name: "desc sort",
			args: args{
				model: struct {
					ID uint `json:"id,omitempty"`
				}{},
				sort: "-id",
			},
			want: []clause.OrderByColumn{
				getOrderByColumn("id", true),
			},
		},
		{
			name: "gorm column name sort",
			args: args{
				model: struct {
					ID uint `json:"id,omitempty" gorm:"column:post_id" sort:"id"`
				}{},
				sort: "id",
			},
			want: []clause.OrderByColumn{
				getOrderByColumn("post_id", false),
			},
		},
		{
			name: "multi sort",
			args: args{
				model: struct {
					ID        uint      `json:"id,omitempty"`
					CreatedAt time.Time `json:"created_at,omitempty" sort:"created_at"`
				}{},
				sort: "id,created_at",
			},
			want: []clause.OrderByColumn{
				getOrderByColumn("id", false),
				getOrderByColumn("created_at", false),
			},
		},
		{
			name: "nested sort",
			args: args{
				model: struct {
					User *struct {
						ID uint `json:"id,omitempty" sort:"id"`
					} `json:"user,omitempty" sort:"User"`
				}{},
				sort: "User.id",
			},
			want: []clause.OrderByColumn{
				getOrderByColumn("User.id", false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sort2OrderByColumns(tt.args.model, tt.args.sort); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sort2OrderByColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSearchColumns(t *testing.T) {
	type args struct {
		model any
	}
	tests := []struct {
		name string
		args args
		want []clause.Column
	}{
		// TODO: Add test cases.
		{
			name: "current table field",
			args: args{
				model: struct {
					Title   string `json:"title,omitempty" search:"title"`
					Content string `json:"content,omitempty" search:"content"`
				}{},
			},
			want: []clause.Column{
				getColumn("title"),
				getColumn("content"),
			},
		},
		{
			name: "gorm column name",
			args: args{
				model: struct {
					Title   string `json:"title,omitempty" search:"title"`
					Content string `json:"content,omitempty" gorm:"column:content2" search:"content"`
				}{},
			},
			want: []clause.Column{
				getColumn("title"),
				getColumn("content2"),
			},
		},
		{
			name: "nested table field",
			args: args{
				model: struct {
					Title string `json:"title,omitempty" search:"title"`
					User  struct {
						Content string `json:"content,omitempty" search:"content"`
					} `search:"User"`
				}{},
			},
			want: []clause.Column{
				getColumn("title"),
				getColumn("User.content"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSearchColumns(tt.args.model); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSearchColumns() = %v, want %v", got, tt.want)
			}
		})
	}
}
