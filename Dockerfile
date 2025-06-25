FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

# Install TailwindCSS
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 \
  && chmod +x tailwindcss-linux-x64 \
  && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Install Templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install Make
RUN apt-get update && apt-get install -y make

WORKDIR /app