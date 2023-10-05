package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateUniqueDirName() string {
    // Generate a random string of 8 characters
    rand.Seed(time.Now().UnixNano())
    letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    b := make([]rune, 8)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    randomString := string(b)

    // Get the current timestamp
    timestamp := time.Now().Unix()

    // Combine the timestamp and random string to create the directory name
    dirName := fmt.Sprintf("%d_%s", timestamp, randomString)

    return dirName
}