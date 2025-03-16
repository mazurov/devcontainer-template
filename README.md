# Devcontainer Template

This project provides a CLI tool to apply devcontainer templates to your workspace.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Installation of Go Module

To install the CLI tool using `go get`, run the following command:

```sh
go get github.com/mazurov/devcontainer-template
```

This will download and install the `devctmpl` binary to your `$GOPATH/bin` directory.

### Installation from source

To install the CLI tool, you can build it from source:

```sh
git clone https://github.com/mazurov/devcontainer-template.git
cd devcontainer-template
make build
```

This will create the `devctmpl` binary in the `bin/` directory.

## Usage

To use the CLI tool, run the following command:

```sh
devctmpl -w /path/to/workspace -t template-id --template-args '{"key": "value"}' --log-level info
```

### Flags

- `-w, --workspace-folder`: Target workspace folder (required)
- `-t, --template-id`: Source template directory (required)
- `-a, --template-args`: Template arguments as JSON string
- `--tmp-dir`: Directory to use for temporary files. If not provided, the system default will be used.
- `--keep-tmp-dir`: Keep temporary directory after execution
- `--omit-paths`: List of paths within the Template to omit applying, provided as JSON. To ignore a directory append '/*'
- `-l, --log-level`: Log level (debug, info, warn, error)

## Development

To contribute to the project, follow these steps:

1. Fork the repository
2. Create a new branch (`git checkout -b feature-branch`)
3. Make your changes
4. Commit your changes (`git commit -am 'Add new feature'`)
5. Push to the branch (`git push origin feature-branch`)
6. Create a new Pull Request

## Testing

To run the tests, use the following command:

```sh
make test
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
```

This README.md provides an overview of the project, installation instructions, usage examples, development guidelines, testing instructions, and contribution guidelines. Adjust the content as needed to fit your project's specifics.
