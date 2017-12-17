#!/usr/bin/env bats

setup() {
    targetdir=$(mktemp -d)
}

teardown() {
    rm -rf $targetdir
}

@test "lndir" {
  run go-lndir $PWD/sample-dir $targetdir
  [ "$status" -eq 0 ]
  [ "$(readlink $targetdir/ignored-file)" == "$PWD/sample-dir/ignored-file" ]
  [ "$(readlink -n $targetdir/included-file)" == "$PWD/sample-dir/included-file" ]
  [ "$(readlink -n $targetdir/dir1/ignored-file)" == "$PWD/sample-dir/dir1/ignored-file" ]
  [ "$(readlink -n $targetdir/dir1/included-file)" == "$PWD/sample-dir/dir1/included-file" ]
}

@test "lndir -gitignore" {
  run go-lndir -gitignore $PWD/sample-dir $targetdir
  [ "$status" -eq 0 ]
  [ "$(readlink -n $targetdir/included-file)" == "$PWD/sample-dir/included-file" ]
  [ "$(readlink -n $targetdir/dir1/included-file)" == "$PWD/sample-dir/dir1/included-file" ]
  [ ! -e "$targetdir/sample-dir/ignored-file" ]
  [ ! -e "$targetdir/sample-dir/dir1/ignored-file" ]
}
