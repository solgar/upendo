package security

import (
	"encoding/base64"
	"testing"
)

func assert(trueStatement bool, msg string) {
	if !trueStatement {
		//_t.Error(msg)
		panic(msg)
	}
}

var (
	_t *testing.T = nil
)

func TestGenerateRandomPassword(t *testing.T) {
	pass := GenerateRandomPassword()
	assert(len(pass) == 8, "Password length should be 8.")
}

func TestGenerateRandomSalt(t *testing.T) {
	_t = t
	salt := GenerateRandomSalt()
	bytes, err := base64.URLEncoding.DecodeString(salt)

	assert(err == nil, "Error should be nil.")
	assert(len(bytes) == 8, "Decoded string should produce byte slice of length 32.")
}

func TestEncryptPassword(t *testing.T) {
	_t = t
	salt := GenerateRandomSalt()
	plaintextPassword1 := "JustSome1pa$$w0rd"
	plaintextPassword2 := "JustSome2pa$$w0rd"
	shortPassword := "root"

	// two time the same password - should be identical
	pass1 := EncryptPassword(plaintextPassword1, salt)
	pass2 := EncryptPassword(plaintextPassword1, salt)
	assert(pass1 == pass2, "Passwords should be the same.")

	// two times same other password - should be identical
	pass1 = EncryptPassword(plaintextPassword2, salt)
	pass2 = EncryptPassword(plaintextPassword2, salt)
	assert(pass1 == pass2, "Passwords should be the same.")

	pass1 = EncryptPassword(shortPassword, salt)
	pass2 = EncryptPassword("root", "69XpoNSrjCVN-2ODxAUVfEsWtGInvc1RIWadYql5maQ=")
	assert(pass1 != pass2, "Passwords should be different.")
}
