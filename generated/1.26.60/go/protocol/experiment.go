// Code generated from canonical protocol manifest v2. DO NOT EDIT.

package protocol

// ExperimentData holds data on an experiment that is either enabled or disabled.
type ExperimentToggle struct {
	// Name is the name of the experiment.
	Name string
	// Enabled specifies if the experiment is enabled. Vanilla typically always sets this to true for any
	// experiments sent.
	Enabled bool
}

// Marshal reads or writes ExperimentToggle using its canonical wire layout.
func (x *ExperimentToggle) Marshal(io IO) {
	io.String(&x.Name)
	io.Bool(&x.Enabled)
}

type LevelExperimentToggle struct {
	Name    string
	Enabled bool
}

// Marshal reads or writes LevelExperimentToggle using its canonical wire layout.
func (x *LevelExperimentToggle) Marshal(io IO) {
	io.String(&x.Name)
	io.Bool(&x.Enabled)
}
