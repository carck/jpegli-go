package main

import (
	"image/jpeg"
	"os"

	"github.com/carck/jpegli"
)

func main() {
	file, err := os.Open("./abc.jpg")
	if err != nil {
		panic(err)
	}
	img, err := jpegli.Decode(file)
	if err != nil {
		panic(err)
	}

	out2, err := os.OpenFile("456.jpg", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = jpeg.Encode(out2, img, &jpeg.Options{Quality: 90})
	if err != nil {
		panic(err)
	}

	out, err := os.OpenFile("123.jpg", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}

	err = jpegli.Encode(out, img)
	if err != nil {
		panic(err)
	}

	err = jpeg.Encode(out2, img, &jpeg.Options{Quality: 90})
	if err != nil {
		panic(err)
	}

}
