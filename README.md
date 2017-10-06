# tracing

sample apps to test propagating traces between processes.

To run, start `auth`, then `server`, then `client`.

```
cd auth
go run main.go &

cd ../server
go run main.go &

cd ../client
go run main.go
```