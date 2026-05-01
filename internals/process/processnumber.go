package process

import (
	"errors"
	"fmt"
)

func ProcessNumber(numbers []int) (err error) {
	// Handle any panic that occurs inside this function.
	defer func() {
		if r := recover(); r != nil {
			// Recover from panic, print the panic message, and set the returned error.
			fmt.Printf("Recovered from panic: %v\n", r)
			err = fmt.Errorf("%v", r)
		}
	}()

	// Rule a: nil input → return error
	if numbers == nil {
		return errors.New("No data provided")
	}

	// Rule b: empty list (but not nil) → panic
	if len(numbers) == 0 {
		panic("empty list provided")
	}

	// Rule c: valid input → print each number × 2
	for _, num := range numbers {
		fmt.Println(num * 2)
	}

	return nil
}
