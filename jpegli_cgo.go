package jpegli

/*
#cgo LDFLAGS: -ljpegli
#cgo CFLAGS: -O3
#include "jpegli.h"
*/
import "C"
import (
	"fmt"
	"image"
	"image/color"
	"io"
	"unsafe"
)

const (
	jcsGrayscale = iota + 1
	jcsRGB
	jcsYCbCr
	jcsCMYK
	jcsYCCK
)

func decode(r io.Reader, configOnly, fancyUpsampling, blockSmoothing, arithCode bool, dctMethod DCTMethod, tw, th int) (image.Image, image.Config, error) {

	var err error
	var cfg image.Config
	var data []byte

	if configOnly {
		data = make([]byte, 1024)
		_, err = r.Read(data)
		if err != nil {
			return nil, cfg, fmt.Errorf("read: %w", err)
		}
	} else {
		data, err = io.ReadAll(r)
		if err != nil {
			return nil, cfg, fmt.Errorf("read: %w", err)
		}
	}

	inSize := len(data)

	var widthPtr C.uint32_t
	var heightPtr C.uint32_t
	var colorspacePtr C.uint32_t
	var chromaPtr C.uint32_t

	fancyUpsamplingVal := 0
	if fancyUpsampling {
		fancyUpsamplingVal = 1
	}

	blockSmoothingVal := 0
	if blockSmoothing {
		blockSmoothingVal = 1
	}

	arithCodeVal := 0
	if arithCode {
		arithCodeVal = 1
	}

	res := C.Decode((*C.uint8_t)(unsafe.Pointer(&data[0])), C.int(inSize), C.int(1), &widthPtr, &heightPtr, &colorspacePtr, &chromaPtr, nil,
		C.int(fancyUpsamplingVal), C.int(blockSmoothingVal), C.int(arithCodeVal), C.int(dctMethod), C.int(tw), C.int(th))

	if res == 0 {
		return nil, cfg, ErrDecode
	}

	width := int(widthPtr)

	height := int(heightPtr)

	colorspace := uint32(colorspacePtr)

	chroma := uint32(chromaPtr)

	cfg.Width = int(width)
	cfg.Height = int(height)

	var size, w, h, cw, ch, i0, i1, i2 int
	switch colorspace {
	case jcsGrayscale:
		cfg.ColorModel = color.GrayModel
		size = cfg.Width * cfg.Height * 1
	case jcsRGB:
		cfg.ColorModel = color.RGBAModel
		size = cfg.Width * cfg.Height * 4
	case jcsYCbCr:
		cfg.ColorModel = color.YCbCrModel
		w, h, cw, ch = yCbCrSize(image.Rect(0, 0, cfg.Width, cfg.Height), image.YCbCrSubsampleRatio(chroma))
		i0 = w*h + 0*cw*ch
		i1 = w*h + 1*cw*ch
		i2 = w*h + 2*cw*ch
		size = i2
	case jcsCMYK, jcsYCCK:
		cfg.ColorModel = color.CMYKModel
		size = cfg.Width * cfg.Height * 4
	default:
		return nil, cfg, fmt.Errorf("unsupported colorspace %d", colorspace)
	}

	if configOnly {
		return nil, cfg, nil
	}

	out := make([]byte, size)

	res = C.Decode((*C.uint8_t)(unsafe.Pointer(&data[0])), C.int(inSize), C.int(0), &widthPtr, &heightPtr, &colorspacePtr, &chromaPtr, (*C.uint8_t)(unsafe.Pointer(&out[0])),
		C.int(fancyUpsamplingVal), C.int(blockSmoothingVal), C.int(arithCodeVal), C.int(dctMethod), C.int(tw), C.int(th))

	if res == 0 {
		return nil, cfg, ErrDecode
	}

	var img image.Image

	switch colorspace {
	case jcsGrayscale:
		i := image.NewGray(image.Rect(0, 0, cfg.Width, cfg.Height))
		i.Pix = out
		img = i
	case jcsRGB:
		i := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
		i.Pix = out
		img = i
	case jcsYCbCr:
		img = &image.YCbCr{
			Y:              out[:i0:i0],
			Cb:             out[i0:i1:i1],
			Cr:             out[i1:i2:i2],
			SubsampleRatio: image.YCbCrSubsampleRatio(chroma),
			YStride:        w,
			CStride:        cw,
			Rect:           image.Rect(0, 0, cfg.Width, cfg.Height),
		}
	case jcsCMYK, jcsYCCK:
		i := image.NewCMYK(image.Rect(0, 0, cfg.Width, cfg.Height))
		i.Pix = out
		img = i
	default:
		return nil, cfg, fmt.Errorf("unsupported colorspace %d", colorspace)
	}

	return img, cfg, nil
}

func encode(w io.Writer, m image.Image, quality, chromaSubsampling, progressiveLevel int, optimizeCoding, adaptiveQuantization,
	standardQuantTables, fancyDownsampling bool, dctMethod DCTMethod) error {

	var data []byte
	var colorspace int
	var chroma int
	var inputSize int

	switch img := m.(type) {
	case *image.Gray:
		data = img.Pix
		colorspace = jcsGrayscale
	case *image.RGBA:
		data = img.Pix
		colorspace = jcsRGB
		chroma = chromaSubsampling
	case *image.NRGBA:
		data = img.Pix
		colorspace = jcsRGB
		chroma = chromaSubsampling
	case *image.CMYK:
		data = img.Pix
		colorspace = jcsCMYK
	case *image.YCbCr:
		colorspace = jcsYCbCr
		chroma = int(img.SubsampleRatio)
	default:
		i := imageToRGBA(img)
		data = i.Pix
		colorspace = jcsRGB
	}

	var in *C.uint8_t
	var inU *C.uint8_t
	var inV *C.uint8_t
	if colorspace == jcsYCbCr {
		yuv := m.(*image.YCbCr)
		in = (*C.uint8_t)(unsafe.Pointer(&yuv.Y[0]))
		inU = (*C.uint8_t)(unsafe.Pointer(&yuv.Cb[0]))
		inV = (*C.uint8_t)(unsafe.Pointer(&yuv.Cr[0]))
		inputSize = len(yuv.Y) + len(yuv.Cb) + len(yuv.Cr)
	} else {
		in = (*C.uint8_t)(unsafe.Pointer(&data[0]))
		inputSize = len(data)
	}

	fmt.Print(inputSize)

	var sizePtr C.size_t

	optimizeCodingVal := 0
	if optimizeCoding {
		optimizeCodingVal = 1
	}

	adaptiveQuantizationVal := 0
	if adaptiveQuantization {
		adaptiveQuantizationVal = 1
	}

	standardQuantTablesVal := 0
	if standardQuantTables {
		standardQuantTablesVal = 1
	}

	fancyDownsamplingVal := 0
	if fancyDownsampling {
		fancyDownsamplingVal = 1
	}

	res := C.Encode(in, inU, inV, C.int(m.Bounds().Dx()), C.int(m.Bounds().Dy()), C.int(colorspace), C.int(chroma), &sizePtr, C.int(quality),
		C.int(progressiveLevel), C.int(optimizeCodingVal), C.int(adaptiveQuantizationVal), C.int(standardQuantTablesVal),
		C.int(fancyDownsamplingVal), C.int(dctMethod))

	defer C.free(unsafe.Pointer((res)))

	size := int(sizePtr)

	if size == 0 {
		return ErrEncode
	}

	out := C.GoBytes(unsafe.Pointer(res), sizePtr)
	//cfs := out[:]
	_, err := w.Write(out)
	if err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}
