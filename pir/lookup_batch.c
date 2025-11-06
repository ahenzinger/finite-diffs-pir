// lookup_batch.c

#include "lookup_batch.h"
#include <immintrin.h>
#include <stdint.h>
#include <stddef.h>
#include <string.h>

void lookupBatch(uint64_t * restrict idx, 
                 uint8_t * restrict data, 
                 uint8_t * restrict ans, 
		 		 uint64_t n,
		 		 uint64_t record_len,
		 		 uint64_t query) {

    uint64_t cloudPoint0, cloudPoint1, cloudPoint2, cloudPoint3;
    uint64_t cloudPoint4, cloudPoint5, cloudPoint6, cloudPoint7;
    uint8_t *src0, *src1, *src2, *src3;
    uint8_t *src4, *src5, *src6, *src7;
    uint8_t *dst0, *dst1, *dst2, *dst3;
    uint8_t *dst4, *dst5, *dst6, *dst7;

    size_t i = 0;

    for ( ; i + 7 < n; i += 8) {
		cloudPoint0 = query ^ idx[i+0];
		cloudPoint1 = query ^ idx[i+1];
		cloudPoint2 = query ^ idx[i+2];
		cloudPoint3 = query ^ idx[i+3];
		cloudPoint4 = query ^ idx[i+4];
		cloudPoint5 = query ^ idx[i+5];
		cloudPoint6 = query ^ idx[i+6];
		cloudPoint7 = query ^ idx[i+7];

		src0 = data + cloudPoint0 * record_len;
		src1 = data + cloudPoint1 * record_len;
		src2 = data + cloudPoint2 * record_len;
		src3 = data + cloudPoint3 * record_len;
		src4 = data + cloudPoint4 * record_len;
		src5 = data + cloudPoint5 * record_len;
		src6 = data + cloudPoint6 * record_len;
		src7 = data + cloudPoint7 * record_len;

		dst0 = ans + (i+0) * record_len;
		dst1 = ans + (i+1) * record_len;
		dst2 = ans + (i+2) * record_len;
		dst3 = ans + (i+3) * record_len;
		dst4 = ans + (i+4) * record_len;
		dst5 = ans + (i+5) * record_len;
		dst6 = ans + (i+6) * record_len;
		dst7 = ans + (i+7) * record_len;

        memcpy(dst0, src0, record_len);
        memcpy(dst1, src1, record_len);
        memcpy(dst2, src2, record_len);
        memcpy(dst3, src3, record_len);
        memcpy(dst4, src4, record_len);
        memcpy(dst5, src5, record_len);
        memcpy(dst6, src6, record_len);
        memcpy(dst7, src7, record_len);
    }

    for (; i < n; i++) {
		cloudPoint0 = query ^ idx[i];
		src0 = data + cloudPoint0 * record_len;
		dst0 = ans + i * record_len;
		memcpy(dst0, src0, record_len);
    }
}

// Wrappers that bake in the constant
void lookupBatch_len1(uint64_t*q, uint8_t*d,uint8_t*a,uint64_t n,uint64_t x)
{ lookupBatch(q,d,a,n,1,x);  }

void lookupBatch_len10(uint64_t*q, uint8_t*d,uint8_t*a,uint64_t n,uint64_t x)
{ lookupBatch(q,d,a,n,10,x); }

void lookupBatch_len64(uint64_t*q, uint8_t*d,uint8_t*a,uint64_t n,uint64_t x)
{ lookupBatch(q,d,a,n,64,x); }

void lookupBatch_len100(uint64_t*q, uint8_t*d,uint8_t*a,uint64_t n,uint64_t x)
{ lookupBatch(q,d,a,n,100,x); }

void lookupBatch_len1024(uint64_t*q,uint8_t*d,uint8_t*a,uint64_t n,uint64_t x)
{ lookupBatch(q,d,a,n,1024,x); }

void lookupBatch_len102400(uint64_t*q,uint8_t*d,uint8_t*a,uint64_t n,uint64_t x)
{ lookupBatch(q,d,a,n,102400,x); }
