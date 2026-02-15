package ulidutil

import (
	"auth/internal/apperror"
	"strings"

	"github.com/oklog/ulid/v2"
)

func FromBytes(b []byte) (ulid.ULID, error) {
	var u ulid.ULID
	if len(b) != 16 {
		return u, ulid.ErrDataSize
	}
	copy(u[:], b)
	return u, nil
}

func MustFromBytes(b []byte) ulid.ULID {
	u, err := FromBytes(b)
	if err != nil {
		panic(err)
	}
	return u
}

func ToPrefixed(prefix string, id ulid.ULID) string {
	return prefix + "_" + id.String()
}

func FromPrefixed(prefix string, prefixedID string) (ulid.ULID, error) {
	withoutPrefix := strings.TrimPrefix(prefixedID, prefix+"_")
	if withoutPrefix == prefixedID {
		return ulid.Zero, apperror.NewBadRequest("Invalid ID")
	}
	id, err := ulid.Parse(withoutPrefix)
	if err != nil {
		return ulid.Zero, apperror.NewBadRequest("Invalid ID")
	}

	return id, nil
}
