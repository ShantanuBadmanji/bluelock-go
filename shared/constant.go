package shared

import (
	"log"
	"os"
)

var (
	err     error
	RootDir string
)

func init() {
	RootDir, err = os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	} else {
		log.Printf("root path: %s", RootDir)
	}
}
