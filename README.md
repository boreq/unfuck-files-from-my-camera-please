# Unfuck Files From My Camera Please

This program renames the files produced by my camera so that each one is unique.
This allows me to put them all in the same directory when copying them to my
server.

As an example a file called `DSC_1234.EXT` would be renamed to something along
the lines of `2024-12-06 14:41:23.ext` or the same with a suffix e.g.
`2024-12-06 14:41:23 00005.ext` if there is a conflict in the timestamps.

I guess this could actually be useful to someone who has an old Nikon camera.
See [Compatiblity](#compatibility), it's easy to add new ones.

## Installation

    $ go install github.com/boreq/unfuck-files-from-my-camera-please@latest

## Dependencies 

Needs `mediainfo` and `exiv2` as it just calls those.

### Arch Linux

Install [`mediainfo`](https://archlinux.org/packages/extra/x86_64/mediainfo/)
and [`exiv2`](https://archlinux.org/packages/extra/x86_64/exiv2/).


## Problem

My camera produces files similar to `DSC_0001.MOV`, `DSC_0001.NEF` or
`DSC_0001.NEF`. This is annoying as it's impossible to keep dumping all files
into the same directory due to filename will conflicts. Doing that however makes
a ton of sense to me if I took the photos or videos in several batches
throughout the day and dumped them gradually as I keep photographing the same
thing. 

Another unrelated question: why are we using all caps in the extension names?

## Solution

This program scans the target directory for all supported file types, presumably
files that came from a digital camera, and extracts timestamps from them.  Once
that's done it prepares a plan for renaming the files and presents it to the
user. Once approved all files are renamed.

What if the file names conflict because there isn't enough granularity in the
timestamps? In that case the files will indeed get a suffix. The program tries
to keep the filenames lexicographically sorted in the same order in which they
came from the camera.

The work flow could be as follows:
- Copy all files from the SD card into some directory which may already contain
other files.
- Run the program on that directory to rename them.

My camera produces incomplete timestamps as it doesn't save the timezone in the
EXIF data so that's a problem.

The program renames the files to the time and date in UTC. Reasoning:
- from what I'm seeing the timestamps in the files are garbage most of the time
  and either contain timestamps in UTC or are without time zone information so
  technically we are not erasing any data
- if you travel a lot I feel like that's actually less confusing

If there is no zime zone info in the file then the program just guesses the
local time to be the local time zone. Reasoning:
- I never change the time and date on my camera and it's set to where I live
- I almost exclusively dump files on my desktop at home so it works 99% of the
  time for me

## Compatibility

If the camera/format combo is here then the files from it it are supported, if
not I can easily add support for a camera if you send me an example file that
you would like to rename. 

Time zone info:
- `present` means that the time zone info is present and the timestamp can be parsed correctly
- `missing` means that the time zone info is missing and the timestamp will be interpreted as being in the local time zone

| Camera | File format | Time zone info | 
| --- | --- | --- |
| Nikon D5300 | `nef` | missing |
| Nikon D5300 | `mov` | present |
| Nikon D5300 | `jpg` | missing |


## Design goals

Make this program as painless to run as possible. No configs, no flags. Just
run `unfuck-files-from-my-camera-please /path/to/directory` or run it without
the arguments to process the current directory.

There should be no need to remember anything e.g. where the config is or do
anything e.g. adjust the config or set the flags. The program should just
unfuck the files files from my camera please. 

The only potentially allowed flag is to override the user's confirmation.

The format `YYYY-MM-DD HH:MM:SS` is the only valid output timestamp format.
