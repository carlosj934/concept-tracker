package main

import (
	"context"
	"fmt"
	
	"concept-tracker/config"

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

	return s, nil
}

func (s *Server) Start() {
	s.router.Run(fmt.Sprintf(":%d", s.cfg.ServerPort))
}
