//! The vectors `differential/retail/fields_test.go` checks its from_utf8_lossy
//! port against, run through the real thing.

#[test]
fn from_utf8_lossy_replaces_each_maximal_subpart() {
    for (input, want) in [
        (&b"\xff\xff"[..], "\u{FFFD}\u{FFFD}"),
        (b"a\xe2\x82b", "a\u{FFFD}b"),
        (b"\xe2\x82\xac\xe2", "\u{20ac}\u{FFFD}"),
        (b"\xf0\x80\x80", "\u{FFFD}\u{FFFD}\u{FFFD}"),
        (b"\xed\xa0\x80", "\u{FFFD}\u{FFFD}\u{FFFD}"),
        (b"\xf4\x8f\xbf", "\u{FFFD}"),
        (b"\xc0\xaf", "\u{FFFD}\u{FFFD}"),
        (b"ok", "ok"),
    ] {
        assert_eq!(String::from_utf8_lossy(input), want, "{input:x?}");
    }
}
