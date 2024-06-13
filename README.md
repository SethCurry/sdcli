# sdcli

`sdcli` is a CLI application for generating images with the Stable Diffusion API.

## Features

- Generate and save images with the Stable Diffusion API
- Saves the prompt you used to generate the image in exif metadata so you can reference it
- Can invoke commands after generating an image to open it in an image viewer or editor

## Configuration

Config is stored at ~/.config/sdcli/config.json (even on Windows, sorry ya'll).

Example config:

```json
{
  // The Stability API key to use for billing
  "api_key": "YourAPIkeyHere",

  // The absolute or relative path to save images at when generating.
  // This does not expand ~ or environment variables.
  "output_directory": "/path/to/directory/to/store/files",

  // The command to run after generating an image.
  // The command will be invoked with a single positional argument,
  // the path to the file that was generated.
  //
  // I use this to open the images with Firefox after generation,
  // but a real image viewer or editor could also work.  You can
  // pass a path to a script if you need to provide additional
  // arguments to the application you want to run.
  "post_generation_command": "/usr/bin/firefox"
}
```

## Usage

Generate a basic image with Stable Diffusion 3:

```bash
sdcli gen-3 A bear riding a unicycle in space
```

Generate an image with a different aspect ratio:

```bash
sdcli gen-3 --ratio 3:4 A bear riding a unicycle in space
```

You can also quote the prompt if you need to or don't want to escape characters:

```bash
sdcli gen-3 "A bear eating another bear's porridge"
```
