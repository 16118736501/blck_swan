# gdu - Go Disk Usage

<a href="https://repology.org/project/gdu/versions">
    <img src="https://repology.org/badge/vertical-allrepos/gdu.svg" alt="Packaging status" align="right">
</a>

[![Build Status](https://travis-ci.com/dundee/gdu.svg?branch=master)](https://travis-ci.com/dundee/gdu)
[![codecov](https://codecov.io/gh/dundee/gdu/branch/master/graph/badge.svg)](https://codecov.io/gh/dundee/gdu)
[![Go Report Card](https://goreportcard.com/badge/github.com/dundee/gdu)](https://goreportcard.com/report/github.com/dundee/gdu)

Pretty fast disk usage analyzer written in Go.

Gdu is intended primarily for SSD disks where it can fully utilize parallel processing.
However HDDs work as well, but the performance gain is not so huge.

[![asciicast](https://asciinema.org/a/382738.svg)](https://asciinema.org/a/382738)

## Installation

Head for the [releases](https://github.com/dundee/gdu/releases) and download binary for your system.

Using curl:

    curl -L https://github.com/dundee/gdu/releases/latest/download/gdu_linux_amd64.tgz | tar xz
    chmod +x gdu_linux_amd64
    mv gdu_linux_amd64 /usr/bin/gdu

[Arch Linux](https://aur.archlinux.org/packages/gdu/):

    yay -S gdu

Debian:

    dpkg -i gdu_*_all.deb

[NixOS](https://search.nixos.org/packages?channel=unstable&show=gdu&query=gdu):

    nix-env -iA nixos.gdu

[Homebrew](https://formulae.brew.sh/formula/gdu):

    brew install gdu

[Snap](https://snapcraft.io/gdu-disk-usage-analyzer):

    snap install gdu-disk-usage-analyzer
    snap alias gdu-disk-usage-analyzer.gdu gdu

[Go](https://pkg.go.dev/github.com/dundee/gdu):

    go get -u github.com/dundee/gdu


## Usage

```
  gdu [flags] [directory_to_scan]

Flags:
  -h, --help                  help for gdu
  -i, --ignore-dirs strings   Absolute paths to ignore (separated by comma) (default [/proc,/dev,/sys,/run])
  -l, --log-file string       Path to a logfile (default "/dev/null")
  -c, --no-color              Do not use colorized output
  -p, --no-progress           Do not show progress in non-interactive mode
  -n, --non-interactive       Do not run in interactive mode
  -d, --show-disks            Show all mounted disks
  -v, --version               Print version
```

## Examples

    gdu                                   # analyze current dir
    gdu <some_dir_to_analyze>             # analyze given dir
    gdu -d                                # show all mounted disks
    gdu -l ./gdu.log <some_dir>           # write errors to log file
    gdu -i /sys,/proc /                   # ignore some paths
    gdu -c /                              # use only white/gray/black colors

    gdu -n /                              # only print stats, do not start interactive mode
    gdu -np /                             # do not show progress, useful when using its output in a script
    gdu / > file                          # write stats to file, do not start interactive mode

Gdu has two modes: interactive (default) and non-interactive.

Non-interactive mode is started automtically when TTY is not detected (using [go-isatty](https://github.com/mattn/go-isatty)), for example if the output is being piped to a file, or it can be started explicitly by using a flag.

## Running tests

    make test


## Benchmark

Benchmarks performed on 50G directory (100k directories, 400k files) on 500 GB SSD using [hyperfine](https://github.com/sharkdp/hyperfine).
See `benchmark` target in [Makefile](Makefile) for more info.

### Cold cache

Filesystem cache was cleared using `sync; echo 3 | sudo tee /proc/sys/vm/drop_caches`.

| Command | Mean [s] | Min [s] | Max [s] | Relative |
|:---|---:|---:|---:|---:|
| `gdu -npc ~` | 3.634 ± 0.016 | 3.613 | 3.657 | 1.13 ± 0.02 |
| `dua ~` | 4.029 ± 0.051 | 3.993 | 4.160 | 1.25 ± 0.03 |
| `duc index ~` | 27.731 ± 0.436 | 27.128 | 28.283 | 8.61 ± 0.22 |
| `ncdu -0 -o /dev/null ~` | 27.238 ± 0.198 | 26.908 | 27.604 | 8.45 ± 0.18 |
| `diskus -b ~` | 3.222 ± 0.063 | 3.149 | 3.351 | 1.00 |
| `du -hs ~` | 25.966 ± 0.910 | 24.056 | 26.997 | 8.06 ± 0.32 |
| `dust -d0 ~` | 18.661 ± 1.629 | 14.461 | 19.672 | 5.79 ± 0.52 |

### Warm cache

| Command | Mean [ms] | Min [ms] | Max [ms] | Relative |
|:---|---:|---:|---:|---:|
| `gdu -npc ~` | 619.6 ± 28.0 | 574.0 | 657.7 | 2.57 ± 0.25 |
| `dua ~` | 344.0 ± 10.8 | 327.3 | 358.8 | 1.43 ± 0.13 |
| `duc index ~` | 1092.4 ± 11.3 | 1073.9 | 1103.6 | 4.54 ± 0.39 |
| `ncdu -0 -o /dev/null ~` | 1512.5 ± 18.4 | 1488.4 | 1546.5 | 6.28 ± 0.54 |
| `diskus -b ~` | 240.8 ± 20.6 | 207.8 | 289.1 | 1.00 |
| `du -hs ~` | 876.9 ± 19.5 | 843.9 | 910.3 | 3.64 ± 0.32 |
| `dust -d0 ~` | 7614.2 ± 45.6 | 7557.0 | 7687.5 | 31.61 ± 2.72 |



Gdu is inspired by [ncdu](https://dev.yorhel.nl/ncdu), [godu](https://github.com/viktomas/godu), [dua](https://github.com/Byron/dua-cli) and [df](https://www.gnu.org/software/coreutils/manual/html_node/df-invocation.html).