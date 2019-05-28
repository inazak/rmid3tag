package main

import (
  "flag"
  "fmt"
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

  OPTION:
    -c or -check  ... nothing is changed. dump id3tag info.
`

var optionsCheck bool

func main() {

  // options parse
  flag.BoolVar(&optionsCheck, "check", false, "nothing is changed. dump id3tag info.")
  flag.BoolVar(&optionsCheck, "c",     false, "nothing is changed. dump id3tag info.")
  flag.Parse()

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

  tempname, err := CopyMPEGFrame(filename, stat.OffsetMPEGFrame, stat.SizeOfMPEGFrame())
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

func CopyMPEGFrame(filename string, offset, size int64) (tempfile string, err error) {

  // open original file
  f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
  if err != nil {
    return "", err
  }
  defer f.Close()

  // create temporary file in the same location as Windows can not 
  // rename between different drives
  tf, err := ioutil.TempFile(filepath.Dir(filepath.Clean(filename)), "tmp_")
  if err != nil {
    return "", err
  }
  defer tf.Close()

  tempfile = tf.Name()

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

func errorExit(f string, v ...interface{}) {
  fmt.Fprintf(os.Stderr, f, v...)
  os.Exit(1)
}

