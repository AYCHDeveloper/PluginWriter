package plugin

import (
	"fmt"
	"io"
	"os"

	"github.com/netdata/go-plugin/module"

	"gopkg.in/yaml.v2"
)

func newModuleConfig() *moduleConfig {
	return &moduleConfig{
		UpdateEvery:        1,
		AutoDetectionRetry: 0,
	}
}

type moduleConfig struct {
	UpdateEvery        int                      `yaml:"update_every"`
	AutoDetectionRetry int                      `yaml:"autodetection_retry"`
	Jobs               []map[string]interface{} `yaml:"jobs"`

	name string
}

func (m *moduleConfig) updateJobs(moduleUpdateEvery, pluginUpdateEvery int) {
	if moduleUpdateEvery > 0 {
		m.UpdateEvery = moduleUpdateEvery
	}

	for _, job := range m.Jobs {
		if _, ok := job["update_every"]; !ok {
			job["update_every"] = m.UpdateEvery
		}

		if _, ok := job["autodetection_retry"]; !ok {
			job["autodetection_retry"] = m.AutoDetectionRetry
		}

		if v, ok := job["update_every"].(int); ok && v < pluginUpdateEvery {
			job["update_every"] = pluginUpdateEvery
		}
	}
}

func (p *Plugin) loadModuleConfig(name string) *moduleConfig {

	log.Infof("loading '%s' configuration", name)

	configPath, err := p.ConfigPath.Find(fmt.Sprintf("%s/%s.conf", p.Name, name))
	if err != nil {
		log.Warningf("skipping '%s': %v", name, err)
		return nil
	}

	modConf := newModuleConfig()
	modConf.name = name

	if err = loadYAML(modConf, configPath); err != nil {
		log.Warningf("skipping '%s': %v", name, err)
		return nil
	}

	if len(modConf.Jobs) == 0 {
		log.Warningf("skipping '%s': config 'jobs' section is empty or not exist", name)
		return nil
	}

	return modConf
}

func (p *Plugin) createModuleJobs(modConf *moduleConfig) []Job {
	var jobs []Job

	creator := p.Registry[modConf.name]
	modConf.updateJobs(creator.UpdateEvery, p.Option.UpdateEvery)

	jobName := func(conf map[string]interface{}) interface{} {
		if name, ok := conf["name"]; ok {
			return name
		}
		return "unnamed"
	}

	for _, conf := range modConf.Jobs {
		mod := creator.Create()

		if err := unmarshal(conf, mod); err != nil {
			log.Errorf("skipping %s[%s]: %s", modConf.name, jobName(conf), err)
			continue
		}

		job := module.NewJob(p.Name, modConf.name, mod, p.Out, p)

		if err := unmarshal(conf, job); err != nil {
			log.Errorf("skipping %s[%s]: %s", modConf.name, jobName(conf), err)
			continue
		}

		jobs = append(jobs, job)
	}

	return jobs
}

func (p *Plugin) createJobs() []Job {
	var jobs []Job

	for name := range p.modules {
		conf := p.loadModuleConfig(name)
		if conf == nil {
			continue
		}

		for _, job := range p.createModuleJobs(conf) {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

func unmarshal(conf interface{}, module interface{}) error {
	b, _ := yaml.Marshal(conf)
	return yaml.Unmarshal(b, module)
}

func loadYAML(conf interface{}, filename string) error {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		log.Debug("open file ", filename, ": ", err)
		return err
	}

	if err = yaml.NewDecoder(file).Decode(conf); err != nil {
		if err == io.EOF {
			log.Debug("config file is empty")
			return nil
		}
		log.Debug("read YAML ", filename, ": ", err)
		return err
	}

	return nil
}
