package rmid3tag

import (
  "fmt"
  "os"
  "io"
  "bytes"
  "golang.org/x/text/transform"
  "golang.org/x/text/encoding/unicode"
)

// Stat structure has MPEG frame offset and size.
// +---------------+
// |  ID3v2tag     |
// |  (optional)   |
// +---------------+ <-- offset
// |               |   |
// |  MPEG Frames  |   | size of mpeg frame
// |               |   |
// +---------------+ <-+ 
// |  ID3v1tag     |
// |  (optional)   |
// +---------------+
//
type Stat struct {
  Size            int64
  V1TagExist      bool
  V2TagExist      bool
  OffsetMPEGFrame int64
}

func (s *Stat) SizeOfMPEGFrame() int64 {
  if s.V1TagExist {
    return s.Size - s.OffsetMPEGFrame - 128
  }
  return s.Size - s.OffsetMPEGFrame
}

func GetStat(filename string) (stat *Stat, err error) {

  stat = &Stat{}

  // open original file
  f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
  if err != nil {
    return stat, err
  }
  defer f.Close()

  // filesize
  info, err := f.Stat()
  if err != nil {
    return stat, err
  }
  stat.Size = info.Size()

  // v2tag
  stat.V2TagExist, stat.OffsetMPEGFrame, err = getID3v2TagSize(f)
  if err != nil {
    return stat, err
  }

  // v1tag
  stat.V1TagExist, err = isExistID3v1Tag(f, stat.Size -128)
  if err != nil {
    return stat, err
  }

  return stat, nil
}

func getID3v2TagSize(r io.ReaderAt) (isExist bool, size int64, err error) {

  marker := make([]byte, 4)
  n, err := r.ReadAt(marker, 0)
  if n != 4 || err != nil {
    return false, 0, err
  }

  var offset int64 = 0 //header start position

  if string(marker[:3]) != "ID3" {

     // some mp3 file has irregular byte at the beginning,
     // so ignore them.
     if string(marker[1:]) == "ID3" {
       offset += 1
     } else {
       // then marker not found
       return false, 0, nil
     }
  }

  isExist = true

  data := make([]byte, 4)
  n, err = r.ReadAt(data, 6 + offset)
  if n != 4 || err != nil {
    return isExist, 0, err
  }

  size = int64(decodeSynchsafe(data, 4))
  size += 10 + offset //add v2 header size

  // Some files have padding greater than the specified size,
  // so search until the mpeg frame is found.
  for ;; size++ {
    ok, err := isExistMP3Frame(r, size)

    if err != nil {
      return isExist, size, fmt.Errorf("mpeg frame not found")
    }

    if ok { // found
      return isExist, size, nil
    }
  }

  return isExist, size, nil // do not reach here
}

func isExistID3v1Tag(r io.ReaderAt, offset int64) (bool, error) {

  marker := make([]byte, 3)
  n, err := r.ReadAt(marker, offset)
  if n != 3 || err != nil {
    return false, err
  }

  if string(marker) != "TAG" {
    return false, nil
  }

  return true, nil
}

func isExistMP3Frame(r io.ReaderAt, offset int64) (bool, error) {

  data := make([]byte, 2)
  n, err := r.ReadAt(data, offset)
  if n != 2 || err != nil {
    return false, err
  }

  // beginning byte pattern of mpeg frame
  pattern := [][]byte{
    {0xff,0xfb}, {0xff,0xfa},
  }

  for _, p := range pattern {
    if bytes.HasPrefix(data, p) {
      return true, nil
    }
  }

  return false, nil
}

// utilites

func decodeSynchsafe(data []byte, size int) int {

  result := 0

  for i:=0; i<size; i++ {
    result += (int(data[i]) & 0x7f) << uint(7 * (size-1-i))
  }

  return result
}

func encodeSynchsafe(data int, size int) []byte {

  result := make([]byte, size)

  for i:=0; i<size; i++ {
    result[i] = byte((data & 0x7f) >> uint(7 * (size-1-i)))
  }

  return result
}

// for create

func CreateMinimumTag(title, artist string) ([]byte, error) {

  tf, err := CreateTitleFrame(title)
  if err != nil {
    return []byte{}, err
  }

  af, err := CreateArtistFrame(artist)
  if err != nil {
    return []byte{}, err
  }

  return CreateID3V2Tag(tf, af), nil
}

func CreateID3V2Tag(frames ...[]byte) []byte {

  size := 0
  for _, frame := range frames {
    size += len(frame)
  }

  buf := &bytes.Buffer{}
  buf.WriteString("ID3")
  buf.Write([]byte{0x3,0x0,0x0}) //version 2.3
  buf.Write(encodeSynchsafe(size, 4))
  for _, frame := range frames {
    buf.Write(frame)
  }

  return buf.Bytes()
}

func CreateTitleFrame(text string) ([]byte, error) {
  return CreateTextFrame("TIT2", text)
}

func CreateArtistFrame(text string) ([]byte, error) {
  return CreateTextFrame("TPE1", text)
}

func CreateTextFrame(id, text string) ([]byte, error) {

  data, err := encodeTextFrameData(text)
  if err != nil {
    return []byte{}, err
  }
  size := len(data)

  buf := &bytes.Buffer{}
  buf.WriteString(id)
  buf.WriteByte(byte(0xff&(size>>24)))
  buf.WriteByte(byte(0xff&(size>>16)))
  buf.WriteByte(byte(0xff&(size>>8)))
  buf.WriteByte(byte(0xff&size))
  buf.Write([]byte{0x0,0x0})
  buf.Write(data)

  return buf.Bytes(), nil
}

func encodeTextFrameData(s string) ([]byte, error) {

  u16bytes,err := toUTF16BEWithBOM(s)
  if err != nil {
    return []byte{}, err
  }

  buf := &bytes.Buffer{}
  buf.Write([]byte{0x1}) //encoding UTF16/useBOM
  buf.Write(u16bytes)
  buf.Write([]byte{0x0,0x0}) //null terminater

  return buf.Bytes(), nil
}

func toUTF16BEWithBOM(s string) ([]byte, error) {

  u16str, _, err := transform.String(
    unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewEncoder(), s)

  if err != nil {
    return []byte{}, err
  }

  return []byte(u16str), nil
}

