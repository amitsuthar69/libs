### LRU Cache Usage

```go
package main

import (
	cache "github.com/amitsuthar69/libs/http"
)

func main() {
	lc := cache.NewLRUCache[any, any](3)
	lc.Set("name", "amit")
	lc.Set("age", 20)
	lc.Set("height", 6)

	fmt.Println(lc.Get("name"))   // amit, true
	fmt.Println(lc.Get("age"))    // 20, true
	fmt.Println(lc.Get("height")) // 6, true

	fmt.Println("name", lc.Contains("name"))     // true
	fmt.Println("age", lc.Contains("age"))       // true
	fmt.Println("height", lc.Contains("height")) // true

	fmt.Println(lc.Len())

	lc.Set("package", 11)

	fmt.Println("name", lc.Contains("name"))       // false
	fmt.Println("age", lc.Contains("age"))         // true
	fmt.Println("height", lc.Contains("height"))   // true
	fmt.Println("package", lc.Contains("package")) // true
}
```
