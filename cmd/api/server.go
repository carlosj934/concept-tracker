package main

import (
	"context"
	"fmt"
	
	"concept-tracker/config"
	"concept-tracker/internal/handler"
	"concept-tracker/internal/service"
	"concept-tracker/internal/repository"
	"concept-tracker/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/clerk/clerk-sdk-go/v2"
)

type Server struct {
	cfg *config.Config
	router *gin.Engine
	db *pgxpool.Pool
}

func New(c *config.Config) (*Server, error) {
	clerk.SetKey(c.ClerkSecretKey)

	p, err := pgxpool.New(context.Background(), c.DSN())
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg: c,
		router: gin.Default(),
		db: p,
	}

	v1 := s.router.Group("/api/v1", middleware.ClerkAuth())

	r := repository.New(p)
	svc := service.NewConceptService(r)
	h := handler.NewConceptHandler(svc)

	handler.RegisterHealthRoutes(s.router)
	handler.RegisterConceptRoutes(v1, h)
	handler.RegisterMeRoutes(v1)

	return s, nil
}

func (s *Server) Start() {
	s.router.Run(fmt.Sprintf(":%d", s.cfg.ServerPort))
}
