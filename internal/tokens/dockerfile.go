package tokens

// RenderDockerfile returns the canonical multi-stage Dockerfile body.
func RenderDockerfile() string {
	return "" +
		"# syntax=docker/dockerfile:1\n" +
		"\n" +
		"FROM " + ImageGoBuild + " AS build\n" +
		"WORKDIR /src\n" +
		"RUN apk add --no-cache ca-certificates\n" +
		"COPY go.mod go.sum ./\n" +
		"RUN go mod download\n" +
		"COPY . .\n" +
		"RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags=\"-s -w\" -o /out/" + BinaryAPI + " ./cmd/app\n" +
		"RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags=\"-s -w\" -o /out/" + BinaryWorker + " ./cmd/worker\n" +
		"\n" +
		"FROM " + ImageRuntime + " AS api\n" +
		"COPY --from=build /out/" + BinaryAPI + " /" + BinaryAPI + "\n" +
		"USER nonroot:nonroot\n" +
		"EXPOSE " + DefaultHTTPPort + "\n" +
		"ENTRYPOINT [\"/" + BinaryAPI + "\"]\n" +
		"\n" +
		"FROM " + ImageRuntime + " AS worker\n" +
		"COPY --from=build /out/" + BinaryWorker + " /" + BinaryWorker + "\n" +
		"USER nonroot:nonroot\n" +
		"ENTRYPOINT [\"/" + BinaryWorker + "\"]\n"
}
