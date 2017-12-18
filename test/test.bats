#!/usr/bin/env bats

setup() {
  targetdir=$(mktemp -d)
}

teardown() {
    rm -rf $targetdir
}

@test "lndir creates links to original files" {
  run go-lndir $PWD/sample-dir $targetdir
  [ "$status" -eq 0 ]
  [ "$(readlink -n $targetdir/ignored-file)" == "$PWD/sample-dir/ignored-file" ]
  [ "$(readlink -n $targetdir/included-file)" == "$PWD/sample-dir/included-file" ]
  [ "$(readlink -n $targetdir/dir1/ignored-file)" == "$PWD/sample-dir/dir1/ignored-file" ]
  [ "$(readlink -n $targetdir/dir1/included-file)" == "$PWD/sample-dir/dir1/included-file" ]
  linked_files=$(cd $targetdir/dir1 && echo .* *)
  echo "Linked files: $linked_files"
  [ "$linked_files" == ". .. .gitignore ignored-file ignored-nested-file included-file relative-link" ]
}

@test "lndir creates relative links given a relative path to the source" {
  cp -r $PWD/sample-dir $targetdir/sample-dir
  mkdir -p $targetdir/relative-test-dir
  cd $targetdir && run go-lndir ../sample-dir $targetdir/relative-test-dir
  [ "$status" -eq 0 ]
  [ "$(readlink -n $targetdir/relative-test-dir/included-file)" == "../sample-dir/included-file" ]
  [ "$(readlink -n $targetdir/relative-test-dir/dir1/included-file)" == "../../sample-dir/dir1/included-file" ]
  linked_files=$(cd $targetdir/relative-test-dir/dir1 && echo .* *)
  echo "Linked files: $linked_files"
  [ "$linked_files" == ". .. .gitignore ignored-file ignored-nested-file included-file relative-link" ]
}

@test "lndir -gitignore ignores files specified in .gitignore files" {
  run go-lndir -gitignore $PWD/sample-dir $targetdir
  [ "$status" -eq 0 ]
  [ "$(readlink -n $targetdir/included-file)" == "$PWD/sample-dir/included-file" ]
  [ "$(readlink -n $targetdir/dir1/included-file)" == "$PWD/sample-dir/dir1/included-file" ]
  [ ! -e "$targetdir/sample-dir/ignored-file" ]
  [ ! -e "$targetdir/sample-dir/dir1/ignored-file" ]
  [ ! -e "$targetdir/sample-dir/dir1/ignored-nested-file" ]
  linked_files=$(cd $targetdir/dir1 && echo .* *)
  echo "Linked files: $linked_files"
  [ "$linked_files" == ". .. .gitignore included-file relative-link" ]
}

@test "lndir ignores revinfo files" {
  run go-lndir $PWD/sample-dir $targetdir
  [ "$status" -eq 0 ]
  linked_files=$(cd $targetdir/revinfo-files && echo .* *)
  echo "Linked files: $linked_files"
  [ "$linked_files" == ". .. *" ]
}

@test "lndir can include revinfo files" {
  run go-lndir -withrevinfo $PWD/sample-dir $targetdir
  [ "$status" -eq 0 ]
  linked_files=$(cd $targetdir/revinfo-files && echo .* *)
  echo "Linked files: $linked_files"
  [ "$linked_files" == ". .. .git .hg .svn BigKeeper CVS CVS.adm RCS SCCS" ]
}
