package rmid3tag

import (
  "fmt"
  "os"
  "io"
  "bytes"
  "golang.org/x/text/transform"
  "golang.org/x/text/encoding/unicode"
)

//// for delete

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

// GetStat() provide mp3 file imformation.
// Stat structure has MPEG frame offset and size.
// +---------------+
// |  ID3v2tag     |
// |  (optional)   |
// +---------------+ <-- offset
// |               |   |
// |  MPEG Frames  |   | size
// |               |   |
// +---------------+ <-+ 
// |  ID3v1tag     |
// |  (optional)   |
// +---------------+
//
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
  stat.V2TagExist, err = isExistID3v2Tag(f)
  if err != nil {
    return stat, err
  }

  if stat.V2TagExist {
    stat.OffsetMPEGFrame, err = getID3v2TagSize(f)
    if err != nil {
      return stat, err
    }
    stat.OffsetMPEGFrame += 10 // add v2tag header size
  }

  // v1tag
  stat.V1TagExist, err = isExistID3v1Tag(f, stat.Size -128)
  if err != nil {
    return stat, err
  }

  // check mpeg frame
  ok, err := isExistMP3Frame(f, stat.OffsetMPEGFrame)
  if err != nil {
    return stat, err
  }
  if ! ok {
    return stat, fmt.Errorf("mpeg frame not found")
  }

  return stat, nil
}

func isExistID3v2Tag(r io.ReaderAt) (bool, error) {

  data := make([]byte, 3)

  n, err := r.ReadAt(data, 0)
  if n != 3 || err != nil {
    return false, err
  }

  if string(data) != "ID3" {
    return false, nil
  }

  return true, nil
}

func getID3v2TagSize(r io.ReaderAt) (int64, error) {

  data := make([]byte, 4)

  n, err := r.ReadAt(data, 6)
  if n != 4 || err != nil {
    return 0, err
  }

  return int64(decodeSynchsafe(data, 4)), nil
}

func isExistID3v1Tag(r io.ReaderAt, offset int64) (bool, error) {

  data := make([]byte, 3)

  n, err := r.ReadAt(data, offset)
  if n != 3 || err != nil {
    return false, err
  }

  if string(data) != "TAG" {
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

  if ! bytes.HasPrefix(data, []byte{0xff,0xfb}) {
    return false, nil
  }

  return true, nil
}

//// utilites

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

//// for create

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

