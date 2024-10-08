package models

import "time"

type TeamFilters struct {
	Name     string
	IsPublic *bool
}

type InvitationFilters struct {
	TeamName  string
	IsInviter *bool
}

type MembershipFilters struct {
	MemberUsername string
	MemberRoles    []MemberRole
}

type TaskFilters struct {
	CreatedBefore    *time.Time
	CreatedAfter     *time.Time
	DueBefore        *time.Time
	DueAfter         *time.Time
	Status           []TaskStatus
	Priority         []TaskPriority
	CreatorUsername  string
	AssigneeUsername string
}
