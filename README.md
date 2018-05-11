# mapfs-release

This is a bosh release that packages
[mapfs](https://github.com/cloudfoundry/mapfs) used by volume drivers to map
gid/uid of the process.

## Usage

Collocate mapfs job onto diego cell job via operations file
[add-mapfs.yml](operations/add-mapfs.yml). See [BOSH operations
file](https://bosh.io/docs/cli-ops-files/). Once deployed mapfs executable will
be available on VM at `/usr/bin/mapfs`.
