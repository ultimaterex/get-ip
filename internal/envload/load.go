package envload

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// DotEnv loads `.env` from the current working directory when present.
// Variables already set in the process environment are not overridden (godotenv semantics).
// A missing `.env` file is ignored; other read errors are logged and do not stop the program.
func DotEnv() {
	if err := godotenv.Load(); err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Printf("warning: load .env: %v", err)
	}
}
