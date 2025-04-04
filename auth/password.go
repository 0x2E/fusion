package auth

import (
	"crypto/sha256"
	"crypto/subtle"

	"golang.org/x/crypto/pbkdf2"
)

type HashedPassword struct {
	hash []byte
}

func (hp HashedPassword) Bytes() []byte {
	return hp.hash
}

func (hp HashedPassword) Equals(other HashedPassword) bool {
	return subtle.ConstantTimeCompare(hp.hash, other.hash) != 0
}

func HashPassword(password string) (HashedPassword, error) {
	if len(password) == 0 {
		return HashedPassword{}, nil
	}

	// These bytes are chosen at random. It's insecure to use a static salt to
	// hash a set of passwords, but since we're only ever hashing a single
	// password, using a static salt is fine. The salt prevents an attacker from
	// using a rainbow table to retrieve the plaintext password from the hashed
	// version, and that's all that's necessary for fusion's needs.
	staticSalt := []byte{36, 129, 1, 54}
	iter := 100
	keyLen := 32
	hash := pbkdf2.Key([]byte(password), staticSalt, iter, keyLen, sha256.New)

	return HashedPassword{
		hash: hash,
	}, nil
}
