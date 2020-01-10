package io

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/gildub/phronetic/pkg/env"
)

// ReadFile reads a file from WorkDir and returns its contents
func ReadFile(file string) ([]byte, error) {
	src := filepath.Join(env.Config().GetString("WorkDir"), file)
	return ioutil.ReadFile(src)
}

// WriteFile writes data to a file into WorkDir
func WriteFile(content []byte, file string) error {
	dst := filepath.Join(env.Config().GetString("WorkDir"), file)
	os.MkdirAll(path.Dir(dst), 0750)
	return ioutil.WriteFile(dst, content, 0640)
}

