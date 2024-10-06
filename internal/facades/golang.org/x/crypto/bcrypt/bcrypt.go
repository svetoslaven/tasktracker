package bcrypt

import originalbcrypt "golang.org/x/crypto/bcrypt"

type Facade interface {
	CompareHashAndPassword(hashedPassword, password []byte) error
	GenerateFromPassword(password []byte, cost int) ([]byte, error)
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

func (f *facade) CompareHashAndPassword(hashedPassword, password []byte) error {
	return originalbcrypt.CompareHashAndPassword(hashedPassword, password)
}

func (f *facade) GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	return originalbcrypt.GenerateFromPassword(password, cost)
}
