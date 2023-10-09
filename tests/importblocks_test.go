package tests

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	type testCase struct {
		dir    string
		config string
		lookup string
		args   []string
		cmdDir string
	}

	tests := []testCase{
		{dir: "std_first", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "std_last", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "std_first_if_not_set", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "any", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "block_sorted", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "multi_rules_block", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "multiple_import_blocks", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "original_content_not_changed", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "deep_file_structure", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "single_file", config: "importblocks.yaml", args: []string{"file.go"}},
		{dir: "multiple_files", config: "importblocks.yaml", args: []string{"file.go", "file2.go"}},
		{dir: "single_dir", config: "importblocks.yaml", args: []string{"dir1"}},
		{dir: "multiple_dirs", config: "importblocks.yaml", args: []string{"dir1", "dir2/..."}},
		{dir: "bad_ast", args: []string{"./..."}},
		{dir: "module", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "comments", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "named_imports", config: "importblocks.yaml", args: []string{"./..."}},
		{dir: "default_config_if_no_config_passed", args: []string{"./..."}},
		{dir: "lookup", lookup: "lookup.yaml", args: []string{"./..."}, cmdDir: filepath.Join("work", "dir1", "dir2")},
	}

	testsDir, err := os.Getwd()
	require.NoError(t, err)

	executable := filepath.Join(testsDir, "importblocks")

	_, err = exec.Command("go", "build", "-o", executable, filepath.Dir(testsDir)).Output()
	require.NoError(t, err)

	for _, test := range tests {
		t.Run(test.dir, func(t *testing.T) {
			testDir := filepath.Join(testsDir, test.dir)

			startDir := filepath.Join(testDir, "start")

			workDir := filepath.Join(testDir, "work")
			expectedDir := filepath.Join(testDir, "expected")

			err = copyDir(startDir, workDir)
			require.NoError(t, err)

			var args []string
			if test.config != "" {
				args = append(args, "-config", test.config)
			}
			if test.lookup != "" {
				args = append(args, "-lookup", test.lookup)
			}
			args = append(args, "-v")
			args = append(args, test.args...)

			cmd := exec.Command(executable, args...)
			if test.cmdDir != "" {
				cmd.Dir = filepath.Join(testDir, test.cmdDir)
			} else {
				cmd.Dir = workDir
			}
			cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
			cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", os.Getenv("HOME")))

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err = cmd.Run()
			if err != nil {
				t.Errorf("%s", err)
			}

			compareDirs(t, workDir, expectedDir)

			err = os.RemoveAll(workDir)
			require.NoError(t, err)
		})
	}

	err = os.Remove(executable)
	require.NoError(t, err)
}

func copyDir(from string, to string) error {
	return filepath.Walk(from, func(path string, info fs.FileInfo, err error) error {
		dest := strings.Replace(path, from, to, 1)
		if info.IsDir() {
			err := os.MkdirAll(dest, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			err = os.WriteFile(strings.TrimSuffix(dest, ".txt"), src, fs.ModePerm)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func compareDirs(t *testing.T, dir1, dir2 string) {
	_ = filepath.Walk(dir1, func(f1Path string, f1Info fs.FileInfo, err error) error {
		require.NoError(t, err)

		f2Path := filepath.Join(dir2, strings.TrimPrefix(f1Path, dir1))
		if strings.HasSuffix(f2Path, ".go") {
			f2Path += ".txt"
		}

		f2, err := os.Open(f2Path)
		require.NoError(t, err)

		f2Info, err := f2.Stat()
		require.NoError(t, err)

		if f1Info.IsDir() && f2Info.IsDir() {
			return nil
		}

		if f1Info.IsDir() && !f2Info.IsDir() {
			t.Error(fmt.Errorf("%s not dir", f2Path))
		}

		if !f1Info.IsDir() && f2Info.IsDir() {
			t.Error(fmt.Errorf("%s not file", f2Path))
		}

		f1Bytes, err := os.ReadFile(f1Path)
		require.NoError(t, err)

		f2Bytes, err := os.ReadFile(f2Path)
		require.NoError(t, err)

		require.Equal(t, strings.TrimSpace(string(f1Bytes)), strings.TrimSpace(string(f2Bytes)))

		return nil
	})
}
