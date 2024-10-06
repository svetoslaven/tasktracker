package time

import (
	originaltime "time"
)

type Facade interface {
	Now() originaltime.Time
}

type facade struct{}

var instance Facade

func Instance() Facade {
	if instance == nil {
		instance = &facade{}
	}

	return instance
}

func SetInstance(f Facade) {
	instance = f
}

func (f *facade) Now() originaltime.Time {
	return originaltime.Now()
}
