package v1

import (
	"net/http"

	"github.com/rocketb/asperitas/internal/handlers/v1/postgrp"
	"github.com/rocketb/asperitas/internal/handlers/v1/usergrp"
	"github.com/rocketb/asperitas/internal/usecase/post"
	postrepo "github.com/rocketb/asperitas/internal/usecase/post/repo"
	"github.com/rocketb/asperitas/internal/usecase/user"
	userrepo "github.com/rocketb/asperitas/internal/usecase/user/repo"
	"github.com/rocketb/asperitas/internal/web/auth"
	"github.com/rocketb/asperitas/internal/web/middleware"
	"github.com/rocketb/asperitas/pkg/logger"
	"github.com/rocketb/asperitas/pkg/web"

	"github.com/jmoiron/sqlx"
)

// Config represents routes configuration.
type Config struct {
	Build string
	Log   *logger.Logger
	Auth  auth.Auth
	DB    *sqlx.DB
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	usersRepo := userrepo.NewPostgres(cfg.DB, cfg.Log)
	postsRepo := postrepo.NewPostgres(cfg.DB, cfg.Log)

	postsHandler := &postgrp.PostsHandler{
		Posts: post.NewCore(postsRepo),
		Users: user.NewCore(usersRepo),
	}

	usersHandler := &usergrp.UserHandler{
		Logger: cfg.Log,
		Users:  user.NewCore(usersRepo),
		Auth:   cfg.Auth,
	}

	authen := middleware.Authenticate(cfg.Auth)
	ruleAdmin := middleware.Authorize(cfg.Auth, auth.RuleAdminOnly)
	ruleAdminOrSubject := middleware.Authorize(cfg.Auth, auth.RuleAdminOrSubject)

	// =============================================================
	// user account endpoints
	app.Handle(http.MethodPost, version, "/api/register", usersHandler.Register)
	app.Handle(http.MethodPost, version, "/api/login", usersHandler.Login)
	app.Handle(http.MethodGet, version, "/api/user_info/:user_id", usersHandler.GetByID, authen, ruleAdminOrSubject)
	app.Handle(http.MethodGet, version, "/api/users/", usersHandler.List, authen, ruleAdmin)

	// =============================================================
	// posts endpoints
	app.Handle(http.MethodPost, version, "/api/posts", postsHandler.AddPost, authen)
	app.Handle(http.MethodGet, version, "/api/posts/", postsHandler.List)
	app.Handle(http.MethodGet, version, "/api/post/:post_id", postsHandler.GetByID)
	app.Handle(http.MethodGet, version, "/api/posts/:category_name", postsHandler.ListByCatName)
	app.Handle(http.MethodGet, version, "/api/user/:user_name", postsHandler.ListByUsername)
	app.Handle(http.MethodDelete, version, "/api/post/:post_id", postsHandler.DeleteByID, authen)

	app.Handle(http.MethodPost, version, "/api/post/:post_id/comment", postsHandler.AddComment, authen)
	app.Handle(http.MethodDelete, version, "/api/post/:post_id/:comment_id", postsHandler.DeleteComment, authen)

	app.Handle(http.MethodGet, version, "/api/post/:post_id/upvote", postsHandler.UpVote, authen)
	app.Handle(http.MethodGet, version, "/api/post/:post_id/downvote", postsHandler.DownVote, authen)
}
