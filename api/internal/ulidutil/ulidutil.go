package ulidutil

import "github.com/oklog/ulid/v2"

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
