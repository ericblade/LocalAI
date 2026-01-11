#include <cstdint>
#include "stable-diffusion.h"

// Cross-platform export macro for Windows DLL and other platforms
#ifdef _WIN32
    #define GOSD_EXPORT extern "C" __declspec(dllexport)
#else
    #define GOSD_EXPORT
#endif

#ifdef __cplusplus
extern "C" {
#endif

GOSD_EXPORT
void sd_tiling_params_set_enabled(sd_tiling_params_t *params, bool enabled);
GOSD_EXPORT
void sd_tiling_params_set_tile_sizes(sd_tiling_params_t *params, int tile_size_x, int tile_size_y);
GOSD_EXPORT
void sd_tiling_params_set_rel_sizes(sd_tiling_params_t *params, float rel_size_x, float rel_size_y);
GOSD_EXPORT
void sd_tiling_params_set_target_overlap(sd_tiling_params_t *params, float target_overlap);
GOSD_EXPORT
sd_tiling_params_t* sd_img_gen_params_get_vae_tiling_params(sd_img_gen_params_t *params);

GOSD_EXPORT
sd_img_gen_params_t* sd_img_gen_params_new(void);
GOSD_EXPORT
void sd_img_gen_params_set_prompts(sd_img_gen_params_t *params, const char *prompt, const char *negative_prompt);
GOSD_EXPORT
void sd_img_gen_params_set_dimensions(sd_img_gen_params_t *params, int width, int height);
GOSD_EXPORT
void sd_img_gen_params_set_seed(sd_img_gen_params_t *params, int64_t seed);

GOSD_EXPORT
int load_model(const char *model, char *model_path, char* options[], int threads, int diffusionModel);
GOSD_EXPORT
int gen_image(sd_img_gen_params_t *p, int steps, char *dst, float cfg_scale, char *src_image, float strength, char *mask_image, char* ref_images[], int ref_images_count);
#ifdef __cplusplus
}
#endif
