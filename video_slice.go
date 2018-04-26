//+build support_video

package pic2ascii

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"
	"time"

	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/cgo/ffmpeg"
	"github.com/nareix/joy4/format"
)

var rc io.Reader
var iotoken = "mem://video"

// 兼容 github.com/nareix/joy4 一个强制断言的 bug
type noper struct {
	io.ReadSeeker
}

func (noper) Close() error { return nil }

func init() {
	avutil.DefaultHandlers.Add(func(r *avutil.RegisterHandler) {
		r.UrlReader = func(f string) (bool, io.ReadCloser, error) {
			if f != iotoken {
				return false, nil, nil
			}
			data, err := ioutil.ReadAll(rc)
			if err != nil {
				return false, nil, err
			}

			if rc0 := rc; rc0 != nil {
				rc = nil
				return true, noper{bytes.NewReader(data)}, nil
			}
			return false, nil, nil
		}
	})
	format.RegisterAll()
}

func VideoSlice(read io.Reader, f func(time.Duration, image.Image)) error {
	rc = read
	file, err := avutil.Open(iotoken)
	if err != nil {
		return err
	}
	defer file.Close()

	streams, err := file.Streams()
	if err != nil {
		return err
	}

	var vd *ffmpeg.VideoDecoder
	var inx int
	for ind, s := range streams {
		if s.Type().IsVideo() {
			vd0, err := ffmpeg.NewVideoDecoder(s)
			if err != nil {
				return err
			}
			vd = vd0
			inx = ind
			break
		}
	}

	//	buf := bytes.NewBuffer(nil)
	for {
		pkt, err := file.ReadPacket()
		if err != nil {
			if io.EOF == err {
				break
			}
			return err
		}

		if pkt.Idx != int8(inx) {
			continue
		}

		vf, err := vd.Decode(pkt.Data)
		if err != nil {
			return err
		}

		f(pkt.Time, &vf.Image)
	}
	return nil
}