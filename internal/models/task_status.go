package models

import (
	"errors"
	"strconv"
	"strings"
)

type TaskStatus int

const (
	TaskStatusOpen       TaskStatus = 1
	TaskStatusInProgress TaskStatus = 2
	TaskStatusCompleted  TaskStatus = 3
	TaskStatusCancelled  TaskStatus = 4
)

func NewTaskStatus(status string) (TaskStatus, error) {
	switch strings.ToLower(status) {
	case TaskStatusOpen.String():
		return TaskStatusOpen, nil
	case TaskStatusInProgress.String():
		return TaskStatusInProgress, nil
	case TaskStatusCompleted.String():
		return TaskStatusCompleted, nil
	case TaskStatusCancelled.String():
		return TaskStatusCancelled, nil
	default:
		return TaskStatusOpen, errors.New("models: invalid task status")
	}
}

func (s TaskStatus) String() string {
	switch s {
	case TaskStatusOpen:
		return "open"
	case TaskStatusInProgress:
		return "in-progress"
	case TaskStatusCompleted:
		return "completed"
	case TaskStatusCancelled:
		return "candelled"
	default:
		panic("invalid task status")
	}
}

func (s TaskStatus) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.String())), nil
}
