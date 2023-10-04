package repo

import (
	"database/sql"
	"testing"
	"time"

	"github.com/rocketb/asperitas/internal/usecase/post"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToCorePost(t *testing.T) {
	id := uuid.New()
	date := time.Now()
	dbP := dbPost{
		ID:          id,
		Type:        "text",
		Title:       "title",
		Category:    "category",
		Body:        "body",
		Score:       sql.NullInt32{Int32: 1, Valid: true},
		Views:       1,
		DateCreated: date,
		UserID:      id,
	}
	corePost := post.Post{
		ID:          id,
		Type:        "text",
		Title:       "title",
		Body:        "body",
		Category:    "category",
		Score:       1,
		Views:       1,
		DateCreated: date,
		UserID:      id,
	}

	assert.Equal(t, corePost, toCorePost(dbP))
}

func TestToDBPost(t *testing.T) {
	id := uuid.New()
	date := time.Now()
	corePost := post.Post{
		ID:          id,
		Type:        "text",
		Title:       "title",
		Body:        "body",
		Category:    "category",
		Views:       1,
		DateCreated: date,
		UserID:      id,
	}
	dbP := dbPost{
		ID:          id,
		Type:        "text",
		Title:       "title",
		Category:    "category",
		Body:        "body",
		Views:       1,
		DateCreated: date,
		UserID:      id,
	}
	assert.Equal(t, dbP, toDBPost(corePost))
}

func TestToCorePosts(t *testing.T) {
	id := uuid.New()
	date := time.Now()
	corePosts := []post.Post{{
		ID:          id,
		Type:        "text",
		Title:       "title",
		Body:        "body",
		Category:    "category",
		Views:       1,
		DateCreated: date,
		UserID:      id,
	}}
	dbPosts := []dbPost{{
		ID:          id,
		Type:        "text",
		Title:       "title",
		Category:    "category",
		Body:        "body",
		Views:       1,
		DateCreated: date,
		UserID:      id,
	}}
	assert.Equal(t, corePosts, toCorePosts(dbPosts))
}

func TestToCoreComment(t *testing.T) {
	dbComm := dbComment{
		ID:          [16]byte{},
		PostID:      [16]byte{},
		Body:        "",
		DateCreated: time.Time{},
	}
	comm := post.Comment{
		ID:          [16]byte{},
		PostID:      [16]byte{},
		DateCreated: time.Time{},
		Body:        "",
	}

	assert.Equal(t, comm, toCoreComment(dbComm))
}

func TestToDBComment(t *testing.T) {
	dbComm := dbComment{
		ID:     [16]byte{},
		PostID: [16]byte{},
		// Author:      [16]byte{},
		Body:        "",
		DateCreated: time.Time{},
	}
	comm := post.Comment{
		ID:          [16]byte{},
		PostID:      [16]byte{},
		DateCreated: time.Time{},
		Body:        "",
	}

	assert.Equal(t, dbComm, toDBComment(comm))
}

func TestToCoreComments(t *testing.T) {
	dbComms := []dbComment{{
		ID:          [16]byte{},
		PostID:      [16]byte{},
		Body:        "",
		DateCreated: time.Time{},
	}}
	comms := []post.Comment{{
		ID:          [16]byte{},
		PostID:      [16]byte{},
		DateCreated: time.Time{},
		Body:        "",
	}}

	assert.Equal(t, comms, toCoreComments(dbComms))
}

func TestToCoreVote(t *testing.T) {
	dbVote := dbVote{
		PostID: [16]byte{},
		UserID: [16]byte{},
		Vote:   0,
	}
	vote := post.Vote{
		Vote: 0,
		User: [16]byte{},
	}

	assert.Equal(t, vote, toCoreVote(dbVote))
}

func TestToCoreVotes(t *testing.T) {
	dbVotes := []dbVote{{
		PostID: [16]byte{},
		UserID: [16]byte{},
		Vote:   0,
	}}
	votes := []post.Vote{{
		Vote: 0,
		User: [16]byte{},
	}}

	assert.Equal(t, votes, toCoreVotes(dbVotes))
}

func TestToDBVote(t *testing.T) {
	id := uuid.New()
	vote := post.Vote{
		Vote: 0,
		User: [16]byte{},
	}
	dbVote := dbVote{
		PostID: id,
		UserID: [16]byte{},
		Vote:   0,
	}
	assert.Equal(t, dbVote, toDBVote(id, vote))
}
