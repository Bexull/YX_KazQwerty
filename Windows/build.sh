#!/bin/bash
# Builds dist/YX_KazQwerty.exe for 64-bit Windows from any OS with Go: the autocorrect data from
# ../Resources/kazakh goes inside the exe, so it is a single file.
set -euo pipefail
cd "$(dirname "$0")"

rm -rf data/kazakh
mkdir -p data/kazakh dist
for file in dictionary.tsv context.tsv pairs.tsv endings.tsv russian.txt LICENSE.md; do
    cp "../Resources/kazakh/$file" data/kazakh/
done

go test ./kazakh/
# unsafeptr: hook data and version strings live in memory Windows owns.
GOOS=windows GOARCH=amd64 go vet -unsafeptr=false .
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-H windowsgui -s -w" -o dist/YX_KazQwerty.exe .
echo "Built dist/YX_KazQwerty.exe ($(du -h dist/YX_KazQwerty.exe | cut -f1))"
