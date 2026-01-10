/*
Package api exposes the main API engine. All HTTP APIs are handled here - so-called "business logic" should be here, or
in a dedicated package (if that logic is complex enough).

To use this package, you should create a new instance with New() passing a valid Config. The resulting Router will have
the Router.Handler() function that returns a handler that can be used in a http.Server (or in other middlewares).

Example:

	// Create the API router
	apirouter, err := api.New(api.Config{
		Logger:   logger,
		Database: appdb,
	})
	if err != nil {
		logger.WithError(err).Error("error creating the API server instance")
		return fmt.Errorf("error creating the API server instance: %w", err)
	}
	router := apirouter.Handler()

	// ... other stuff here, like middleware chaining, etc.

	// Create the API server
	apiserver := http.Server{
		Addr:              cfg.Web.APIHost,
		Handler:           router,
		ReadTimeout:       cfg.Web.ReadTimeout,
		ReadHeaderTimeout: cfg.Web.ReadTimeout,
		WriteTimeout:      cfg.Web.WriteTimeout,
	}

	// Start the service listening for requests in a separate goroutine
	apiserver.ListenAndServe()

See the `main.go` file inside the `cmd/webapi` for a full usage example.
*/
package api

import (
	"errors"
	"net/http"

	"git.phoebe2z/WASAText/service/database"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Config is used to provide dependencies and configuration to the New function.
type Config struct {
	// Logger where log entries are sent
	Logger logrus.FieldLogger

	// Database is the instance of database.AppDatabase where data are saved
	Database database.AppDatabase
}

// Router is the package API interface representing an API handler builder
type Router interface {
	// Handler returns an HTTP handler for APIs provided in this package
	Handler() http.Handler

	// Close terminates any resource used in the package
	Close() error
}

// New returns a new Router instance
func New(cfg Config) (Router, error) {
	// Check if the configuration is correct
	if cfg.Logger == nil {
		return nil, errors.New("logger is required")
	}
	if cfg.Database == nil {
		return nil, errors.New("database is required")
	}

	// Create a new router where we will register HTTP endpoints. The server will pass requests to this router to be
	// handled.
	router := httprouter.New()
	router.RedirectTrailingSlash = false
	router.RedirectFixedPath = false

	r := &_router{
		router:     router,
		baseLogger: cfg.Logger,
		db:         cfg.Database,
	}

	// Register Routes
	router.POST("/session", r.doLogin)
	router.PUT("/user/name", r.setMyUserName)
	router.PUT("/user/photo", r.setMyPhoto)
	router.GET("/user/me", r.getMyProfile)
	router.GET("/users", r.listUsers)

	router.POST("/conversations", r.createConversation)
	router.GET("/conversations", r.getMyConversations)
	router.GET("/conversations/:conversationId", r.getConversation)

	router.POST("/messages", r.sendMessage)
	router.DELETE("/messages/:messageId", r.deleteMessage)
	router.POST("/messages/:messageId/forward", r.forwardMessage)
	router.POST("/messages/:messageId/reaction", r.commentMessage)
	router.DELETE("/messages/:messageId/reaction", r.uncommentMessage)

	router.POST("/groups", r.createGroup)
	router.POST("/groups/:groupId/members", r.addToGroup)
	router.DELETE("/groups/:groupId/me", r.leaveGroup)
	router.PUT("/groups/:groupId/name", r.setGroupName)
	router.PUT("/groups/:groupId/photo", r.setGroupPhoto)

	// Serve static files
	router.ServeFiles("/static/*filepath", http.Dir("./static"))

	return r, nil
}

type _router struct {
	router *httprouter.Router

	// baseLogger is a logger for non-requests contexts, like goroutines or background tasks not started by a request.
	// Use context logger if available (e.g., in requests) instead of this logger.
	baseLogger logrus.FieldLogger

	db database.AppDatabase
}
