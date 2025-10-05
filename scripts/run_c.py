import sys
import os

if len(sys.argv) == 1:
    print("no file")

file = sys.argv[1]
TEMP = "__temp.exe"


if os.system(f"gcc -o {TEMP} {file}") != 0:
    exit()

try:
    os.system(TEMP)
finally:
    os.unlink(TEMP)
