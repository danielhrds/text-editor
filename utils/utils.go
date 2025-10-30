package utils

import (
	"bufio"
	"log"
	"os"
	"strconv"
)

var Logger = log.New(os.Stdout, "\033[33m[DEBUG]\033[0m ", 0)

func ReadFile(path string) ([]byte, int) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	bytes := make([]byte, 0)
	i := 0
	for scanner.Scan() {
		bytes = append(bytes, scanner.Bytes()...)
		bytes = append(bytes, '\n')
		i++
	}
	return bytes, i
}

func IntToString(num int) string {
	return strconv.Itoa(num)
}
