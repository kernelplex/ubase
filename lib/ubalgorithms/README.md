# Data Structures and Algorithms

## Least recently used cache (LRUCache)

```go
    // Create a cache with 10 items
    cache := algorithms.NewLRUCache[string, string](10)

    for x := 1; x <= 11; x++ {
        key := fmt.Sprintf("key%d", x)
        value := fmt.Sprintf("value %d", x)
        cache.Put(key, value)
    }

    // Try getting the last item
    item, found := cache.Get("key11")
    if found {
        fmt.Printf("Got value: %s\n", item)
    }

    // The first item ought to have rolled off
    item, found = cache.Get("key1")

    if found {
        fmt.Printf("Ooops, we shouldn't be here.\n")
    } else {
        fmt.Printf("Cache miss.\n")
    }
```

## PriorityQueue

A queue which returns the higher priority first (those with smaller priority values).

```go
    // Generate a priority queue with an initial capacity of 10
    //
    // The first type parameter is the value, the second is the
    // priority.
    queue := algorithms.NewPriorityQueue[string, int](10)

    // Push some items with their priority. Smaller values have
    // higher priority.
    queue.Push("Three", 3)
    queue.Push("One", 1)
    queue.Push("Two", 2)

    // Should print:
    // One
    // Two
    // Three
    for queue.Size() > 0 {
        value, err := queue.Pop()
        if err != nil {
            panic(err)
        }
        fmt.Println(*value)
    }
```
