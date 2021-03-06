package image

import (
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/server/httputils"
	"github.com/docker/docker/api/server/router"
)

// imageRouter is a router to talk with the image controller
type imageRouter struct {
	backend Backend
	decoder httputils.ContainerDecoder
	routes  []router.Route
}

// NewRouter initializes a new image router
func NewRouter(backend Backend, decoder httputils.ContainerDecoder) router.Router {
	logrus.Debugf("Executing api/server/router/image/image.go NewRouter(%s,%s)", backend, decoder)
	r := &imageRouter{
		backend: backend,
		decoder: decoder,
	}
	r.initRoutes()
	return r
}

// Routes returns the available routes to the image controller
func (r *imageRouter) Routes() []router.Route {
	return r.routes
}

// initRoutes initializes the routes in the image router
func (r *imageRouter) initRoutes() {
	logrus.Debug("Executing api/server/router/image/image.go initRoutes()")
	r.routes = []router.Route{
		// GET
		router.NewGetRoute("/images/json", r.getImagesJSON),
		router.NewGetRoute("/images/search", r.getImagesSearch),
		router.NewGetRoute("/images/get", r.getImagesGet),
		router.NewGetRoute("/images/{name:.*}/get", r.getImagesGet),
		router.NewGetRoute("/images/{name:.*}/history", r.getImagesHistory),
		router.NewGetRoute("/images/{name:.*}/json", r.getImagesByName),
		// POST
		router.NewPostRoute("/commit", r.postCommit),
		router.NewPostRoute("/images/load", r.postImagesLoad),
		router.Cancellable(router.NewPostRoute("/images/create", r.postImagesCreate)),
		router.Cancellable(router.NewPostRoute("/images/{name:.*}/push", r.postImagesPush)),
		router.NewPostRoute("/images/{name:.*}/tag", r.postImagesTag),
		// DELETE
		router.NewDeleteRoute("/images/{name:.*}", r.deleteImages),
	}
}
