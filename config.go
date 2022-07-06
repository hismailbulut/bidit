package main

import (
	"bidit/bidutil"
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	PRIVATE_CONFIG_FILE_NAME = "config.bin"
	PUBLIC_CONFIG_FILE_NAME  = "config.json"
)

func FileExists(path string) bool {
	_, err := os.Stat(PrivateConfigFilePath())
	return err == nil
}

func ConfigFolderPath() string {
	userConfigDir, err := os.UserConfigDir()
	Assert(err == nil, "Error when trying to get UserConfigDir:", err)
	return filepath.Join(userConfigDir, NAME)
}

func PrivateConfigFilePath() string {
	return filepath.Join(ConfigFolderPath(), PRIVATE_CONFIG_FILE_NAME)
}

func PublicConfigFilePath() string {
	return filepath.Join(ConfigFolderPath(), PUBLIC_CONFIG_FILE_NAME)
}

func CheckConfigFolder() bool {
	configFolderPath := ConfigFolderPath()
	print("ConfigFolderPath:", configFolderPath)
	// check if the path exists
	_, err := os.Stat(configFolderPath)
	if err != nil {
		print("Config folder not found")
		err := os.Mkdir(configFolderPath, 0666)
		Assert(err == nil, "Config folder creation failed:", err)
		print("Config folder created successfully")
	}
	// check private config file, public one is not so important
	if !FileExists(PrivateConfigFilePath()) {
		print("Private config file not found")
		return false
	}
	print("PrivateConfigFilePath:", PrivateConfigFilePath())
	if FileExists(PublicConfigFilePath()) {
		print("PublicConfigFilePath:", PublicConfigFilePath())
	}
	return true
}

type Config struct {
	// May not safe
	keyHash []byte
	// This field encrypted with aes 256 times
	Private struct {
		PrivateKey string
	}
	// This field is not encrypted
	Public struct {
		MainnetWorkers []WorkerParams
		TestnetWorkers []WorkerParams
	}
}

func DefaultConfig() *Config {
	config := new(Config)
	return config
}

func LoadConfig(key []byte) (*Config, error) {
	config := DefaultConfig()
	// First load private file
	{
		// Load file
		data, err := os.ReadFile(PrivateConfigFilePath())
		if err != nil {
			return config, err
		}
		// Decrypt with key
		jsonData, err := bidutil.Decrypt(data, key)
		if err != nil {
			return config, err
		}
		// Unmarshall json
		err = json.Unmarshal(jsonData, &config.Private)
		if err != nil {
			return config, err
		}
		// Store key hash because we need it when saving
		config.keyHash = key
	}
	// Then load public one
	func() {
		if FileExists(PublicConfigFilePath()) {
			jsonData, err := os.ReadFile(PublicConfigFilePath())
			if err != nil {
				print("Error when reading public config:", err)
				return
			}
			err = json.Unmarshal(jsonData, &config.Public)
			if err != nil {
				print("Error when unmarshalling public config:", err)
				return
			}
		}
	}()
	return config, nil
}

func (config *Config) Save() error {
	// Private
	{
		// Marshall struct
		jsonData, err := json.Marshal(config.Private)
		if err != nil {
			return err
		}
		// Encrypt with password
		data, err := bidutil.Encrypt(jsonData, config.keyHash)
		if err != nil {
			return err
		}
		err = os.WriteFile(PrivateConfigFilePath(), data, 0666)
		if err != nil {
			return err
		}
	}
	// Public
	{
		jsonData, err := json.Marshal(config.Public)
		if err != nil {
			return err
		}
		err = os.WriteFile(PublicConfigFilePath(), jsonData, 0666)
		if err != nil {
			return err
		}
	}
	return nil
}
