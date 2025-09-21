package utils

import (
	"log"
	"os"
)

var Logger = log.New(os.Stdout, "\033[33m[DEBUG]\033[0m ", 0)
