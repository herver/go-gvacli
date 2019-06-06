workflow "Build" {
  on = "push"
  resolved = ["build"]
}

action "build" {
  uses = "sosedoff/actions/golang-build@master"
  args = "linux/amd64 windows/386 windows/amd64"
}
