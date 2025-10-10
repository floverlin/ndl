import os
import subprocess
import pathlib
import re

TEMP_NAME = "__test_build.exe"
TEST_FOLDER = "tests"

os.system(f"go build -o {TEMP_NAME} .")


def read_expected(file: pathlib.Path) -> list[str]:
    with open(file, "r", encoding="utf-8") as f:
        expected = [
            re.search("//#(.+)", line).group(1).strip()
            for line in f.readlines()
            if "//#" in line
        ]
    return expected


try:
    ok_flag = True
    for file in pathlib.Path(TEST_FOLDER).rglob("*.ndl"):
        expected = read_expected(file)
        result = subprocess.run(
            [TEMP_NAME, file],
            capture_output=True,
            text=True,
        )
        if result.stderr != "":
            ok_flag = False
            print(file, "-> EXCEPTION:", result.stderr)
            break
        out = result.stdout.splitlines()
        start, end = out.index("== Output =="), out.index("== Results ==")
        if out[start + 1 : end] == expected:
            print(file, "-> OK")
        else:
            ok_flag = False
            if len(out[start + 1 : end]) != len(expected):
                print(
                    file,
                    f"-> ERROR: want {len(expected)} lines, got {len(out[start + 1 : end])}",
                )
                continue
            for i, line in enumerate(out[start + 1 : end]):
                if line != expected[i]:
                    print(file, f"-> ERROR: expected {expected[i]}, got {line}")

    print("=" * 16)
    print("OK!" if ok_flag else "ERROR!")
finally:
    os.unlink(TEMP_NAME)
