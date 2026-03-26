package postgres

import (
	"github.com/rockkley/pushpost/services/common_service/database"
)

type ProfileRepository struct {
	exec database.Executor
}

func NewProfileRepository(exec database.Executor) *ProfileRepository {
	return &ProfileRepository{exec: exec}
}
