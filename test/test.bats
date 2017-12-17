#!/usr/bin/env bats

@test "lndir" {
  tmpdir=$(mktemp -d)
  # trap "status=$?; rm -f $tmpdir; exit $status" EXIT
  run go-lndir $PWD/sample-dir $tmpdir
  [ "$status" -eq 0 ]
  echo "file1: $(readlink $tmpdir/file1)"
  [ "$(readlink $tmpdir/file1)" == "$PWD/sample-dir/file1" ]
  [ "$(readlink $tmpdir/fileA)" == "$PWD/sample-dir/fileA" ]
  [ "$(readlink $tmpdir/dir1/file1)" == "$PWD/sample-dir/dir1/file1" ]
  [ "$(readlink $tmpdir/dir1/fileA)" == "$PWD/sample-dir/dir1/fileA" ]
}

@test "lndir -gitignore" {
  tmpdir=$(mktemp -d)
  # trap "status=$?; rm -f $tmpdir; exit $status" EXIT
  run go-lndir -gitignore $PWD/sample-dir $tmpdir
  [ "$status" -eq 0 ]
  [ "$(readlink $tmpdir/fileA)" == "$PWD/sample-dir/fileA" ]
  [ "$(readlink $tmpdir/dir1/fileA)" == "$PWD/sample-dir/dir1/fileA" ]
  [ ! -e "$tmpdir/sample-dir/file1" ]
  [ ! -e "$tmpdir/sample-dir/dir1/file1" ]
}
