package helpers

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateReference() string {
	now := time.Now()
	nowFormat := now.Format("20060102150405")
	randomNumber := rand.Intn(100)
	reference := fmt.Sprintf("%s%d", nowFormat, randomNumber)
	return reference
}
