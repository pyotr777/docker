package git

import (
	"bufio"
	"bytes"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver"
	//"github.com/docker/docker/daemon/graphdriver/aufs"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/chrootarchive"
	"github.com/docker/docker/pkg/directory"
	"github.com/docker/docker/pkg/idtools"
	"os"
	"os/exec"
	"path"
	"strings"
)

const gitDir string = "git"

func init() {
	logrus.Debugf("Executing init")
	graphdriver.Register("git", Init)
}

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
	logrus.Debugf("Executing Init with %s, %s", home, strings.Join(options, " "))
	if err := os.MkdirAll(home, 0700); err != nil {
		return nil, err
	}

	var innerDriverStr string

	if len(options) >= 1 {
		innerDriverStr = options[0]
	} else {
		innerDriverStr = "aufs"
	}

	innerDriver, err := graphdriver.GetDriver(innerDriverStr, home, options, uidMaps, gidMaps)
	//driver, err := aufs.Init(home, options, uidMaps, gidMaps)
	if err != nil {
		return nil, err
	}

	d := &Driver{
		home:        home,
		innerDriver: innerDriver,
		uidMaps:     uidMaps,
		gidMaps:     gidMaps,
	}

	return d, nil
}

// We want to change AUFS-specific functions of graphDriver, so we have to
// implement the whole set of AUFS features required for the functions we need.
// Here we added uid and gid mappings used in AUFS.
type Driver struct {
	home        string
	innerDriver graphdriver.Driver
	uidMaps     []idtools.IDMap
	gidMaps     []idtools.IDMap
}

func (d *Driver) rootPath() string {
	return d.home
}

// Return absolute path to git repository.
// Creaete if the directory doesn't exist.
func (d *Driver) gitPath() (string, error) {
	git_path, err := path.Join(d.rootPath(), gitDir)
	if err != nil {
		logrus.Errorf("Cannot join %s and %s", d.rootPath(), gitDir)
		return nil, err
	}
	_, err = os.Stat(git_path)
	if err != nil {
		if os.IsNotExist(err) { // Directory doesn't exist. Create.
			if err := os.MkdirAll(git_path, 0755); err != nil {
				logrus.Errorf("Cannot create directory %s", git_path)
				return nil, err
			}
			logrus.Debugf("Directory created: %s", git_path)
		} else {
			return nil, err
		}
	}
	return git_path, nil
}

// Return string with git-dir and work-tree options for calling git
func (d *Driver) getGitPaths(id string) (string, error) {
	work_dir, err := d.getAUFSpath(id)
	if err != nil {
		logrus.Errorf("Cannot get AUFS dirs for %s", id)
		return nil, err
	}
	git_path, err = d.gitPath()
	if err != nil {
		logrus.Errorf("Cannot get git dirs for %s", id)
		return nil, err
	}
	return strings.Join([]string{"--git-dir ", git_path, "--work-tree", work_dir}, ""), nil
}

// Call AUFS to apply diff
func (d *Driver) ApplyDiff(id, parent string, diff archive.Reader) (size int64, err error) {
	logrus.Debugf("(git) Executing ApplyDiff with %s, %s", id, parent)
	return d.innerDriver.ApplyDiff(id, parent, diff)
}

// Use git logic for obtaining changes.
// Can use AUFS also.
func (d *Driver) Changes(id, parent string) ([]archive.Change, error) {
	logrus.Debugf("(git) Executing Changes with %s, %s", id, parent)
	// Inner driver (AUFS) changes
	changes, err := d.innerDriver.Changes(id, parent)
	// Get changes from git
	commits_arr := []string{parent, "..", id}
	commits := strings.Join(commits_arr, "")
	git_paths, err := d.getGitPaths()
	if err != nil {
		return nil, err
	}
	logrus.Debugf("(git) Execute command: git %s diff --name-status %s", git_paths, commits)
	output, err := exec.Command("git", git_paths, "diff", "--name-status", commits).CombinedOutput()
	if err != nil {
		logrus.Errorf("Error trying to take diff from GIT repository: %s (%s)", err, output)
		return nil, err
	}
	// convert output []byte to output_s string
	n := bytes.IndexByte(output, 0) // length
	output_s := string(output[:n])
	// output_array - output of "git diff" split by lines
	output_array := strings.Split(output_s, "\n")

	// Loop through lines
	var change archive.Change
	var git_changes []archive.Change
	for _, line := range output_array {
		var change_kind string
		// Split line by space. First element - modification, second - relative path
		line_split := strings.Split(line, " ")
		change_kind = line_split[0]
		change.Path = line_split[1]
		// Set change.kind to int depending on change_kind string (letter)
		switch change_kind {
		case "M":
			change.Kind = 0
		case "A":
			change.Kind = 1
		case "D":
			change.Kind = 2
		}
		git_changes = append(git_changes, change)
	}

	// Compare output from AUFS and git
	logrus.Debug("AUFS changes:")
	for _, change := range changes {
		logrus.Debug(change)
	}
	logrus.Debug("")
	logrus.Debug("GIT changes:")
	for _, change := range git_changes {
		logrus.Debug(change)
	}

	return git_changes, err
}

// TODO Use git logic insead of inner driver
func (d *Driver) Diff(id, parent string) (archive.Archive, error) {
	logrus.Debugf("Executing Diff with %s, %s", id, parent)
	return d.innerDriver.Diff(id, parent)
}

// Use gitDir directory
func (d *Driver) DiffSize(id, parent string) (size int64, err error) {
	logrus.Debugf("(git) Executing DiffSize with %s, %s", id, parent)
	return d.innerDriver.DiffSize(id, parent)
}

// TODO Use git logic insead of inner driver
func (d *Driver) GetMetadata(id string) (map[string]string, error) {
	logrus.Debugf("Executing GetMetadata with %s", id)
	return d.innerDriver.GetMetadata(id)
}

func (d *Driver) String() string {
	return "git"
}

func (d *Driver) Status() [][2]string {
	logrus.Debugf("Executing Status")
	return d.innerDriver.Status()
}

func (d *Driver) Cleanup() error {
	logrus.Debugf("Executing Cleanup")
	return d.innerDriver.Cleanup()
}

// Use same logic as in aufs.Get(),
// but do not use reference count.
// Use it instead of aufs.Get() when do not need to create AUFS mount.
func (d *Driver) getAUFSpath(id string) (string, error) {
	parents, err := d.getParentLayerPaths(id)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// If a dir does not have a parent ( no layers )do not try to mount
	// just return the diff path to the data
	layer_path := path.Join(d.rootPath(), "diff", id)
	if len(parents) > 0 {
		layer_path = path.Join(d.rootPath(), "mnt", id)
	}
	return layer_path, nil
}

func (d *Driver) getParentLayerPaths(id string) ([]string, error) {
	parentIds, err := getParentIds(d.rootPath(), id)
	if err != nil {
		return nil, err
	}
	layers := make([]string, len(parentIds))

	// Get the diff paths for all the parent ids
	for i, p := range parentIds {
		layers[i] = path.Join(d.rootPath(), "diff", p)
	}
	return layers, nil
}

// Read the layers file for the current id and return all the
// layers represented by new lines in the file
//
// If there are no lines in the file then the id has no parent
// and an empty slice is returned.
//
// Function copied from aufs/dirs.go
func getParentIds(root, id string) ([]string, error) {
	f, err := os.Open(path.Join(root, "layers", id))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := []string{}
	s := bufio.NewScanner(f)

	for s.Scan() {
		if t := s.Text(); t != "" {
			out = append(out, s.Text())
		}
	}
	return out, s.Err()
}

// Created directories for AUFS and git
func (d *Driver) Create(id, parent string, mountLabel string, storageOpt map[string]string) error {
	logrus.Debugf("Executing Create with %s, %s, %s", id, parent, mountLabel)
	if err := d.innerDriver.Create(id, parent, mountLabel, storageOpt); err != nil {
		return err
	}
	dirStr, err := d.getAUFSpath(id)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Error trying to get the current directory: (%s)", err)
		return err
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(d.rootPath()); err != nil {
		logrus.Errorf("Error trying to change to %s (%s)", d.rootPath(), err)
		return err
	}

	if output, err := exec.Command("git", git_options, "add", "-A").CombinedOutput(); err != nil {
		logrus.Errorf("Error trying to add files to GIT repository: %s (%s)", err, output)
		return err
	}
	// Init git repository in rootPath/gitDir
	git_path := d.gitPath()
	if err != nil {
		return err
	}
	logrus.Debugf("Initialise GIT in %s", git_path)
	git_options, err := d.getGitPaths(id)
	if err != nil {
		return err
	}
	if output, err := exec.Command("git", git_options, "init").CombinedOutput(); err != nil {
		logrus.Errorf("Error trying to init GIT repository: %s (%s)", err, output)
		return nil
	}
	if output, err := exec.Command("git", git_options, "config", "user.email", "docker@example.com").CombinedOutput(); err != nil {
		logrus.Errorf("Error trying to configure user.email on GIT repository: %s (%s)", err, output)
		return err
	}

	if output, err := exec.Command("git", git_options, "config", "user.name", "Docker").CombinedOutput(); err != nil {
		logrus.Errorf("Error trying to configure user.name on GIT repository: %s (%s)", err, output)
		return err
	}

	return nil
}

func (d *Driver) CreateAndMerge(id, parent1, parent2 string) error {
	logrus.Debugf("Executing CreateAndMerge with %s, %s, %s", id, parent1, parent2)
	if err := d.Create(id, parent1, "", nil); err != nil {
		return err
	}
	dirStr, err := d.innerDriver.Get(id, "")
	if err != nil {
		return err
	}
	defer d.innerDriver.Put(id)

	parent2dirStr, err := d.innerDriver.Get(parent2, "")
	if err != nil {
		return err
	}
	defer d.innerDriver.Put(parent2)

	cwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Error trying to get the current directory: (%s)", err)
		return err
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(dirStr); err != nil {
		logrus.Errorf("Error trying to change the current directory to %s (%s)", dirStr, err)
		return err
	}

	if output, err := exec.Command("git", "pull", parent2dirStr).CombinedOutput(); err != nil {
		logrus.Errorf("Error trying to pull GIT repository: %s (%s)", err, output)
		return err
	}

	return nil
}

func (d *Driver) Remove(id string) error {
	logrus.Debugf("Executing Remove with %s", id)
	return d.innerDriver.Remove(id)
}

func (d *Driver) Get(id, mountLabel string) (string, error) {
	logrus.Debugf("Executing Get with %s, %s", id, mountLabel)
	dirStr, err := d.innerDriver.Get(id, mountLabel)
	if err != nil {
		return dirStr, err
	}
	//return dirStr, nil
	logrus.Debug("(git) Get returned %s", dirStr)
	return dirStr, nil
}

func (d *Driver) Put(id string) error {
	logrus.Debugf("(git) Executing Put with %s", id)
	git_options, err := d.getGitPaths(id)
	if err != nil {
		logrus.Errorf("Error trying to Get the directory for %s (%s)", id, err)
		return err
	}
	// DO NOT use gitDir path for creating git commit
	// X dirStr = path.Join(dirStr, gitDir) - Do not do it. Git repository will be created in this directory
	defer d.innerDriver.Put(id)

	cwd, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Error trying to get the current directory: (%s)", err)
		return err
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(d.rootPath()); err != nil {
		logrus.Errorf("Error trying to change to %s (%s)", d.rootPath(), err)
		return err
	}

	if output, err := exec.Command("git", git_options, "add", "-A").CombinedOutput(); err != nil {
		logrus.Errorf("Error trying to add files to GIT repository: %s (%s)", err, output)
		return err
	}

	if output, err := exec.Command("git", git_options, "commit", "-m", "Commit by Docker: "+id).CombinedOutput(); err != nil {
		logrus.Debugf("Error trying to commit GIT repository: %s (%s)", err, output)
		// Skip "nothing to commit errors"
		if strings.Contains(string(output), "nothing to commit") {
			// Substring "nothing to commit" is in output
			logrus.Debug("--- Diagnostic info:")
			logrus.Debugf("Current directory: %s", dirStr)
			output_c, err_c := exec.Command("ls", "-la").CombinedOutput()
			if err_c != nil {
				logrus.Errorf("Cannot list files")
			}
			logrus.Debug(string(output_c))
			logrus.Debugf("ls -la %s", gitDir)
			output_c, err_c = exec.Command("ls", "-la", gitDir).CombinedOutput()
			if err_c != nil {
				logrus.Errorf("Cannot list files")
			}
			logrus.Debug(string(output_c))

			output_c, err_c = exec.Command("git", "status").CombinedOutput()
			if err_c != nil {
				logrus.Errorf("Cannot execute git status")
			}
			logrus.Debug(string(output_c))
			output_c, err_c = exec.Command("git", "log", "--pretty=format:'%h %cd | %s %d'", "--date=short").CombinedOutput()
			if err_c != nil {
				logrus.Errorf("Cannot execute git log")
			}
			logrus.Debug(string(output_c))
			logrus.Debugf("Came from: %s", cwd)
			logrus.Debug("--- end diagnostic info")
			return nil
		} else {
			return err
		}
	}
	logrus.Debugf("Created commit from %s with options %s", d.rootPath(), d.getGitPaths(id))
	return nil
}

func (d *Driver) Exists(id string) bool {
	logrus.Debugf("Executing Exists with %s", id)
	return d.innerDriver.Exists(id)
}
