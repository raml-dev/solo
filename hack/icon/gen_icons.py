#!/usr/bin/env python3
"""
gen_icons.py — Generate Wails icon files from an SVG source.

Produces:
  build/windows/icon.ico   (16,32,48,64,256 px)
  build/darwin/icon.icns   (16,32,64,128,256,512,1024 px + @2x variants)
  build/linux/icon.png     (256 px)

Requirements:
  pip install cairosvg pillow

Usage:
  python gen_icons.py icon.svg
  python gen_icons.py icon.svg --out ./build
"""

import argparse
import os
import struct
import zlib
import sys
from pathlib import Path

try:
    import cairosvg
except ImportError:
    sys.exit("Missing dependency: pip install cairosvg")

try:
    from PIL import Image
except ImportError:
    sys.exit("Missing dependency: pip install pillow")

import io


# ── Rasterise ────────────────────────────────────────────────────────────────

def svg_to_png_bytes(svg_path: Path, size: int) -> bytes:
    return cairosvg.svg2png(
        url=str(svg_path),
        output_width=size,
        output_height=size,
    )


def png_bytes_to_image(data: bytes) -> Image.Image:
    return Image.open(io.BytesIO(data)).convert("RGBA")


# ── .ico ─────────────────────────────────────────────────────────────────────
# ICO format: header + directory entries + image data blobs.
# Sizes > 48px are stored as embedded PNGs (standard since Vista).

ICO_SIZES = [16, 32, 48, 64, 256]

def build_ico(svg_path: Path) -> bytes:
    images = []  # (size, png_bytes)
    for size in ICO_SIZES:
        png = svg_to_png_bytes(svg_path, size)
        images.append((size, png))

    count = len(images)
    # ICO header: 6 bytes
    header = struct.pack("<HHH", 0, 1, count)

    # Directory entries: 16 bytes each
    # Image data starts after header + all directory entries
    data_offset = 6 + count * 16
    directory = b""
    blobs = b""
    for size, png in images:
        w = 0 if size == 256 else size   # 256 is stored as 0 in ICO spec
        h = w
        data_size = len(png)
        directory += struct.pack(
            "<BBBBHHII",
            w, h,       # width, height
            0,          # color count (0 = no palette)
            0,          # reserved
            1,          # color planes
            32,         # bits per pixel
            data_size,
            data_offset,
        )
        data_offset += data_size
        blobs += png

    return header + directory + blobs


# ── .icns ────────────────────────────────────────────────────────────────────
# ICNS format: 4-byte magic + 4-byte total size + concatenated chunks.
# Each chunk: 4-byte OSType + 4-byte chunk size (including 8-byte header) + data.
# Modern macOS only needs the 'ic' series (PNG-compressed).

ICNS_TYPES = [
    # (ostype, pixel_size)   — covers 1x and 2x retina
    (b"icp4",   16),   # 16×16
    (b"icp5",   32),   # 32×32  (also serves as 16@2x)
    (b"icp6",   64),   # 64×64  (32@2x)
    (b"ic07",  128),   # 128×128
    (b"ic08",  256),   # 256×256 (128@2x)
    (b"ic09",  512),   # 512×512
    (b"ic10", 1024),   # 1024×1024 (512@2x)
]

def build_icns(svg_path: Path) -> bytes:
    chunks = b""
    for ostype, size in ICNS_TYPES:
        png = svg_to_png_bytes(svg_path, size)
        chunk_size = 8 + len(png)
        chunks += ostype + struct.pack(">I", chunk_size) + png

    magic = b"icns"
    total_size = 8 + len(chunks)
    return magic + struct.pack(">I", total_size) + chunks


# ── Main ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Generate Wails icons from SVG.")
    parser.add_argument("svg", help="Path to source SVG file")
    parser.add_argument("--out", help="Output root directory (default: current dir)")
    args = parser.parse_args()

    svg_path = Path(args.svg).resolve()
    if not svg_path.exists():
        print(f"SVG not found: {svg_path}")
        sys.exit(1)

    if not args.out:
        print("You must provide the --out param")
        sys.exit(1)

    out = Path(args.out).resolve()
    if not out.exists():
        print(f"Output folder does not exist: {out}")
        sys.exit(1)

    targets = {
        "windows": out / "windows",
        "darwin":  out / "darwin",
        "linux":   out / "linux",
    }
    for p in targets.values():
        p.mkdir(parents=True, exist_ok=True)

    print(f"Source: {svg_path}")

    # Windows .ico
    ico_path = targets["windows"] / "icon.ico"
    print(f"Building {ico_path} ({', '.join(str(s) for s in ICO_SIZES)}px)...")
    ico_path.write_bytes(build_ico(svg_path))
    print(f"  ✓ {ico_path} ({ico_path.stat().st_size // 1024} KB)")

    # macOS .icns
    icns_path = targets["darwin"] / "icon.icns"
    sizes_str = ", ".join(str(s) for _, s in ICNS_TYPES)
    print(f"Building {icns_path} ({sizes_str}px)...")
    icns_path.write_bytes(build_icns(svg_path))
    print(f"  ✓ {icns_path} ({icns_path.stat().st_size // 1024} KB)")

    # Linux .png (256px)
    linux_path = targets["linux"] / "icon.png"
    print(f"Building {linux_path} (512px)...")
    linux_path.write_bytes(svg_to_png_bytes(svg_path, 512))
    print(f"  ✓ {linux_path} ({linux_path.stat().st_size // 1024} KB)")

    # Generic appicon.png (256px)
    generic_appicon_path = out / "appicon.png"
    print(f"Building {generic_appicon_path} (512px)...")
    generic_appicon_path.write_bytes(svg_to_png_bytes(svg_path, 512))
    print(f"  ✓ {generic_appicon_path} ({generic_appicon_path.stat().st_size // 1024} KB)")

    print(f"\nDone. Go to {out} to check results.")


if __name__ == "__main__":
    main()
