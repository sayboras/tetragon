// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

//go:build !windows
// +build !windows

package btf

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/cilium/ebpf/btf"
	api "github.com/cilium/tetragon/pkg/api/tracingapi"
	"github.com/cilium/tetragon/pkg/defaults"
	"github.com/cilium/tetragon/pkg/kernels"
	"github.com/cilium/tetragon/pkg/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

var testBTFFiles = []struct {
	btf     string
	create  string
	wantbtf string
	err     error
}{
	{"", "", defaults.DefaultBTFFile, nil},
	{defaults.DefaultBTFFile, "", defaults.DefaultBTFFile, nil},
	{"invalid-btf-file", "", "", errors.New("BTF file 'invalid-btf-file' does not exist should fail")},
	{"valid-btf-file", "valid-btf-file", "valid-btf-file", nil},
}

func setupfiles() func(*testing.T, string, ...string) {
	return func(t *testing.T, param string, files ...string) {
		for _, f := range files {
			switch param {
			case "create":
				h, e := os.Create(f)
				require.NoError(t, e)
				h.Close()
			case "remove":
				os.Remove(f)
			}
		}
	}
}

func listBTFFiles() ([]string, error) {
	_, kernelVersion, err := kernels.GetKernelVersion(option.Config.KernelVersion, option.Config.ProcFS)
	if err != nil {
		return nil, err
	}

	_, testFname, _, _ := runtime.Caller(0)
	testdataPath := filepath.Join(filepath.Dir(testFname), "..", "..", "testdata")

	btfFiles := []string{
		defaults.DefaultBTFFile,
		path.Join(defaults.DefaultTetragonLib, "btf"),
		path.Join(defaults.DefaultTetragonLib, "metadata", "vmlinux-"+kernelVersion),
		filepath.Join(testdataPath, "btf", "vmlinux-5.4.104+"),
	}

	return btfFiles, nil
}

func TestObserverFindBTF(t *testing.T) {
	tmpdir := t.TempDir()

	old := os.Getenv("TETRAGON_BTF")
	defer os.Setenv("TETRAGON_BTF", old)

	handlefiles := setupfiles()
	for _, test := range testBTFFiles {
		if test.create != "" {
			handlefiles(t, "create", test.create, filepath.Join(tmpdir, test.create))
			defer handlefiles(t, "remove", test.create, filepath.Join(tmpdir, test.create))
		}

		_, err := os.Stat(defaults.DefaultBTFFile)
		if err != nil && test.wantbtf == defaults.DefaultBTFFile {
			continue
		}

		btf, err := observerFindBTF(tmpdir, test.btf)
		if test.err != nil {
			require.Errorf(t, err, "observerFindBTF() on '%s'  -  want:%v  -  got:no error", test.btf, test.err)
			continue
		}
		require.NoErrorf(t, err, "observerFindBTF() on '%s'  - want:no error  -  got:%v", test.btf, err)
		assert.Equalf(t, test.wantbtf, btf, "observerFindBTF() on '%s'  -  want:'%s'  -  got:'%s'", test.btf, test.wantbtf, btf)

		// Test now without lib set
		btf, err = observerFindBTF("", test.btf)
		if test.err != nil {
			require.Errorf(t, err, "observerFindBTF() on '%s'  -  want:%v  -  got:no error", test.btf, test.err)
			continue
		}
		require.NoErrorf(t, err, "observerFindBTF() on '%s'  -   want:no error  -  got:%v", test.btf, err)
		assert.Equalf(t, test.wantbtf, btf, "observerFindBTF() on '%s'  -  want:'%s'  -  got:'%s'", test.btf, test.wantbtf, btf)
	}
}

func TestObserverFindBTFEnv(t *testing.T) {
	old := os.Getenv("TETRAGON_BTF")
	defer os.Setenv("TETRAGON_BTF", old)

	lib := defaults.DefaultTetragonLib
	btffile := defaults.DefaultBTFFile
	_, err := os.Stat(btffile)
	if err != nil {
		/* No default vmlinux file */
		btf, err := observerFindBTF("", "")
		if old != "" {
			require.NoError(t, err)
			assert.NotEmpty(t, btf)
		} else {
			require.Error(t, err)
			assert.Empty(t, btf)
		}
		/* Let's clear up environment vars */
		os.Setenv("TETRAGON_BTF", "")
		btf, err = observerFindBTF("", "")
		require.Error(t, err)
		assert.Empty(t, btf)

		/* Let's try provided path to lib but tests put the btf inside /boot/ */
		btf, err = observerFindBTF(lib, "")
		require.Error(t, err)
		assert.Empty(t, btf)

		/* Let's try out the btf file that is inside /boot/ */
		var uname unix.Utsname
		err = unix.Uname(&uname)
		require.NoError(t, err)
		kernelVersion := unix.ByteSliceToString(uname.Release[:])
		os.Setenv("TETRAGON_BTF", filepath.Join("/boot/", "btf-"+kernelVersion))
		btf, err = observerFindBTF(lib, "")
		require.NoError(t, err)
		assert.NotEmpty(t, btf)

		btffile = btf
		err = os.Setenv("TETRAGON_BTF", btffile)
		require.NoError(t, err)
		btf, err = observerFindBTF(lib, "")
		require.NoError(t, err)
		assert.Equal(t, btffile, btf)
	} else {
		btf, err := observerFindBTF("", "")
		require.NoError(t, err)
		assert.Equal(t, btffile, btf)

		err = os.Setenv("TETRAGON_BTF", btffile)
		require.NoError(t, err)
		btf, err = observerFindBTF(lib, "")
		require.NoError(t, err)
		assert.Equal(t, btffile, btf)
	}

	/* Following should fail */
	err = os.Setenv("TETRAGON_BTF", "invalid-btf-file")
	require.NoError(t, err)
	btf, err := observerFindBTF(lib, "")
	require.Error(t, err)
	assert.Empty(t, btf)
}

func TestInitCachedBTF(t *testing.T) {
	_, err := os.Stat(defaults.DefaultBTFFile)
	if err != nil {
		btffile := os.Getenv("TETRAGON_BTF")
		err = InitCachedBTF(defaults.DefaultTetragonLib, "")
		if btffile != "" {
			require.NoError(t, err)
			file := GetCachedBTFFile()
			assert.Equal(t, btffile, file, "GetCachedBTFFile()  -  want:'%s'  - got:'%s'", btffile, file)
		} else {
			require.Error(t, err)
		}
	} else {
		err = InitCachedBTF(defaults.DefaultTetragonLib, "")
		require.NoError(t, err)

		btffile := GetCachedBTFFile()
		assert.Equal(t, defaults.DefaultBTFFile, btffile, "GetCachedBTFFile()  -  want:'%s'  - got:'%s'", defaults.DefaultBTFFile, btffile)
	}
}

func genericTestFindBTFFuncParamFromHook(t *testing.T, spec *btf.Spec, hook string, argIndex int, expectedName string) error {
	param, err := findBTFFuncParamFromHookWithSpec(spec, hook, argIndex)
	if err != nil {
		return err
	}

	assert.NotNil(t, param)
	assert.Equal(t, expectedName, param.Name)

	return nil
}

func testFindBTFFuncParamFromHook(btfFName string) func(*testing.T) {
	return func(t *testing.T) {
		spec, err := btf.LoadSpec(btfFName)
		if err != nil {
			t.Skipf("%q not found", btfFName)
		}

		// Assert no errors on Kprobe
		hook := "wake_up_new_task"
		argIndex := 0
		expectedName := "p"
		err = genericTestFindBTFFuncParamFromHook(t, spec, hook, argIndex, expectedName)
		require.NoError(t, err)

		// Assert error raises with invalid hook
		hook = "fake_hook"
		argIndex = 0
		expectedName = "p"
		err = genericTestFindBTFFuncParamFromHook(t, spec, hook, argIndex, expectedName)
		require.ErrorContains(t, err, fmt.Sprintf("failed to find BTF type for hook %q", hook))

		// Assert error raises when hook is a valid BTF type but not btf.Func
		hook = "linux_binprm"
		argIndex = 0
		expectedName = "p"
		err = genericTestFindBTFFuncParamFromHook(t, spec, hook, argIndex, expectedName)
		require.ErrorContains(t, err, fmt.Sprintf("failed to find BTF type for hook %q", hook))

		// Assert error raises when argIndex is out of scope
		hook = "wake_up_new_task"
		argIndex = 10
		expectedName = "p"
		err = genericTestFindBTFFuncParamFromHook(t, spec, hook, argIndex, expectedName)
		require.ErrorContains(t, err, fmt.Sprintf("index %d is out of range", argIndex))
	}
}

func TestFindBTFFuncParamFromHook(t *testing.T) {
	btfFiles, err := listBTFFiles()
	fatalOnError(t, err)

	for _, btfFile := range btfFiles {
		t.Run(btfFile, testFindBTFFuncParamFromHook(btfFile))
	}
}

func fatalOnError(t *testing.T, err error) {
	if err != nil {
		require.Error(t, err)
		t.Fatal(err.Error())
	}
}

func getBTFStruct(ty btf.Type) (*btf.Struct, error) {
	t, ok := ty.(*btf.Struct)
	if ok {
		return t, nil
	}
	return nil, fmt.Errorf("Invalid type for \"%v\", expected \"*btf.Struct\", got %q", t, reflect.TypeOf(ty).String())
}

func getBTFPointer(ty btf.Type) (*btf.Pointer, error) {
	t, ok := ty.(*btf.Pointer)
	if ok {
		return t, nil
	}
	return nil, fmt.Errorf("Invalid type for \"%v\", expected \"*btf.Pointer\", got %q", t, reflect.TypeOf(ty).String())
}

func findMemberInBTFStruct(structTy *btf.Struct, memberName string) (*btf.Member, error) {
	for _, member := range structTy.Members {
		if member.Name == memberName {
			return &member, nil
		}

		if anonymousStructTy, ok := member.Type.(*btf.Struct); ok && len(member.Name) == 0 {
			for _, m := range anonymousStructTy.Members {
				if m.Name == memberName {
					return &m, nil
				}
			}
		}

		if unionTy, ok := member.Type.(*btf.Union); ok {
			for _, m := range unionTy.Members {
				if m.Name == memberName {
					return &m, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("Member %q not found in struct %v", memberName, structTy)
}

func getBTFPointerAndSetConfig(ty btf.Type, btfConfig *api.ConfigBTFArg) (*btf.Pointer, error) {
	ptr, err := getBTFPointer(ty)
	if err != nil {
		return nil, err
	}
	btfConfig.IsInitialized = uint16(1)
	btfConfig.IsPointer = uint16(1)

	return ptr, nil
}

func getConfigAndNextType(structTy *btf.Struct, memberName string) (*btf.Type, *api.ConfigBTFArg, error) {
	btfConfig := api.ConfigBTFArg{}

	member, err := findMemberInBTFStruct(structTy, memberName)
	if err != nil {
		return nil, nil, err
	}

	btfConfig.Offset = uint32(member.Offset.Bytes())
	btfConfig.IsInitialized = uint16(1)

	ty := ResolveNestedTypes(member.Type)

	ptr, _ := getBTFPointerAndSetConfig(ty, &btfConfig)
	if ptr != nil {
		return &ptr.Target, &btfConfig, err
	}
	if _, ok := ty.(*btf.Int); ok {
		btfConfig.IsPointer = uint16(1)
	}
	return &ty, &btfConfig, err
}

func getConfigAndNextStruct(structTy *btf.Struct, memberName string) (*btf.Struct, *api.ConfigBTFArg, error) {
	btfConfig := api.ConfigBTFArg{}

	member, err := findMemberInBTFStruct(structTy, memberName)
	if err != nil {
		return nil, nil, err
	}

	btfConfig.Offset = uint32(member.Offset.Bytes())
	btfConfig.IsInitialized = uint16(1)

	ty := ResolveNestedTypes(member.Type)

	ptr, _ := getBTFPointerAndSetConfig(ty, &btfConfig)
	if ptr != nil {
		t, err := getBTFStruct(ptr.Target)
		return t, &btfConfig, err
	}
	t, err := getBTFStruct(ty)
	return t, &btfConfig, err
}

func addPaddingOnNestedPtr(ty btf.Type, path []string) []string {
	if t, ok := ty.(*btf.Pointer); ok {
		updatedPath := append([]string{""}, path...)
		return addPaddingOnNestedPtr(t.Target, updatedPath)
	}
	return path
}

func resolveNestedPtr(rootType btf.Type, btfArgs *[api.MaxBTFArgDepth]api.ConfigBTFArg, i int) (btf.Type, int) {
	if ptr, ok := rootType.(*btf.Pointer); ok {
		btfArgs[i] = api.ConfigBTFArg{}
		ty, err := getBTFPointerAndSetConfig(ptr, &btfArgs[i])
		if err != nil {
			return ty.Target, i
		}
		return resolveNestedPtr(ty.Target, btfArgs, i+1)
	}
	return rootType, i
}

func manuallyResolveBTFPath(t *testing.T, rootType btf.Type, p []string) [api.MaxBTFArgDepth]api.ConfigBTFArg {
	var btfArgs [api.MaxBTFArgDepth]api.ConfigBTFArg
	var i int

	rootType, i = resolveNestedPtr(rootType, &btfArgs, 0)

	currentStruct, err := getBTFStruct(rootType)
	fatalOnError(t, err)

	for ; i < len(p); i++ {
		step := p[i]
		if len(step) == 0 {
			btfArgs[i] = api.ConfigBTFArg{}
			ptr, err := getBTFPointerAndSetConfig(ResolveNestedTypes(rootType), &btfArgs[i])
			fatalOnError(t, err)
			currentStruct, err = getBTFStruct(ptr.Target)
			fatalOnError(t, err)
		} else if i < len(p)-1 {
			ty, nextConfig, err := getConfigAndNextStruct(currentStruct, step)
			fatalOnError(t, err)
			currentStruct = ty
			btfArgs[i] = *nextConfig
		} else {
			_, nextConfig, err := getConfigAndNextType(currentStruct, step)
			fatalOnError(t, err)
			btfArgs[i] = *nextConfig
			return btfArgs
		}
	}
	return btfArgs
}

func buildPathFromString(t *testing.T, rootType btf.Type, pathStr string) []string {
	pathBase := strings.Split(pathStr, ".")
	path := addPaddingOnNestedPtr(rootType, pathBase)
	if len(path) > api.MaxBTFArgDepth {
		t.Errorf("Unable to resolve %q. The maximum depth allowed is %d", pathStr, api.MaxBTFArgDepth)
	}
	return path
}

func buildResolveBTFConfig(t *testing.T, rootType btf.Type, pathStr string) [api.MaxBTFArgDepth]api.ConfigBTFArg {
	var btfArgs [api.MaxBTFArgDepth]api.ConfigBTFArg

	path := buildPathFromString(t, rootType, pathStr)
	_, err := ResolveBTFPath(&btfArgs, rootType, path, 0)
	fatalOnError(t, err)

	return btfArgs
}

func buildExpectedBTFConfig(t *testing.T, rootType btf.Type, pathStr string) [api.MaxBTFArgDepth]api.ConfigBTFArg {
	path := buildPathFromString(t, rootType, pathStr)
	return manuallyResolveBTFPath(t, rootType, path)
}

func testPathIsAccessible(rootType btf.Type, strPath string) (*[api.MaxBTFArgDepth]api.ConfigBTFArg, *btf.Type, error) {
	var btfArgs [api.MaxBTFArgDepth]api.ConfigBTFArg
	path := strings.Split(strPath, ".")

	lastBTFType, err := ResolveBTFPath(&btfArgs, ResolveNestedTypes(rootType), path, 0)
	if err != nil {
		return nil, nil, err
	}

	return &btfArgs, lastBTFType, nil
}

func testAssertEqualPath(spec *btf.Spec) func(*testing.T) {
	return func(t *testing.T) {
		hook := "security_bprm_check"
		argIndex := 0 // struct linux_binprm *bprm
		funcParamTy, err := findBTFFuncParamFromHookWithSpec(spec, hook, argIndex)
		fatalOnError(t, err)

		bprmTy := funcParamTy.Type
		if ty, ok := bprmTy.(*btf.Pointer); ok {
			bprmTy = ty.Target
		}

		// Test default behaviour
		path := "file.f_path.dentry.d_name.name"
		assert.Equal(t,
			buildExpectedBTFConfig(t, bprmTy, path),
			buildResolveBTFConfig(t, bprmTy, path),
		)

		// Test anonymous struct
		path = "mm.arg_start"
		assert.Equal(t,
			buildExpectedBTFConfig(t, bprmTy, path),
			buildResolveBTFConfig(t, bprmTy, path),
		)

		// Test Union
		path = "file.f_inode.i_dir_seq"
		assert.Equal(t,
			buildExpectedBTFConfig(t, bprmTy, path),
			buildResolveBTFConfig(t, bprmTy, path),
		)

		// Test if param is double ptr
		hook = "security_inode_copy_up"
		argIndex = 1 // struct cred **new
		funcParamTy, err = findBTFFuncParamFromHookWithSpec(spec, hook, argIndex)
		fatalOnError(t, err)

		newTy := funcParamTy.Type
		if ty, ok := newTy.(*btf.Pointer); ok {
			newTy = ty.Target
		}
		path = "uid.val"
		assert.Equal(t,
			buildExpectedBTFConfig(t, newTy, path),
			buildResolveBTFConfig(t, newTy, path),
		)
	}
}

func testAssertPathIsAccessible(spec *btf.Spec) func(*testing.T) {
	return func(t *testing.T) {
		hook := "wake_up_new_task"
		argIndex := 0 //struct task_struct *p
		funcParamTy, err := findBTFFuncParamFromHookWithSpec(spec, hook, argIndex)
		fatalOnError(t, err)

		taskStructTy := funcParamTy.Type
		if ty, ok := taskStructTy.(*btf.Pointer); ok {
			taskStructTy = ty.Target
		}

		_, _, err = testPathIsAccessible(taskStructTy, "sched_task_group.css.id")
		require.NoError(t, err)

		hook = "security_bprm_check"
		argIndex = 0 // struct linux_binprm *bprm
		funcParamTy, err = findBTFFuncParamFromHookWithSpec(spec, hook, argIndex)
		fatalOnError(t, err)

		bprmTy := funcParamTy.Type
		if ty, ok := bprmTy.(*btf.Pointer); ok {
			bprmTy = ty.Target
		}

		_, _, err = testPathIsAccessible(bprmTy, "mm.pgd.pgd")
		require.NoError(t, err)
	}
}

func testAssertErrorOnInvalidPath(spec *btf.Spec) func(*testing.T) {
	return func(t *testing.T) {
		hook := "security_bprm_check"
		argIndex := 0 // struct linux_binprm *bprm
		funcParamTy, err := findBTFFuncParamFromHookWithSpec(spec, hook, argIndex)
		fatalOnError(t, err)

		rootType := funcParamTy.Type
		if rootTy, ok := rootType.(*btf.Pointer); ok {
			rootType = rootTy.Target
		}

		// Assert an error is raised when attribute does not exists
		_, _, err = testPathIsAccessible(rootType, "fail")
		require.ErrorContains(t, err, "attribute \"fail\" not found in structure")

		_, _, err = testPathIsAccessible(rootType, "mm.fail")
		require.ErrorContains(t, err, "attribute \"fail\" not found in structure")

		_, _, err = testPathIsAccessible(rootType, "mm.pgd.fail")
		require.ErrorContains(t, err, "attribute \"fail\" not found in structure")

		hook = "do_sys_open"
		argIndex = 0 // int dfd
		funcParamTy, err = findBTFFuncParamFromHookWithSpec(spec, hook, argIndex)
		fatalOnError(t, err)

		rootType = funcParamTy.Type

		// Assert an error is raised when attribute has invalid type
		_, _, err = testPathIsAccessible(rootType, "fail")
		require.ErrorContains(t, err, fmt.Sprintf("unexpected type : \"fail\" has type %q", rootType.TypeName()))
	}
}

func testResolveBTFPath(btfFName string) func(t *testing.T) {
	return func(t *testing.T) {
		spec, err := btf.LoadSpec(btfFName)
		if err != nil {
			t.Skipf("%q not found", btfFName)
		}
		t.Run("PathIsAccessible", testAssertPathIsAccessible(spec))
		t.Run("AssertErrorOnInvalidPath", testAssertErrorOnInvalidPath(spec))
		t.Run("AssertEqualPath", testAssertEqualPath(spec))
	}
}

func TestResolveBTFPath(t *testing.T) {
	btfFiles, err := listBTFFiles()
	fatalOnError(t, err)

	for _, btfFile := range btfFiles {
		t.Run(btfFile, testResolveBTFPath(btfFile))
	}
}
