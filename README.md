# jellyzset - Redis-Compatible Sorted Set Library in Go

`jellyzset` is a high-performance implementation of a Redis-compatible ZSet (sorted set) data structure in Go. This library is designed to efficiently manage ordered collections of elements associated with scores, offering logarithmic-time complexity for key operations.

## Overview

The `jellyzset` package employs a skip list-based indexing structure to optimize insertion, deletion, and range query operations inherent to sorted sets. Each `ZSet` instance manages multiple sorted sets, uniquely identified by keys. The underlying skip list, represented by the `zskiplist` struct, serves as the backbone for quick and effective access to elements within a sorted set.

### Key Features

- **Dynamic Balancing:** The skip list dynamically balances itself, ensuring rapid insertion and deletion operations.

- **Ordered Elements:** Element order is maintained based on scores, facilitating efficient range queries.

- **Logarithmic Time Complexity:** Skip list levels expedite searches, resulting in logarithmic time complexity for key operations.

## Implementation Details

- **Skip List:** The skip list is a fundamental part of the implementation, and its structure is based on research and insights from various resources, including:
  - [Skip Lists: A Probabilistic Alternative to Balanced Trees](https://www.epaperpress.com/sortsearch/download/skiplist.pdf)
  - [Redis Internal Data Structure: Skip List](http://blog.wjin.org/posts/redis-internal-data-structure-skiplist.html)
  - [Skip List Lookup Tree (BTree) in Redis & MySQL](https://developpaper.com/skip-list-lookup-tree-btree-in-redis-mysql/)

- **Inspiration:** The `jellyzset` library draws inspiration from the Redis source code and other relevant projects, including:
  - [Redis Source Code](https://github.com/redis/redis/blob/unstable/src/server.h)
  - [arriqaaq/zset](https://github.com/arriqaaq/zset)

## Installation

```bash
go get github.com/davidandw190/jellyzset
```

## Usage

### Creating a New ZSet

```go
import "github.com/davidandw190/jellyzset"

// Create a new instance of a ZSet
zset := jellyzset.New()
```

### Key Operations

```go
// ZAdd adds one or more members with associated scores to a sorted set.
zset := jellyzset.New()
zset.ZAdd("mySortedSet", 3.5, "member1", "value1")
zset.ZAdd("mySortedSet", 2.0, "member2", "value2")
zset.ZAdd("mySortedSet", 4.2, "member1", "updatedValue1")


// ZScore retrieves the score of a member in a sorted set.
exists, score := zset.ZScore("mySortedSet", "member1")


// ZCard returns the number of members in a sorted set.
count := zset.ZCard("mySortedSet")


// ZRank returns the rank of a member in a sorted set, with the scores ordered from low to high.
rank := zset.ZRank("mySortedSet", "member2")


// ZRevRank returns the rank of a member in a sorted set, with the scores ordered from high to low.
revRank := zset.ZRevRank("mySortedSet", "member1")


// ZRem removes one or more members from a sorted set.
removed := zset.ZRem("mySortedSet", "member1")


// ZScoreRange returns members with scores within the specified range in a sorted set.
results := zset.ZScoreRange("mySortedSet", 2.5, 4.0)


// ZRevScoreRange returns members with scores within the specified range in a sorted set, in reverse order.
result := zset.ZRevScoreRange("mySortedSet", 4.0, 2.0)


// ZKeyExists checks if a key exists in the ZSet.
exists := zset.ZKeyExists("mySortedSet")


// ZClear removes all members from a sorted set.
zset.ZClear("mySortedSet")
```