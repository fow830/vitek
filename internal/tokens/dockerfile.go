package tokens

// RenderDockerfile returns the canonical multi-stage Dockerfile body.
func RenderDockerfile() string {
	return "" +
		"# syntax=docker/dockerfile:1\n" +
		"\n" +
		"FROM " + ImageGoBuild() + " AS build\n" +
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
		"FROM " + ImageWorkerRuntime + " AS worker\n" +
		"ENV DEBIAN_FRONTEND=noninteractive\n" +
		"RUN apt-get update && apt-get install -y --no-install-recommends \\\n" +
		"    ca-certificates \\\n" +
		"    curl \\\n" +
		"    fonts-liberation \\\n" +
		"    fonts-noto-core \\\n" +
		"    gnupg \\\n" +
		"    xauth \\\n" +
		"    xvfb \\\n" +
		"  && curl -fsSL https://dl.google.com/linux/linux_signing_key.pub \\\n" +
		"    | gpg --dearmor -o /usr/share/keyrings/google-chrome.gpg \\\n" +
		"  && echo \"deb [arch=amd64 signed-by=/usr/share/keyrings/google-chrome.gpg] http://dl.google.com/linux/chrome/deb/ stable main\" \\\n" +
		"    > /etc/apt/sources.list.d/google-chrome.list \\\n" +
		"  && apt-get update && apt-get install -y --no-install-recommends google-chrome-stable \\\n" +
		"  && rm -rf /var/lib/apt/lists/*\n" +
		"COPY --from=build /out/" + BinaryWorker + " /" + BinaryWorker + "\n" +
		"ENV " + EnvRodBrowser + "=" + DefaultRodChromeBin + "\n" +
		"ENV " + RodChromeEnvTZ + "=" + RodChromeTimezone + "\n" +
		"ENV " + RodChromeEnvLANG + "=" + RodChromeLANGValue + "\n" +
		"ENTRYPOINT [\"/" + BinaryWorker + "\"]\n"
}
