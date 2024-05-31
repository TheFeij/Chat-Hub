package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// different groups of alphabets
const (
	LOWERCASE    = "abcdefghijklmnopqrstuvwxyz"
	UPPERCASE    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NUMBERS      = "0123456789"
	SPECIALS     = "_!@$%&*^"
	ALPHANUMERIC = LOWERCASE + UPPERCASE + NUMBERS
	ALPHABETS    = LOWERCASE + UPPERCASE
	ALL          = ALPHANUMERIC + SPECIALS
)

// random used to generate random numbers
var random *rand.Rand

// init initializes random
func init() {
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// RandomInt generates a random int64 integer
func RandomInt(min int64, max int64) int64 {
	return random.Int63n(max-min+1) + min
}

// RandomString generates a random string from the given alphabet and the given length
func RandomString(length int, alphabet string) string {
	alphabetLength := len(alphabet)
	var randomString strings.Builder

	for i := 0; i < length; i++ {
		randomByte := alphabet[random.Intn(alphabetLength)]
		randomString.WriteByte(randomByte)
	}

	return randomString.String()
}

// RandomText generates a random text
func RandomText() string {
	return RandomString(int(RandomInt(1, 1024)), ALL)
}

// RandomUsername generates a random username with length of 4-64
func RandomUsername() string {
	return RandomString(4, ALPHABETS)
}

// RandomPassword generates a random password with the length of 8-64
func RandomPassword() string {
	return RandomString(int(RandomInt(8, 60)), ALL)
}

// RandomIPv4 generates a random IPv4
func RandomIPv4() string {
	return fmt.Sprintf("%d.%d.%d.%d\n",
		RandomInt(0, 255),
		RandomInt(0, 255),
		RandomInt(0, 255),
		RandomInt(0, 255),
	)
}
