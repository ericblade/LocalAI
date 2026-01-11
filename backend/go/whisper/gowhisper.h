#include <cstddef>
#include <cstdint>

#ifdef _WIN32
#define GOWHISPER_EXPORT extern "C" __declspec(dllexport)
#else
#define GOWHISPER_EXPORT
#endif

extern "C" {
GOWHISPER_EXPORT
int load_model(const char *const model_path);
GOWHISPER_EXPORT
int load_model_vad(const char *const model_path);
GOWHISPER_EXPORT
int vad(float pcmf32[], size_t pcmf32_size, float **segs_out,
        size_t *segs_out_len);
GOWHISPER_EXPORT
int transcribe(uint32_t threads, char *lang, bool translate, bool tdrz,
               float pcmf32[], size_t pcmf32_len, size_t *segs_out_len,
               char *prompt);
GOWHISPER_EXPORT
const char *get_segment_text(int i);
GOWHISPER_EXPORT
int64_t get_segment_t0(int i);
GOWHISPER_EXPORT
int64_t get_segment_t1(int i);
GOWHISPER_EXPORT
int n_tokens(int i);
GOWHISPER_EXPORT
int32_t get_token_id(int i, int j);
GOWHISPER_EXPORT
bool get_segment_speaker_turn_next(int i);
}
