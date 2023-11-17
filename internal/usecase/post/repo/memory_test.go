package repo

import (
	"context"
	"sync"
	"testing"

	"github.com/rocketb/asperitas/internal/usecase/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMemory_GetAll(t *testing.T) {
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	tests := []struct {
		name   string
		posts  []*post.Post
		err    error
		fields fields
	}{
		{
			name:   "get items from empty list should return []",
			fields: fields{posts: map[uuid.UUID]*PostDB{}},
			posts:  []*post.Post{},
		},
		{
			name: "get one item should return that item",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					uuid.New(): {
						data: &post.Post{},
					},
				},
			},
			posts: []*post.Post{{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			got, err := r.GetAll(context.Background())
			assert.Equal(t, tt.err, err)
			assert.Equalf(t, tt.posts, got, "GetAll()")
		})
	}
}

func TestMemory_GetByCatName(t *testing.T) {
	type fields struct{ posts map[uuid.UUID]*PostDB }
	type args struct {
		catName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		posts  []*post.Post
		err    error
	}{
		{
			name: "get item should return that item",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					uuid.New(): {
						data: &post.Post{
							Category: "category_0",
						},
					},
				},
			},
			args: args{
				catName: "category_0",
			},
			posts: []*post.Post{
				{
					Category: "category_0",
				},
			},
		},
		{
			name: "get items from empty repo should not produce an error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				catName: "not_exist",
			},
			posts: []*post.Post{},
		},
		{
			name: "get not existing category should not produce error",
			args: args{catName: "not_exist"},
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					uuid.New(): {
						data: &post.Post{
							Category: "category_1",
						},
					},
				},
			},
			posts: []*post.Post{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			got, err := r.GetByCatName(context.Background(), tt.args.catName)
			assert.Equal(t, tt.err, err)
			assert.Equalf(t, tt.posts, got, "GetByCatName(%v)", tt.args.catName)
		})
	}
}

func TestMemory_GetByUserID(t *testing.T) {
	uid := uuid.New()
	type fields struct{ posts map[uuid.UUID]*PostDB }
	type args struct {
		userID uuid.UUID
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*post.Post
		err    error
	}{
		{
			name: "get item should return that item",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					uuid.New(): {
						data: &post.Post{
							UserID: uid,
						},
					},
				},
			},
			args: args{
				userID: uid,
			},
			want: []*post.Post{
				{
					UserID: uid,
				},
			},
		},
		{
			name: "get items with empty repo should not produce an error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				userID: uid,
			},
			want: []*post.Post{},
		},
		{
			name: "get post of not existing uid should not produce an error",
			args: args{userID: uid},
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					uuid.New(): {
						data: &post.Post{
							UserID: uuid.New(),
						},
					},
				},
			},
			want: []*post.Post{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			got, err := r.GetByUserID(context.Background(), tt.args.userID)
			assert.Equal(t, tt.err, err)
			assert.Equalf(t, tt.want, got, "GetByCatName(%v)", tt.args.userID)
		})
	}
}

func TestMemory_GetByID(t *testing.T) {
	pid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *post.Post
		err    error
	}{
		{
			name:   "get item from empty repo should return err",
			fields: fields{posts: map[uuid.UUID]*PostDB{}},
			want:   nil,
			err:    post.ErrNotFound,
		},
		{
			name: "get one item should return that item",
			args: args{postID: pid},
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{
							ID: pid,
						},
					},
				},
			},
			want: &post.Post{
				ID: pid,
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
			assert.Equal(t, tt.err, err)
			assert.Equalf(t, tt.want, got, "GetByID(ctx, %v)", tt.args.postID)
		})
	}
}

func TestMemory_Add(t *testing.T) {
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		newPost *post.Post
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		err    error
	}{
		{
			name:   "create post should not produce error",
			fields: fields{posts: make(map[uuid.UUID]*PostDB)},
			args: args{
				newPost: &post.Post{
					ID: uuid.New(),
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

			_, err := r.Add(context.Background(), tt.args.newPost)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestMemory_Delete(t *testing.T) {
	pid := uuid.New()
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
					pid: {
						data: &post.Post{
							ID: pid,
						},
					},
				},
			},
			args: args{
				postID: pid,
			},
			wantErr: assert.NoError,
			want:    map[uuid.UUID]*PostDB{},
		},
		{
			name: "delete from from empty repo should not produce an error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {},
				},
			},
			wantErr: assert.NoError,
			want: map[uuid.UUID]*PostDB{
				pid: {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			tt.wantErr(t, r.Delete(context.Background(), tt.args.postID))
			assert.Equal(t, tt.want, r.posts)
		})
	}
}

func TestMemory_GetVotes(t *testing.T) {
	pid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		posts  []*post.Vote
		err    error
	}{
		{
			name: "get votes",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data:  &post.Post{ID: pid},
						votes: []*post.Vote{{Vote: 1}},
					},
				},
			},
			args:  args{postID: pid},
			posts: []*post.Vote{{Vote: 1}},
		},
		{
			name: "err if post not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: pid,
			},
			err: post.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			got, err := r.GetVotes(context.Background(), tt.args.postID)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.posts, got)
		})
	}
}

func TestMemory_AddVote(t *testing.T) {
	pid := uuid.New()
	uid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
		Vote   *post.Vote
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		posts  map[uuid.UUID]*PostDB
		err    error
	}{
		{
			name: "vote with empty repo is empty should produce error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: pid,
				Vote: &post.Vote{
					Vote: 1,
					User: uid,
				},
			},
			posts: make(map[uuid.UUID]*PostDB),
			err:   post.ErrNotFound,
		},
		{
			name: "upvote should be counted and not produce error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{
							ID:     pid,
							UserID: uid,
							Score:  0,
						},
						votes: []*post.Vote{},
					},
				},
			},
			args: args{
				postID: pid,
				Vote: &post.Vote{
					Vote: 1,
					User: uid,
				},
			},
			posts: map[uuid.UUID]*PostDB{
				pid: {
					data: &post.Post{
						ID:     pid,
						UserID: uid,
						Score:  1,
					},
					votes: []*post.Vote{
						{
							Vote: 1,
							User: uid,
						},
					},
				},
			},
		},
		{
			name: "previous vote should be updated if exists",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{
							ID:     pid,
							UserID: uid,
							Score:  2,
						},
						votes: []*post.Vote{
							{Vote: 1, User: uid},
						},
					},
				},
			},
			args: args{
				postID: pid,
				Vote: &post.Vote{
					Vote: -1,
					User: uid,
				},
			},
			posts: map[uuid.UUID]*PostDB{
				pid: {
					data: &post.Post{
						ID:     pid,
						UserID: uid,
						Score:  1,
					},
					votes: []*post.Vote{
						{Vote: -1, User: uid},
					},
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

			err := r.AddVote(context.Background(), tt.args.postID, tt.args.Vote)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.posts, r.posts)
		})
	}
}

func TestMemory_GetComments(t *testing.T) {
	pid := uuid.New()
	cid := uuid.New()
	type fields struct {
		posts map[uuid.UUID]*PostDB
	}
	type args struct {
		postID uuid.UUID
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		posts  []*post.Comment
		err    error
	}{
		{
			name: "get comments",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{ID: pid},
						comments: map[uuid.UUID]*post.Comment{
							cid: {ID: cid},
						},
					},
				},
			},
			args:  args{postID: pid},
			posts: []*post.Comment{{ID: cid}},
		},
		{
			name: "err if post not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{postID: pid},
			err:  post.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			got, err := r.GetComments(context.Background(), tt.args.postID)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.posts, got)
		})
	}
}

func TestMemory_GetCommentByID(t *testing.T) {
	pid := uuid.New()
	cid := uuid.New()
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
		err error
	}{
		{
			name: "get comment",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{ID: pid},
						comments: map[uuid.UUID]*post.Comment{
							cid: {ID: cid},
						},
					},
				},
			},
			args: args{
				postID:    pid,
				commentID: cid,
			},
			want: &post.Comment{ID: cid},
		},
		{
			name: "err if post not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{postID: pid},
			err: post.ErrCommentNotFound,
		},
		{
			name: "err if comment not exist",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{ID: pid},
						comments: map[uuid.UUID]*post.Comment{
							cid: {ID: cid},
						},
					},
				},
			},
			args: args{
				postID:    pid,
				commentID: uuid.New(),
			},
			err: post.ErrCommentNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			got, err := r.GetCommentByID(context.Background(), tt.args.commentID)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemory_AddComment(t *testing.T) {
	pid := uuid.New()
	uid := uuid.New()
	cid := uuid.New()
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
		posts map[uuid.UUID]*PostDB
		err   error
	}{
		{
			name: "add comment should not produce error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{
							ID:     pid,
							UserID: uid,
						},
						comments: make(map[uuid.UUID]*post.Comment),
					},
				},
			},
			args: args{
				postID: pid,
				newComment: &post.Comment{
					ID:          cid,
				},
			},
			want: cid,
			posts: map[uuid.UUID]*PostDB{
				pid: {
					data: &post.Post{
						ID:     pid,
						UserID: uid,
					},
					comments: map[uuid.UUID]*post.Comment{
						cid: {
							ID:          cid,
						},
					},
				},
			},
		},
		{
			name: "comment should produce error if post is not exist",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID: pid,
				newComment: &post.Comment{},
			},
			posts: make(map[uuid.UUID]*PostDB),
			err: post.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Memory{
				mu:    &sync.RWMutex{},
				posts: tt.fields.posts,
			}

			got, err := r.AddComment(context.Background(), tt.args.postID, tt.args.newComment)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.posts, r.posts)
		})
	}
}

func TestMemory_DeleteComment(t *testing.T) {
	pid := uuid.New()
	cid := uuid.New()
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
		err     error
	}{
		{
			name: "delete comment of not existing post should produce an error",
			fields: fields{
				posts: make(map[uuid.UUID]*PostDB),
			},
			args: args{
				postID:    pid,
				commentID: cid,
			},
			want:    make(map[uuid.UUID]*PostDB),
			err: post.ErrNotFound,
		},
		{
			name: "delete not existing comment should not produce an error",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{
							ID: pid,
						},
						comments: make(map[uuid.UUID]*post.Comment),
					},
				},
			},
			args: args{
				postID:    pid,
				commentID: cid,
			},
			want: map[uuid.UUID]*PostDB{
				pid: {
					data: &post.Post{ID: pid},
					comments: make(map[uuid.UUID]*post.Comment),
				},
			},
		},
		{
			name: "delete existing comment should remove it",
			fields: fields{
				posts: map[uuid.UUID]*PostDB{
					pid: {
						data: &post.Post{ID: pid},
						comments: map[uuid.UUID]*post.Comment{cid: {ID: cid}},
					},
				},
			},
			args: args{
				postID:    pid,
				commentID: cid,
			},
			want: map[uuid.UUID]*PostDB{
				pid: {
					data: &post.Post{
						ID: pid,
					},
					comments: make(map[uuid.UUID]*post.Comment),
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

			err := r.DeleteComment(context.Background(), tt.args.postID, tt.args.commentID)
			assert.Equal(t, tt.err, err)
			assert.Equalf(t, tt.want, r.posts, "Posts in repo should be equal")
		})
	}
}

func TestNewMemory(t *testing.T) {
	r := NewMemory()
	assert.NotNil(t, r)
}
