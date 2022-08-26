package main

import (
  "flag"
  "fmt"
  "strings"
  "os"
  "io"
  "io/ioutil"
  "path/filepath"
  "github.com/inazak/rmid3tag"
)

var usage =`
This program deletes the id3 tag contained in the mp3 file.
By default the original file is left as a backup.

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
`

var optionsCheck  bool
var optionsGuess  bool
var optionsSet    bool
var optionsTitle  string
var optionsArtist string

func main() {

  flag.BoolVar(&optionsCheck, "check", false, "nothing is changed. dump id3tag info.")
  flag.BoolVar(&optionsCheck, "c",     false, "nothing is changed. dump id3tag info.")
  flag.BoolVar(&optionsGuess, "guess", false, "set tag. guess title and artist from filename.")
  flag.BoolVar(&optionsGuess, "g",     false, "set tag. guess title and artist from filename.")
  flag.BoolVar(&optionsSet,   "set",   false, "set title and artist with -t and -a")
  flag.BoolVar(&optionsSet,   "s",     false, "set title and artist with -t and -a")
  flag.StringVar(&optionsTitle,  "title",  "", "title")
  flag.StringVar(&optionsTitle,  "t",      "", "title")
  flag.StringVar(&optionsArtist, "artist", "", "artist")
  flag.StringVar(&optionsArtist, "a",      "", "artist")
  flag.Parse()

  if optionsSet {
    if optionsTitle == "" || optionsArtist == "" || optionsGuess {
      fmt.Printf("%s", usage)
      os.Exit(1)
    }
  }

  if len(flag.Args()) != 1 {
    fmt.Printf("%s", usage)
    os.Exit(1)
  }
  filename := flag.Args()[0]

  // get mpeg frame infomation
  stat, err := rmid3tag.GetStat(filename)
  if err != nil {
    errorExit("%v", err)
  }

  if optionsCheck {
    fmt.Printf("[rmid3tag] V1Tag Exist       = %v\n", stat.V1TagExist)
    fmt.Printf("[rmid3tag] V2Tag Exist       = %v\n", stat.V2TagExist)
    fmt.Printf("[rmid3tag] File Size         = %v\n", stat.Size)
    fmt.Printf("[rmid3tag] MPEG Frame Offset = %v\n", stat.OffsetMPEGFrame)
    fmt.Printf("[rmid3tag] MPEG Frame Size   = %v\n", stat.SizeOfMPEGFrame())
    os.Exit(0)
  }

  if optionsGuess {
    _, name, _ := splitFilepath(filename)
    optionsTitle, optionsArtist = guessFromFilename(name)
    optionsSet = true
  }

  tag := []byte{}

  // create id3v2.3 tag
  if optionsSet {
    tag, err = rmid3tag.CreateMinimumTag(optionsTitle, optionsArtist)
    if err != nil {
      errorExit("%v", err)
    }
  }

  // copy mpeg fraem from original file with minimum tags.
  tempname, err := Copy(filename, stat.OffsetMPEGFrame, stat.SizeOfMPEGFrame(), tag)
  if err != nil {
    errorExit("%v", err)
  }

  // rename original file to xxx.backup
  err = os.Rename(filename, filename + ".backup")
  if err != nil {
    _ = os.Remove(tempname)
    errorExit("%v", err)
  }

  // rename temp file to original filename
  err = os.Rename(tempname, filename)
  if err != nil {
    _ = os.Remove(tempname)
    errorExit("%v", err)
  }

  os.Exit(0)
}

// Copy mpeg frame from the original file. If a tag is set, add it at the beginning.
func Copy(filename string, offset, size int64, tag []byte) (tempfile string, err error) {

  // open original file
  f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
  if err != nil {
    return "", err
  }
  defer f.Close()

  // create temporary file in the same location for os.Rename.
  tf, err := ioutil.TempFile(filepath.Dir(filepath.Clean(filename)), "tmp_")
  if err != nil {
    return "", err
  }
  defer tf.Close()

  tempfile = tf.Name()

  // write tag
  if len(tag) > 0 {
    _, err = tf.Write(tag)
    if err != nil {
      return "", err
    }
  }

  // seek mpeg frame head. 0 means relative to the origin of the file
  _, err = f.Seek(offset, 0)
  if err != nil {
    return tempfile, err
  }

  // copy to new file
  n, err := io.CopyN(tf, f, size)
  if n != size || err != nil {
    return tempfile, err
  }

  return tempfile, nil
}

func guessFromFilename(filename string) (title, artist string) {

  index := strings.Index(filename, " - ")
  if index == -1 || index == 0 || (index+3) == len(filename) {
    return "-", "-"
  }

  artist = filename[0:index]
  title  = filename[index+3:len(filename)]

  return title, artist
}

func splitFilepath(fpath string) (dir, name, ext string) {

  dir   = filepath.Dir(fpath)
  base := filepath.Base(fpath)

  dotindex := strings.LastIndex(base, ".")

  // file has no extention
  if dotindex == -1 {
    name = base
    ext  = ""
    return
  }

  name = base[0:dotindex]
  ext  = base[dotindex:len(base)]
  return
}

func errorExit(f string, v ...interface{}) {
  fmt.Fprintf(os.Stderr, f, v...)
  os.Exit(1)
}

