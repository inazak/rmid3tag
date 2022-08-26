# rmid3tag

This program deletes the id3 tag contained in the mp3 file.
By default the original file is left as a backup.
After deletion, you can set the tag of title and artist.
If the title and the artist are included in the file name, 
you can use `-g` option to set it automatically.


## Usage

```
Usage:
  rmid3tag [OPTION] FILENAME

  when the option is not set, only delete tags.

  OPTION:
    -c or -check  ... nothing is changed. dump id3tag info.
    -s or -set    ... set tag. must be used with -t and -a.
    -t or -title  ... title.  use -t="SONG NAME"
    -a or -artist ... artist. use -a="ARTIST NAME"
    -g or -guess  ... set tag. guess title and artist from filename.
                      file must be named as "Artist - Title.mp3".

```

## Example

This is an example of tag deletion.

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


This is an example of setting title and artist after erasing all tags.

```
$ rmid3tag.exe -s -t="TITLE" -a="ARTIST" sample.mp3
```


This is automatically guessing from the file name the same operation as above.

```
$ rmid3tag.exe -g "ARTIST - TITLE.mp3"
```


## Installation

windows binary is [here](https://github.com/inazak/rmid3tag/releases)

