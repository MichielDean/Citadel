// zeroproposalsagent is a minimal fake agent binary used in tests to simulate
// an LLM filter that returns zero proposals (an empty JSON array).
// This exercises the error path in importCmd when the filter produces no output.
package main

import "fmt"

func main() {
	fmt.Println("[]")
}
