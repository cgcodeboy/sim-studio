FROM golang:1.25-alpine AS build
WORKDIR /src
COPY des-engine /des-engine
COPY sim-studio/go.mod sim-studio/go.sum /src/
COPY sim-studio/cmd /src/cmd
COPY sim-studio/api /src/api
COPY sim-studio/internal /src/internal
RUN go mod edit -replace github.com/cgcodeboy/des-engine=/des-engine \
  && go mod download \
  && CGO_ENABLED=0 go build -o /out/api ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY --from=build /out/api /api
ENV SIM_MONGO=mongodb://mongo:27017
EXPOSE 8080
HEALTHCHECK CMD wget -q -O - http://127.0.0.1:8080/api/health || exit 1
ENTRYPOINT ["/api"]
