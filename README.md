# gRPC Chat with docker

### Instruction:

1. Build with `docker build --tag=docker_chat_grpc .`

2. Run with `docker run-it -p 8080:8080 docker_chat_grpc`

### Launch Chating

cmd := `go run client/main.go -N <name>`

1. Open tab #1 for first user, ex : `go run client/main.go -N foo`

2. Open tab #2 for second user, ex : `go run client/main.go -N bar`

3. Try it to Chat (bi-directional streaming)