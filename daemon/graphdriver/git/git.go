package git

import (
	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/pkg/log"
	"os/exec"
	"os"
	"path"
)

func init() {
	graphdriver.Register("git", Init)
}

func Init(home string, options []string) (graphdriver.Driver, error) {
        if err := os.MkdirAll(home, 0700); err != nil {
                return nil, err
        }

	var innerDriverStr string

	if len(options) >= 1 {
		innerDriverStr = options[0]
	} else {
		innerDriverStr = "aufs"
	}

	innerDriver, err := graphdriver.GetDriver(innerDriverStr, home, options)
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

func (d *Driver) String() string {
	return "git"
}

func (d *Driver) Status() [][2]string {
	return d.innerDriver.Status()
}

func (d *Driver) Cleanup() error {
	return d.innerDriver.Cleanup()
}

func (d *Driver) Create(id, parent string) error {
	if err := d.innerDriver.Create(id, parent); err != nil {
		return err
	}
	dirStr, err := d.innerDriver.Get(id, "")
	if err != nil {
		return err
	}
	defer d.innerDriver.Put(id)

	if output, err := exec.Command("git", "init", dirStr).CombinedOutput(); err != nil {
		log.Errorf("Error trying to init GIT repository: %s (%s)", err, output)
		return nil
	}

	return nil
}

func (d *Driver) CreateAndMerge(id, parent1, parent2 string) error {
	if err := d.Create(id, parent1); err != nil {
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
		log.Errorf("Error trying to get the current directory: (%s)", err)
		return err
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(dirStr); err != nil {
		log.Errorf("Error trying to change the current directory to %s (%s)", dirStr, err)
		return err
	}

	if output, err := exec.Command("git", "pull", parent2dirStr).CombinedOutput(); err != nil {
		log.Errorf("Error trying to pull GIT repository: %s (%s)", err, output)
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

func (d *Driver) Put(id string) {
	dirStr, err := d.innerDriver.Get(id, "")
	if err != nil {
		log.Errorf("Error trying to Get the directory: %s (%s)", dirStr, err)
		return
	}
	defer d.innerDriver.Put(id)
	defer d.innerDriver.Put(id) // Need twice!

	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Error trying to get the current directory: (%s)", err)
		return
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(dirStr); err != nil {
		log.Errorf("Error trying to change the current directory to %s (%s)", dirStr, err)
		return
	}

	if output, err := exec.Command("git", "add", "-A").CombinedOutput(); err != nil {
		log.Errorf("Error trying to add files to GIT repository: %s (%s)", err, output)
		return
	}

	if output, err := exec.Command("git", "config", "user.email", "docker@example.com").CombinedOutput(); err != nil {
		log.Errorf("Error trying to configure user.email on GIT repository: %s (%s)", err, output)
		return
	}

	if output, err := exec.Command("git", "config", "user.name", "Docker").CombinedOutput(); err != nil {
		log.Errorf("Error trying to configure user.name on GIT repository: %s (%s)", err, output)
		return
	}

	if output, err := exec.Command("git", "commit", "-a", "-m", "Commit by Docker: " + id).CombinedOutput(); err != nil {
		log.Errorf("Error trying to commit GIT repository: %s (%s)", err, output)
		return
	}
}

func (d *Driver) Exists(id string) bool {
	return d.innerDriver.Exists(id)
}
