/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package utils

import (
	"bytes"
	crypto "crypto/rand"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Data map[string]interface{}

func ResetTmpDir(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().SudoCmd(fmt.Sprintf(
		"if [ -d %s ]; then rm -rf %s ;fi && mkdir -m 777 -p %s",
		common.TmpDir, common.TmpDir, common.TmpDir), false, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "reset tmp dir failed")
	}
	return nil
}

func ToYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

func Indent(n int, text string) string {
	startOfLine := regexp.MustCompile(`(?m)^`)
	indentation := strings.Repeat(" ", n)
	return startOfLine.ReplaceAllLiteralString(text, indentation)
}

// Render text template with given `variables` Render-context
func Render(tmpl *template.Template, variables map[string]interface{}) (string, error) {

	var buf strings.Builder

	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", errors.Wrap(err, "Failed to render template")
	}
	return buf.String(), nil
}

func WorkDir() (string, error) {
	return os.Getwd()
}

// Home returns the home directory for the executing user.
func Home() (string, error) {
	u, err := user.Current()
	if nil == err {
		return u.HomeDir, nil
	}

	if "windows" == runtime.GOOS {
		return homeWindows()
	}

	return homeUnix()
}

func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

func GetArgs(argsMap map[string]string, args []string) ([]string, map[string]string) {
	targetMap := make(map[string]string, len(argsMap))
	for k, v := range argsMap {
		targetMap[k] = v
	}
	targetSlice := make([]string, len(args))
	copy(targetSlice, args)

	for _, arg := range targetSlice {
		splitArg := strings.SplitN(arg, "=", 2)
		if len(splitArg) < 2 {
			continue
		}
		targetMap[splitArg[0]] = splitArg[1]
	}

	for arg, value := range targetMap {
		cmd := fmt.Sprintf("%s=%s", arg, value)
		targetSlice = append(targetSlice, cmd)
	}
	sort.Strings(targetSlice)
	return targetSlice, targetMap
}

// Round returns the result of rounding 'val' according to the specified 'precision' precision (the number of digits after the decimal point)。
// and precision can be negative number or zero
func Round(val float64, precision int) float64 {
	p := math.Pow10(precision)
	return math.Floor(val*p+0.5) / p
}

func FormatBytes(bytes int64) string {
	const (
		KB = 1 << 10 // 1024
		MB = 1 << 20 // 1024 * 1024
		GB = 1 << 30 // 1024 * 1024 * 1024
		TB = 1 << 40 // 1024 * 1024 * 1024 * 1024
	)

	var result string
	switch {
	case bytes >= TB:
		result = fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		result = fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		result = fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		result = fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		result = fmt.Sprintf("%d Byte", bytes)
	}

	return result
}

func GenerateNumberWithProbability(p float64) int {
	rand.Seed(time.Now().UnixNano())
	randomFloat := rand.Float64()
	if randomFloat < p {
		return 2 * rand.Intn(50)
	} else {
		return 2*rand.Intn(50) + 1
	}
}

func GeneratePassword(length int) (string, error) {
	password := make([]byte, length)
	for i := range password {
		index, err := crypto.Int(crypto.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[index.Int64()]
	}
	return string(password), nil
}

func RemoveAnsiCodes(input string) string {
	ansiEscape := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiEscape.ReplaceAllString(input, "")
}

func ArchAlias(arch string) string {
	switch arch {
	case "aarch64", "armv7l", "arm64", "arm":
		return "arm64"
	case "x86_64", "amd64":
		fallthrough
	case "ppc64le":
		fallthrough
	case "s390x":
		fallthrough
	default:
		return "amd64"
	}
}