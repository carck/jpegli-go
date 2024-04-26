
#ifndef jepgli_h_
#define jepgli_h_

#include <stdint.h>
#include <stdlib.h>
#include <string.h>

int Decode(int fd, uint8_t *jpeg_in, int jpeg_in_size, int config_only, uint32_t *width, uint32_t *height, uint32_t *colorspace, uint32_t *chroma, uint8_t *out,
        int fancy_upsampling, int block_smoothing, int arith_code, int dct_method, int tw, int th);

uint8_t* Encode(uint8_t *in, uint8_t *inU, uint8_t *inV, int width, int height, int colorspace, int chroma, size_t *size, int quality, int progressive_level, int optimize_coding,
        int adaptive_quantization, int standard_quant_tables, int fancy_downsampling, int dct_method);

#endif // jepgli_h_