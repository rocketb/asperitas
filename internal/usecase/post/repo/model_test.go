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
	dbPost := dbPost{
		ID:          uuid.UUID{},
		Type:        "text",
		Title:       "title",
		Category:    "category",
		Body:        "body",
		Score:       sql.NullInt32{Int32: 1, Valid: true},
		Views:       1,
		DateCreated: time.Time{},
		UserID:      uuid.UUID{},
	}
	corePost := post.Post{
		ID:          uuid.UUID{},
		Type:        "text",
		Title:       "title",
		Body:        "body",
		Category:    "category",
		Score:       1,
		Views:       1,
		DateCreated: time.Time{},
		UserID:      uuid.UUID{},
	}

	assert.Equal(t, corePost, toCorePost(dbPost))
}

func TestToDBPost(t *testing.T) {
	corePost := post.Post{
		ID:          uuid.UUID{},
		Type:        "text",
		Title:       "title",
		Body:        "body",
		Category:    "category",
		Views:       1,
		DateCreated: time.Time{},
		UserID:      uuid.UUID{},
	}
	dbP := dbPost{
		ID:          uuid.UUID{},
		Type:        "text",
		Title:       "title",
		Category:    "category",
		Body:        "body",
		Views:       1,
		DateCreated: time.Time{},
		UserID:      uuid.UUID{},
	}
	assert.Equal(t, dbP, toDBPost(corePost))
}

func TestToCorePosts(t *testing.T) {
	corePosts := []post.Post{{
		ID:          uuid.UUID{},
		Type:        "text",
		Title:       "title",
		Body:        "body",
		Category:    "category",
		Views:       1,
		DateCreated: time.Time{},
		UserID:      uuid.UUID{},
	}}
	dbPosts := []dbPost{{
		ID:          uuid.UUID{},
		Type:        "text",
		Title:       "title",
		Category:    "category",
		Body:        "body",
		Views:       1,
		DateCreated: time.Time{},
		UserID:      uuid.UUID{},
	}}
	assert.Equal(t, corePosts, toCorePosts(dbPosts))
}

func TestToCoreComment(t *testing.T) {
	dbComm := dbComment{
		ID:          uuid.UUID{},
		PostID:      uuid.UUID{},
		Body:        "",
		DateCreated: time.Time{},
	}
	comm := post.Comment{
		ID:          uuid.UUID{},
		PostID:      uuid.UUID{},
		DateCreated: time.Time{},
		Body:        "",
	}

	assert.Equal(t, comm, toCoreComment(dbComm))
}

func TestToDBComment(t *testing.T) {
	dbComm := dbComment{
		ID:     uuid.UUID{},
		PostID: uuid.UUID{},
		Body:        "",
		DateCreated: time.Time{},
	}
	comm := post.Comment{
		ID:          uuid.UUID{},
		PostID:      uuid.UUID{},
		DateCreated: time.Time{},
		Body:        "",
	}

	assert.Equal(t, dbComm, toDBComment(comm))
}

func TestToCoreComments(t *testing.T) {
	dbComms := []dbComment{{
		ID:          uuid.UUID{},
		PostID:      uuid.UUID{},
		Body:        "",
		DateCreated: time.Time{},
	}}
	comms := []post.Comment{{
		ID:          uuid.UUID{},
		PostID:      uuid.UUID{},
		DateCreated: time.Time{},
		Body:        "",
	}}

	assert.Equal(t, comms, toCoreComments(dbComms))
}

func TestToCoreVote(t *testing.T) {
	dbVote := dbVote{
		PostID: uuid.UUID{},
		UserID: uuid.UUID{},
		Vote:   0,
	}
	vote := post.Vote{
		Vote: 0,
		User: uuid.UUID{},
	}

	assert.Equal(t, vote, toCoreVote(dbVote))
}

func TestToCoreVotes(t *testing.T) {
	dbVotes := []dbVote{{
		PostID: uuid.UUID{},
		UserID: uuid.UUID{},
		Vote:   0,
	}}
	votes := []post.Vote{{
		Vote: 0,
		User: uuid.UUID{},
	}}

	assert.Equal(t, votes, toCoreVotes(dbVotes))
}

func TestToDBVote(t *testing.T) {
	id := uuid.New()
	vote := post.Vote{
		Vote: 0,
		User: uuid.UUID{},
	}
	dbVote := dbVote{
		PostID: id,
		UserID: uuid.UUID{},
		Vote:   0,
	}

	assert.Equal(t, dbVote, toDBVote(id, vote))
}
