//! Criterion benches of the contract crate over the committed synthetic
//! corpus. Groups and case names match `differential/retail/bench_test.go`
//! (`Decode/StartGame` here is `BenchmarkDecode/StartGame` there).

use std::hint::black_box;

use bedrock_protocol_1_26_51::packets::Packet;
use bedrock_protocol_1_26_51::wire::Frames;
use criterion::{Criterion, Throughput, criterion_group, criterion_main};
use retail_bench::{Case, SET_BATCH, load};

fn encode_batch(case: &Case, out: &mut Vec<u8>) {
    out.clear();
    out.push(0xfe);
    for packet in &case.packets {
        packet.encode_frame(case.sender, 0, 0, out).expect("encode");
    }
}

fn benches(c: &mut Criterion) {
    let cases = load();
    let (batches, singles): (Vec<&Case>, Vec<&Case>) =
        cases.iter().partition(|case| case.file.set == SET_BATCH);

    let mut group = c.benchmark_group("Decode");
    for case in &singles {
        group.throughput(Throughput::Bytes(case.entry_len(0) as u64));
        group.bench_function(case.file.name.as_str(), |b| {
            b.iter(|| {
                Packet::decode_frame(black_box(&case.frames[0]), case.sender, &case.limits)
                    .expect("decode")
            })
        });
    }
    group.finish();

    let mut group = c.benchmark_group("Encode");
    for case in &singles {
        let mut out = Vec::new();
        case.packets[0]
            .encode_frame(case.sender, 0, 0, &mut out)
            .expect("encode");
        assert_eq!(
            out.as_slice(),
            &case.frames[0].frame()[..],
            "{} does not re-encode byte for byte",
            case.file.file
        );
        group.throughput(Throughput::Bytes(case.entry_len(0) as u64));
        group.bench_function(case.file.name.as_str(), |b| {
            b.iter(|| {
                out.clear();
                black_box(&case.packets[0])
                    .encode_frame(case.sender, 0, 0, &mut out)
                    .expect("encode");
            })
        });
    }
    group.finish();

    let mut group = c.benchmark_group("DecodeBatch");
    for case in &batches {
        group.throughput(Throughput::Bytes(case.batch.len() as u64));
        group.bench_function(case.file.name.as_str(), |b| {
            b.iter(|| {
                for frame in
                    Frames::from_batch(black_box(case.batch.clone()), usize::MAX).expect("frame")
                {
                    let frame = frame.expect("frame entry");
                    black_box(
                        Packet::decode_frame(&frame, case.sender, &case.limits).expect("decode"),
                    );
                }
            })
        });
    }
    group.finish();

    let mut group = c.benchmark_group("EncodeBatch");
    for case in &batches {
        let mut out = Vec::new();
        encode_batch(case, &mut out);
        assert_eq!(
            out.as_slice(),
            &case.batch[..],
            "{} does not re-encode byte for byte",
            case.file.file
        );
        group.throughput(Throughput::Bytes(case.batch.len() as u64));
        group.bench_function(case.file.name.as_str(), |b| {
            b.iter(|| encode_batch(black_box(case), &mut out))
        });
    }
    group.finish();
}

criterion_group!(codec, benches);
criterion_main!(codec);
