<a href="https://answer.apache.org">
    <img alt="logo" src="docs/img/logo.svg" height="99px">
</a>

# Apache Answer - Build Q&A platform

A Q&A platform software for teams at any scales. Whether it’s a community forum, help center, or knowledge management platform, you can always count on Answer.

To learn more about the project, visit [answer.apache.org](https://answer.apache.org).

[![LICENSE](https://img.shields.io/github/license/apache/answer)](https://github.com/apache/answer/blob/main/LICENSE)
[![Language](https://img.shields.io/badge/language-go-blue.svg)](https://golang.org/)
[![Language](https://img.shields.io/badge/language-react-blue.svg)](https://reactjs.org/)
[![Go Report Card](https://goreportcard.com/badge/github.com/apache/answer)](https://goreportcard.com/report/github.com/apache/answer)
[![Discord](https://img.shields.io/badge/discord-chat-5865f2?logo=discord&logoColor=f5f5f5)](https://discord.gg/Jm7Y4cbUej)

## Screenshots

![screenshot](docs/img/screenshot.png)

## Quick start

### Running with docker

```bash
docker run -d -p 9080:80 -v answer-data:/data --name answer apache/answer:1.7.1
```

For more information, see [Installation](https://answer.apache.org/docs/installation).

### Plugins

Answer provides a plugin system for developers to create custom plugins and expand Answer’s features. You can find the [plugin documentation here](https://answer.apache.org/community/plugins).

We value your feedback and suggestions to improve our documentation. If you have any comments or questions, please feel free to contact us. We’re excited to see what you can create using our plugin system!

You can also check out the [plugins here](https://answer.apache.org/plugins).

## Building from Source

### Prerequisites

- Golang >= 1.23
- Node.js >= 20
- pnpm >= 9
- [mockgen](https://github.com/uber-go/mock?tab=readme-ov-file#installation) >= 0.6.0
- [wire](https://github.com/google/wire/) >= 0.5.0

### Environment Configuration

    You may create a `.env` file in the project root.
    Copy `.env.example` and fill required values.

    ```bash
    cp .env.example .env


### Build

```bash
# Install wire and mockgen for building. You can run `make check` to check if they are installed.
$ make generate
# Install frontend dependencies and build
$ make ui
# Install backend dependencies and build
$ make build
```

## Contributing

Contributions are always welcome!

See [CONTRIBUTING](https://answer.apache.org/community/contributing) for ways to get started.

## License

[Apache License 2.0](https://github.com/apache/answer/blob/main/LICENSE)
