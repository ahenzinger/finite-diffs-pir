#include <stdint.h>
#include <stddef.h>

void lookupBatch(uint64_t *idx,
                 uint8_t *table,
                 uint8_t *out,
                 uint64_t n,
                 uint64_t record_len,
                 uint64_t query);

void lookupBatch_len1(uint64_t *idx, uint8_t *table, uint8_t *out, uint64_t n, uint64_t query);
void lookupBatch_len10(uint64_t *idx, uint8_t *table, uint8_t *out, uint64_t n, uint64_t query);
void lookupBatch_len64(uint64_t *idx, uint8_t *table, uint8_t *out, uint64_t n, uint64_t query);
void lookupBatch_len100(uint64_t *idx, uint8_t *table, uint8_t *out, uint64_t n, uint64_t query);
void lookupBatch_len1024(uint64_t *idx, uint8_t *table, uint8_t *out, uint64_t n, uint64_t query);
void lookupBatch_len102400(uint64_t *idx, uint8_t *table, uint8_t *out, uint64_t n, uint64_t query);
