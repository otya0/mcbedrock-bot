package names

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Mojang's data (every block with its property values, every biome's number)
// is not part of mcbedrock-bot: download-mojang-data takes it from Mojang's
// bedrock-samples into the user's cache folder. It is Mojang's, under the EULA.

// MojangDataURL is where the files are taken from: Mojang's reference data
// for add-on makers, the latest release.
const MojangDataURL = "https://raw.githubusercontent.com/Mojang/bedrock-samples/main/metadata/vanilladata_modules/"

// The two files, by the names Mojang gives them.
const (
	MojangBlocksFile = "mojang-blocks.json"
	MojangBiomesFile = "mojang-biomes.json"
)

// DownloadCommand is what a user runs once to take the files.
const DownloadCommand = "go run github.com/otya0/mcbedrock-bot/cmd/download-mojang-data@latest"

// MojangDataFolder is where the files are kept: mcbedrock-bot/mojang in the
// user's cache folder (on Windows, in AppData\Local), outside any repository.
func MojangDataFolder() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("find the cache folder: %w", err)
	}
	return filepath.Join(cache, "mcbedrock-bot", "mojang"), nil
}

// blockPropertyValueType says how a property's values are typed in the
// block state: it decides how the value is written into the NBT that is
// hashed (block_hash.go).
type blockPropertyValueType int

const (
	blockPropertyBool blockPropertyValueType = iota
	blockPropertyInt
	blockPropertyString
)

// blockProperty is one property with every value it can take, as text.
type blockProperty struct {
	valueType blockPropertyValueType
	values    []string
}

// blockListEntry is one block: its name and the names of its properties,
// in the order Mojang lists them.
type blockListEntry struct {
	name       string
	properties []string
}

// mojangData is what New takes from the two files.
type mojangData struct {
	blockProperties map[string]blockProperty
	blockList       []blockListEntry
	biomeList       map[int32]string
}

// readMojangData reads the two files from a folder. A file that is not
// there gives an error that says how to take it.
func readMojangData(folder string) (mojangData, error) {
	var data mojangData
	blocks, err := os.ReadFile(filepath.Join(folder, MojangBlocksFile))
	if errors.Is(err, os.ErrNotExist) {
		return data, fmt.Errorf("no Mojang data in %s: take it once with %s", folder, DownloadCommand)
	}
	if err != nil {
		return data, err
	}
	biomes, err := os.ReadFile(filepath.Join(folder, MojangBiomesFile))
	if errors.Is(err, os.ErrNotExist) {
		return data, fmt.Errorf("no Mojang data in %s: take it once with %s", folder, DownloadCommand)
	}
	if err != nil {
		return data, err
	}
	if data.blockProperties, data.blockList, err = readMojangBlocks(blocks); err != nil {
		return data, fmt.Errorf("%s: %w", MojangBlocksFile, err)
	}
	if data.biomeList, err = readMojangBiomes(biomes); err != nil {
		return data, fmt.Errorf("%s: %w", MojangBiomesFile, err)
	}
	return data, nil
}

// readMojangBlocks reads mojang-blocks.json: "block_properties" (each
// property, its type and its values) and "data_items" (each block and the
// names of its properties). The rest of the file is not used.
func readMojangBlocks(raw []byte) (map[string]blockProperty, []blockListEntry, error) {
	var file struct {
		BlockProperties []struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Values []struct {
				Value any `json:"value"`
			} `json:"values"`
		} `json:"block_properties"`
		DataItems []struct {
			Name       string `json:"name"`
			Properties []struct {
				Name string `json:"name"`
			} `json:"properties"`
		} `json:"data_items"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, nil, err
	}
	if len(file.DataItems) == 0 {
		return nil, nil, errors.New("no blocks in it")
	}
	properties := make(map[string]blockProperty, len(file.BlockProperties))
	for _, property := range file.BlockProperties {
		var valueType blockPropertyValueType
		switch property.Type {
		case "bool":
			valueType = blockPropertyBool
		case "int":
			valueType = blockPropertyInt
		case "string":
			valueType = blockPropertyString
		default:
			return nil, nil, fmt.Errorf("property %s has the unknown type %q", property.Name, property.Type)
		}
		values := make([]string, len(property.Values))
		for i, value := range property.Values {
			values[i] = fmt.Sprint(value.Value)
		}
		properties[property.Name] = blockProperty{valueType: valueType, values: values}
	}
	blocks := make([]blockListEntry, len(file.DataItems))
	for i, item := range file.DataItems {
		blocks[i].name = item.Name
		for _, property := range item.Properties {
			if _, known := properties[property.Name]; !known {
				return nil, nil, fmt.Errorf("block %s has the property %s, which is not listed", item.Name, property.Name)
			}
			blocks[i].properties = append(blocks[i].properties, property.Name)
		}
	}
	return properties, blocks, nil
}

// readMojangBiomes reads mojang-biomes.json: "data_items", each biome's
// number and name.
func readMojangBiomes(raw []byte) (map[int32]string, error) {
	var file struct {
		DataItems []struct {
			ID   int32  `json:"id"`
			Name string `json:"name"`
		} `json:"data_items"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, err
	}
	if len(file.DataItems) == 0 {
		return nil, errors.New("no biomes in it")
	}
	biomes := make(map[int32]string, len(file.DataItems))
	for _, item := range file.DataItems {
		biomes[item.ID] = item.Name
	}
	return biomes, nil
}
