// Command retail prepares the retail differential inputs, runs the Go
// decoders over them and compares every decoder's verdicts.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"protocolgen/differential/retail"
)

const usage = `usage: retail <command> [flags]

  synth    regenerate the committed synthetic corpus
  prepare  verify the pinned fixtures and write every input
  decode   run the generated Go codec and gophertunnel over the inputs
  compare  compare all four decoders' verdicts against the exceptions
  bench-check  check both benches reported exactly the synthetic corpus's paths`

var errFailed = errors.New("comparison failed")

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	commands := map[string]func([]string) error{"synth": synth, "prepare": prepare, "bench-check": benchCheck, "decode": decode, "compare": compare}
	command, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if err := command(os.Args[2:]); err != nil {
		if !errors.Is(err, errFailed) {
			fmt.Fprintln(os.Stderr, "retail:", err)
		}
		os.Exit(1)
	}
}

func synth(args []string) error {
	flags := flag.NewFlagSet("synth", flag.ExitOnError)
	out := flags.String("out", retail.SyntheticDir, "synthetic corpus directory")
	encodeOut := flags.String("encode-out", retail.EncodeDir, "encode-from-struct entries directory")
	flags.Parse(args)
	files, err := retail.SyntheticCorpus()
	if err != nil {
		return err
	}
	if err := retail.WriteSyntheticCorpus(*out, files); err != nil {
		return err
	}
	if err := retail.WriteEncodeCases(*encodeOut); err != nil {
		return err
	}
	fmt.Printf("wrote %d synthetic files to %s\n", len(files), *out)
	return nil
}

func benchCheck(args []string) error {
	flags := flag.NewFlagSet("bench-check", flag.ExitOnError)
	synthetic := flags.String("synthetic", retail.SyntheticDir, "committed synthetic corpus")
	goOut := flags.String("go", "", "go test -bench output")
	count := flags.Int("count", 1, "the -count the Go benches ran with")
	criterion := flags.String("criterion", "", "criterion output directory")
	flags.Parse(args)
	files, err := retail.ReadSyntheticCorpus(*synthetic)
	if err != nil {
		return err
	}
	want := retail.BenchKeys(files)
	goKeys, err := retail.GoBenchKeys(*goOut, *count)
	if err != nil {
		return err
	}
	if err := retail.CheckKeys("go benches", want, goKeys); err != nil {
		return err
	}
	rustKeys, err := retail.CriterionKeys(*criterion)
	if err != nil {
		return err
	}
	if err := retail.CheckKeys("criterion benches", want, rustKeys); err != nil {
		return err
	}
	fmt.Printf("both benches report the %d paths of the synthetic corpus\n", len(want))
	return nil
}

func prepare(args []string) error {
	flags := flag.NewFlagSet("prepare", flag.ExitOnError)
	corpusPath := flags.String("corpus", "corpus.json", "retail corpus")
	cinnabar := flags.String("cinnabar", "", "Cinnabar checkout at the corpus commit")
	synthetic := flags.String("synthetic", retail.SyntheticDir, "committed synthetic corpus")
	out := flags.String("out", "", "output directory")
	flags.Parse(args)
	if *cinnabar == "" || *out == "" {
		return errors.New("prepare needs -cinnabar and -out")
	}
	built, err := retail.BuildAllInputs(*corpusPath, *cinnabar, *synthetic)
	if err != nil {
		return err
	}
	if err := retail.WriteInputs(*out, built.Inputs); err != nil {
		return err
	}
	fmt.Printf("prepared %d inputs from %d retail and %d synthetic fixtures; unknown-id uses packet id %d\n",
		len(built.Inputs), built.Retail, built.Synthetic, built.UnknownID)
	return nil
}

func decode(args []string) error {
	flags := flag.NewFlagSet("decode", flag.ExitOnError)
	out := flags.String("out", "", "directory prepare wrote")
	flags.Parse(args)
	inputs, err := retail.ReadInputs(*out)
	if err != nil {
		return err
	}
	var generated, oracle []retail.Verdict
	for _, input := range inputs {
		generated = append(generated, retail.DecodeGenerated(input)...)
		oracle = append(oracle, retail.DecodeGophertunnel(input)...)
	}
	if err := retail.WriteVerdicts(*out, retail.DecoderGo, generated); err != nil {
		return err
	}
	return retail.WriteVerdicts(*out, retail.DecoderGophertunnel, oracle)
}

func compare(args []string) error {
	flags := flag.NewFlagSet("compare", flag.ExitOnError)
	corpusPath := flags.String("corpus", "corpus.json", "retail corpus")
	exceptionsPath := flags.String("exceptions", "exceptions.json", "reviewed exceptions")
	goMod := flags.String("gomod", "go.mod", "this module's go.mod, which pins the oracle")
	out := flags.String("out", "", "directory holding expected.tsv, verdicts/ and fields/")
	cinnabar := flags.String("cinnabar", "", "Cinnabar checkout at the corpus commit, to rebuild the expected inputs")
	synthetic := flags.String("synthetic", retail.SyntheticDir, "committed synthetic corpus")
	rustSource := flags.String("rust-source", "../../generated/1.26.51/rust-contract/src", "the contract crate's sources, for enum and union tables")
	goSource := flags.String("go-source", "../../generated/1.26.51/go/protocol", "the generated Go protocol package, for union tables")
	flags.Parse(args)
	corpus, err := retail.LoadCorpus(*corpusPath)
	if err != nil {
		return err
	}
	oracle, err := retail.GophertunnelReplacement(*goMod)
	if err != nil {
		return err
	}
	exceptions, err := retail.LoadExceptions(*exceptionsPath, corpus, oracle)
	if err != nil {
		return err
	}
	if *cinnabar == "" {
		return errors.New("compare needs -cinnabar")
	}
	built, err := retail.BuildAllInputs(*corpusPath, *cinnabar, *synthetic)
	if err != nil {
		return err
	}
	inputs, err := retail.ReadInputs(*out)
	if err != nil {
		return err
	}
	if err := retail.CheckKeys("inputs", retail.InputKeys(built.Inputs), retail.InputKeys(inputs)); err != nil {
		return err
	}
	expected, err := retail.ReadExpected(*out)
	if err != nil {
		return err
	}
	wantExpected, err := retail.ExpectedEntries(built.Inputs)
	if err != nil {
		return err
	}
	if err := retail.CheckKeys("expected entries", retail.ExpectedKeys(wantExpected), retail.ExpectedKeys(expected)); err != nil {
		return err
	}
	verdicts := map[string][]retail.Verdict{}
	for _, decoder := range retail.Decoders {
		if verdicts[decoder], err = retail.ReadVerdicts(*out, decoder); err != nil {
			return err
		}
	}
	schema, err := retail.LoadSchema(*rustSource, *goSource)
	if err != nil {
		return err
	}
	checks, err := retail.CompareFields(*out, inputs, verdicts[retail.DecoderRust], schema)
	if err != nil {
		return err
	}
	if unused := schema.UnusedAliases(); len(unused) != 0 {
		return fmt.Errorf("field aliases no entry needed: %v", unused)
	}
	found := append(retail.Compare(expected, verdicts), retail.FieldDisagreements(checks)...)
	report := retail.Apply(found, exceptions)
	summary := retail.Summarize(verdicts) + fmt.Sprintf("rust fields agree with go on %d of %d decoded entries\n",
		len(checks)-len(retail.FieldDisagreements(checks)), len(checks))
	text := formatReport(summary, report)
	fmt.Print(text)
	if err := os.WriteFile(filepath.Join(*out, "report.txt"), []byte(text), 0o644); err != nil {
		return err
	}
	if report.Failed() {
		return errFailed
	}
	return nil
}

func formatReport(summary string, report retail.Report) string {
	var b strings.Builder
	b.WriteString(summary)
	fmt.Fprintf(&b, "\n%d unexpected disagreements, %d accepted by exceptions, %d stale exceptions\n",
		len(report.Unexpected), len(report.Accepted), len(report.Stale))
	for _, d := range report.Unexpected {
		fmt.Fprintf(&b, "UNEXPECTED\t%s\t%s\t%s\t%d\t%s\t%s\n", d.Decoder, d.Case, d.Fixture, d.Entry, d.Kind, d.Detail)
	}
	for _, d := range report.Accepted {
		fmt.Fprintf(&b, "accepted\t%s\t%s\t%s\t%d\t%s\t%s\n", d.Decoder, d.Case, d.Fixture, d.Entry, d.Kind, d.Detail)
	}
	for _, stale := range report.Stale {
		e := stale.Exception
		fmt.Fprintf(&b, "STALE\t%s\t%s\t%s\t%s\t%s\n", e.Decoder, e.Case, stale.Fixture, e.Kind, e.Reason)
	}
	return b.String()
}
