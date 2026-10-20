//! Prints, as a markdown table, the heap allocations and bytes one decode and
//! one encode of each synthetic case make, which criterion cannot measure.

use std::alloc::{GlobalAlloc, Layout, System};
use std::sync::atomic::{AtomicUsize, Ordering};

use bedrock_protocol_1_26_51::packets::Packet;
use retail_bench::{SET_BATCH, load};

struct Counting;

static ALLOCS: AtomicUsize = AtomicUsize::new(0);
static BYTES: AtomicUsize = AtomicUsize::new(0);

unsafe impl GlobalAlloc for Counting {
    unsafe fn alloc(&self, layout: Layout) -> *mut u8 {
        ALLOCS.fetch_add(1, Ordering::Relaxed);
        BYTES.fetch_add(layout.size(), Ordering::Relaxed);
        unsafe { System.alloc(layout) }
    }
    unsafe fn dealloc(&self, ptr: *mut u8, layout: Layout) {
        unsafe { System.dealloc(ptr, layout) }
    }
    unsafe fn realloc(&self, ptr: *mut u8, layout: Layout, new_size: usize) -> *mut u8 {
        ALLOCS.fetch_add(1, Ordering::Relaxed);
        BYTES.fetch_add(new_size, Ordering::Relaxed);
        unsafe { System.realloc(ptr, layout, new_size) }
    }
}

#[global_allocator]
static GLOBAL: Counting = Counting;

const RUNS: usize = 64;

/// Allocations and bytes per call of `f`, averaged over `RUNS` calls.
fn measure(mut f: impl FnMut()) -> (usize, usize) {
    f();
    let (allocs, bytes) = (
        ALLOCS.load(Ordering::Relaxed),
        BYTES.load(Ordering::Relaxed),
    );
    for _ in 0..RUNS {
        f();
    }
    (
        (ALLOCS.load(Ordering::Relaxed) - allocs) / RUNS,
        (BYTES.load(Ordering::Relaxed) - bytes) / RUNS,
    )
}

fn main() {
    // `cargo bench` passes harness flags such as --bench; only run when asked
    // for everything or for this target by name.
    if std::env::args()
        .skip(1)
        .any(|arg| !arg.starts_with('-') && arg != "allocs")
    {
        return;
    }
    println!(
        "| case | entry bytes | decode allocs/op | decode B/op | encode allocs/op | encode B/op |"
    );
    println!("| --- | ---: | ---: | ---: | ---: | ---: |");
    for case in load().iter().filter(|case| case.file.set != SET_BATCH) {
        let frame = &case.frames[0];
        let decode = measure(|| {
            std::hint::black_box(
                Packet::decode_frame(frame, case.sender, &case.limits).expect("decode"),
            );
        });
        let mut out = Vec::with_capacity(case.batch.len());
        let encode = measure(|| {
            out.clear();
            case.packets[0]
                .encode_frame(case.sender, 0, 0, &mut out)
                .expect("encode");
        });
        println!(
            "| {} | {} | {} | {} | {} | {} |",
            case.file.name,
            case.entry_len(0),
            decode.0,
            decode.1,
            encode.0,
            encode.1
        );
    }
}
