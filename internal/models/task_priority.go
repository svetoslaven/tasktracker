package models

import (
	"errors"
	"strconv"
	"strings"
)

type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = 1
	TaskPriorityMedium TaskPriority = 2
	TaskPriorityHigh   TaskPriority = 3
)

func NewTaskPriority(priority string) (TaskPriority, error) {
	switch strings.ToLower(priority) {
	case TaskPriorityLow.String():
		return TaskPriorityLow, nil
	case TaskPriorityMedium.String():
		return TaskPriorityMedium, nil
	case TaskPriorityHigh.String():
		return TaskPriorityHigh, nil
	default:
		return TaskPriorityLow, errors.New("models: invalid task status")
	}
}

func (p TaskPriority) String() string {
	switch p {
	case TaskPriorityLow:
		return "low"
	case TaskPriorityMedium:
		return "medium"
	case TaskPriorityHigh:
		return "high"
	default:
		panic("invalid task priority")
	}
}

func (p TaskPriority) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(p.String())), nil
}
