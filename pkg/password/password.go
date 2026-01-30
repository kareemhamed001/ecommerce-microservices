package password

import (
	"github.com/kareemhamed001/e-commerce/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

func Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		logger.Errorf("Error While Hashing password: %s", err.Error())
		return "", err
	}

	return string(hashedPassword), nil

}

func Verify(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
