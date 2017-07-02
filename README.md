# githuuk
A GitHub webhook receiver written in Go.

This project was originally a fork of [phayes/hookserve](https://github.com/phayes/hookserve), but has been rewritten nearly completely.

```go
import "maunium.net/go/githuuk"

func main() {
  server := githuuk.NewServer()
  server.Port = 8888
  server.Secret = "GitHub webhook secret"
  server.AsyncListenAndServe()

  for event := range server.Events {

  }
}
```
