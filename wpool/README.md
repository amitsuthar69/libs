> ### Usage

```
go get github.com/amitsuthar69/libs/wpool
```

```go
func makeReq() string {
	res, err := http.Get("https://jsonplaceholder.typicode.com/todos/1")
	if err != nil {
		return err.Error()
	}
	return res.Status
}

func main() {
  // initialize the worker pool
	wpool := wpool.NewWPool(2)

  // collect result
	go func() {
		for res := range wpool.Result() {
			fmt.Println("Res: ", res)
		}
	}()

  // spawn workers
	for range 10 {
		wpool.AddWork(func() any {
			return makeReq()
		})
	}

  // wait and close the pool
	wpool.Close()
}
```
