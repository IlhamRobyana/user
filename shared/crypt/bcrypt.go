package crypt

import "golang.org/x/crypto/bcrypt"

func HashByBcrypt(plaintext string) (hashed string, err error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.MinCost)
	return string(hashedBytes), err
}

func CompareBcrypt(plaintext, hash string) (match bool, err error) {
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	if err == nil {
		match = true
	}
	return
}
