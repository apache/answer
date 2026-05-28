package fake_username

import (
	"fmt"
	"math/rand"
)

type FakeUsernameGenerator struct{}

func NewFakeUsernameGenerator() *FakeUsernameGenerator {
	return &FakeUsernameGenerator{}
}

func (fg *FakeUsernameGenerator) GenerateFakeName() string {
	firstParts := []string{
		"Cool",
		"Dark",
		"Fast",
		"Silent",
		"Crazy",
	}

	secondParts := []string{
		"Tiger",
		"Wolf",
		"Ninja",
		"Dragon",
		"Hawk",
	}

	number := rand.Intn(10000)

	first := firstParts[rand.Intn(len(firstParts))]
	second := secondParts[rand.Intn(len(secondParts))]

	return fmt.Sprintf("%s%s%d", first, second, number)
}
