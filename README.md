# Unfuck Files From My Camera Please

This program unfucks the filenames of files produced by my camera so that each
file gets a unique filename.

As an example a file called `DSC_1234.EXT` could be renamed to `2024-12-06
14:41:23 UTC.EXT` or the same with a suffix e.g. `2024-12-06 14:41:23
00005.EXT` if there is a conflict in the timestamps.

## Problem

My camera produces files with names similar to `DSC_0001.MOV` or
`DSC_0001.NEF`. This is annoying. For example it's impossible to keep dumping
all files into the same directory because the names will conflict. This however
makes a ton of sense to me if I took the photos or videos in several batches
throughout the day and dumped them gradually as I keep photographing the same
thing. 

Also why are we using all caps in the extension name but that's unrelated.

## Solution

This program scans the target directory for all supported file types,
presumably files that came from a digital camera. It then attempts to extract
the timestamp from them. Once that's done it will prepare the plan for renaming
the files and present it to the user. Once approved all files are renamed.

What if the file names conflict because there isn't enough granularity in the
timestamps contained within the files? In that case the files will get a suffix
which does indeed go up consistently because that's when adding one makes sense
to me as we need to differentiate the files somehow and also keep them ordered.
The suffixes will be generated based on the original sequence numbers so that
the files are still in order and the relative position will be kept. The
filenames should still be easy enough to parse if someone wants to later.

The work flow should be as follows:
- Copy all files from the SD card into the target directory which may already
  contain other files.
- Run the program on that directory to unfuck them.

My camera produces incomoplete timestamps as it doesn't save the timezone in
the EXIF data so that's a problem.

The program renames the files to the time and date in UTC. Reasoning:
- from what I'm seeing the timestamps in the files are garbage most of the time
  and either contain timestamps in UTC or without time zone information so
  technically we are not erasing any data
- if you travel a lot I feel like that's actually less confusing

If there is no zime zone info in the file then the program just guesses the
local time to be the local time zone. Reasoning:
- I never change the time and date on my camera and it's set to where I live
- I almost exclusively dump files on my desktop at home so it works 99% of the
  time for me

## Design goals

Make this program as painless to run as possible. No configs, no flags. Just
run `unfuck-files-from-my-camera-please /path/to/directory` and the files in
that directory are unfucked.
pl
I don't want to remember anything e.g. where the config is, I don't want to do
anything e.g. adjust the config or set the flags. I just want this program to
unfuck the files files from my camera please. I don't even want to be able to
configure anything.

The only allowed flag is to override the user confirmation.

The format `YYYY-MM-DD HH:MM:SS` is the only allowed timestamp format.
