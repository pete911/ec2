# ec2
simple ci to create and delete aws ec2 instance

## usage
- `ec2 create <name>` (name has to be unique per region)
- `ec2 delete`
 
## build/install

You can either build from source, or install. Pick any of the options:
- build `go build` (or use `Taskfile.yml` - `task` command)
- [download binary](https://github.com/pete911/ec2/releases)
- install via brew
    - add tap `brew tap pete911/tap`
    - install `brew install ec2`
 
## releases
Releases are published when the new tag is created e.g. `git tag -m "add some feature" v0.0.1 && git push --follow-tags`
