package security

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	saltSizeBits  = 64
	saltSizeBytes = int(saltSizeBits / 8)
)

var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	initDone    = false
)

func Initialize() {
	if initDone {
		panic("Initialization already done.")
	}
	initDone = true

	rand.Seed(time.Now().Unix())
}

func GenerateRandomPassword() string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GenerateRandomSalt() string {
	buf := new(bytes.Buffer)
	for i := 0; i < (saltSizeBytes / 4); i++ {
		err := binary.Write(buf, binary.LittleEndian, rand.Uint32())
		if err != nil {
			panic("binary.Write failed: " + err.Error())
		}
	}
	salt := base64.URLEncoding.EncodeToString(buf.Bytes())
	return salt
}

func EncryptPassword(plaintextPassword, salt string) string {
	k := pbkdf2.Key([]byte(plaintextPassword), []byte(salt), 10000, 32, sha256.New)
	return hex.EncodeToString(k)
}
