// Command download-mojang-data takes Mojang's block and biome lists from
// Mojang's bedrock-samples into the user's cache folder, where mcbedrock-bot
// reads them to turn block and biome numbers into names. Run it once, and
// again after Minecraft updates:
//
//	go run github.com/otya0/mcbedrock-bot/cmd/download-mojang-data@latest
//
// The files come straight from Mojang to this computer; mcbedrock-bot does not
// contain them. They belong to Mojang AB and are subject to the Minecraft EULA.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/otya0/mcbedrock-bot/internal/record/names"
)

// maxFileSize stops a download that is far bigger than the files are
// (about 400 KB and 6 KB): something else is being sent.
const maxFileSize = 16 << 20

func main() {
	defaultFolder, err := names.MojangDataFolder()
	if err != nil {
		log.Fatal(err)
	}
	folder := flag.String("folder", defaultFolder, "where to save the files")
	flag.Parse()

	fmt.Println("Downloading Mojang's block and biome lists from " + names.MojangDataURL)
	fmt.Println("These files belong to Mojang AB and are subject to the Minecraft EULA: https://www.minecraft.net/en-us/eula")
	if err := os.MkdirAll(*folder, 0o755); err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Timeout: time.Minute}
	for _, name := range []string{names.MojangBlocksFile, names.MojangBiomesFile} {
		if err := download(client, names.MojangDataURL+name, filepath.Join(*folder, name)); err != nil {
			log.Fatalf("%s: %v", name, err)
		}
	}
	// Reading them as mcbedrock-bot does checks that what arrived is the real thing.
	if _, err := names.New(*folder); err != nil {
		log.Fatalf("the downloaded files cannot be read: %v", err)
	}
	fmt.Println("Saved in " + *folder)
}

// download saves one file. It is written next to its place first and moved
// in only when it arrived whole, so a broken download never replaces a good file.
func download(client *http.Client, url, path string) error {
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("the server answered %s", response.Status)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxFileSize+1))
	if err != nil {
		return err
	}
	if len(body) > maxFileSize {
		return errors.New("the file is far bigger than expected")
	}
	temporary := path + ".download"
	if err := os.WriteFile(temporary, body, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
