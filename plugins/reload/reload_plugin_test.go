package reload

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/roadrunner-server/config/v2"
	endure "github.com/roadrunner-server/endure/pkg/container"
	"github.com/roadrunner-server/errors"
	httpPlugin "github.com/roadrunner-server/http/v2"
	"github.com/roadrunner-server/logger/v2"
	"github.com/roadrunner-server/reload/v2"
	"github.com/roadrunner-server/resetter/v2"
	mocklogger "github.com/roadrunner-server/rr-e2e-tests/mock"
	"github.com/roadrunner-server/server/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testDir string = "unit_tests"
const testCopyToDir string = "unit_tests_copied"
const dir1 string = "dir1"
const hugeNumberOfFiles uint = 500

func TestReloadInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-reload.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	t.Run("ReloadTestInit", reloadTestInit)
	time.Sleep(time.Second * 3)
	stopCh <- struct{}{}
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("http server was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("file was created").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("file was added to watcher").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("reset signal was received").Len())

	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
	})
}

func reloadTestInit(t *testing.T) {
	err := ioutil.WriteFile(filepath.Join(testDir, "file.txt"), //nolint:gosec
		[]byte{}, 0755)
	assert.NoError(t, err)
}

func TestReloadBadWorker(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-reload-bad-worker.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Error(t, e.Error)
				return
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)

	orig, err := os.ReadFile("../../php_test_files/psr-worker-bench-reload.php")
	require.NoError(t, err)

	f, err := os.OpenFile("../../php_test_files/psr-worker-bench-reload.php", os.O_RDWR, os.ModePerm)
	require.NoError(t, err)
	defer func() {
		_ = f.Truncate(0)
		_, _ = f.Seek(0, 0)
		_, _ = f.Write(orig)
		_ = f.Close()
	}()

	_, err = f.Write([]byte("foo"))
	require.NoError(t, err)

	t.Run("ReloadTestInit", reloadTestInit)
	time.Sleep(time.Second * 3)
	wg.Wait()

	require.Equal(t, 1, oLogger.FilterMessageSnippet("http server was started").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("file was created").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("file was added to watcher").Len())
	require.Equal(t, 1, oLogger.FilterMessageSnippet("reset signal was received").Len())

	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
	})
}

func TestReloadHugeNumberOfFiles(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-reload.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))

	assert.NoError(t, os.Mkdir(testDir, 0755))
	assert.NoError(t, os.Mkdir(testCopyToDir, 0755))

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	t.Run("ReloadTestHugeNumberOfFiles", reloadHugeNumberOfFiles)
	t.Run("ReloadRandomlyChangeFile", randomlyChangeFile)
	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()

	require.GreaterOrEqual(t, oLogger.Len(), 500)
	require.Equal(t, 1, oLogger.FilterMessageSnippet("http server was started").Len())

	require.Greater(t, oLogger.FilterMessageSnippet("file was created").Len(), 1)
	require.Greater(t, oLogger.FilterMessageSnippet("file was updated").Len(), 1)
	require.Greater(t, oLogger.FilterMessageSnippet("file was added to watcher").Len(), 10)
	require.GreaterOrEqual(t, oLogger.FilterMessageSnippet("reset signal was received").Len(), 1)

	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
		_ = os.RemoveAll(testCopyToDir)
	})
}

func randomlyChangeFile(t *testing.T) {
	// we know, that directory contains 500 files (0-499)
	// let's try to randomly change it
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(500) //nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(hugeNumberOfFiles))                                                                            //nolint:gosec
		err := ioutil.WriteFile(filepath.Join(testDir, "file_"+strconv.Itoa(int(rNum))+".txt"), []byte("Hello, Gophers!"), 0755) //nolint:gosec
		assert.NoError(t, err)
	}
}

func reloadHugeNumberOfFiles(t *testing.T) {
	for i := uint(0); i < hugeNumberOfFiles; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
}

// Should be events only about creating files with txt ext
func TestReloadFilterFileExt(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-reload-2.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, os.Mkdir(testDir, 0755))

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	t.Run("ReloadMakeFiles", reloadMakeFiles)
	time.Sleep(time.Second * 2)
	t.Run("ReloadFilteredExt", reloadFilteredExt)
	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()

	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
	})
}

func reloadMakeFiles(t *testing.T) {
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".abc"))
	}
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".def"))
	}
}

func reloadFilteredExt(t *testing.T) {
	// change files with abc extension
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) //nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(hugeNumberOfFiles))                                                                            //nolint:gosec
		err := ioutil.WriteFile(filepath.Join(testDir, "file_"+strconv.Itoa(int(rNum))+".abc"), []byte("Hello, Gophers!"), 0755) //nolint:gosec
		assert.NoError(t, err)
	}

	// change files with def extension
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) //nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(hugeNumberOfFiles))                                                                            //nolint:gosec
		err := ioutil.WriteFile(filepath.Join(testDir, "file_"+strconv.Itoa(int(rNum))+".def"), []byte("Hello, Gophers!"), 0755) //nolint:gosec
		assert.NoError(t, err)
	}
}

// Should be events only about creating files with txt ext
func TestReloadCopy100(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-reload-3.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))
	assert.NoError(t, freeResources(dir1))

	assert.NoError(t, os.Mkdir(testDir, 0755))
	assert.NoError(t, os.Mkdir(testCopyToDir, 0755))
	assert.NoError(t, os.Mkdir(dir1, 0755))

	l, oLogger := mocklogger.ZapTestLogger(zap.DebugLevel)
	err = cont.RegisterAll(
		cfg,
		l,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	// Scenario
	// 1
	// Create 100 files with txt, abc, def extensions
	// Copy files to the unit_tests_copy dir
	// 2
	// Delete both dirs, recreate
	// Create 100 files with txt, abc, def extensions
	// Move files to the unit_tests_copy dir
	// 3
	// Recursive

	time.Sleep(time.Second * 3)
	t.Run("ReloadMake100Files", reloadMake100Files)
	time.Sleep(time.Second * 2)
	t.Run("ReloadCopyFiles", reloadCopyFiles)
	time.Sleep(time.Second * 2)
	t.Run("ReloadRecursiveDirsSupport", copyFilesRecursive)
	time.Sleep(time.Second * 2)
	t.Run("RandomChangesInRecursiveDirs", randomChangesInRecursiveDirs)
	time.Sleep(time.Second * 2)
	t.Run("RemoveFilesSupport", removeFilesSupport)
	time.Sleep(time.Second * 2)
	t.Run("ReloadMoveSupport", reloadMoveSupport)
	time.Sleep(time.Second * 10)

	t.Cleanup(func() {
		_ = os.RemoveAll(testDir)
		_ = os.RemoveAll(testCopyToDir)
		_ = os.RemoveAll(dir1)
	})

	time.Sleep(time.Second * 5)

	stopCh <- struct{}{}
	wg.Wait()

	require.Greater(t, oLogger.FilterMessageSnippet("file was updated").Len(), 0)
	require.Greater(t, oLogger.FilterMessageSnippet("file added to the list of removed files").Len(), 0)
	require.Greater(t, oLogger.FilterMessageSnippet("file was added to watcher").Len(), 0)
	require.GreaterOrEqual(t, oLogger.FilterMessageSnippet("reset signal was received").Len(), 0)
}

func reloadMoveSupport(t *testing.T) {
	t.Run("MoveSupportCopy", copyFilesRecursive)
	// move some files
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(500) //nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(33)) //nolint:gosec
		rDir := rand.Int63n(9)         //nolint:gosec
		rExt := rand.Int63n(3)         //nolint:gosec

		ext := []string{
			".txt",
			".abc",
			".def",
		}

		// change files with def extension
		dirs := []string{
			"dir1",
			"dir1/dir2",
			"dir1/dir2/dir3",
			"dir1/dir2/dir3/dir4",
			"dir1/dir2/dir3/dir4/dir5",
			"dir1/dir2/dir3/dir4/dir5/dir6",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10",
		}

		// move file
		_ = os.Rename(filepath.Join(dirs[rDir], "file_"+strconv.Itoa(int(rNum))+ext[rExt]), filepath.Join(dirs[rDir+1], "file_"+strconv.Itoa(int(rNum))+ext[rExt]))
	}
}

func removeFilesSupport(t *testing.T) {
	// remove some files
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(500) //nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(100)) //nolint:gosec
		rDir := rand.Int63n(10)         //nolint:gosec
		rExt := rand.Int63n(3)          //nolint:gosec

		ext := []string{
			".txt",
			".abc",
			".def",
		}

		// change files with def extension
		dirs := []string{
			"dir1",
			"dir1/dir2",
			"dir1/dir2/dir3",
			"dir1/dir2/dir3/dir4",
			"dir1/dir2/dir3/dir4/dir5",
			"dir1/dir2/dir3/dir4/dir5/dir6",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10",
		}
		// here can be a situation, when file already deleted
		_ = os.Remove(filepath.Join(dirs[rDir], "file_"+strconv.Itoa(int(rNum))+ext[rExt]))
	}
}

func randomChangesInRecursiveDirs(t *testing.T) {
	// change files with def extension
	dirs := []string{
		"dir1",
		"dir1/dir2",
		"dir1/dir2/dir3",
		"dir1/dir2/dir3/dir4",
		"dir1/dir2/dir3/dir4/dir5",
		"dir1/dir2/dir3/dir4/dir5/dir6",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10",
	}

	ext := []string{
		".txt",
		".abc",
		".def",
	}

	filenames := []string{
		"file_", // should be updated
		"foo_",  // should be created
		"bar_",  // should be created
	}
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(100) //nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(100)) //nolint:gosec
		rDir := rand.Int63n(10)         //nolint:gosec
		rExt := rand.Int63n(3)          //nolint:gosec
		rName := rand.Int63n(3)         //nolint:gosec

		_ = ioutil.WriteFile(filepath.Join(dirs[rDir], filenames[rName]+strconv.Itoa(int(rNum))+ext[rExt]), []byte("Hello, Gophers!"), 0755) //nolint:gosec
	}
}

func copyFilesRecursive(t *testing.T) {
	err := copyDir(testDir, "dir1")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10")
	assert.NoError(t, err)
}

func reloadCopyFiles(t *testing.T) {
	err := copyDir(testDir, testCopyToDir)
	assert.NoError(t, err)

	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))

	assert.NoError(t, os.Mkdir(testDir, 0755))
	assert.NoError(t, os.Mkdir(testCopyToDir, 0755))

	// recreate files
	for i := uint(0); i < 33; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
	for i := uint(0); i < 33; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".abc"))
	}
	for i := uint(0); i < 34; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".def"))
	}

	err = copyDir(testDir, testCopyToDir)
	assert.NoError(t, err)
}

func reloadMake100Files(t *testing.T) {
	for i := uint(0); i < 33; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
	for i := uint(0); i < 33; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".abc"))
	}
	for i := uint(0); i < 34; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".def"))
	}
}

func TestReloadNoRecursion(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Plugin{
		Path:   "configs/.rr-reload-4.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))
	assert.NoError(t, freeResources(dir1))

	assert.NoError(t, os.Mkdir(testDir, 0755))
	assert.NoError(t, os.Mkdir(dir1, 0755))
	assert.NoError(t, os.Mkdir(testCopyToDir, 0755))

	err = cont.RegisterAll(
		cfg,
		&logger.Plugin{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	assert.NoError(t, err)

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	stopCh := make(chan struct{}, 1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-stopCh:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	time.Sleep(time.Second * 3)
	t.Run("ReloadMakeFiles", reloadMakeFiles) // make files in the testDir
	time.Sleep(time.Second * 2)
	t.Run("ReloadCopyFilesRecursive", reloadCopyFiles)
	time.Sleep(time.Second * 3)
	assert.NoError(t, freeResources(testDir))
	time.Sleep(time.Second * 10)

	stopCh <- struct{}{}
	wg.Wait()

	t.Cleanup(func() {
		_ = os.RemoveAll(testCopyToDir)
		_ = os.RemoveAll(dir1)
	})
}

// ========================================================================

func freeResources(path string) error {
	return os.RemoveAll(path)
}

func makeFile(filename string) error {
	return ioutil.WriteFile(filepath.Join(testDir, filename), []byte{}, 0755) //nolint:gosec
}

func copyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return errors.E(errors.Str("source is not a directory"))
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.E(err)
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dst)
	if err != nil {
		return errors.E(err)
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.E(err)
	}

	err = out.Sync()
	if err != nil {
		return errors.E(err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return errors.E(err)
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return errors.E(err)
	}
	return nil
}
