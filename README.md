# rmid3tag

This program deletes the id3 tag contained in the mp3 file.
By default the original file is left as a backup.

## Usage

```
  rmid3tag [OPTION] FILENAME

  OPTION:
    -c or -check  ... nothing is changed. dump id3tag info.
```

## Example

like this

```
$ rmid3tag.exe -check sample.mp3
[rmid3tag] V1Tag Exist       = false
[rmid3tag] V2Tag Exist       = true
[rmid3tag] File Size         = 12685181
[rmid3tag] MPEG Frame Offset = 2209
[rmid3tag] MPEG Frame Size   = 12682972

$ rmid3tag.exe sample.mp3

$ dir /b sample.mp3*
sample.mp3
sample.mp3.backup

$ rmid3tag.exe -check sample.mp3
[rmid3tag] V1Tag Exist       = false
[rmid3tag] V2Tag Exist       = false
[rmid3tag] File Size         = 12682972
[rmid3tag] MPEG Frame Offset = 0
[rmid3tag] MPEG Frame Size   = 12682972

$ 
```

