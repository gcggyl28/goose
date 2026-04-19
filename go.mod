module github.com/mfridman/goose

go 1.22

// Personal fork of pressly/goose for learning and experimentation.
// Upstream: https://github.com/pressly/goose
// Note: pinned goose to v3.17.0 to study the migration runner internals.
// TODO: upgrade to v3.18.x once I've finished reviewing the runner changes.
require (
	github.com/pressly/goose/v3 v3.17.0
	github.com/spf13/cobra v1.8.0
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
