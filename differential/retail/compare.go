package retail

import (
	"fmt"
	"sort"
	"strings"
)

// Disagreement kinds.
const (
	KindMissing       = "missing"        // no verdict for an entry the harness framed
	KindExtra         = "extra"          // a verdict for an entry the harness did not frame
	KindFrame         = "frame"          // the decoder could not frame the batch, or framed past its end
	KindPanic         = "panic"          // the decoder panicked, whatever the input
	KindHeader        = "header"         // the decoder read different header bits
	KindPayload       = "payload"        // the decoder framed different entry bytes
	KindDecode        = "decode"         // not decoded, or for unknown-id not kept as unknown
	KindReencode      = "reencode"       // re-encoding did not reproduce the entry
	KindRoundTrip     = "roundtrip"      // decoding the re-encoding changed the decoded fields
	KindAccepted      = "accepted"       // a malformed or misdirected entry was not rejected
	KindClass         = "class"          // rejected for a reason other than the probed one
	KindLimitMismatch = "limit-mismatch" // Rust and Go disagree on which entries exceed a zero element limit
	KindZeroCopy      = "zero-copy"      // a frame or unknown body copied the batch, or compaction kept it
	KindUnexercised   = "unexercised"    // decoded although its large reservation was refused
	KindDecision      = "decision"       // accepted what gophertunnel rejects, or rejected what it accepts
	KindFields        = "fields"         // Rust decoded different field values than the generated Go
)

var kinds = []string{
	KindMissing, KindExtra, KindFrame, KindPanic, KindHeader, KindPayload, KindDecode, KindReencode, KindRoundTrip,
	KindAccepted, KindClass, KindLimitMismatch, KindZeroCopy, KindUnexercised, KindDecision, KindFields,
}

// FieldDisagreements turns failed field checks into Rust disagreements.
func FieldDisagreements(checks []FieldCheck) []Disagreement {
	var out []Disagreement
	for _, c := range checks {
		if c.Diff != "" {
			out = append(out, Disagreement{DecoderRust, c.Case, c.Fixture, c.Entry, KindFields, c.Diff})
		}
	}
	return out
}

// Disagreement is one departure from a case's expectation.
type Disagreement struct {
	Decoder string
	Case    string
	Fixture string
	Entry   int
	Kind    string
	Detail  string
}

type expectation int

const (
	expectRoundTrip expectation = iota // decodes, re-encodes byte for byte, reads back the same
	expectUnknown                      // kept as unknown and forwarded byte for byte
	expectReject                       // any decode error
	expectDirection                    // a direction error before the body is read
	expectLimit                        // decodes, or fails on the element limit
	expectAlloc                        // fails on the refused allocation
	expectShared                       // frame and unknown bodies share the batch
	expectDecision                     // accepts or rejects as gophertunnel does
)

type rule struct {
	source   string // input case whose entries the verdicts describe
	decoders []string
	expect   expectation
}

var rules = map[string]rule{
	CaseRetail:        {CaseRetail, Decoders, expectRoundTrip},
	CaseCrafting:      {CaseCrafting, Decoders, expectRoundTrip},
	CaseUnknownID:     {CaseUnknownID, Decoders, expectUnknown},
	CaseWrongSender:   {CaseWrongSender, Decoders, expectDirection},
	CaseTruncated:     {CaseTruncated, Decoders, expectReject},
	CaseTrailing:      {CaseTrailing, Decoders, expectReject},
	CaseLarge:         {CaseLarge, Decoders, expectRoundTrip},
	CaseOversized:     {CaseOversized, Decoders, expectDecision},
	ProbeElementLimit: {CaseRetail, []string{DecoderRust, DecoderGo}, expectLimit},
	ProbeAllocFailure: {CaseLarge, []string{DecoderRust}, expectAlloc},
	ProbeZeroCopy:     {CaseUnknownID, []string{DecoderRust}, expectShared},
}

type verdictKey struct {
	decoder  string
	caseName string
	fixture  string
	entry    int
}

// Compare checks every decoder's verdicts against each case's expectation and
// the entries the harness framed.
func Compare(expected []Expected, verdicts map[string][]Verdict) []Disagreement {
	var out []Disagreement
	index := map[verdictKey]Verdict{}
	frameAt := map[verdictKey]int{} // entry -1 keys the fixture
	for decoder, lines := range verdicts {
		for _, v := range lines {
			key := verdictKey{decoder, v.Case, v.Fixture, v.Entry}
			if _, repeated := index[key]; repeated {
				out = append(out, Disagreement{decoder, v.Case, v.Fixture, v.Entry, KindExtra, "repeated verdict"})
				continue
			}
			index[key] = v
			fixtureKey := verdictKey{decoder, v.Case, v.Fixture, -1}
			if at, ok := frameAt[fixtureKey]; v.Outcome == OutcomeFrameError && (!ok || v.Entry < at) {
				frameAt[fixtureKey] = v.Entry
			}
		}
	}
	bySource := map[string][]Expected{}
	for _, want := range expected {
		bySource[want.Case] = append(bySource[want.Case], want)
	}
	baseHolds := func(decoder, source, fixture string, entry int) bool {
		v, ok := index[verdictKey{decoder, source, fixture, entry}]
		if rules[source].expect == expectUnknown {
			return ok && v.Outcome == OutcomeUnknown
		}
		return ok && v.Outcome == OutcomeOK
	}
	seen := map[verdictKey]bool{}
	for caseName, r := range rules {
		probe := caseName != r.source
		for _, decoder := range r.decoders {
			for _, want := range bySource[r.source] {
				key := verdictKey{decoder, caseName, want.Fixture, want.Entry}
				v, ok := index[key]
				seen[key] = ok
				if probe && !baseHolds(decoder, r.source, want.Fixture, want.Entry) {
					if ok && v.Outcome == OutcomePanic {
						out = append(out, Disagreement{decoder, caseName, want.Fixture, want.Entry, KindPanic, v.Detail})
					}
					continue
				}
				if !ok {
					if at, failed := frameAt[verdictKey{decoder, caseName, want.Fixture, -1}]; !failed || want.Entry < at {
						out = append(out, Disagreement{decoder, caseName, want.Fixture, want.Entry, KindMissing, ""})
					}
					continue
				}
				for _, found := range check(r.expect, want, v) {
					out = append(out, Disagreement{decoder, caseName, want.Fixture, want.Entry, found[0], found[1]})
				}
			}
		}
	}
	for key, v := range index {
		switch {
		case seen[key]:
		case v.Outcome == OutcomePanic:
			out = append(out, Disagreement{key.decoder, key.caseName, key.fixture, key.entry, KindPanic, v.Detail})
		case v.Outcome == OutcomeFrameError:
			out = append(out, Disagreement{key.decoder, key.caseName, key.fixture, key.entry, KindFrame, "after the last framed entry: " + v.Detail})
		default:
			out = append(out, Disagreement{key.decoder, key.caseName, key.fixture, key.entry, KindExtra, "no such framed entry"})
		}
	}
	out = append(out, compareLimits(bySource[CaseRetail], index, baseHolds)...)
	out = append(out, compareDecisions(bySource[CaseOversized], index)...)
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Decoder != b.Decoder {
			return a.Decoder < b.Decoder
		}
		if a.Case != b.Case {
			return a.Case < b.Case
		}
		if a.Fixture != b.Fixture {
			return a.Fixture < b.Fixture
		}
		if a.Entry != b.Entry {
			return a.Entry < b.Entry
		}
		return a.Kind < b.Kind
	})
	return out
}

// check returns the kind and detail of each way v misses its expectation.
func check(expect expectation, want Expected, v Verdict) [][2]string {
	switch v.Outcome {
	case OutcomePanic:
		return [][2]string{{KindPanic, v.Detail}}
	case OutcomeFrameError:
		return [][2]string{{KindFrame, v.Detail}}
	}
	var out [][2]string
	add := func(kind, detail string) { out = append(out, [2]string{kind, detail}) }
	if v.Header != want.Header {
		add(KindHeader, fmt.Sprintf("read %s, framed %s", v.Header, want.Header))
	}
	if v.Payload != want.Payload {
		add(KindPayload, fmt.Sprintf("read %s, framed %s", v.Payload, want.Payload))
	}
	switch expect {
	case expectRoundTrip, expectUnknown:
		outcome := OutcomeOK
		if expect == expectUnknown {
			outcome = OutcomeUnknown
		}
		if v.Outcome != outcome {
			add(KindDecode, describe(v))
			break
		}
		// Bytes and field stability are separate dimensions, so an exception
		// for one cannot hide a failure in the other.
		if v.Encoded != want.Payload {
			add(KindReencode, fmt.Sprintf("encoded %s: %s", orNone(v.Encoded), v.Detail))
		}
		if expect == expectRoundTrip && v.RoundTrip != RoundTripSame {
			add(KindRoundTrip, describe(v))
		}
	case expectReject, expectDirection:
		if v.Outcome != OutcomeError {
			add(KindAccepted, describe(v))
		} else if expect == expectDirection && v.Class != ClassDirection {
			add(KindClass, describe(v))
		}
	case expectLimit:
		if v.Outcome != OutcomeOK && (v.Outcome != OutcomeError || v.Class != ClassLimit) {
			add(KindClass, describe(v))
		}
	case expectAlloc:
		if v.Outcome == OutcomeOK {
			add(KindUnexercised, describe(v))
		} else if v.Outcome != OutcomeError || v.Class != ClassAlloc {
			add(KindClass, describe(v))
		}
	case expectShared:
		if v.Outcome != OutcomeOK {
			add(KindZeroCopy, describe(v))
		}
	}
	return out
}

func compareLimits(retail []Expected, index map[verdictKey]Verdict, baseHolds func(string, string, string, int) bool) []Disagreement {
	var out []Disagreement
	for _, want := range retail {
		rust, rustOK := index[verdictKey{DecoderRust, ProbeElementLimit, want.Fixture, want.Entry}]
		goV, goOK := index[verdictKey{DecoderGo, ProbeElementLimit, want.Fixture, want.Entry}]
		if !rustOK || !goOK || !baseHolds(DecoderRust, CaseRetail, want.Fixture, want.Entry) || !baseHolds(DecoderGo, CaseRetail, want.Fixture, want.Entry) {
			continue
		}
		if (rust.Outcome == OutcomeOK) != (goV.Outcome == OutcomeOK) {
			out = append(out, Disagreement{DecoderRust, ProbeElementLimit, want.Fixture, want.Entry, KindLimitMismatch,
				fmt.Sprintf("rust %s; go %s", describe(rust), describe(goV))})
		}
	}
	return out
}

// compareDecisions checks that every decoder accepts exactly the oversized
// entries gophertunnel accepts; the reason for a rejection may differ.
func compareDecisions(oversized []Expected, index map[verdictKey]Verdict) []Disagreement {
	var out []Disagreement
	accepts := func(v Verdict) bool { return v.Outcome == OutcomeOK || v.Outcome == OutcomeUnknown }
	for _, want := range oversized {
		oracle, ok := index[verdictKey{DecoderGophertunnel, CaseOversized, want.Fixture, want.Entry}]
		if !ok || oracle.Outcome == OutcomePanic {
			continue
		}
		for _, decoder := range rules[CaseOversized].decoders {
			if decoder == DecoderGophertunnel {
				continue
			}
			v, ok := index[verdictKey{decoder, CaseOversized, want.Fixture, want.Entry}]
			if ok && v.Outcome != OutcomePanic && accepts(v) != accepts(oracle) {
				out = append(out, Disagreement{decoder, CaseOversized, want.Fixture, want.Entry, KindDecision,
					fmt.Sprintf("%s; gophertunnel %s", describe(v), describe(oracle))})
			}
		}
	}
	return out
}

func describe(v Verdict) string {
	text := v.Outcome
	if v.Class != "" {
		text += "/" + v.Class
	}
	if v.Outcome == OutcomeOK && v.RoundTrip != "" {
		text += " roundtrip " + v.RoundTrip
	}
	if v.Detail != "" {
		text += ": " + v.Detail
	}
	return text
}

// Summarize counts outcomes per decoder and case.
func Summarize(verdicts map[string][]Verdict) string {
	counts := map[[2]string]map[string]int{}
	for decoder, lines := range verdicts {
		for _, v := range lines {
			key := [2]string{decoder, v.Case}
			if counts[key] == nil {
				counts[key] = map[string]int{}
			}
			counts[key][v.Outcome]++
		}
	}
	keys := make([][2]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i][0] != keys[j][0] {
			return keys[i][0] < keys[j][0]
		}
		return keys[i][1] < keys[j][1]
	})
	var b strings.Builder
	for _, key := range keys {
		c := counts[key]
		fmt.Fprintf(&b, "%-13s %-14s ok=%d unknown=%d error=%d frame-error=%d\n", key[0], key[1],
			c[OutcomeOK], c[OutcomeUnknown], c[OutcomeError], c[OutcomeFrameError])
	}
	return b.String()
}
