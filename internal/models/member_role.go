package models

import (
	"errors"
	"strconv"
	"strings"
)

type MemberRole int

const (
	MemberRoleRegular MemberRole = 1
	MemberRoleLeader  MemberRole = 2
	MemberRoleAdmin   MemberRole = 3
	MemberRoleOwner   MemberRole = 4
)

func (r MemberRole) String() string {
	switch r {
	case MemberRoleRegular:
		return "regular"
	case MemberRoleLeader:
		return "leader"
	case MemberRoleAdmin:
		return "admin"
	case MemberRoleOwner:
		return "owner"
	default:
		panic("invalid member role")
	}
}

func (r MemberRole) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(r.String())), nil
}

func NewMemberRole(role string) (MemberRole, error) {
	switch strings.ToLower(role) {
	case MemberRoleRegular.String():
		return MemberRoleRegular, nil
	case MemberRoleLeader.String():
		return MemberRoleLeader, nil
	case MemberRoleAdmin.String():
		return MemberRoleAdmin, nil
	case MemberRoleOwner.String():
		return MemberRoleOwner, nil
	default:
		return MemberRoleRegular, errors.New("models: invalid member role")
	}
}
