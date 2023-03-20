package enviroment

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func GetEnvAsInt(key string) (int, error) {
	keyStr := os.Getenv(key)
	if keyStr == "" {
		return 0, fmt.Errorf("empty '%s'", key)
	}

	value, err := strconv.Atoi(keyStr)
	if err != nil {
		log.Fatalf("Failed to convert '%s' to integer", key)
	}

	return value, nil
}
