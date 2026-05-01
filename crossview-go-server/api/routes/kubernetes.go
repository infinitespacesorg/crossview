package routes

import (
	"crossview-go-server/api/controllers/kubernetes"
	"crossview-go-server/api/middlewares"
	"crossview-go-server/lib"
	"crossview-go-server/models"
)

type KubernetesRoutes struct {
	logger          lib.Logger
	handler         lib.RequestHandler
	controller      kubernetes.KubernetesController
	watchController *kubernetes.WatchController
	userRepo        *models.UserRepository
}

func NewKubernetesRoutes(
	logger lib.Logger,
	handler lib.RequestHandler,
	controller kubernetes.KubernetesController,
	watchController *kubernetes.WatchController,
	db lib.Database,
) KubernetesRoutes {
	return KubernetesRoutes{
		logger:          logger,
		handler:         handler,
		controller:      controller,
		watchController: watchController,
		userRepo:        models.NewUserRepository(db.DB),
	}
}

func (r KubernetesRoutes) Setup() {
	r.logger.Info("Setting up Kubernetes routes")
	api := r.handler.Gin.Group("/api")
	admin := middlewares.RequireAdmin(r.userRepo)
	{
		api.GET("/kubernetes/status", r.controller.GetStatus)
		api.POST("/kubernetes/context", admin, r.controller.SetContext)
		api.PUT("/kubernetes/context", admin, r.controller.SetContext)
		api.GET("/kubernetes/context", admin, r.controller.GetCurrentContext)
		api.GET("/kubernetes/contexts", admin, r.controller.GetContexts)
		api.GET("/kubernetes/connection", admin, r.controller.CheckConnection)
		api.POST("/kubernetes/kubeconfig", admin, r.controller.AddKubeConfig)
		api.GET("/contexts", admin, r.controller.GetContexts)
		api.GET("/contexts/current", admin, r.controller.GetCurrentContext)
		api.POST("/contexts/current", admin, r.controller.SetContext)
		api.POST("/contexts/add", admin, r.controller.AddKubeConfig)
		api.DELETE("/contexts", admin, r.controller.RemoveContext)
		api.GET("/resources", admin, r.controller.GetResources)
		api.GET("/resource", admin, r.controller.GetResource)
		api.GET("/events", admin, r.controller.GetEvents)
		api.GET("/managed", admin, r.controller.GetManagedResources)
		api.GET("/watch", admin, r.watchController.WatchResources)
	}
}
