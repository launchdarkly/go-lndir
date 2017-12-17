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
  [ "$(readlink $targetdir/file1)" == "$PWD/sample-dir/file1" ]
  [ "$(readlink $targetdir/fileA)" == "$PWD/sample-dir/fileA" ]
  [ "$(readlink $targetdir/dir1/file1)" == "$PWD/sample-dir/dir1/file1" ]
  [ "$(readlink $targetdir/dir1/fileA)" == "$PWD/sample-dir/dir1/fileA" ]
}

@test "lndir -gitignore" {
  run go-lndir -gitignore $PWD/sample-dir $targetdir
  [ "$status" -eq 0 ]
  [ "$(readlink $targetdir/fileA)" == "$PWD/sample-dir/fileA" ]
  [ "$(readlink $targetdir/dir1/fileA)" == "$PWD/sample-dir/dir1/fileA" ]
  [ ! -e "$targetdir/sample-dir/file1" ]
  [ ! -e "$targetdir/sample-dir/dir1/file1" ]
}
