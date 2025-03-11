package config

import "runtime"

type BuildDirType struct {
	NavtiveBuildDir string
	JarBuildDir     string
}

type ConfigType struct {
	SrcDir     string
	BinDir     string
	PackageDir string
	BuildDir   *BuildDirType
	CacheDir   string
}

var config = &ConfigType{
	SrcDir:     "src",
	BinDir:     ".jpkg/bin",
	PackageDir: "lib",
	BuildDir:   &BuildDirType{},
	CacheDir:   "",
}

func GetConfig() *ConfigType {
	config.BuildDir.JarBuildDir = ".jets/build/jar"
	config.BuildDir.NavtiveBuildDir = ".jets/build/" + getOS()
	return config
}

func getOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "mac"
	case "freebsd":
		return "freebsd"
	default:
		return "unknown"
	}
}
