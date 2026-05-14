FROM node:24-alpine AS frontend-build
WORKDIR /src/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend ./
RUN npm run build

FROM golang:1.25-alpine AS backend-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY backend ./backend
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/nowhere ./backend/cmd

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=backend-build /out/nowhere /app/nowhere
COPY --from=frontend-build /src/frontend/dist /app/frontend/dist
ENV FRONTEND_DIST_DIR=/app/frontend/dist
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/nowhere"]
