// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var fs http.FileSystem = http.Dir("staticpage")

	err := vfsgen.Generate(fs, vfsgen.Options{
		PackageName: "reportoutput",
	})

	if err != nil {
		log.Fatalln(err)
	}
}
