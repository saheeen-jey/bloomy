# bloomy

## A space-efficient set for when "maybe" is a good enough answer.

Checking whether a value exists in a huge set — a billion crawled URLs, every
username ever taken, a blocklist of ten million hashes — usually means either
a database round-trip or holding the whole thing in memory. Most of the time
you don't need a definitive answer that fast. You need a very fast **"almost
certainly not"** so you can skip the expensive check entirely.

That's what a Bloom filter is for.

**`bloomy` is a small, dependency-free Go implementation of a Bloom filter: a
probabilistic set that can tell you an item is *definitely not* in the set,
or *probably* is.** No false negatives, ever. A tunable, small rate of false
positives. A fraction of the memory of a real set or map.

## How it works

A Bloom filter is a bit array of size `m`, plus `k` hash functions. Adding an
item hashes it `k` times and flips those `k` bits on. Checking an item hashes
it the same way and checks whether all `k` bits are already set — if even one
isn't, the item was never added. If all are set, it probably was (or it's a
collision).

`bloomy` picks sane values for `m` and `k` automatically based on how many
items you expect to add and how many false positives you're willing to
tolerate, using the standard sizing formulas. Two independent FNV hashes are
combined via double hashing (Kirsch–Mitzenmacher) to cheaply simulate `k`
hash functions from just two.

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

Two filters built with the same `m` and `k` can be merged with `Union`,
which is handy for combining filters built independently (e.g. across
shards):

```go
merged := bloom.NewWithEstimates(10000, 0.01)
_ = merged.Union(otherFilter)
```

## CLI demo

The `cmd/bloomy` binary loads a newline-delimited wordlist into a filter and
checks a comma-separated list of words against it:

```bash
go run ./cmd/bloomy -wordlist examples/wordlist.txt -check "apple,pineapple,banana"
```

```
loaded 24 words (m=230 bits, k=7 hashes, est. false positive rate 0.9967%)

  apple                -> possibly in set
  pineapple            -> definitely NOT in set
  banana               -> possibly in set
```

## Testing

```bash
go test ./...
```

Tests cover: no false negatives ever occur, the observed false-positive rate
stays within a generous margin of the target rate, unioned filters contain
both source sets, and mismatched filters refuse to union.

## Why not just use a map?

You can, until the set gets big enough that memory matters. A Bloom filter
sized for a 1% false-positive rate uses roughly 9.6 bits per item — regardless
of how large or small the items themselves are. A `map[string]struct{}` with
the same items costs the size of every key, plus Go's map overhead. For use
cases like "have I already crawled this URL" or "is this cache key definitely
a miss," that trade-off is usually worth it.

## License

MIT — see [LICENSE](LICENSE).
