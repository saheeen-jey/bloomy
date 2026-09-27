# bloomy

## A space-efficient set for when "maybe" is a good enough answer.

You've got a huge set of things to check against. A billion crawled URLs, every username that's ever been taken, a blocklist with ten million entries. The obvious way to check membership is a database round trip or holding the whole thing in memory, and both get expensive fast once the set is large enough.

Most of the time, though, you don't need a guaranteed answer. You just need a fast, reliable "almost certainly not," so you can skip the expensive check for the 99% of lookups that would have come back empty anyway.

That's what a Bloom filter is for.

`bloomy` is a small, dependency-free Go implementation of one: a probabilistic set that can tell you an item is definitely not in the set, or probably is. No false negatives, ever. A false positive rate you control. A fraction of the memory a real set or map would cost you.

## How it works

A Bloom filter is a bit array of size `m`, plus `k` hash functions. Adding an item hashes it `k` times and flips those `k` bits on. Checking an item hashes it the same way, then checks whether all `k` bits are already set. If even one isn't, the item was never added. If all of them are, it probably was (or you've hit a collision).

`bloomy` picks sane values for `m` and `k` automatically, based on how many items you expect to add and how many false positives you're willing to tolerate, using the standard sizing formulas. Two independent FNV hashes are combined via double hashing (Kirsch-Mitzenmacher) to cheaply simulate `k` hash functions from just two.

## Install

```bash
go get github.com/saheeen-jey/bloomy
```

## Library usage

```go
package main

import (
	"fmt"

	"github.com/saheeen-jey/bloomy/bloom"
)

func main() {
	// Expect ~10,000 items, tolerate a 1% false positive rate.
	filter := bloom.NewWithEstimates(10000, 0.01)

	filter.Add([]byte("alice@example.com"))
	filter.Add([]byte("bob@example.com"))

	fmt.Println(filter.Test([]byte("alice@example.com"))) // true
	fmt.Println(filter.Test([]byte("carol@example.com")))  // almost certainly false
}
```

Two filters built with the same `m` and `k` can be merged with `Union`, which is handy for combining filters built independently, for example across shards:

```go
merged := bloom.NewWithEstimates(10000, 0.01)
_ = merged.Union(otherFilter)
```

## CLI demo

The `cmd/bloomy` binary loads a newline-delimited wordlist into a filter and checks a comma-separated list of words against it:

```bash
go run ./cmd/bloomy -wordlist examples/wordlist.txt -check "apple,pineapple,banana"
```
