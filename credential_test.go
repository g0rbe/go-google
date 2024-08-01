package google_test

import (
	"fmt"
	"testing"

	"github.com/g0rbe/go-google"
)

func TestRandomApiKeys(t *testing.T) {

	var res []string

	c := google.RandomApiKeys("one", "two", "three", "four", "five")

	for i := 0; i < 50; i++ {

		k, err := c.Token()
		if err != nil {
			t.Fatalf("Token error: %s\n", err)
		}

		res = append(res, k)
	}

	t.Logf("%v\n", res)
}

func ExampleRandomApiKeys() {

	c := google.RandomApiKeys("one", "two", "three", "four", "five")

	for i := 0; i < 5; i++ {

		k, err := c.Token()
		if err != nil {
			// Handle error
		}

		fmt.Printf("%s\n", k)
	}
	// Output:
	// one
	// four
	// two
	// one
	// three
}
func TestRotatingApiKeys(t *testing.T) {

	var res []string

	c := google.RotatingApiKeys("one", "two", "three", "four", "five")

	for i := 0; i < 50; i++ {

		k, err := c.Token()
		if err != nil {
			t.Fatalf("Token error: %s\n", err)
		}

		res = append(res, k)
	}

	t.Logf("%v\n", res)
}

func ExampleRotatingApiKeys() {

	c := google.RotatingApiKeys("one", "two", "three", "four", "five")

	for i := 0; i < 5; i++ {

		k, err := c.Token()
		if err != nil {
			// Handle error
		}

		fmt.Printf("%s\n", k)
	}
	// Output:
	// one
	// two
	// three
	// four
	// five
}
