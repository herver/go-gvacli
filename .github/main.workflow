workflow "Build" {
  on = "push"
  resolves = [
    " linux/amd64"
    " windows/amd64"
  ]
}

action "build linux/amd64" {
  uses = "actions-contrib/go@master"
  env = {
    GOOS = "linux"
    GOARCH = "amd64"
  }
  args = "build ./..."
}

actions "build windows/amd64" {
  env = {
    GOOS = "windows"
    GOARCH = "amd64"
  }
  args = "build ./..."
}

workflow "Release" {
  on = "release"
  resolves = [
    "release linux/amd64"
    "release windows/amd64"
  ]
}

action "release linux/amd64" {
  uses = "ngs/go-release.action@v1.0.1"
  env = {
    GOOS = "linux"
    GOARCH = "amd64"
  }
  secrets = ["GITHUB_TOKEN"]
}

action "release windows/amd64" {
  uses = "ngs/go-release.action@v1.0.1"
  env = {
    GOOS = "windows"
    GOARCH = "amd64"
  }
  secrets = ["GITHUB_TOKEN"]
}
