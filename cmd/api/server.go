package main

import (
	"context"
	"fmt"
	
	"concept-tracker/config"
	"concept-tracker/internal/handler"

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

	handler.RegisterHealthRoutes(s.router)

	return s, nil
}

func (s *Server) Start() {
	s.router.Run(fmt.Sprintf(":%d", s.cfg.ServerPort))
}
