// Helpers to run bash commands in containers
package bash

import (
	"dagger.io/dagger"
	"dagger.io/dagger/core"

	"universe.dagger.io/docker"
)

// Run a bash script in a Docker container
//  Since this is a thin wrapper over docker.#Run, we embed it.
//  Whether to embed or wrap is a case-by-case decision, like in Go.
#Run: {
	// The script to execute
	script: {
		// A directory containing one or more bash scripts
		directory: dagger.#FS

		// Name of the file to execute
		filename: string

		_directory: directory
		_filename:  filename
	} | {
		// Script contents
		contents: string

		_filename: "run.sh"
		_write:    core.#WriteFile & {
			input:      dagger.#Scratch
			path:       _filename
			"contents": contents
		}
		_directory: _write.output
	}

	// Arguments to the script
	args: [...string]

	// Where in the container to mount the scripts directory
	_mountpoint: "/bash/scripts"

	docker.#Run & {
		entrypoint: ["/bin/bash"]
		command: {
			name:   "\(_mountpoint)/\(script._filename)"
			"args": args
			// FIXME: make default flags overrideable
			flags: {
				"--norc": true
				"-e":     true
				"-o":     "pipefail"
			}
		}
		mounts: "Bash scripts": {
			contents: script._directory
			dest:     _mountpoint
		}
	}
}
