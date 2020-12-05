package main

import (
	"log"
	"os"
	"strconv"
)

// MustGetenv returns the value of the environment variable with the given name,
// or exits with an error if the environment variable is unset or empty.
func MustGetenv(key string) string {
	retv := os.Getenv(key)
	if retv == "" {
		log.Fatalf("required environment variable '%s' is missing", key)
	}
	return retv
}

// GetenvBool returns the given environment variable as a bool, or returns the given default value if the env variable is not set.
func GetenvBool(name string, defaultVal bool) bool {
	valStr := os.Getenv(name)
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

// WriteFileExcl writes the given data to a file named by filename.
// The file is created with the given permissions (before umask).
// If the file already exists, an error is returned (for which os.IsExist
// will return true).
func WriteFileExcl(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
