## Go Websocket

### How to use:

1. Import this package
2. Create a basic http server using `net/http` package.
3. Create a route of your choosing and pass in following arguments into `WSUpgrade`
   - `http.ResponseWriter` from `net/http`.
   - `*http.Request` from `net/http`.
   - Your callback function for when messages arrive.

### Example:

```go
func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		WSUpgrade(w, r, func(message Message) {
			// Your messages will be shown here.
		})
	})
	log.Fatal(http.ListenAndServe(":3000", nil))
}
```

### Messages

Messages sent from the client to server will be shown in a `Message` struct.

```go
type Message struct {
	Type string
	Data interface{}
}
```

If `Message.Type` is "string", then `Message.Data` will have a `string`.\
If `Message.Typoe` is "binary", then `Message.Data` will have `[]byte`.
