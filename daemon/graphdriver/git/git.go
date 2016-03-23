package git

import (
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/idtools"
	"github.com/docker/docker/pkg/archive"
	"github.com/Sirupsen/logrus"
	"os/exec"
	"os"
	"path"
	"strings"
	"bytes"
)

func init() {
	graphdriver.Register("git", Init)
}

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {
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
	if err != nil {
		return nil, err
	}

	d := &Driver{
		home: home,
		innerDriver: innerDriver,
	}

	return d, nil
}

type Driver struct {
	home string
	innerDriver graphdriver.Driver
}


// TODO Use git logic insead of inner driver
func (d *Driver) ApplyDiff(id, parent string, diff archive.Reader) (size int64, err error) {
	return d.innerDriver.ApplyDiff(id, parent, diff)
}

// TODO Use git logic insead of inner driver
func (d *Driver) Changes(id, parent string) ([]archive.Change, error) {
	// Inner driver (AUFS)
	changes, err := d.innerDriver.Changes(id, parent)
	// GIT
	commits_arr := []string{ parent,"..", id}
	commits := strings.Join(commits_arr,"")
	output, err := exec.Command("git", "diff", "--name-status", commits).CombinedOutput()
	if err != nil {
		logrus.Error("Error trying to take diff from GIT repository: %s (%s)", err, output)
		return nil, err
	}
	// convert output []byte to output_s string
	n := bytes.IndexByte(output, 0) // length
	output_s := string(output[:n]) 
	// output_array - output of "git diff" split by lines
	output_array := strings.Split(output_s,"\n")
	// Loop through lines
	var change archive.Change
	for _, line := range output_array {
		var change_kind string			
		// Split line by space. First element - modification, second - relative path
		line_split := strings.Split(line," ")
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
		changes = append(changes, change)
	}
	return changes, err
}

// TODO Use git logic insead of inner driver
func (d *Driver) Diff(id, parent string) (archive.Archive, error) {
	return d.innerDriver.Diff(id,parent)
}

// TODO Use git logic insead of inner driver
func (d *Driver) DiffSize(id, parent string) (size int64, err error) {
	return d.innerDriver.DiffSize(id, parent)
}

// TODO Use git logic insead of inner driver
func (d *Driver) GetMetadata(id string) (map[string]string, error) {
	return d.innerDriver.GetMetadata(id)
}

func (d *Driver) String() string {
	return "git"
}

func (d *Driver) Status() [][2]string {
	return d.innerDriver.Status()
}

func (d *Driver) Cleanup() error {
	return d.innerDriver.Cleanup()
}



func (d *Driver) Create(id, parent string, mountLabel string) error {
	if err := d.innerDriver.Create(id, parent, mountLabel); err != nil {
		return err
	}
	dirStr, err := d.innerDriver.Get(id, "")
	if err != nil {
		return err
	}
	defer d.innerDriver.Put(id)

	if output, err := exec.Command("git", "init", dirStr).CombinedOutput(); err != nil {
		logrus.Error("Error trying to init GIT repository: %s (%s)", err, output)
		return nil
	}

	return nil
}

func (d *Driver) CreateAndMerge(id, parent1, parent2 string) error {
	// TODO Use 3rd argument of d.Create (mountLabel) with meaning
	if err := d.Create(id, parent1, parent1); err != nil {
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
		logrus.Error("Error trying to get the current directory: (%s)", err)
		return err
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(dirStr); err != nil {
		logrus.Error("Error trying to change the current directory to %s (%s)", dirStr, err)
		return err
	}

	if output, err := exec.Command("git", "pull", parent2dirStr).CombinedOutput(); err != nil {
		logrus.Error("Error trying to pull GIT repository: %s (%s)", err, output)
		return err
	}

	return nil
}

func (d *Driver) Remove(id string) error {
	return d.innerDriver.Remove(id)
}

func (d *Driver) Get(id, mountLabel string) (string, error) {
	dirStr, err := d.innerDriver.Get(id, mountLabel)
	if err != nil {
		return dirStr, err
	}
	return path.Join(dirStr, "git-repo"), nil
}

func (d *Driver) Put(id string) error {
	dirStr, err := d.innerDriver.Get(id, "")
	if err != nil {
		logrus.Error("Error trying to Get the directory: %s (%s)", dirStr, err)
		return err
	}
	defer d.innerDriver.Put(id)
	defer d.innerDriver.Put(id) // Need twice!

	cwd, err := os.Getwd()
	if err != nil {
		logrus.Error("Error trying to get the current directory: (%s)", err)
		return err
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(dirStr); err != nil {
		logrus.Error("Error trying to change the current directory to %s (%s)", dirStr, err)
		return err
	}

	if output, err := exec.Command("git", "add", "-A").CombinedOutput(); err != nil {
		logrus.Error("Error trying to add files to GIT repository: %s (%s)", err, output)
		return err
	}

	if output, err := exec.Command("git", "config", "user.email", "docker@example.com").CombinedOutput(); err != nil {
		logrus.Error("Error trying to configure user.email on GIT repository: %s (%s)", err, output)
		return err
	}

	if output, err := exec.Command("git", "config", "user.name", "Docker").CombinedOutput(); err != nil {
		logrus.Error("Error trying to configure user.name on GIT repository: %s (%s)", err, output)
		return err
	}

	if output, err := exec.Command("git", "commit", "-a", "-m", "Commit by Docker: " + id).CombinedOutput(); err != nil {
		logrus.Error("Error trying to commit GIT repository: %s (%s)", err, output)
		return err
	}
	return nil
}

func (d *Driver) Exists(id string) bool {
	return d.innerDriver.Exists(id)
}
