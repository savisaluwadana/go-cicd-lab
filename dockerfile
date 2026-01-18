FROM golang:1.19-alpine
WORKDIR /app
# build
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o main .

FROM alpine:3.18
WORKDIR /app
COPY --from=0 /app/main /app/main
COPY --from=0 /app/frontend /app/frontend
EXPOSE 8080
ENV PORT=8080
CMD ["/app/main"]