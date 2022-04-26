package ci

import (
	"dagger.io/dagger"
	"universe.dagger.io/go"
	"tom.chauveau.pro@icloud.com/golangci"
)

dagger.#Plan & {
	client: filesystem: ".": read: {
		include: ["**/*.go", "go.mod", "go.sum", ".golangci.yaml"]
	}

	actions: {
		_code: client.filesystem.".".read.contents

		test: go.#Test & {
			source:  _code
			package: "./..."
			command: flags: "-race": true
		}

		lint: golangci.#Lint & {
			source: _code
		}
	}
}
