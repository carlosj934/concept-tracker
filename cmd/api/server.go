package main

import (
	"context"
	"fmt"
	
	"concept-tracker/config"
	"concept-tracker/internal/handler"
	"concept-tracker/internal/service"
	"concept-tracker/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg *config.Config
	router *gin.Engine
	db *pgxpool.Pool
}

func New(c *config.Config) (*Server, error) {
	p, err := pgxpool.New(context.Background(), c.DSN())
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg: c,
		router: gin.Default(),
		db: p,
	}

	r := repository.New(p)
	svc := service.NewConceptService(r)
	h := handler.NewConceptHandler(svc)

	handler.RegisterHealthRoutes(s.router)
	handler.RegisterConceptRoutes(s.router, h)

	return s, nil
}

func (s *Server) Start() {
	s.router.Run(fmt.Sprintf(":%d", s.cfg.ServerPort))
}
