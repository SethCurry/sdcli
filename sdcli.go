package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SethCurry/sdcli/internal/exif"
	"github.com/SethCurry/sdcli/pkg/stability"
	"github.com/alecthomas/kong"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
)

const defaultBaseURL = "https://api.stability.ai"

func getExifAdder(format string) (func([]byte, string) ([]byte, error), error) {
	switch format {
	case "jpeg":
		return exif.AddToJPEG, nil
	case "png":
		return exif.AddToPNG, nil
	}

	return nil, fmt.Errorf("unknown output format %q", format)
}

type Gen3Command struct {
	Model          string   `optional:"model" default:"sd3" help:"The model to use.  Must be either sd3 or sd3turbo"`
	Ratio          string   `optional:"ratio" help:"The aspect ratio to use when generating."`
	OutputFormat   string   `optional:"format" default:"png" help:"The format of the returned image.  Must be either png or jpeg."`
	NegativePrompt string   `optional:"negative" help:"The negative prompt to use during generation."`
	Strength       float32  `optional:"strength" help:"The strength to use when doing image-to-image generation."`
	Image          string   `optional:"image" type:"path" help:"The image to use for image-to-image generation."`
	PromptParts    []string `arg:"" help:"The prompt to use for generation."`
}

func (g Gen3Command) Run(ctx *Context) error {
	prompt := strings.Join(g.PromptParts, " ")

	if prompt == "" {
		ctx.Logger.Fatal("prompt is empty, exiting")
	}

	opts := []stability.Generate3Option{stability.WithPrompt(prompt)}

	if g.Ratio != "" {
		opts = append(opts, stability.WithAspectRatio(g.Ratio))
	}

	if g.Model != "" {
		opts = append(opts, stability.WithModel(g.Model))
	}

	if g.OutputFormat != "" {
		opts = append(opts, stability.WithOutputFormat(g.OutputFormat))
	}

	if g.NegativePrompt != "" {
		opts = append(opts, stability.WithNegativePrompt(g.NegativePrompt))
	}

	if g.Strength != 0 {
		opts = append(opts, stability.WithStrength(g.Strength))
	}

	if g.Image != "" {
		fd, err := os.Open(g.Image)
		if err != nil {
			ctx.Logger.Fatal("failed to open image", zap.String("path", g.Image), zap.Error(err))
		}
		defer fd.Close()

		opts = append(opts, stability.WithImage(fd))
	}

	gotImage, err := stability.Generate3(context.Background(), defaultBaseURL, ctx.Config.APIKey, opts...)
	if err != nil {
		ctx.Logger.Fatal("failed to generate image", zap.Error(err))
	}

	exifAdder, err := getExifAdder(g.OutputFormat)
	if err != nil {
		ctx.Logger.Fatal("failed to find Exif adder", zap.Error(err))
	}

	imageWithNewExif, err := exifAdder(gotImage, prompt)
	if err != nil {
		ctx.Logger.Fatal("failed to add new exif metadata", zap.Error(err))
	}

	currentTime := strconv.FormatInt(time.Now().Unix(), 10)

	outputFile := filepath.Join(ctx.Config.OutputDirectory, fmt.Sprintf("%s.%s", currentTime, g.OutputFormat))
	if _, err := os.Stat(outputFile); err == nil {
		ctx.Logger.Fatal("output file already exists", zap.String("path", outputFile))
	}

	err = os.WriteFile(outputFile, imageWithNewExif, 0o644)
	if err != nil {
		ctx.Logger.Fatal("failed while writing to output file", zap.String("path", outputFile), zap.Error(err))
	}

	if ctx.Config.PostGenerationCommand != "" {
		cmd := exec.Command(ctx.Config.PostGenerationCommand, outputFile)
		err = cmd.Run()
		if err != nil {
			ctx.Logger.Error(
				"post-generation command failed",
				zap.String("command", fmt.Sprintf("%s %q", ctx.Config.PostGenerationCommand, outputFile)))
		}
	}

	return nil
}

type CLI struct {
	Gen3 Gen3Command `cmd:"" help:"Generate an image with Stable Diffusion 3"`
}

type Context struct {
	Logger *zap.Logger
	Config Config
}

type Config struct {
	APIKey                string `json:"api_key"`
	OutputDirectory       string `json:"output_directory"`
	PostGenerationCommand string `json:"post_generation_command"`
}

func getConfigDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "sdcli"), nil
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Errorf("failed to create logger: %w", err))
	}

	configDir, err := getConfigDir()
	if err != nil {
		logger.Fatal("failed to get config directory", zap.Error(err))
	}

	configData, err := os.ReadFile(filepath.Join(configDir, "config.json"))
	if err != nil {
		logger.Fatal("failed to read config data", zap.Error(err))
	}

	var config Config

	err = json.Unmarshal(configData, &config)
	if err != nil {
		logger.Fatal("failed to unmarshal config JSON", zap.Error(err))
	}

	cli := &CLI{}

	ctx := kong.Parse(cli)

	err = ctx.Run(&Context{
		Logger: logger,
		Config: config,
	})
	if err != nil {
		logger.Fatal("failed to execute command", zap.Error(err))
	}
}
