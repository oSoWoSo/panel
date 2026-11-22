package panel

import (
	"fmt"
	"testing"
)

func parseArgs(n int) []string {
	args := []string{
		"--single", "--border", "rounded", "--gap", "0", "--height", "12",
		"--reverse", "--no-show-help", "--", "filter", "--no-strict", "--header", "Choose", "--",
	}
	for i := range n {
		args = append(args, fmt.Sprintf("pkg-%05d", i))
	}
	return args
}

func BenchmarkParseArgs(b *testing.B) {
	for _, n := range []int{100, 1000, 5000, 15000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			args := parseArgs(n)
			for b.Loop() {
				parser, err := newParser("test", func(int) {})
				if err != nil {
					b.Fatal(err)
				}
				if _, err := parser.Parse(args); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkParseFilterBlock(b *testing.B) {
	for _, n := range []int{100, 1000, 5000, 15000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			args := []string{"--no-strict", "--header", "Choose", "--"}
			for i := range n {
				args = append(args, fmt.Sprintf("pkg-%05d", i))
			}
			for b.Loop() {
				if _, err := parseFilterBlock(args); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
