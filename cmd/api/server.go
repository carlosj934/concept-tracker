package main

import (
	"context"
	"fmt"
	"log"

	"concept-tracker/config"
	"concept-tracker/internal/handler"
	"concept-tracker/internal/middleware"
	"concept-tracker/internal/repository"
	"concept-tracker/internal/service"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
	db     *pgxpool.Pool
}

func New(c *config.Config) (*Server, error) {
	clerk.SetKey(c.ClerkSecretKey)

	p, err := pgxpool.New(context.Background(), c.DSN())
	if err != nil {
		return nil, err
	}

	if err := p.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("database unreachable: %w", err)
	}
	log.Printf("database connected successfully")

	s := &Server{
		cfg:    c,
		router: gin.Default(),
		db:     p,
	}

	v1 := s.router.Group("/api/v1", middleware.ClerkAuth())

	// concept repository / svc / handler
	cr := repository.New(p)
	csvc := service.NewConceptService(cr)
	ch := handler.NewConceptHandler(csvc)

	// resource repository / svc / handler
	rr := repository.NewResource(p)
	rsvc := service.NewResourceService(rr)
	rh := handler.NewResourceHandler(rsvc)

	// activity log repository / svc / handler
	ar := repository.NewActivityLog(p)
	asvc := service.NewActivityLogService(ar)
	ah := handler.NewActivityLogHandler(asvc)

	// register routes
	handler.RegisterHealthRoutes(s.router)
	handler.RegisterConceptRoutes(v1, ch)
	handler.RegisterResourceRoutes(v1, rh)
	handler.RegisterActivityLogRoutes(v1, ah)
	handler.RegisterMeRoutes(v1)

	return s, nil
}

func (s *Server) Start() {
	s.router.Run(fmt.Sprintf(":%d", s.cfg.ServerPort))
}
