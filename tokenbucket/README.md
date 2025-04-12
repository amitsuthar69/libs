> ### Usage

```
go get github.com/amitsuthar69/libs/tokenbucket
```

```go
func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	// wrap a regular http handler with this Limiter
	lHandler := tokenbucket.Limiter(handler, 2, time.Minute)

	fmt.Println("http://127.0.0.1:8080")
	http.ListenAndServe(":8080", lHandler)
}
```
