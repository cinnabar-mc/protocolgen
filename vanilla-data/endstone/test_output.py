#!/usr/bin/env python3
"""Test the adapter's save helpers with CXX and ZSTR_INCLUDE_DIR configured."""
import gzip
import os
from pathlib import Path
import resource
import signal
import subprocess
import tempfile
import unittest


def save_helpers():
    """Extract the adapter's local save helpers for a standalone C++ harness."""
    source = Path(__file__).with_name("headless_vanilla_data.cpp").read_text()
    start = source.index("    const auto save_json")
    end = source.index("\n    if (!save_json", start)
    return source[start:end]


def refuse_file_output():
    """Make buffered regular-file writes fail when the child flushes them."""
    signal.signal(signal.SIGXFSZ, signal.SIG_IGN)
    resource.setrlimit(resource.RLIMIT_FSIZE, (0, 0))


class OutputTests(unittest.TestCase):
    def test_finalization(self):
        """Successful exports are complete, and deferred write failures are reported."""
        include = os.environ.get("ZSTR_INCLUDE_DIR")
        if not include:
            self.skipTest("set ZSTR_INCLUDE_DIR to the zstr header directory")
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / "test.cpp"
            source.write_text(r'''
#include <filesystem>
#include <fstream>
#include <iostream>
#include <string>
#include <zstr.hpp>
namespace fs = std::filesystem;
namespace nlohmann { using json = std::string; }
using CompoundTag = std::string;
struct BigEndianStringByteOutput { std::string &buffer; explicit BigEndianStringByteOutput(std::string &b): buffer(b) {} };
struct NbtIo { static void writeNamedTag(const char *, const CompoundTag &value, BigEndianStringByteOutput &out) { out.buffer = value; } };
int main(int argc, char **argv) {
    const fs::path base_path = argv[1];
''' + save_helpers() + r'''
    std::cout << save_json("{}", "value.json") << " " << save_nbt("payload", "value.nbt") << "\n";
}
''')
            executable = root / "test"
            subprocess.run([os.environ.get("CXX", "c++"), "-std=c++17", "-I", include,
                            str(source), "-lz", "-o", str(executable)], check=True)
            good = subprocess.run([str(executable), str(root)], check=True, capture_output=True, text=True)
            self.assertEqual(good.stdout.strip(), "1 1")
            self.assertEqual((root / "value.json").read_text(), "{}")
            self.assertEqual(gzip.decompress((root / "value.nbt").read_bytes()), b"payload")
            failed = subprocess.run([str(executable), str(root)], capture_output=True, text=True,
                                    preexec_fn=refuse_file_output)
            self.assertEqual(failed.returncode, 0, failed.stderr)
            self.assertEqual(failed.stdout.strip(), "0 0")


if __name__ == "__main__":
    unittest.main()
