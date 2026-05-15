package routes

import (
	"crossview-go-server/api/controllers/isp"
	"crossview-go-server/api/middlewares"
	"crossview-go-server/lib"
	"crossview-go-server/models"
)

type ISPRoutes struct {
	logger     lib.Logger
	handler    lib.RequestHandler
	controller isp.ISPController
	userRepo   *models.UserRepository
}

func NewISPRoutes(
	logger lib.Logger,
	handler lib.RequestHandler,
	controller isp.ISPController,
	db lib.Database,
) ISPRoutes {
	return ISPRoutes{
		logger:     logger,
		handler:    handler,
		controller: controller,
		userRepo:   models.NewUserRepository(db.DB),
	}
}

func (r ISPRoutes) Setup() {
	r.logger.Info("Setting up ISP fleet routes")
	api := r.handler.Gin.Group("/api")
	admin := middlewares.RequireAdmin(r.userRepo)
	{
		api.GET("/isp/environments/:env/sites", admin, r.controller.ListSites)
		api.GET("/isp/sites/:site/nodes",       admin, r.controller.ListNodes)
		api.GET("/isp/sites/:site/pen",         admin, r.controller.ListPen)
		api.POST("/isp/nodes/muster",           admin, r.controller.Muster)
		api.POST("/isp/herds",                  admin, r.controller.CreateHerd)
		api.DELETE("/isp/herds/:herd",          admin, r.controller.RecallHerd)
		api.GET("/isp/fleet/status",            admin, r.controller.FleetStatus)
	}
}
