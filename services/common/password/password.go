package password

import "golang.org/x/crypto/bcrypt"

func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func Compare(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
