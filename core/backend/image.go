package backend

import (
	"path/filepath"
	"strings"

	"github.com/mudler/LocalAI/core/config"

	"github.com/mudler/LocalAI/pkg/grpc/proto"
	model "github.com/mudler/LocalAI/pkg/model"
)

func ImageGeneration(height, width, step, seed int, positive_prompt, negative_prompt, src, dst string, loader *model.ModelLoader, modelConfig config.ModelConfig, appConfig *config.ApplicationConfig, refImages []string) (func() error, error) {

	opts := ModelOptions(modelConfig, appConfig)
	inferenceModel, err := loader.Load(
		opts...,
	)
	if err != nil {
		return nil, err
	}

	// Convert Windows paths to WSL paths when the backend shell is wsl-bash
	modelID := modelConfig.Name
	if modelID == "" {
		modelID = modelConfig.Model
	}
	shellKind := loader.GetShellKindForModel(modelID)
	toShellPath := func(p string) string {
		if p == "" {
			return p
		}
		if shellKind == "wsl-bash" {
			pp := filepath.ToSlash(p)
			if len(pp) > 2 && pp[1] == ':' { // e.g., G:/...
				drive := string(pp[0])
				rest := pp[2:]
				return "/mnt/" + strings.ToLower(drive) + rest
			}
			return pp
		}
		return p
	}

	convDst := toShellPath(dst)
	convSrc := toShellPath(src)
	convRef := make([]string, 0, len(refImages))
	for _, r := range refImages {
		convRef = append(convRef, toShellPath(r))
	}

	fn := func() error {
		_, err := inferenceModel.GenerateImage(
			appConfig.Context,
			&proto.GenerateImageRequest{
				Height:           int32(height),
				Width:            int32(width),
				Step:             int32(step),
				Seed:             int32(seed),
				CLIPSkip:         int32(modelConfig.Diffusers.ClipSkip),
				PositivePrompt:   positive_prompt,
				NegativePrompt:   negative_prompt,
				Dst:              convDst,
				Src:              convSrc,
				EnableParameters: modelConfig.Diffusers.EnableParameters,
				RefImages:        convRef,
			})
		return err
	}

	return fn, nil
}

// ImageGenerationFunc is a test-friendly indirection to call image generation logic.
// Tests can override this variable to provide a stub implementation.
var ImageGenerationFunc = ImageGeneration
