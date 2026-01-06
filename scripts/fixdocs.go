// Исправляет docs.go, удаляя LeftDelim и RightDelim из шаблонов.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	data, err := os.ReadFile("docs/docs.go")
	if err != nil {
		log.Fatal(err)
	}

	content := string(data)
	content = strings.ReplaceAll(content, "\tLeftDelim:        \"{{\",\n", "")
	content = strings.ReplaceAll(content, "\tRightDelim:       \"}}\",\n", "")

	err = os.WriteFile(".\\docs\\docs.go", []byte(content), 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("docs.go fixed successfully")
}
