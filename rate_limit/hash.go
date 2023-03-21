package rate_limit

import (
	"crypto/md5"
	"github.com/google/uuid"
)

func newMd5UUID(namespace, name string) (uuid.UUID, error) {
	hash := md5.New()
	_, err := hash.Write([]byte(namespace))
	if err != nil {
		return uuid.Nil, err
	}
	_, err = hash.Write([]byte(name))
	if err != nil {
		return uuid.Nil, err
	}

	sum := hash.Sum(nil)
	return uuid.FromBytes(sum)
}
