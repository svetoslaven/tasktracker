package models

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
