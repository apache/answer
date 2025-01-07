/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cli

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
	"github.com/apache/answer/pkg/dir"
	"github.com/apache/answer/pkg/writer"
	"github.com/apache/answer/ui"
	"github.com/segmentfault/pacman/log"
	"gopkg.in/yaml.v3"
)

const (
	mainGoTpl = `package main

import (
	answercmd "github.com/apache/answer/cmd"

  // remote plugins
	{{- range .remote_plugins}}
	_ "{{.}}"
	{{- end}}

  // local plugins
	{{- range .local_plugins}}
	_ "answer/{{.}}"
	{{- end}}
)

func main() {
	answercmd.Main()
}
`
	goModTpl = `module answer

go 1.22
`
)

type answerBuilder struct {
	buildingMaterial *buildingMaterial
	BuildError       error
}

type buildingMaterial struct {
	answerModuleReplacement string
	plugins                 []*pluginInfo
	outputPath              string
	tmpDir                  string
	originalAnswerInfo      OriginalAnswerInfo
}

type OriginalAnswerInfo struct {
	Version  string
	Revision string
	Time     string
}

type pluginInfo struct {
	// Name of the plugin e.g. github.com/apache/answer-plugins/github-connector
	Name string
	// Path to the plugin. If path exist, read plugin from local filesystem
	Path string
	// Version of the plugin
	Version string
}

func newAnswerBuilder(buildDir, outputPath string, plugins []string, originalAnswerInfo OriginalAnswerInfo) *answerBuilder {
	material := &buildingMaterial{originalAnswerInfo: originalAnswerInfo}
	parentDir, _ := filepath.Abs(".")
	if buildDir != "" {
		material.tmpDir = filepath.Join(parentDir, buildDir)
	} else {
		material.tmpDir, _ = os.MkdirTemp(parentDir, "answer_build")
	}
	if len(outputPath) == 0 {
		outputPath = filepath.Join(parentDir, "new_answer")
	}
	material.outputPath = outputPath
	material.plugins = formatPlugins(plugins)
	material.answerModuleReplacement = os.Getenv("ANSWER_MODULE")
	return &answerBuilder{
		buildingMaterial: material,
	}
}

func (a *answerBuilder) DoTask(task func(b *buildingMaterial) error) {
	if a.BuildError != nil {
		return
	}
	a.BuildError = task(a.buildingMaterial)
}

// BuildNewAnswer builds a new answer with specified plugins
func BuildNewAnswer(buildDir, outputPath string, plugins []string, originalAnswerInfo OriginalAnswerInfo) (err error) {
	builder := newAnswerBuilder(buildDir, outputPath, plugins, originalAnswerInfo)
	builder.DoTask(createMainGoFile)
	builder.DoTask(downloadGoModFile)
	builder.DoTask(copyUIFiles)
	builder.DoTask(buildUI)
	builder.DoTask(mergeI18nFiles)
	builder.DoTask(buildBinary)
	builder.DoTask(cleanByproduct)
	return builder.BuildError
}

func formatPlugins(plugins []string) (formatted []*pluginInfo) {
	for _, plugin := range plugins {
		plugin = strings.TrimSpace(plugin)
		// plugin description like this 'github.com/apache/answer-plugins/github-connector@latest=/local/path'
		info := &pluginInfo{}
		plugin, info.Path, _ = strings.Cut(plugin, "=")
		info.Name, info.Version, _ = strings.Cut(plugin, "@")
		formatted = append(formatted, info)
	}
	return formatted
}

// createMainGoFile creates main.go file in tmp dir that content is mainGoTpl
func createMainGoFile(b *buildingMaterial) (err error) {
	fmt.Printf("[build] build dir: %s\n", b.tmpDir)
	err = dir.CreateDirIfNotExist(b.tmpDir)
	if err != nil {
		return err
	}

	var (
		remotePlugins []string
	)
	for _, p := range b.plugins {
		remotePlugins = append(remotePlugins, versionedModulePath(p.Name, p.Version))
	}

	mainGoFile := &bytes.Buffer{}
	tmpl, err := template.New("main").Parse(mainGoTpl)
	if err != nil {
		return err
	}
	err = tmpl.Execute(mainGoFile, map[string]any{
		"remote_plugins": remotePlugins,
	})
	if err != nil {
		return err
	}

	err = writer.WriteFile(filepath.Join(b.tmpDir, "main.go"), mainGoFile.String())
	if err != nil {
		return err
	}

	err = writer.WriteFile(filepath.Join(b.tmpDir, "go.mod"), goModTpl)
	if err != nil {
		return err
	}

	for _, p := range b.plugins {
		// If user set a path, use it to replace the module with local path
		if len(p.Path) > 0 {
			replacement := fmt.Sprintf("%s@%s=%s", p.Name, p.Version, p.Path)
			err = b.newExecCmd("go", "mod", "edit", "-replace", replacement).Run()
		} else if len(p.Version) > 0 {
			// If user specify a version, use it to get specific version of the module
			err = b.newExecCmd("go", "get", fmt.Sprintf("%s@%s", p.Name, p.Version)).Run()
		}
		if err != nil {
			return err
		}
	}
	return
}

// downloadGoModFile run go mod commands to download dependencies
func downloadGoModFile(b *buildingMaterial) (err error) {
	// If user specify a module replacement, use it. Otherwise, use the latest version.
	if len(b.answerModuleReplacement) > 0 {
		replacement := fmt.Sprintf("%s=%s", "github.com/apache/answer", b.answerModuleReplacement)
		err = b.newExecCmd("go", "mod", "edit", "-replace", replacement).Run()
		if err != nil {
			return err
		}
	}

	err = b.newExecCmd("go", "mod", "tidy").Run()
	if err != nil {
		return err
	}

	err = b.newExecCmd("go", "mod", "vendor").Run()
	if err != nil {
		return err
	}
	return
}

// copyUIFiles copy ui files from answer module to tmp dir
func copyUIFiles(b *buildingMaterial) (err error) {
	goListCmd := b.newExecCmd("go", "list", "-mod=mod", "-m", "-f", "{{.Dir}}", "github.com/apache/answer")
	buf := new(bytes.Buffer)
	goListCmd.Stdout = buf
	if err = goListCmd.Run(); err != nil {
		return fmt.Errorf("failed to run go list: %w", err)
	}

	answerDir := strings.TrimSpace(buf.String())
	goModUIDir := filepath.Join(answerDir, "ui")
	localUIBuildDir := filepath.Join(b.tmpDir, "vendor/github.com/apache/answer/ui/")
	// The node_modules folder generated during development will interfere packaging, so it needs to be ignored.
	if err = copyDirEntries(os.DirFS(goModUIDir), ".", localUIBuildDir, "node_modules"); err != nil {
		return fmt.Errorf("failed to copy ui files: %w", err)
	}

	pluginsDir := filepath.Join(b.tmpDir, "vendor/github.com/apache/answer-plugins/")
	localUIPluginDir := filepath.Join(localUIBuildDir, "src/plugins/")

	// copy plugins dir
	fmt.Printf("try to copy dir from %s to %s\n", pluginsDir, localUIPluginDir)

	// if plugins dir not exist means no plugins
	if !dir.CheckDirExist(pluginsDir) {
		return nil
	}

	pluginsDirEntries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return fmt.Errorf("failed to read plugins dir: %w", err)
	}
	for _, entry := range pluginsDirEntries {
		if !entry.IsDir() {
			continue
		}
		sourcePluginDir := filepath.Join(pluginsDir, entry.Name())
		// check if plugin is a ui plugin
		packageJsonPath := filepath.Join(sourcePluginDir, "package.json")
		fmt.Printf("check if %s is a ui plugin\n", packageJsonPath)
		if !dir.CheckFileExist(packageJsonPath) {
			continue
		}

		pnpmInstallCmd := b.newExecCmd("pnpm", "install")
		pnpmInstallCmd.Dir = sourcePluginDir
		if err = pnpmInstallCmd.Run(); err != nil {
			return fmt.Errorf("failed to install plugin dependencies: %w", err)
		}

		localPluginDir := filepath.Join(localUIPluginDir, entry.Name())
		fmt.Printf("try to copy dir from %s to %s\n", sourcePluginDir, localPluginDir)
		if err = copyDirEntries(os.DirFS(sourcePluginDir), ".", localPluginDir, "node_modules"); err != nil {
			return fmt.Errorf("failed to copy ui files: %w", err)
		}
	}
	formatUIPluginsDirName(localUIPluginDir)
	return nil
}

// overwriteIndexTs overwrites index.ts file in ui/src/plugins/ dir
func overwriteIndexTs(b *buildingMaterial) (err error) {
	localUIPluginDir := filepath.Join(b.tmpDir, "vendor/github.com/apache/answer/ui/src/plugins/")

	folders, err := getFolders(localUIPluginDir)
	if err != nil {
		return fmt.Errorf("failed to get folders: %w", err)
	}

	content := generateIndexTsContent(folders)
	err = os.WriteFile(filepath.Join(localUIPluginDir, "index.ts"), []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write index.ts: %w", err)
	}
	return nil
}

func getFolders(dir string) ([]string, error) {
	var folders []string
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() && file.Name() != "builtin" {
			folders = append(folders, file.Name())
		}
	}
	return folders, nil
}

func generateIndexTsContent(folders []string) string {
	builder := &strings.Builder{}
	builder.WriteString("export default null;\n")
	// Line 2:1:  Delete `⏎`  prettier/prettier
	if len(folders) > 0 {
		builder.WriteString("\n")
	}
	for _, folder := range folders {
		builder.WriteString(fmt.Sprintf("export { default as %s } from '%s';\n", folder, folder))
	}
	return builder.String()
}

// buildUI run pnpm install and pnpm build commands to build ui
func buildUI(b *buildingMaterial) (err error) {
	localUIBuildDir := filepath.Join(b.tmpDir, "vendor/github.com/apache/answer/ui")

	pnpmInstallCmd := b.newExecCmd("pnpm", "pre-install")
	pnpmInstallCmd.Dir = localUIBuildDir
	if err = pnpmInstallCmd.Run(); err != nil {
		return err
	}

	pnpmBuildCmd := b.newExecCmd("pnpm", "build")
	pnpmBuildCmd.Dir = localUIBuildDir
	if err = pnpmBuildCmd.Run(); err != nil {
		return err
	}
	return nil
}

func replaceNecessaryFile(b *buildingMaterial) (err error) {
	fmt.Printf("try to replace ui build directory\n")
	uiBuildDir := filepath.Join(b.tmpDir, "vendor/github.com/apache/answer/ui")
	err = copyDirEntries(ui.Build, ".", uiBuildDir)
	return err
}

// mergeI18nFiles merge i18n files
func mergeI18nFiles(b *buildingMaterial) (err error) {
	fmt.Printf("try to merge i18n files\n")

	type YamlPluginContent struct {
		Plugin map[string]any `yaml:"plugin"`
	}

	pluginAllTranslations := make(map[string]*YamlPluginContent)
	for _, plugin := range b.plugins {
		i18nDir := filepath.Join(b.tmpDir, fmt.Sprintf("vendor/%s/i18n", plugin.Name))
		fmt.Println("i18n dir: ", i18nDir)
		if !dir.CheckDirExist(i18nDir) {
			continue
		}

		entries, err := os.ReadDir(i18nDir)
		if err != nil {
			return err
		}

		for _, file := range entries {
			// ignore directory
			if file.IsDir() {
				continue
			}
			// ignore non-YAML file
			if filepath.Ext(file.Name()) != ".yaml" {
				continue
			}
			buf, err := os.ReadFile(filepath.Join(i18nDir, file.Name()))
			if err != nil {
				log.Debugf("read translation file failed: %s %s", file.Name(), err)
				continue
			}

			translation := &YamlPluginContent{}
			if err = yaml.Unmarshal(buf, translation); err != nil {
				log.Debugf("unmarshal translation file failed: %s %s", file.Name(), err)
				continue
			}

			if pluginAllTranslations[file.Name()] == nil {
				pluginAllTranslations[file.Name()] = &YamlPluginContent{Plugin: make(map[string]any)}
			}
			for k, v := range translation.Plugin {
				pluginAllTranslations[file.Name()].Plugin[k] = v
			}
		}
	}

	originalI18nDir := filepath.Join(b.tmpDir, "vendor/github.com/apache/answer/i18n")
	entries, err := os.ReadDir(originalI18nDir)
	if err != nil {
		return err
	}

	for _, file := range entries {
		// ignore directory
		if file.IsDir() {
			continue
		}
		// ignore non-YAML file
		filename := file.Name()
		if filepath.Ext(filename) != ".yaml" && filename != "i18n.yaml" {
			continue
		}

		// if plugin don't have this translation file, ignore it
		if pluginAllTranslations[filename] == nil {
			continue
		}

		out, _ := yaml.Marshal(pluginAllTranslations[filename])

		buf, err := os.OpenFile(filepath.Join(originalI18nDir, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Debugf("read translation file failed: %s %s", filename, err)
			continue
		}

		_, _ = buf.WriteString("\n")
		_, _ = buf.Write(out)
		_ = buf.Close()
	}
	return err
}

func copyDirEntries(sourceFs fs.FS, sourceDir, targetDir string, ignoreDir ...string) (err error) {
	err = dir.CreateDirIfNotExist(targetDir)
	if err != nil {
		return err
	}
	ignoreThisDir := func(path string) bool {
		for _, s := range ignoreDir {
			if strings.HasPrefix(path, s) {
				return true
			}
		}
		return false
	}

	err = fs.WalkDir(sourceFs, sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ignoreThisDir(path) {
			return nil
		}

		// Convert the path to use forward slashes, important because we use embedded FS which always uses forward slashes
		path = filepath.ToSlash(path)

		// Construct the absolute path for the source file/directory
		srcPath := filepath.Join(sourceDir, path)

		// Construct the absolute path for the destination file/directory
		dstPath := filepath.Join(targetDir, path)

		if d.IsDir() {
			// Create the directory in the destination
			err := os.MkdirAll(dstPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
			}
		} else {
			// Open the source file
			srcFile, err := sourceFs.Open(srcPath)
			if err != nil {
				return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
			}
			defer srcFile.Close()

			// Create the destination file
			dstFile, err := os.Create(dstPath)
			if err != nil {
				return fmt.Errorf("failed to create destination file %s: %w", dstPath, err)
			}
			defer dstFile.Close()

			// Copy the file contents
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return fmt.Errorf("failed to copy file contents from %s to %s: %w", srcPath, dstPath, err)
			}
		}

		return nil
	})

	return err
}

// format plugins dir name from dash to underline
func formatUIPluginsDirName(dirPath string) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fmt.Printf("read ui plugins dir failed: [%s] %s\n", dirPath, err)
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() || !strings.Contains(entry.Name(), "-") {
			continue
		}
		newName := strings.ReplaceAll(entry.Name(), "-", "_")
		if err := os.Rename(filepath.Join(dirPath, entry.Name()), filepath.Join(dirPath, newName)); err != nil {
			fmt.Printf("rename ui plugins dir failed: [%s] %s\n", dirPath, err)
		} else {
			fmt.Printf("rename ui plugins dir success: [%s] -> [%s]\n", entry.Name(), newName)
		}
	}
}

// buildBinary build binary file
func buildBinary(b *buildingMaterial) (err error) {
	versionInfo := b.originalAnswerInfo
	cmdPkg := "github.com/apache/answer/cmd"
	ldflags := fmt.Sprintf("-X %s.Version=%s -X %s.Revision=%s -X %s.Time=%s",
		cmdPkg, versionInfo.Version, cmdPkg, versionInfo.Revision, cmdPkg, versionInfo.Time)
	err = b.newExecCmd("go", "build",
		"-ldflags", ldflags, "-o", b.outputPath, ".").Run()
	if err != nil {
		return err
	}
	return
}

// cleanByproduct delete tmp dir
func cleanByproduct(b *buildingMaterial) (err error) {
	return os.RemoveAll(b.tmpDir)
}

func (b *buildingMaterial) newExecCmd(command string, args ...string) *exec.Cmd {
	cmd := exec.Command(command, args...)
	fmt.Println(cmd.Args)
	cmd.Dir = b.tmpDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func versionedModulePath(modulePath, moduleVersion string) string {
	if moduleVersion == "" {
		return modulePath
	}
	ver, err := semver.StrictNewVersion(strings.TrimPrefix(moduleVersion, "v"))
	if err != nil {
		return modulePath
	}
	major := ver.Major()
	if major > 1 {
		modulePath += fmt.Sprintf("/v%d", major)
	}
	return path.Clean(modulePath)
}
