package utils

import (
	"github.com/google/uuid"
)

type UUIDGenerator interface {
	UUIDGenerate() string
}

type UUIDGeneratorImpl struct{}

func (*UUIDGeneratorImpl) UUIDGenerate() string {
	return uuid.New().String()
}
