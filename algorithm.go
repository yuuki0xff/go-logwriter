package logwriter

import (
	"bytes"
	"compress/gzip"
	"github.com/klauspost/compress/zstd"
	"io"
	"sync"
)

type Algorithm interface {
	Compress(in []byte, out *bytes.Buffer) error
	Decompress(in []byte, out *bytes.Buffer) error
}

var _ Algorithm = &NopAlgorithm{}

type NopAlgorithm struct{}

func (n *NopAlgorithm) Compress(in []byte, out *bytes.Buffer) error {
	_, err := out.Write(in)
	return err
}

func (n *NopAlgorithm) Decompress(in []byte, out *bytes.Buffer) error {
	_, err := out.Write(in)
	return err
}

var _ Algorithm = &GzipAlgorithm{}

type GzipAlgorithm struct {
	once sync.Once
	gw   *gzip.Writer
}

func (g *GzipAlgorithm) Compress(in []byte, out *bytes.Buffer) error {
	g.once.Do(g.init)
	g.gw.Reset(out)
	_, err := g.gw.Write(in)
	if err != nil {
		return err
	}
	return g.gw.Close()
}

func (g *GzipAlgorithm) Decompress(in []byte, out *bytes.Buffer) error {
	r, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(out, r)
	return err
}

func (g *GzipAlgorithm) init() {
	g.gw = gzip.NewWriter(Discard)
}

var _ Algorithm = &ZstdAlgorithm{}

type ZstdAlgorithm struct {
	once sync.Once
	zw   *zstd.Encoder
	err  error
}

func (z *ZstdAlgorithm) Compress(in []byte, out *bytes.Buffer) error {
	z.once.Do(z.init)
	if z.err != nil {
		return z.err
	}
	z.zw.Reset(out)
	_, err := z.zw.Write(in)
	if err != nil {
		return err
	}
	return z.zw.Close()
}

func (z *ZstdAlgorithm) Decompress(in []byte, out *bytes.Buffer) error {
	r, err := zstd.NewReader(bytes.NewReader(in))
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = io.Copy(out, r)
	return err
}

func (z *ZstdAlgorithm) init() {
	z.zw, z.err = zstd.NewWriter(Discard, zstd.WithEncoderConcurrency(1))
}
