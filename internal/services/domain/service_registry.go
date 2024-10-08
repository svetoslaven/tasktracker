package domain

import (
	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/services"
)

func NewServiceRegistry(repos repositories.RepositoryRegistry) services.ServiceRegistry {
	return services.ServiceRegistry{
		UserService:  &UserService{UserRepo: repos.UserRepo},
		TokenService: &TokenService{TokenRepo: repos.TokenRepo},
		TeamService:  &TeamService{TeamRepo: repos.TeamRepo},
		TaskService: &TaskService{
			TaskRepo: repos.TaskRepo,
			TeamRepo: repos.TeamRepo,
		},
	}
}
