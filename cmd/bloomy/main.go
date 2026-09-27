// Command bloomy is a small CLI demo for the bloom package. It loads a
// newline-delimited wordlist into a Bloom filter and checks whether a
// comma-separated list of words is (possibly) present.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/saheeen-jey/bloomy/bloom"
)

func main() {
	wordlistPath := flag.String("wordlist", "", "path to a newline-delimited wordlist file to load into the filter")
	checkWords := flag.String("check", "", "comma-separated list of words to check against the filter")
	fpRate := flag.Float64("fp", 0.01, "target false positive rate")
	flag.Parse()

	if *wordlistPath == "" || *checkWords == "" {
		fmt.Fprintln(os.Stderr, `usage: bloomy -wordlist words.txt -check "word1,word2,word3"`)
		flag.PrintDefaults()
		os.Exit(1)
	}

	words, err := readLines(*wordlistPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading wordlist: %v\n", err)
		os.Exit(1)
	}

	f := bloom.NewWithEstimates(uint(len(words)), *fpRate)
	for _, w := range words {
		f.Add([]byte(w))
	}

	fmt.Printf("loaded %d words (m=%d bits, k=%d hashes, est. false positive rate %.4f%%)\n\n",
		len(words), f.M(), f.K(), f.EstimatedFalsePositiveRate()*100)

	for _, w := range strings.Split(*checkWords, ",") {
		w = strings.TrimSpace(w)
		if w == "" {
			continue
		}
		if f.Test([]byte(w)) {
			fmt.Printf("  %-20s -> possibly in set\n", w)
		} else {
			fmt.Printf("  %-20s -> definitely NOT in set\n", w)
		}
	}
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
