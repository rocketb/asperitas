package repo

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var curDate = time.Now()

func TestMemory_GetAll(t *testing.T) {
	puuid := uuid.New()
	uuuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	tests := []struct {
		name    string
		want    []*post.Post
		fields  fields
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "get items from empty list should return []",
			fields:  fields{posts: map[uuid.UUID]*PostDB{}},
			want:    []*post.Post{},
			wantErr: assert.NoError,
		},
		{
			name: "get one item should return that item",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:          puuid,
							Type:        "text",
							Title:       "title",
							Category:    "category",
							Body:        "body",
							Score:       1,
							Views:       1,
							DateCreated: curDate,
							UserID:      uuuid,
						},
					},
				},
			},
			wantErr: assert.NoError,
			want: []*post.Post{
				{
					ID:          puuid,
					Type:        "text",
					Title:       "title",
					Category:    "category",
					Body:        "body",
					Score:       1,
					Views:       1,
					DateCreated: curDate,
					UserID:      uuuid,
				},
			},
		},
		{
			name: "get post of not existing author should not produce error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:          puuid,
							Type:        "text",
							Title:       "title",
							Category:    "category",
							Body:        "body",
							Score:       1,
							Views:       1,
							DateCreated: curDate,
							UserID:      uuuid,
						},
					},
				},
			},
			wantErr: assert.NoError,
			want: []*post.Post{
				{
					ID:          puuid,
					Type:        "text",
					Title:       "title",
					Category:    "category",
					Body:        "body",
					Score:       1,
					Views:       1,
					DateCreated: curDate,
					UserID:      uuuid,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.GetAll(context.Background())
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetAll()")
		})
	}
}

func TestMemory_GetByCatName(t *testing.T) {
	puuid := uuid.New()
	uuuid := uuid.New()
	type fields struct{ posts map[uuid.UUID]*PostDB }
	type args struct {
		catName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*post.Post
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "get one item should return that item",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:          puuid,
							Type:        "type",
							Title:       "title",
							Category:    "category_0",
							Body:        "body",
							Score:       1,
							Views:       1,
							DateCreated: curDate,
							UserID:      uuuid,
						},
					},
				},
			},
			args: args{
				catName: "category_0",
			},
			want: []*post.Post{
				{
					ID:          puuid,
					Type:        "type",
					Title:       "title",
					Body:        "body",
					Category:    "category_0",
					Score:       1,
					Views:       1,
					DateCreated: curDate,
					UserID:      uuuid,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "get items from empty repo should not produce an error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				catName: "not_exist",
			},
			want:    []*post.Post{},
			wantErr: assert.NoError,
		},
		{
			name: "get post of not existing category should not produce error",
			args: args{catName: "not_exist"},
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:          puuid,
							Type:        "text",
							Title:       "title",
							Category:    "category_1",
							Body:        "body",
							Score:       1,
							Views:       1,
							DateCreated: curDate,
							UserID:      uuuid,
						},
					},
				},
			},
			wantErr: assert.NoError,
			want:    []*post.Post{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.GetByCatName(context.Background(), tt.args.catName)
			if !tt.wantErr(t, err, fmt.Sprintf("GetByCatName(%v)", tt.args.catName)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetByCatName(%v)", tt.args.catName)
		})
	}
}

func TestMemory_GetByUserID(t *testing.T) {
	puuid := uuid.New()
	uuuid := uuid.New()
	type fields struct{ posts map[uuid.UUID]*PostDB }
	type args struct {
		userID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*post.Post
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "get item should return that item",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:     puuid,
							UserID: uuuid,
						},
					},
				},
			},
			args: args{
				userID: uuuid,
			},
			want: []*post.Post{
				{
					ID:     puuid,
					UserID: uuuid,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "get items with empty repo should not produce an error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				userID: uuuid,
			},
			want:    []*post.Post{},
			wantErr: assert.NoError,
		},
		{
			name: "get post of not existing uid should not produce an error",
			args: args{userID: uuuid},
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:     puuid,
							UserID: uuid.New(),
						},
					},
				},
			},
			wantErr: assert.NoError,
			want:    []*post.Post{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.GetByUserID(context.Background(), tt.args.userID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetByCatName(%v)", tt.args.userID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetByCatName(%v)", tt.args.userID)
		})
	}
}

func TestMemory_GetByID(t *testing.T) {
	puuid := uuid.New()
	uuuid := uuid.New()
	u2uuid := uuid.New()

	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *post.Post
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "get item from empty repo should return err",
			fields:  fields{posts: map[uuid.UUID]*PostDB{}},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "get one item should return that item",
			args: args{postID: puuid},
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:          puuid,
							Type:        "text",
							Title:       "title",
							Category:    "category",
							Body:        "body",
							Score:       1,
							Views:       1,
							DateCreated: curDate,
							UserID:      uuuid,
						},
					},
				},
			},
			wantErr: assert.NoError,
			want: &post.Post{
				ID:          puuid,
				Type:        "text",
				Title:       "title",
				Category:    "category",
				Body:        "body",
				Score:       1,
				Views:       1,
				DateCreated: curDate,
				UserID:      uuuid,
			},
		},
		{
			name: "get post of not existing author should not produce error",
			args: args{postID: puuid},
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:          puuid,
							Type:        "text",
							Title:       "title",
							Category:    "category",
							Body:        "body",
							Score:       1,
							Views:       1,
							DateCreated: curDate,
							UserID:      u2uuid,
						},
					},
				},
			},
			wantErr: assert.NoError,
			want: &post.Post{
				ID:          puuid,
				Type:        "text",
				Title:       "title",
				Category:    "category",
				Body:        "body",
				Score:       1,
				Views:       1,
				DateCreated: curDate,
				UserID:      u2uuid,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.GetByID(context.Background(), tt.args.postID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetByID(ctx, %v)", tt.args.postID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetByID(ctx, %v)", tt.args.postID)
		})
	}
}

func TestMemory_Add(t *testing.T) {
	uuuid := uuid.New()

	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		newPost *post.Post
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uuid.UUID
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "create post should not produce error",
			fields: fields{posts: make(map[uuid.UUID]*PostDB)},
			args: args{
				newPost: &post.Post{
					Type:        "text",
					Title:       "title",
					Category:    "category",
					Body:        "body",
					Score:       1,
					Views:       0,
					DateCreated: curDate,
					UserID:      uuuid,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.Add(context.Background(), tt.args.newPost)
			if !tt.wantErr(t, err, fmt.Sprintf("Add(ctx, %v)", tt.args.newPost)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Add(ctx, %v)", tt.args.newPost)
		})
	}
}

func TestMemory_Delete(t *testing.T) {
	puuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
		want    map[uuid.UUID]*PostDB
	}{
		{
			name: "delete should not produce an error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
					},
				},
			},
			args: args{
				postID: puuid,
			},
			wantErr: assert.NoError,
			want:    map[uuid.UUID]*PostDB{},
		},
		{
			name: "delete from from empty repo should not produce an error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {},
				},
			},
			wantErr: assert.NoError,
			want: map[uuid.UUID]*PostDB{
				puuid: {},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			tt.wantErr(t, r.Delete(context.Background(), tt.args.postID), fmt.Sprintf("Delete(ctx, %v)", tt.args.postID))
			assert.Equal(t, tt.want, r.posts)
		})
	}
}

func TestMemory_GetVotes(t *testing.T) {
	puuid := uuid.New()
	uuuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*post.Vote
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "get votes",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
						votes: []*post.Vote{
							{
								Vote: 1,
								User: uuuid,
							},
						},
					},
				},
			},
			args: args{
				postID: puuid,
			},
			want: []*post.Vote{
				{
					Vote: 1,
					User: uuuid,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "err if post not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: puuid,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.GetVotes(context.Background(), tt.args.postID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetVotes(ctx, %v)", tt.args.postID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetVotes(ctx, %v)", tt.args.postID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemory_AddVote(t *testing.T) {
	puuid := uuid.New()
	uuuid := uuid.New()
	u2uuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
		Vote   *post.Vote
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[uuid.UUID]*PostDB
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "vote with empty repo is empty should produce error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: puuid,
				Vote: &post.Vote{
					Vote: 1,
					User: uuuid,
				},
			},
			want:    make(map[uuid.UUID]*PostDB),
			wantErr: assert.Error,
		},
		{
			name: "vote should produce error if post not exist",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
					},
				},
			},
			args: args{
				postID: uuid.New(),
				Vote: &post.Vote{
					Vote: 1,
					User: puuid,
				},
			},
			wantErr: assert.Error,
			want: map[uuid.UUID]*PostDB{
				puuid: {
					data: &post.Post{
						ID: puuid,
					},
				},
			},
		},
		{
			name: "upvote should be counted and not produce error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:     puuid,
							UserID: uuuid,
							Score:  0,
						},
						votes: []*post.Vote{},
					},
				},
			},
			args: args{
				postID: puuid,
				Vote: &post.Vote{
					Vote: 1,
					User: uuuid,
				},
			},
			want: map[uuid.UUID]*PostDB{
				puuid: {
					data: &post.Post{
						ID:     puuid,
						UserID: uuuid,
						Score:  1,
					},
					votes: []*post.Vote{
						{
							Vote: 1,
							User: uuuid,
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "downvote should be counted and not produce error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:     puuid,
							UserID: uuuid,
							Score:  0,
						},
						votes: []*post.Vote{},
					},
				},
			},
			args: args{
				postID: puuid,
				Vote: &post.Vote{
					Vote: -1,
					User: uuuid,
				},
			},
			want: map[uuid.UUID]*PostDB{
				puuid: {
					data: &post.Post{
						ID:     puuid,
						UserID: uuuid,
						Score:  -1,
					},
					votes: []*post.Vote{
						{
							Vote: -1,
							User: uuuid,
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "previous vote should be updated if exists",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:     puuid,
							UserID: uuuid,
							Score:  2,
						},
						votes: []*post.Vote{
							{Vote: 1, User: uuuid},
							{Vote: 1, User: u2uuid},
						},
					},
				},
			},
			args: args{
				postID: puuid,
				Vote: &post.Vote{
					Vote: -1,
					User: uuuid,
				},
			},
			want: map[uuid.UUID]*PostDB{
				puuid: {
					data: &post.Post{
						ID:     puuid,
						UserID: uuuid,
						Score:  1,
					},
					votes: []*post.Vote{
						{Vote: -1, User: uuuid},
						{Vote: 1, User: u2uuid},
					},
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			err := r.AddVote(context.Background(), tt.args.postID, tt.args.Vote)
			if !tt.wantErr(t, err, fmt.Sprintf("AddVote(ctx, %v, %v)", tt.args.postID, tt.args.Vote)) {
				return
			}
			assert.Equalf(t, tt.want, r.posts, "AddVote(%v, %v)", tt.args.postID, tt.args.Vote)
		})
	}
}

func TestMemory_GetComments(t *testing.T) {
	puuid := uuid.New()
	cuuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*post.Comment
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "get comments",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
						comments: map[uuid.UUID]*post.Comment{
							cuuid: {
								ID:          cuuid,
								DateCreated: curDate,
								Body:        "comment body",
							},
						},
					},
				},
			},
			args: args{
				postID: puuid,
			},
			want: []*post.Comment{
				{
					ID:          cuuid,
					DateCreated: curDate,
					Body:        "comment body",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "err if post not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: puuid,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.GetComments(context.Background(), tt.args.postID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetComments(ctx, %v)", tt.args.postID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetCommentByID(ctx, %v)", tt.args.postID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemory_GetCommentByID(t *testing.T) {
	puuid := uuid.New()
	cuuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID    uuid.UUID
		commentID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *post.Comment
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "get comment",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
						comments: map[uuid.UUID]*post.Comment{
							cuuid: {
								ID:          cuuid,
								DateCreated: curDate,
								Body:        "comment body",
							},
						},
					},
				},
			},
			args: args{
				postID:    puuid,
				commentID: cuuid,
			},
			want: &post.Comment{
				ID:          cuuid,
				DateCreated: curDate,
				Body:        "comment body",
			},
			wantErr: assert.NoError,
		},
		{
			name: "err if post not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: puuid,
			},
			wantErr: assert.Error,
		},
		{
			name: "err if comment not exist",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
						comments: map[uuid.UUID]*post.Comment{
							cuuid: {
								ID: cuuid,
							},
						},
					},
				},
			},
			args: args{
				postID:    puuid,
				commentID: uuid.New(),
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.GetCommentByID(context.Background(), tt.args.commentID)
			if !tt.wantErr(t, err, fmt.Sprintf("GetCommentByID(ctx, %v, %v)", tt.args.postID, tt.args.commentID)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetCommentByID(ctx, %v, %v)", tt.args.postID, tt.args.commentID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemory_AddComment(t *testing.T) {
	puuid := uuid.New()
	uuuid := uuid.New()
	cuuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		newComment *post.Comment
		postID     uuid.UUID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      uuid.UUID
		wantPosts map[uuid.UUID]*PostDB
		wantErr   assert.ErrorAssertionFunc
	}{
		{
			name: "add comment should not produce error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID:     puuid,
							UserID: uuuid,
						},
						comments: make(map[uuid.UUID]*post.Comment),
					},
				},
			},
			args: args{
				postID: puuid,
				newComment: &post.Comment{
					ID:          cuuid,
					DateCreated: curDate,
					Body:        "some comment text",
				},
			},
			want: cuuid,
			wantPosts: map[uuid.UUID]*PostDB{
				puuid: {
					data: &post.Post{
						ID:     puuid,
						UserID: uuuid,
					},
					comments: map[uuid.UUID]*post.Comment{
						cuuid: {
							ID:          cuuid,
							DateCreated: curDate,
							Body:        "some comment text",
						},
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "comment should produce error if post is not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: puuid,
				newComment: &post.Comment{
					DateCreated: curDate,
					Body:        "body",
				},
			},
			wantPosts: make(map[uuid.UUID]*PostDB),
			wantErr:   assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			got, err := r.AddComment(context.Background(), tt.args.postID, tt.args.newComment)
			if !tt.wantErr(t, err, fmt.Sprintf("AddComment(ctx, %v, %v)", tt.args.postID, tt.args.newComment)) {
				return
			}
			assert.Equalf(t, tt.want, got, "AddComment(ctx, %v, %v)", tt.args.postID, tt.args.newComment)
			assert.Equal(t, tt.wantPosts, r.posts)
		})
	}
}

func TestMemory_DeleteComment(t *testing.T) {
	puuid := uuid.New()
	cuuid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID    uuid.UUID
		commentID uuid.UUID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[uuid.UUID]*PostDB
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "delete comment of not existing post should produce an error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID:    puuid,
				commentID: cuuid,
			},
			want:    make(map[uuid.UUID]*PostDB),
			wantErr: assert.Error,
		},
		{
			name: "delete not existing comment should not produce an error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
						comments: make(map[uuid.UUID]*post.Comment),
					},
				},
			},
			args: args{
				postID:    puuid,
				commentID: cuuid,
			},
			want: map[uuid.UUID]*PostDB{
				puuid: {
					data: &post.Post{
						ID: puuid,
					},
					comments: make(map[uuid.UUID]*post.Comment),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "delete existing comment should remove it",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					puuid: {
						data: &post.Post{
							ID: puuid,
						},
						comments: map[uuid.UUID]*post.Comment{
							cuuid: {
								ID:          cuuid,
								DateCreated: curDate,
								Body:        "comment text",
							},
						},
					},
				},
			},
			args: args{
				postID:    puuid,
				commentID: cuuid,
			},
			want: map[uuid.UUID]*PostDB{
				puuid: {
					data: &post.Post{
						ID: puuid,
					},
					comments: make(map[uuid.UUID]*post.Comment),
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}
			err := r.DeleteComment(context.Background(), tt.args.postID, tt.args.commentID)
			if !tt.wantErr(t, err, fmt.Sprintf("DeleteComment(ctx, %v, %v)", tt.args.postID, tt.args.commentID)) {
				return
			}
			assert.Equalf(t, tt.want, r.posts, "Posts in repo should be equal")
		})
	}
}

func TestNewMemory(t *testing.T) {
	r := NewMemory()
	assert.NotNil(t, r)
}
