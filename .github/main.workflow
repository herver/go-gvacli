workflow "Build" {
  on = "push"
  resolves = ["build"]
}

action "build" {
  uses = "actions-contrib/go@master"
  args = "build ./..."
}
