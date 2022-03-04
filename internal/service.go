package service

import "NewOne/internal/usecases"

type Implementation struct {
	repo usecases.Repository
}

func New(repo usecases.Repository) Implementation {
	return Implementation{repo: repo}
}
