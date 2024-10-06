package domain

import (
	"context"
	"errors"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/validator"
)

const teamNameField = "name"

type TeamService struct {
	TeamRepo repositories.TeamRepository
}

func (s *TeamService) CreateTeam(
	ctx context.Context,
	name string,
	isPublic bool,
	creatorID int64,
) (*models.Team, *validator.Validator, error) {
	validator := validator.New()

	s.validateTeamName(name, validator)

	if validator.HasErrors() {
		return nil, validator, nil
	}

	team := &models.Team{
		Name:     name,
		IsPublic: isPublic,
	}

	if err := s.TeamRepo.InsertTeam(ctx, team, creatorID); err != nil {
		switch {
		case errors.Is(err, repositories.ErrDuplicateTeamName):
			s.addTeamNameTakenError(validator)
			return nil, validator, nil
		default:
			return nil, nil, err
		}
	}

	return team, nil, nil
}

func (s *TeamService) validateTeamName(name string, validator *validator.Validator) {
	validator.CheckNonZero(name, teamNameField)
	validator.CheckStringMaxLength(name, 32, teamNameField)
	validator.Check(
		s.isValidTeamName(name),
		teamNameField,
		"Must contain only alphanumeric characters or single hyphens, and must not begin or end with a hyphen.",
	)
}

func (s *TeamService) isValidTeamName(name string) bool {
	if len(name) == 0 {
		return false
	}

	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}

	for i := 0; i < len(name); i++ {
		c := name[i]

		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		}

		if c != '-' {
			return false
		}

		if i > 0 && name[i-1] == '-' {
			return false
		}
	}

	return true
}

func (s *TeamService) addTeamNameTakenError(validator *validator.Validator) {
	validator.AddError(teamNameField, "A team with this name already exists.")
}
