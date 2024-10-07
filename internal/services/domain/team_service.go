package domain

import (
	"context"
	"errors"
	"fmt"

	"github.com/svetoslaven/tasktracker/internal/models"
	"github.com/svetoslaven/tasktracker/internal/pagination"
	"github.com/svetoslaven/tasktracker/internal/repositories"
	"github.com/svetoslaven/tasktracker/internal/services"
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

func (s *TeamService) GetTeamByName(ctx context.Context, name string, retrieverID int64) (*models.Team, error) {
	return s.getTeamByName(ctx, name, retrieverID)
}

func (s *TeamService) GetAllTeams(
	ctx context.Context,
	filters models.TeamFilters,
	paginationOpts pagination.Options,
	retrieverID int64,
) ([]*models.Team, pagination.Metadata, error) {
	return s.TeamRepo.GetAllTeams(ctx, filters, paginationOpts, retrieverID)
}

func (s *TeamService) UpdateTeam(
	ctx context.Context,
	newName *string,
	newIsPublic *bool,
	team *models.Team,
	updaterID int64,
) (*validator.Validator, error) {
	canUpdateTeam, err := s.isMemberInRole(ctx, team.ID, updaterID, models.MemberRoleOwner)
	if err != nil {
		return nil, err
	}

	if !canUpdateTeam {
		return nil, services.ErrNoPermission
	}

	var isChanged bool

	if newName != nil {
		validator := validator.New()

		s.validateTeamName(*newName, validator)

		if validator.HasErrors() {
			return validator, nil
		}

		if team.Name != *newName {
			team.Name = *newName
			isChanged = true
		}
	}

	if newIsPublic != nil {
		if team.IsPublic != *newIsPublic {
			team.IsPublic = *newIsPublic
			isChanged = true
		}
	}

	if !isChanged {
		return nil, nil
	}

	if err := s.TeamRepo.UpdateTeam(ctx, team); err != nil {
		return nil, handleRepositoryUpdateError(err)
	}

	return nil, nil
}

func (s *TeamService) DeleteTeam(ctx context.Context, teamID, removerID int64) error {
	canDeleteTeam, err := s.isMemberInRole(ctx, teamID, removerID, models.MemberRoleOwner)
	if err != nil {
		return err
	}

	if !canDeleteTeam {
		return services.ErrNoPermission
	}

	if err := s.TeamRepo.DeleteTeam(ctx, teamID); err != nil {
		switch {
		case errors.Is(err, repositories.ErrNoRecordsFound):
			return services.ErrNoRecordsFound
		default:
			return err
		}
	}

	return nil
}

func (s *TeamService) IsMember(ctx context.Context, teamID, userID int64) (bool, error) {
	return s.isMemberInRole(ctx, teamID, userID, models.MemberRoleRegular)
}

func (s *TeamService) InviteUser(ctx context.Context, teamID, inviterID, inviteeID int64) error {
	canInviteUser, err := s.isMemberInRole(ctx, teamID, inviterID, models.MemberRoleAdmin)
	if err != nil {
		return err
	}

	if !canInviteUser {
		return services.ErrNoPermission
	}

	if err := s.TeamRepo.InsertInvitation(ctx, teamID, inviterID, inviteeID); err != nil {
		switch {
		case errors.Is(err, repositories.ErrInvitationExists):
			return services.ErrInvitationExists
		default:
			return err
		}
	}

	return nil
}

func (s *TeamService) GetAllInvitations(
	ctx context.Context,
	filters models.InvitationFilters,
	paginationOpts pagination.Options,
	retrieverID int64,
) ([]*models.Invitation, pagination.Metadata, error) {
	return s.TeamRepo.GetAllInvitations(ctx, filters, paginationOpts, retrieverID)
}

func (s *TeamService) AcceptInvitation(ctx context.Context, invitationID, inviteeID int64) error {
	if err := s.TeamRepo.AcceptInvitation(ctx, invitationID, inviteeID); err != nil {
		switch {
		case errors.Is(err, repositories.ErrNoRecordsFound):
			return services.ErrNoRecordsFound
		default:
			return err
		}
	}

	return nil
}

func (s *TeamService) RejectInvitation(ctx context.Context, invitationID, inviteeID int64) error {
	if err := s.TeamRepo.RejectInvitation(ctx, invitationID, inviteeID); err != nil {
		switch {
		case errors.Is(err, repositories.ErrNoRecordsFound):
			return services.ErrNoRecordsFound
		default:
			return err
		}
	}

	return nil
}

func (s *TeamService) DeleteInvitation(ctx context.Context, invitationID, removerID int64) error {
	if err := s.TeamRepo.DeleteInvitation(ctx, invitationID, removerID); err != nil {
		switch {
		case errors.Is(err, repositories.ErrNoRecordsFound):
			return services.ErrNoRecordsFound
		default:
			return err
		}
	}

	return nil
}

func (s *TeamService) GetAllTeamMembers(
	ctx context.Context,
	filters models.MembershipFilters,
	roles []string,
	paginationOpts pagination.Options,
	teamID int64,
) ([]*models.Membership, pagination.Metadata, *validator.Validator, error) {
	validator := validator.New()

	for _, role := range roles {
		memberRole, err := models.NewMemberRole(role)
		if err != nil {
			validator.AddError("roles", fmt.Sprintf("Contains an invalid member role %q.", role))
			break
		}

		filters.MemberRoles = append(filters.MemberRoles, memberRole)
	}

	if validator.HasErrors() {
		return nil, pagination.Metadata{}, validator, nil
	}

	memberships, metadata, err := s.TeamRepo.GetAllTeamMembers(ctx, filters, paginationOpts, teamID)
	return memberships, metadata, nil, err
}

func (s *TeamService) UpdateMemberRole(
	ctx context.Context,
	teamID, memberID int64,
	newRole string,
	updaterID int64,
) (*validator.Validator, error) {
	validator := validator.New()

	memberRole, err := models.NewMemberRole(newRole)
	if err != nil {
		validator.AddError("new_role", "Must be a valid member role.")
		return validator, nil
	}

	canUpdateMemberRole, err := s.isMemberInRole(ctx, teamID, updaterID, models.MemberRoleOwner)
	if err != nil {
		return nil, err
	}

	if !canUpdateMemberRole {
		return nil, services.ErrNoPermission
	}

	if updaterID == memberID {
		return nil, services.ErrCannotChangeOwnerRole
	}

	membership, err := s.TeamRepo.GetMembership(ctx, teamID, memberID)
	if err != nil {
		return nil, err
	}

	if membership.MemberRole == memberRole {
		return nil, nil
	}

	membership.MemberRole = memberRole

	if err := s.TeamRepo.UpdateMembership(ctx, membership); err != nil {
		return nil, handleRepositoryUpdateError(err)
	}

	return nil, nil
}

func (s *TeamService) RemoveMemberFromTeam(ctx context.Context, teamID, memberID, removerID int64) error {
	isRemovedMemberOwner, err := s.isMemberInRole(ctx, teamID, memberID, models.MemberRoleOwner)
	if err != nil {
		return err
	}

	if isRemovedMemberOwner {
		return services.ErrCannotRemoveTeamOwner
	}

	canRemoveMember, err := s.isMemberInRole(ctx, teamID, removerID, models.MemberRoleAdmin)
	if err != nil {
		return err
	}

	if canRemoveMember || removerID == memberID {
		err = s.TeamRepo.DeleteMembership(ctx, teamID, memberID)
		return err
	}

	return services.ErrNoPermission
}

func (s *TeamService) isMemberInRole(
	ctx context.Context,
	teamID, memberID int64,
	role models.MemberRole,
) (bool, error) {
	memberRole, err := s.TeamRepo.GetMemberRole(ctx, teamID, memberID)
	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrNoRecordsFound):
			return false, nil
		default:
			return false, err
		}
	}

	return memberRole >= role, nil
}

func (s *TeamService) getTeamByName(ctx context.Context, name string, retrieverID int64) (*models.Team, error) {
	team, err := s.TeamRepo.GetTeamByName(ctx, name, retrieverID)
	if err != nil {
		return nil, handleRepositoryRetrievalError(err)
	}

	return team, nil
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
