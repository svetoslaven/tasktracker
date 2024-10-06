package models

type MemberRole int

const (
	MemberRoleRegular MemberRole = 1
	MemberRoleLeader  MemberRole = 2
	MemberRoleAdmin   MemberRole = 3
	MemberRoleOwner   MemberRole = 4
)
