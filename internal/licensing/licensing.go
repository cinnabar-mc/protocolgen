// Package licensing embeds the notices distributed with both codec outputs.
package licensing

import _ "embed"

// SPDX identifies the license of Cinnabar's authored implementation. Notice
// describes the separate rights in imported documentation, inputs and data.
const SPDX = "MIT"

// MIT is the canonical license for Cinnabar's copyrightable contributions.
//
//go:embed LICENSE-MIT
var MIT string

// Notice preserves the contribution scope and upstream attribution in outputs.
//
//go:embed NOTICE
var Notice string
