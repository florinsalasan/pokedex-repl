package main

import (
	"fmt"
	"testing"
)


func TestCleanInputEasy(t *testing.T) {

    cases := []struct {
        input    string
        expected []string
    } {
        {
            input: "  hello world  ",
            expected: []string{"hello", "world"},
        },
        {
            input: "Charmander Bulbasaur PIKACHU",
            expected: []string{"charmander", "bulbasaur", "pikachu"},
        },
        {
            input: "TEStingSpOnGeBoBCaSE",
            expected: []string{"testingspongebobcase"},
        }, 
    }

    for _, c := range cases {
        actual := cleanInput(c.input)
        if len(actual) != len(c.expected) {
            t.Errorf("not getting the number of expected words we want for %v", c.input)
        }
        for i := range actual {
            word := actual[i]
            expectedWord := c.expected[i]
            fmt.Println(word, expectedWord)
            if word != expectedWord {
                t.Errorf("%v is not equal to %v", word, expectedWord)
            }
        }
    }
}

